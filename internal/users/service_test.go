package users

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type userRepoStub struct {
	created   User
	createErr error
	existsErr error
	get       User
}

func (r *userRepoStub) Create(_ context.Context, u User) error           { r.created = u; return r.createErr }
func (r *userRepoStub) GetByID(context.Context, uuid.UUID) (User, error) { return r.get, nil }
func (r *userRepoStub) Exists(context.Context, uuid.UUID) error          { return r.existsErr }

func TestServiceCreateUserValidatesBeforePersistence(t *testing.T) {
	repo := &userRepoStub{}
	svc := NewService(repo)
	_, err := svc.CreateUser(context.Background(), " ", nil)
	if !errors.Is(err, ErrNameRequired) {
		t.Fatalf("got %v", err)
	}
	if repo.created.ID != uuid.Nil {
		t.Fatal("repository called for invalid user")
	}
}

func TestServiceCreateUserPersistsNormalizedUser(t *testing.T) {
	repo := &userRepoStub{}
	svc := NewService(repo)
	got, err := svc.CreateUser(context.Background(), " Bob ", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Bob" || repo.created.ID != got.ID {
		t.Fatalf("user not persisted: got=%#v stored=%#v", got, repo.created)
	}
}

func TestServiceCreateUserPropagatesRepositoryError(t *testing.T) {
	want := errors.New("database unavailable")
	svc := NewService(&userRepoStub{createErr: want})
	_, err := svc.CreateUser(context.Background(), "Bob", nil)
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want %v", err, want)
	}
}
