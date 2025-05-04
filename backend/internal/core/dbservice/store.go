package dbservice

import (
	"context"
	"github.com/google/uuid"
)

type DBInstanceStore interface {
	Create(ctx context.Context, instance *DBInstance) error
	Get(ctx context.Context, id uuid.UUID) (*DBInstance, error)
	List(ctx context.Context, filters map[string]interface{}) ([]*DBInstance, error)
	Update(ctx context.Context, instance *DBInstance) error
	Delete(ctx context.Context, id uuid.UUID) error
}
