package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Starting Otoria core...")
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Слушаем сигнал для начала чтения аудио из IPC (пока что UNIX сокет)
	go IPCReader.Run(ctx, STTChannel) // Читаем из IPC и формируем батчи аудио для STT

	// Отправляем батч на сервер для STT
	go server.Run(ctx, batch) // Возвращает транскрибированный текст

	// Внутри него:
	// Собираем промпт
	llm_prompt := make_prompt(transcription, context) // Возвращает готовый промпт дйным щелчком выберите поля в следующем порядке:

	// Отправляем целый промпт в API LLM
	llm_response := send_to_llm(llm_prompt) // Возвращает ответ от LLM с возможными инструкциями для агента

	// Получаем ответ от LLM и отправляем его в TTS
	llm_response_audio := send_to_tts(llm_response) // Возвращает аудио с ответом

	// И параллельно отправляем нужные сигналы в MCP агента для выполнения инструкций, полученных от LLM
	send_to_mcp(llm_response.instructions) // Отправляет инструкции в MCP агента

	// И закидываем аудио с ответом в IPC для воспроизведения
	go IPCWriter.Run(llm_response_audio) // Отправляет аудио в IPC для воспроизведения

	fmt.Println("Stopping Otoria core...")
}
