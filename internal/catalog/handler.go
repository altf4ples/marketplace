package catalog

import (
	"errors"
	"github.com/altf4ples/marketplace/internal/platform"
	"github.com/google/uuid"
	"net/http"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }
func (h *Handler) CreateStore(w http.ResponseWriter, r *http.Request) {
	var req CreateStoreRequest
	if err := platform.DecodeJSON(w, r, &req); err != nil {
		platform.WriteError(w, 400, "invalid_request", err.Error())
		return
	}
	s, err := h.svc.CreateStores(r.Context(), req.Name, req.Description)
	switch {
	case errors.Is(err, ErrNameRequired), errors.Is(err, ErrDescriptionInvalid):
		platform.WriteError(w, 400, "validation_error", err.Error())
	case errors.Is(err, ErrConflict):
		platform.WriteError(w, 409, "conflict", err.Error())
	case err != nil:
		platform.LoggerFrom(r.Context()).Error("create store", "error", err)
		platform.WriteError(w, 500, "internal_error", "internal server error")
	default:
		platform.WriteJSON(w, 201, StoreResponse{ID: s.ID.String(), Name: s.Name, Description: s.Description, CreatedAt: s.CreatedAt})
	}
}
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	storeID, err := uuid.Parse(r.PathValue("storeID"))
	if err != nil {
		platform.WriteError(w, 400, "invalid_store_id", "storeID must be a UUID")
		return
	}
	var req CreateProductRequest
	if err := platform.DecodeJSON(w, r, &req); err != nil {
		platform.WriteError(w, 400, "invalid_request", err.Error())
		return
	}
	if req.Price == nil {
		platform.WriteError(w, 400, "validation_error", "price is required")
		return
	}
	p, err := h.svc.CreateProducts(r.Context(), storeID, req.Name, req.Description, *req.Price, req.ImageURL)
	switch {
	case errors.Is(err, ErrNameRequired), errors.Is(err, ErrPriceInvalid), errors.Is(err, ErrStoreIDInvalid), errors.Is(err, ErrDescriptionInvalid), errors.Is(err, ErrImageURLInvalid):
		platform.WriteError(w, 400, "validation_error", err.Error())
	case errors.Is(err, ErrStoreNotFound):
		platform.WriteError(w, 404, "not_found", err.Error())
	case errors.Is(err, ErrConflict):
		platform.WriteError(w, 409, "conflict", err.Error())
	case err != nil:
		platform.LoggerFrom(r.Context()).Error("create product", "error", err)
		platform.WriteError(w, 500, "internal_error", "internal server error")
	default:
		platform.WriteJSON(w, 201, ProductResponse{ID: p.ID.String(), StoreID: p.StoreID.String(), Name: p.Name, Description: p.Description, Price: p.Price, ImageURL: p.ImageURL, CreatedAt: p.CreatedAt})
	}
}
func (h *Handler) GetByStoreProducts(w http.ResponseWriter, r *http.Request) {
	storeID, err := uuid.Parse(r.PathValue("storeID"))
	if err != nil {
		platform.WriteError(w, 400, "invalid_store_id", "storeID must be a UUID")
		return
	}
	products, err := h.svc.GetByStoreProducts(r.Context(), storeID.String())
	switch {
	case errors.Is(err, ErrStoreNotFound):
		platform.WriteError(w, 404, "not_found", err.Error())
		return
	case err != nil:
		platform.LoggerFrom(r.Context()).Error("list products", "error", err)
		platform.WriteError(w, 500, "internal_error", "internal server error")
		return
	}
	out := make([]ProductResponse, 0, len(products))
	for _, p := range products {
		out = append(out, ProductResponse{ID: p.ID.String(), StoreID: p.StoreID.String(), Name: p.Name, Description: p.Description, Price: p.Price, ImageURL: p.ImageURL, CreatedAt: p.CreatedAt})
	}
	platform.WriteJSON(w, 200, ProductListResponse{StoreID: storeID.String(), Products: out})
}
