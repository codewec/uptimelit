package db

import (
	"context"

	"github.com/codewec/uptimelit/backend/internal/model"
)

func (d *DB) GetEndpoint(ctx context.Context, id int64) (*model.Endpoint, error) {
	row := d.conn.QueryRowContext(ctx,
		`SELECT id, owner_id, type, name, url, interval_seconds, fail_threshold, paused, created_at FROM endpoints WHERE id = ? AND (? = '' OR owner_id = ?)`, id, Owner(ctx), Owner(ctx))
	ep := &model.Endpoint{}
	err := row.Scan(&ep.ID, &ep.OwnerID, &ep.Type, &ep.Name, &ep.URL, &ep.IntervalSeconds, &ep.FailThreshold, &ep.Paused, &ep.CreatedAt)
	if err != nil {
		return nil, err
	}
	return ep, nil
}

func (d *DB) DeleteEndpoint(ctx context.Context, id int64) error {
	_, err := d.conn.ExecContext(ctx, `DELETE FROM endpoints WHERE id = ? AND (? = '' OR owner_id = ?)`, id, Owner(ctx), Owner(ctx))
	return err
}

func (d *DB) TogglePause(ctx context.Context, id int64) (bool, error) {
	var paused bool
	if err := d.conn.QueryRowContext(ctx, `SELECT paused FROM endpoints WHERE id = ? AND (? = '' OR owner_id = ?)`, id, Owner(ctx), Owner(ctx)).Scan(&paused); err != nil {
		return false, err
	}
	newPaused := !paused
	_, err := d.conn.ExecContext(ctx, `UPDATE endpoints SET paused = ? WHERE id = ? AND owner_id = ?`, newPaused, id, Owner(ctx))
	return newPaused, err
}

func (d *DB) ListActiveEndpoints(ctx context.Context) ([]model.Endpoint, error) {
	rows, err := d.conn.QueryContext(ctx,
		`SELECT id, owner_id, type, name, url, interval_seconds, fail_threshold, paused, created_at FROM endpoints WHERE paused = 0`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var endpoints []model.Endpoint
	for rows.Next() {
		var ep model.Endpoint
		if err := rows.Scan(&ep.ID, &ep.OwnerID, &ep.Type, &ep.Name, &ep.URL, &ep.IntervalSeconds, &ep.FailThreshold, &ep.Paused, &ep.CreatedAt); err != nil {
			return nil, err
		}
		endpoints = append(endpoints, ep)
	}
	return endpoints, rows.Err()
}

func (d *DB) ListEndpoints(ctx context.Context) ([]model.Endpoint, error) {
	rows, err := d.conn.QueryContext(ctx,
		`SELECT id, owner_id, type, name, url, interval_seconds, fail_threshold, paused, created_at FROM endpoints WHERE (? = '' OR owner_id = ?) ORDER BY created_at DESC`, Owner(ctx), Owner(ctx))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var endpoints []model.Endpoint
	for rows.Next() {
		var ep model.Endpoint
		if err := rows.Scan(&ep.ID, &ep.OwnerID, &ep.Type, &ep.Name, &ep.URL, &ep.IntervalSeconds, &ep.FailThreshold, &ep.Paused, &ep.CreatedAt); err != nil {
			return nil, err
		}
		endpoints = append(endpoints, ep)
	}
	return endpoints, rows.Err()
}
