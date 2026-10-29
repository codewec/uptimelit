package monitor

import (
	"context"
	"testing"
	"time"

	"github.com/codewec/uptimelit/backend/internal/db"
	"github.com/codewec/uptimelit/backend/internal/model"
)

func TestConnectionErrorSurvivesHistoryAndHydration(t *testing.T) {
	database, err := db.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if _, err := database.SQL().Exec(`INSERT INTO users(id,username,role) VALUES('owner','owner','admin')`); err != nil {
		t.Fatal(err)
	}
	ctx := db.WithOwner(context.Background(), "owner")
	ep := model.Endpoint{Type: "tcp", Name: "TCP", URL: "localhost:1", IntervalSeconds: 60, FailThreshold: 3}
	if err := database.SaveEndpoint(ctx, &ep, true); err != nil {
		t.Fatal(err)
	}
	check := &model.Check{EndpointID: ep.ID, CheckedAt: time.Now().UTC(), ErrorCode: "connection_refused"}
	if err := database.RecordCheck(ctx, check); err != nil {
		t.Fatal(err)
	}
	checks, err := database.GetChecksByEndpoint(ctx, ep.ID, 50)
	if err != nil || len(checks) != 1 || checks[0].ErrorCode != "connection_refused" {
		t.Fatalf("checks=%v err=%v", checks, err)
	}
	state := NewStateManager(database)
	if err := state.Hydrate(ctx); err != nil {
		t.Fatal(err)
	}
	stats := state.GetDashboardStats()[ep.ID]
	if stats.LastErrorCode != "connection_refused" || stats.RecentChecks[0].ErrorCode != "connection_refused" {
		t.Fatalf("hydrated stats=%+v", stats)
	}
	state.UpdateOnCheck(&model.Check{EndpointID: ep.ID, CheckedAt: time.Now().UTC(), IsUp: true})
	stats = state.GetDashboardStats()[ep.ID]
	if stats.LastErrorCode != "" || stats.RecentChecks[1].ErrorCode != "" {
		t.Fatalf("recovery retained stale error: %+v", stats)
	}
}
