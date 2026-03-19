package repo_agents

import (
	"context"
	"database/sql"

	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
)

type ICustomQuerier interface {
	Querier
	Update(ctx context.Context, id string, params UpdateRequest) (*models.Agent, error)
	Find(ctx context.Context, params FindParams) (*storecmn.FindResponseWithCount[*models.Agent], error)
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
