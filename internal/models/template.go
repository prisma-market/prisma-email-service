package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Template 이메일 템플릿 모델
type Template struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name"`                 // 템플릿 이름 (e.g., "password-reset")
	Subject     string             `bson:"subject" json:"subject"`           // 이메일 제목
	HTMLContent string             `bson:"html_content" json:"html_content"` // HTML 형식 내용
	TextContent string             `bson:"text_content" json:"text_content"` // 텍스트 형식 내용
	Variables   []string           `bson:"variables" json:"variables"`       // 템플릿 변수 목록
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// EmailRequest 이메일 전송 요청 구조체
type EmailRequest struct {
	TemplateID  string                 `json:"template_id"`
	To          []string               `json:"to"`
	Variables   map[string]interface{} `json:"variables"`
	CallbackURL string                 `json:"callback_url,omitempty"` // 선택적 콜백 URL
}

// EmailResponse 이메일 전송 응답 구조체
type EmailResponse struct {
	MessageID string    `json:"message_id"`
	Status    string    `json:"status"`
	SentAt    time.Time `json:"sent_at"`
}
