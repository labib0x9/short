package queue

import (
	"context"
)

type Queue interface {
	PublishEmail(ctx context.Context, msg Message) error
}
