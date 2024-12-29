package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/kihyun1998/prisma-market/prisma-email-service/internal/config"
	"github.com/kihyun1998/prisma-market/prisma-email-service/internal/handlers"
	"github.com/kihyun1998/prisma-market/prisma-email-service/internal/repository/mongodb"
	"github.com/kihyun1998/prisma-market/prisma-email-service/internal/services"
	"github.com/kihyun1998/prisma-market/prisma-email-service/pkg/queue"
	"github.com/kihyun1998/prisma-market/prisma-email-service/pkg/smtp"
)

func main() {
	// 설정 로드
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// MongoDB 리포지토리 초기화
	templateRepo, err := mongodb.NewTemplateRepository(cfg.MongoURI)
	if err != nil {
		log.Fatalf("Failed to create template repository: %v", err)
	}

	// SMTP 클라이언트 초기화
	smtpClient := smtp.NewSMTPClient(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUsername,
		cfg.SMTPPassword,
		cfg.SMTPFrom,
	)

	// 이메일 서비스 초기화
	emailService := services.NewEmailService(templateRepo, smtpClient)

	// RabbitMQ 소비자 초기화
	consumer, err := queue.NewConsumer(
		cfg.RabbitMQURL,
		cfg.RabbitMQQueue,
		cfg.RabbitMQRetries,
		emailService,
	)
	if err != nil {
		log.Fatalf("Failed to create queue consumer: %v", err)
	}

	// 메시지 큐 소비자 시작
	if err := consumer.Start(context.Background()); err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}

	// HTTP 핸들러 초기화
	emailHandler := handlers.NewEmailHandler(emailService)

	// 라우터 설정
	r := mux.NewRouter()

	// API 엔드포인트 등록
	api := r.PathPrefix("/api/v1").Subrouter()

	// 이메일 전송 엔드포인트
	api.HandleFunc("/email/send", emailHandler.SendEmail).Methods(http.MethodPost)

	// 템플릿 관리 엔드포인트
	api.HandleFunc("/templates", emailHandler.CreateTemplate).Methods(http.MethodPost)
	api.HandleFunc("/templates", emailHandler.ListTemplates).Methods(http.MethodGet)
	api.HandleFunc("/templates/{id}", emailHandler.GetTemplate).Methods(http.MethodGet)
	api.HandleFunc("/templates/{id}", emailHandler.UpdateTemplate).Methods(http.MethodPut)
	api.HandleFunc("/templates/{id}", emailHandler.DeleteTemplate).Methods(http.MethodDelete)

	// HTTP 서버 설정
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 서버를 고루틴으로 시작
	go func() {
		log.Printf("Starting server on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 종료 시그널 처리
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 컨텍스트 타임아웃 설정
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 서버 종료
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// 리소스 정리
	if err := consumer.Close(); err != nil {
		log.Printf("Error closing consumer: %v", err)
	}

	if err := templateRepo.Close(ctx); err != nil {
		log.Printf("Error closing template repository: %v", err)
	}

	log.Println("Server exited properly")
}
