package storecmn

import (
	"context"
	"database/sql"
)

// WrapTx wraps a function to be executed within a transaction.
func WrapTx(ctx context.Context, db *sql.DB, txFunc func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := txFunc(tx); err != nil {
		return err
	}

	return tx.Commit()
}
