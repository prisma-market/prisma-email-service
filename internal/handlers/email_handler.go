package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/kihyun1998/prisma-market/prisma-email-service/internal/models"
	"github.com/kihyun1998/prisma-market/prisma-email-service/internal/services"
)

type EmailHandler struct {
	emailService *services.EmailService
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewEmailHandler(emailService *services.EmailService) *EmailHandler {
	return &EmailHandler{
		emailService: emailService,
	}
}

// SendEmail 이메일 전송 요청 처리
func (h *EmailHandler) SendEmail(w http.ResponseWriter, r *http.Request) {
	var req models.EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.emailService.ProcessEmailRequest(r.Context(), &req); err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Email request accepted",
	})
}

// CreateTemplate 새 템플릿 생성
func (h *EmailHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var template models.Template
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.emailService.CreateTemplate(r.Context(), &template); err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(template)
}

// GetTemplate 템플릿 조회
func (h *EmailHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		h.sendError(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	template, err := h.emailService.GetTemplate(r.Context(), id)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if template == nil {
		h.sendError(w, "Template not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(template)
}

// UpdateTemplate 템플릿 업데이트
func (h *EmailHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		h.sendError(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	var template models.Template
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.emailService.UpdateTemplate(r.Context(), id, &template); err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(template)
}

// DeleteTemplate 템플릿 삭제
func (h *EmailHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		h.sendError(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	if err := h.emailService.DeleteTemplate(r.Context(), id); err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Template deleted successfully",
	})
}

// ListTemplates 모든 템플릿 조회
func (h *EmailHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.emailService.ListTemplates(r.Context())
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(templates)
}

// sendError 에러 응답 전송 헬퍼 함수
func (h *EmailHandler) sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
