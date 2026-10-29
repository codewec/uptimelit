package db

import (
	"context"
	"database/sql"
	"github.com/codewec/uptimelit/backend/internal/model"
)

// SaveEndpoint commits the monitor and all notifier links together.
func (d *DB) SaveEndpoint(ctx context.Context, ep *model.Endpoint, create bool) error {
	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	owner := Owner(ctx)
	if owner == "" {
		return sql.ErrNoRows
	}
	for _, id := range ep.NotifierIDs {
		var found int
		if err = tx.QueryRowContext(ctx, `SELECT 1 FROM notifiers WHERE id=? AND owner_id=?`, id, owner).Scan(&found); err != nil {
			return err
		}
	}
	if create {
		r, e := tx.ExecContext(ctx, `INSERT INTO endpoints(owner_id,type,name,url,interval_seconds,fail_threshold,paused) VALUES(?,?,?,?,?,?,?)`, owner, ep.Type, ep.Name, ep.URL, ep.IntervalSeconds, ep.FailThreshold, ep.Paused)
		if e != nil {
			return e
		}
		ep.ID, err = r.LastInsertId()
		if err != nil {
			return err
		}
	} else {
		r, e := tx.ExecContext(ctx, `UPDATE endpoints SET type=?,name=?,url=?,interval_seconds=?,fail_threshold=? WHERE id=? AND owner_id=?`, ep.Type, ep.Name, ep.URL, ep.IntervalSeconds, ep.FailThreshold, ep.ID, owner)
		if e != nil {
			return e
		}
		n, _ := r.RowsAffected()
		if n == 0 {
			return sql.ErrNoRows
		}
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM endpoint_notifiers WHERE endpoint_id=?`, ep.ID); err != nil {
		return err
	}
	for _, id := range ep.NotifierIDs {
		if _, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO endpoint_notifiers VALUES(?,?)`, ep.ID, id); err != nil {
			return err
		}
	}
	ep.OwnerID = owner
	if err = tx.QueryRowContext(ctx, `SELECT created_at FROM endpoints WHERE id=?`, ep.ID).Scan(&ep.CreatedAt); err != nil {
		return err
	}
	return tx.Commit()
}
