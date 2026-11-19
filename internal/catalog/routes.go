package catalog

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /store", h.CreateStore)
	mux.HandleFunc("POST /stores/{storeID}/products", h.CreateProduct)
	mux.HandleFunc("GET /stores/{storeID}/products", h.GetByStoreProducts)
}
