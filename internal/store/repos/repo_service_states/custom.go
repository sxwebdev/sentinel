package repo_service_states

import (
	"context"
	"database/sql"

	"github.com/sxwebdev/sentinel/internal/models"
)

type ICustomQuerier interface {
	Querier
	Update(ctx context.Context, id string, params UpdateRequest) (*models.ServiceState, error)
}

type CustomQueries struct {
	*Queries
	db DBTX
}

func NewCustom(db DBTX) *CustomQueries {
	return &CustomQueries{
		Queries: New(db),
		db:      db,
	}
}

func (s *CustomQueries) WithTx(tx *sql.Tx) *CustomQueries {
	return &CustomQueries{
		Queries: New(tx),
		db:      tx,
	}
}

var _ ICustomQuerier = (*CustomQueries)(nil)
