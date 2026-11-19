package users

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type handlerUserRepo struct{ user User }

func (r *handlerUserRepo) Create(_ context.Context, u User) error           { r.user = u; return nil }
func (r *handlerUserRepo) GetByID(context.Context, uuid.UUID) (User, error) { return r.user, nil }
func (r *handlerUserRepo) Exists(context.Context, uuid.UUID) error          { return nil }

func TestCreateHandlerReturnsCreatedUser(t *testing.T) {
	repo := &handlerUserRepo{}
	h := NewHandler(NewService(repo))
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"  Alice  "}`))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var response UserResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Name != "Alice" || response.ID == "" {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestCreateHandlerRejectsMalformedJSON(t *testing.T) {
	h := NewHandler(nil) // malformed input must be rejected before the service is accessed
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":`))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"invalid_request"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}
