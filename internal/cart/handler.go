package cart

import (
	"errors"
	"github.com/altf4ples/marketplace/internal/platform"
	"github.com/google/uuid"
	"net/http"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }
func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	var req AddItemRequest
	if err := platform.DecodeJSON(w, r, &req); err != nil {
		platform.WriteError(w, 400, "invalid_request", err.Error())
		return
	}
	c, err := h.svc.AddItem(r.Context(), req.UserID, req.ProductID, req.Quantity)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	platform.WriteJSON(w, 200, toCartResponse(c))
}
func (h *Handler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.URL.Query().Get("user_id"))
	if err != nil {
		platform.WriteError(w, 400, "invalid_user_id", "user_id must be a UUID")
		return
	}
	productID, err := uuid.Parse(r.PathValue("productID"))
	if err != nil {
		platform.WriteError(w, 400, "invalid_product_id", "productID must be a UUID")
		return
	}
	c, err := h.svc.RemoveItem(r.Context(), userID, productID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	platform.WriteJSON(w, 200, toCartResponse(c))
}
func (h *Handler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.URL.Query().Get("user_id"))
	if err != nil {
		platform.WriteError(w, 400, "invalid_user_id", "user_id must be a UUID")
		return
	}
	c, err := h.svc.GetOrCreate(r.Context(), userID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	platform.WriteJSON(w, 200, toCartResponse(c))
}
func (h *Handler) ClearCart(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.URL.Query().Get("user_id"))
	if err != nil {
		platform.WriteError(w, 400, "invalid_user_id", "user_id must be a UUID")
		return
	}
	if err := h.svc.Clear(r.Context(), userID); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrUserNotFound), errors.Is(err, ErrProductNotFound), errors.Is(err, ErrQuantityInvalid):
		platform.WriteError(w, 400, "validation_error", err.Error())
	case errors.Is(err, ErrCartNotFound):
		platform.WriteError(w, 404, "not_found", err.Error())
	default:
		platform.LoggerFrom(r.Context()).Error("cart request failed", "error", err)
		platform.WriteError(w, 500, "internal_error", "internal server error")
	}
}
