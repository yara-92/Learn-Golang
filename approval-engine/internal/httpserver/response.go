package httpserver

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/yorkyu/approval-engine/internal/model"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json response: %v", err)
	}
}

func writeOK(w http.ResponseWriter, v any) {
	writeJSON(w, http.StatusOK, v)
}

func writeCreated(w http.ResponseWriter, v any) {
	writeJSON(w, http.StatusCreated, v)
}

type errorBody struct {
	Error string `json:"error"`
}

// writeErr 把 service/engine 层返回的 error 映射成合适的 HTTP 状态码。
// 这是分层架构里"错误翻译"发生的地方：领域层只关心业务语义
// （ErrNotFound / ErrForbidden 等），完全不知道 HTTP 状态码的存在。
func writeErr(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, model.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, model.ErrInvalidCredential):
		status = http.StatusUnauthorized
	case errors.Is(err, model.ErrUnauthorized):
		status = http.StatusUnauthorized
	case errors.Is(err, model.ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, model.ErrTaskNotPending):
		status = http.StatusConflict
	case errors.Is(err, model.ErrTemplateInvalid):
		status = http.StatusBadRequest
	case errors.Is(err, model.ErrDuplicateUser):
		status = http.StatusConflict
	}
	if status == http.StatusInternalServerError {
		log.Printf("internal error: %v", err)
	}
	writeJSON(w, status, errorBody{Error: err.Error()})
}

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
