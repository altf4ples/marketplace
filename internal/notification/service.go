package notification

import (
	"context"
	"errors"
	"github.com/google/uuid"
)

var ErrInvalidLimit = errors.New("limit must be between 1 and 100")

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) List(ctx context.Context, userID uuid.UUID, limit int) ([]Notification, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user id")
	}
	if limit == 0 {
		limit = 50
	}
	if limit < 1 || limit > 100 {
		return nil, ErrInvalidLimit
	}
	return s.repo.ListByUser(ctx, userID, limit)
}
