package httpserver

import (
	"net/http"

	"github.com/yorkyu/approval-engine/internal/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid request body"})
		return
	}
	result, err := h.svc.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, result)
}

type registerRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid request body"})
		return
	}
	if req.Role == "" {
		req.Role = "employee"
	}
	u, err := h.svc.Register(r.Context(), req.Username, req.Password, req.DisplayName, req.Role)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeCreated(w, u)
}
