package httpserver

import (
	"net/http"
	"strconv"

	"github.com/yorkyu/approval-engine/internal/model"
	"github.com/yorkyu/approval-engine/internal/service"
)

type TemplateHandler struct {
	svc *service.TemplateService
}

func NewTemplateHandler(svc *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{svc: svc}
}

func (h *TemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	var def model.TemplateDef
	if err := decodeJSON(r, &def); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid request body: " + err.Error()})
		return
	}
	id, err := h.svc.Create(r.Context(), &def)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeCreated(w, map[string]int64{"id": id})
}

func (h *TemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.List(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, list)
}

func (h *TemplateHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid id"})
		return
	}
	tpl, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, tpl)
}
