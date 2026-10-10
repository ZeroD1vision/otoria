package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/ZeroD1vision/otoria/anima/internal/domain"
)

type Pipeline struct {
	audioIn  <-chan domain.AudioBatch
	audioOut chan<- domain.AudioBatch

	stt      STTClient
	llm      LLMClient
	tts      TTSClient
	toolExecutor    ToolExecutor // Запускает инструменты (Execute)
	toolProvider 	ToolProvider // Поставляет список (AvailableTools)
	mu 	  	 sync.Mutex
}

func NewPipeline(
	audioIn <-chan domain.AudioBatch,
	stt STTClient,
	llm LLMClient,
	tts TTSClient,
	toolExecutor ToolExecutor,
	toolProvider ToolProvider,
	audioOut chan<- domain.AudioBatch,
) *Pipeline {
	return &Pipeline{
		audioIn:  audioIn,
		stt:      stt,
		llm:      llm,
		tts:      tts,
		toolExecutor: toolExecutor,
		toolProvider: toolProvider,
		audioOut: audioOut,
	}
}

func (p *Pipeline) Run(ctx context.Context) error {
	for {
		select {
		case batch, ok := <-p.audioIn:
			if !ok {
				// Канал закрыт
				slog.Info("audio input channel closed, exiting")
				return nil
			}

			// Отправка в основной цикл пайплайна для одного запроса
			if err := p.Turn(ctx, batch); err != nil {
				slog.Error("failed to process dialogue turn", "error", err)
			}

		case <-ctx.Done():
			// Выход по контексту (завершение приложения)
			slog.Info("pipeline stopped by context")
			return ctx.Err()
		}
	}
}

// Обработка по порядку:
// 1. Отправка батча в STT
// 2. Навешивание контекста для LLM
// 3. Отправка сконфигурированного запроса в LLM
// 4. Отправка тулзовой части ответа в Tools
// 5. Отправка результатов инструментов обратно в LLM
// 6. Отправка устной части ответа в TTS
//
// [TODO: v2 Озвучивание раздумий LLM параллельно с выполнением инструментов]
//
//   Порядок:
//   1. Отправка батча в STT
// 	 2. Навешивание контекста для LLM
// 	 3. Отправка сконфигурированного запроса в LLM
//   4.1 Первичный голосовой ответ типа "Смотрю погоду..." полученный от LLM
// 	 4.2.1 Выполнение инструментов
// 	 4.2.2 Отправка результатов инструментов обратно в LLM
// 	 4.2.3 Отправка устной части ответа в TTS
//   5. Отправка в динамик
//   
//   (4.1 и 4.2 параллельно)
//
//
// [TODO: v2 Прерывание]
//
// Когда пользователь начал запрос №1, перебил себя запросом №2.
// VAD детектирует новое начало речи пока Turn ещё выполняется.
//
// Варианты:
//   A) Cancel текущего Turn через ctx, начать новый. История не пишется.
//   B) Cancel текущего Turn, но сохранить user message в истории.
//   C) Дать №1 доиграть, №2 встаёт в очередь.
//
// Для MVP вариант A, самый простой
func (p *Pipeline) Turn(ctx context.Context, batch domain.AudioBatch) error {
	// 1. Отправка батча в STT
	transcription, err := p.stt.Transcribe(ctx, batch)
	if err != nil {
		slog.Error("failed to transcribe audio batch", "error", err)
		return fmt.Errorf("stt: %w", err)
	}

	// Защита от пустого ввода
	transcription = strings.TrimSpace(transcription)
	if transcription == "" {
		slog.Debug("stt returned empty text, skipping turn")
		return nil
	}

	// 2. Навешивание контекста для LLM
	// Получаем список доступных инструментов для LLM
	availableTools, err := p.toolProvider.AvailableTools(ctx)
	if err != nil {
		slog.Error("failed to get available tools", "error", err)
		return fmt.Errorf("tools: %w", err)
	}
	// Создаем запрос для LLM с учетом транскрипции и доступных инструментов
	llmRequest := p.llm.CreateRequest(transcription, availableTools)

	// 3. Отправка сконфигурированного запроса в LLM
	llmResponse, err := p.llm.SendRequest(ctx, llmRequest)
	if err != nil {
		slog.Error("failed to send request to LLM", "error", err)
		return err
	}
	
	// 4. Отправка тулзовой части ответа в Tools
	const maxToolCalls = 10 // Ограничение на количество вызовов инструментов за раз
	if len(llmResponse.ToolCalls) > maxToolCalls {
		slog.Warn("too many tool calls, limiting to 10")
	}
	
	limit := min(len(llmResponse.ToolCalls), maxToolCalls)
	var toolResultsBuf [maxToolCalls]domain.ToolResult

	toolResults := toolResultsBuf[:limit]
	err = p.ExecuteAllCalls(ctx, llmResponse.ToolCalls, toolResults)
	if err != nil {
		slog.Error("failed to execute tools", "error", err)
		return err
	}

	// 5. Отправка результатов инструментов обратно в LLM
	llmToolResponse, err := p.llm.SendToolResults(ctx, toolResults)
	if err != nil {
		slog.Error("failed to send tool results to LLM", "error", err)
		return err
	}

	// 6. Отправка устной части ответа LLM в TTS
	audioResponse, err := p.tts.Voicing(ctx, llmResponse)
	if err != nil {
		slog.Error("failed to voice LLM response", "error", err)
		return err
	}

	// Отправка аудио с ответом в канал для воспроизведения
	p.audioOut <- audioResponse

	return nil
}

// ExecuteAllCalls выполняет все tool calls из одного ответа LLM.
//
// [MVP]: Все вызовы выполняются последовательно.
//
// [TODO: v2 - деление на non-returning и returning вызовы]
// ToolCall-ы делятся на два типа по полю ToolDefinition.ReturnsData:
//
//   non-returning (ReturnsData: false):
//     open_browser, click_button, write_text например не возвращают payload для LLM
//     Можно запускать параллельно
//     Результат: только status + error например.
//
//   returning (ReturnsData: true):
//     read_file, get_window_state возвращают payload, который LLM использует для 
//     следующего ответа. Для MVP - последовательно.
//     Для v2 - параллельно там где нет зависимостей между вызовами.
func (p *Pipeline) ExecuteAllCalls(ctx context.Context, calls []domain.ToolCall, dst []domain.ToolResult) error {
	for i, call := range calls[:len(dst)] {
		res, err := p.toolExecutor.Execute(ctx, call)
		if err != nil {
			slog.Error("failed to execute tool call", "tool", call.ToolID.Verbose(), "error", err)
			res = domain.ToolResult{
				CallID:   call.ID,
				Error:    domain.CodeInternalErr, // Внутренняя ошибка инструмента
			}
		}

		dst[i] = res
	}
	return nil
}