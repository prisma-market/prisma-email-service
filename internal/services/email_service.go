package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/kihyun1998/prisma-market/prisma-email-service/internal/models"
	"github.com/kihyun1998/prisma-market/prisma-email-service/internal/repository/mongodb"
	"github.com/kihyun1998/prisma-market/prisma-email-service/pkg/smtp"
)

type EmailService struct {
	templateRepo *mongodb.TemplateRepository
	smtpClient   *smtp.SMTPClient
	httpClient   *http.Client
}

func NewEmailService(templateRepo *mongodb.TemplateRepository, smtpClient *smtp.SMTPClient) *EmailService {
	return &EmailService{
		templateRepo: templateRepo,
		smtpClient:   smtpClient,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// ProcessEmailRequest 이메일 요청 처리
func (s *EmailService) ProcessEmailRequest(ctx context.Context, req *models.EmailRequest) error {
	// 템플릿 ID를 ObjectID로 변환
	templateID, err := primitive.ObjectIDFromHex(req.TemplateID)
	if err != nil {
		return fmt.Errorf("invalid template ID: %v", err)
	}

	// 템플릿 조회
	template, err := s.templateRepo.GetTemplateByID(ctx, templateID)
	if err != nil {
		return fmt.Errorf("failed to get template: %v", err)
	}
	if template == nil {
		return errors.New("template not found")
	}

	// 변수 유효성 검사
	if err := s.validateTemplateVariables(template, req.Variables); err != nil {
		return err
	}

	// 이메일 전송
	err = s.smtpClient.SendEmail(
		req.To,
		template.Subject,
		template.HTMLContent,
		template.TextContent,
		req.Variables,
	)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	// 콜백 처리 (있는 경우)
	if req.CallbackURL != "" {
		go s.handleCallback(req.CallbackURL, &models.EmailResponse{
			MessageID: primitive.NewObjectID().Hex(),
			Status:    "delivered",
			SentAt:    time.Now(),
		})
	}

	return nil
}

// 템플릿 변수 유효성 검사
func (s *EmailService) validateTemplateVariables(template *models.Template, variables map[string]interface{}) error {
	for _, required := range template.Variables {
		if _, exists := variables[required]; !exists {
			return fmt.Errorf("missing required variable: %s", required)
		}
	}
	return nil
}

// 콜백 처리
func (s *EmailService) handleCallback(callbackURL string, response *models.EmailResponse) {
	body, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal callback response: %v", err)
		return
	}

	resp, err := s.httpClient.Post(callbackURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Failed to send callback: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Callback failed with status: %d", resp.StatusCode)
	}
}

// Template Management

func (s *EmailService) CreateTemplate(ctx context.Context, template *models.Template) error {
	return s.templateRepo.CreateTemplate(ctx, template)
}

func (s *EmailService) GetTemplate(ctx context.Context, id primitive.ObjectID) (*models.Template, error) {
	return s.templateRepo.GetTemplateByID(ctx, id)
}

func (s *EmailService) UpdateTemplate(ctx context.Context, id primitive.ObjectID, template *models.Template) error {
	return s.templateRepo.UpdateTemplate(ctx, id, template)
}

func (s *EmailService) DeleteTemplate(ctx context.Context, id primitive.ObjectID) error {
	return s.templateRepo.DeleteTemplate(ctx, id)
}

func (s *EmailService) ListTemplates(ctx context.Context) ([]*models.Template, error) {
	return s.templateRepo.ListTemplates(ctx)
}

// Queue Processing Implementation
func (s *EmailService) ProcessMessage(msg *models.EmailRequest) error {
	ctx := context.Background()
	return s.ProcessEmailRequest(ctx, msg)
}
