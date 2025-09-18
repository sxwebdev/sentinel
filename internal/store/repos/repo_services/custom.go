package repo_services

import (
	"context"
	"database/sql"

	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/pkg/dbutils"
)

type ICustomQuerier interface {
	Querier
	GetViewByID(ctx context.Context, id string) (*models.ServiceFullView, error)
	Update(ctx context.Context, id string, service UpdateServiceRequest) (*models.ServiceFullView, error)
	FindView(ctx context.Context, params FindParams) (*dbutils.FindResponseWithCount[*models.ServiceFullView], error)
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
