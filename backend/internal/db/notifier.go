package db

import (
	"context"

	"github.com/codewec/uptimelit/backend/internal/model"
)

func (d *DB) CreateNotifier(ctx context.Context, n *model.Notifier) error {
	res, err := d.conn.ExecContext(ctx,
		`INSERT INTO notifiers (name, type, config_json, owner_id) VALUES (?, ?, ?, ?)`,
		n.Name, n.Type, n.ConfigJSON, Owner(ctx))
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	n.ID = id
	n.OwnerID = Owner(ctx)
	return nil
}

func (d *DB) GetNotifier(ctx context.Context, id int64) (*model.Notifier, error) {
	row := d.conn.QueryRowContext(ctx,
		`SELECT id, owner_id, name, type, config_json, created_at FROM notifiers WHERE id = ? AND (? = '' OR owner_id = ?)`, id, Owner(ctx), Owner(ctx))
	n := &model.Notifier{}
	if err := row.Scan(&n.ID, &n.OwnerID, &n.Name, &n.Type, &n.ConfigJSON, &n.CreatedAt); err != nil {
		return nil, err
	}
	return n, nil
}

func (d *DB) UpdateNotifier(ctx context.Context, n *model.Notifier) error {
	_, err := d.conn.ExecContext(ctx,
		`UPDATE notifiers SET name = ?, type = ?, config_json = ? WHERE id = ? AND owner_id = ?`,
		n.Name, n.Type, n.ConfigJSON, n.ID, Owner(ctx))
	return err
}

func (d *DB) DeleteNotifier(ctx context.Context, id int64) error {
	_, err := d.conn.ExecContext(ctx, `DELETE FROM notifiers WHERE id = ? AND (? = '' OR owner_id = ?)`, id, Owner(ctx), Owner(ctx))
	return err
}

func (d *DB) ListNotifiers(ctx context.Context) ([]model.Notifier, error) {
	rows, err := d.conn.QueryContext(ctx, `SELECT id, owner_id, name, type, config_json, created_at FROM notifiers WHERE (? = '' OR owner_id = ?) ORDER BY id DESC`, Owner(ctx), Owner(ctx))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifiers []model.Notifier
	for rows.Next() {
		var n model.Notifier
		if err := rows.Scan(&n.ID, &n.OwnerID, &n.Name, &n.Type, &n.ConfigJSON, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifiers = append(notifiers, n)
	}
	return notifiers, rows.Err()
}

func (d *DB) GetNotifiersForEndpoint(ctx context.Context, endpointID int64) ([]model.Notifier, error) {
	rows, err := d.conn.QueryContext(ctx, `
		SELECT n.id, n.owner_id, n.name, n.type, n.config_json, n.created_at
		FROM notifiers n
		JOIN endpoint_notifiers en ON n.id = en.notifier_id
		JOIN endpoints e ON e.id = en.endpoint_id AND e.owner_id = n.owner_id
		WHERE en.endpoint_id = ?`, endpointID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifiers []model.Notifier
	for rows.Next() {
		var n model.Notifier
		if err := rows.Scan(&n.ID, &n.OwnerID, &n.Name, &n.Type, &n.ConfigJSON, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifiers = append(notifiers, n)
	}
	return notifiers, rows.Err()
}

// GetEndpointNotifierIDs returns a map of endpoint_id -> []notifier_id
func (d *DB) GetEndpointNotifierIDs(ctx context.Context) (map[int64][]int64, error) {
	rows, err := d.conn.QueryContext(ctx, `SELECT en.endpoint_id, en.notifier_id FROM endpoint_notifiers en JOIN endpoints e ON e.id = en.endpoint_id WHERE (? = '' OR e.owner_id = ?)`, Owner(ctx), Owner(ctx))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[int64][]int64)
	for rows.Next() {
		var eid, nid int64
		if err := rows.Scan(&eid, &nid); err != nil {
			return nil, err
		}
		m[eid] = append(m[eid], nid)
	}
	return m, rows.Err()
}
