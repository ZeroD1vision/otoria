package audio

import (
	"context"

	"github.com/ZeroD1vision/otoria/anima/internal/domain"
)

type Source interface {
	Run(context context.Context, out chan<- domain.AudioBatch) error
}

type Sink interface {
	Run(context context.Context, in <-chan domain.AudioBatch) error
}
