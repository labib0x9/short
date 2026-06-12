package queue

import (
	"context"
)

type Queue interface {
	Publish(ctx context.Context, msg ClickEvent) error
}
