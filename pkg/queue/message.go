package queue

import "github.com/kihyun1998/prisma-market/prisma-email-service/internal/models"

type MessageProcessor interface {
	ProcessMessage(msg *models.EmailRequest) error
}
