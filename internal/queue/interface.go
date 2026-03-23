package queue

import "context"

type MessagePublisher interface {
	PublishMessage(ctx context.Context, payload []byte) error
}
