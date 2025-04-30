package user

import "context"

type Service interface {
	Delete(ctx context.Context, userID string) error
}
