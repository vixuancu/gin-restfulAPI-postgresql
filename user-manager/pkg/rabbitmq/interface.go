package rabbitmq

import "context"


type RabbitMQService interface {
	Puclish(ctx context.Context, queue string, message any) error
	Consume(ctx context.Context, queue string, handler func(message []byte) error) error
	Close() error
}