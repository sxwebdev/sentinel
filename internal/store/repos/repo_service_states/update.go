package repo_service_states

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/samber/lo"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
	"github.com/tkcrm/modules/pkg/db/dbutils"
	"github.com/tkcrm/modules/pkg/utils"
)

var availableUpdateColumns = lo.Filter(ServiceStatesColumnNames(), func(item ColumnName, _ int) bool {
	return !slices.Contains([]ColumnName{
		ColumnNameServiceStatesId,
		ColumnNameServiceStatesCreatedAt,
		ColumnNameServiceStatesUpdatedAt,
	}, item)
})

type UpdateRequest struct {
	models.ServiceState
	FieldMask dbutils.FieldMask[ColumnName]
}

func (s *CustomQueries) Update(ctx context.Context, id string, params UpdateRequest) (*models.ServiceState, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}

	for _, path := range params.FieldMask.Items() {
		if !slices.Contains(availableUpdateColumns, path) {
			return nil, fmt.Errorf("unavailable field %s for this method", path)
		}
	}

	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(TableNameServiceStates.String()).
		Where(ub.Equal("id", id)).
		Set(ub.Assign(ColumnNameServiceStatesUpdatedAt.String(), time.Now()))

	values, err := utils.StructToMap(params.ServiceState, "json")
	if err != nil {
		return nil, err
	}

	for _, path := range params.FieldMask {
		idx := slices.IndexFunc(ServiceStatesColumnNames(), func(i ColumnName) bool {
			return path == i
		})

		if idx == -1 {
			return nil, fmt.Errorf("unavailable path: %s", path)
		}

		value, ok := values[path.String()]
		if !ok {
			return nil, fmt.Errorf("value not found for path: %s", path)
		}

		ub.SetMore(ub.Assign(path.String(), value))
	}

	// execute query
	sql, args := ub.Build()
	if _, err := s.db.ExecContext(ctx, sql, args...); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, id)
}
