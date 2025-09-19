package repo_incidents

import (
	"context"
	"database/sql"
	"time"

	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
)

type ICustomQuerier interface {
	Querier
	Find(ctx context.Context, params FindParams) (*storecmn.FindResponseWithCount[*models.Incident], error)
	StatsByDateRange(ctx context.Context, startTime, endTime time.Time) (StatsByDateRangeData, error)
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
