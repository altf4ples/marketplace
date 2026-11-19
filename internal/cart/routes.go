package cart

import "net/http"

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /cart/items", h.AddItem)
	mux.HandleFunc("DELETE /cart/items/{productID}", h.RemoveItem)
	mux.HandleFunc("GET /cart", h.GetCart)
	mux.HandleFunc("DELETE /cart", h.ClearCart)
}
