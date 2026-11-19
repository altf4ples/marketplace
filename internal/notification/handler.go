package notification

import (
	"github.com/altf4ples/marketplace/internal/platform"
	"github.com/google/uuid"
	"net/http"
	"strconv"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("userID"))
	if err != nil {
		platform.WriteError(w, 400, "invalid_user_id", "userID must be a UUID")
		return
	}
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil {
			platform.WriteError(w, 400, "invalid_limit", "limit must be an integer")
			return
		}
	}
	items, err := h.svc.List(r.Context(), id, limit)
	if err != nil {
		platform.WriteError(w, 400, "invalid_request", err.Error())
		return
	}
	platform.WriteJSON(w, 200, map[string]any{"notifications": items})
}
