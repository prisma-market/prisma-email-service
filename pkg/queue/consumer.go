package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kihyun1998/prisma-market/prisma-email-service/internal/models"
	"github.com/streadway/amqp"
)

type Consumer struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	queue      amqp.Queue
	processor  MessageProcessor
	maxRetries int
}

func NewConsumer(url, queueName string, maxRetries int, processor MessageProcessor) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}

	// 큐 선언
	q, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %v", err)
	}

	// // 데드레터 큐 설정
	// deadLetterQueue, err := ch.QueueDeclare(
	// 	queueName+"_dead_letter",
	// 	true,
	// 	false,
	// 	false,
	// 	false,
	// 	nil,
	// )
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to declare dead letter queue: %v", err)
	// }

	return &Consumer{
		conn:       conn,
		channel:    ch,
		queue:      q,
		processor:  processor,
		maxRetries: maxRetries,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		c.queue.Name,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %v", err)
	}

	go func() {
		for msg := range msgs {
			var emailReq models.EmailRequest
			if err := json.Unmarshal(msg.Body, &emailReq); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				msg.Nack(false, false)
				continue
			}

			// 재시도 횟수 확인
			retryCount := 0
			if xRetry, ok := msg.Headers["x-retry-count"].(int); ok {
				retryCount = xRetry
			}

			if err := c.processor.ProcessMessage(&emailReq); err != nil {
				if retryCount >= c.maxRetries {
					// 최대 재시도 횟수 초과 - 데드레터 큐로 이동
					log.Printf("Message failed after %d retries: %v", retryCount, err)
					msg.Nack(false, false)
				} else {
					// 재시도 큐에 메시지 다시 추가
					retryCount++
					headers := amqp.Table{"x-retry-count": retryCount}
					c.channel.Publish(
						"",
						c.queue.Name,
						false,
						false,
						amqp.Publishing{
							Headers:     headers,
							ContentType: "application/json",
							Body:        msg.Body,
							Expiration:  fmt.Sprintf("%d", int64(time.Second)*5), // 5초 후 재시도
						},
					)
					msg.Ack(false)
				}
				continue
			}

			msg.Ack(false)
		}
	}()

	return nil
}

func (c *Consumer) Close() error {
	if err := c.channel.Close(); err != nil {
		return fmt.Errorf("error closing channel: %v", err)
	}
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("error closing connection: %v", err)
	}
	return nil
}
