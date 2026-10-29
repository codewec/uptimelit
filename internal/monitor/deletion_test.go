package monitor

import (
	"context"
	"testing"
	"time"
	"uptimelit/internal/auth"
	"uptimelit/internal/db"
	"uptimelit/internal/model"
	"uptimelit/internal/notifier"
)

type blockingChecker struct{ started, release chan struct{} }

func (c blockingChecker) Check(context.Context) (*int, int, bool) {
	close(c.started)
	<-c.release
	code := 200
	return &code, 1, true
}
func TestUserDeletionDuringCheck(t *testing.T) {
	d, e := db.New(":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer d.Close()
	a, e := auth.New(d, auth.Config{})
	if e != nil {
		t.Fatal(e)
	}
	ctx := context.Background()
	admin, e := a.Login(ctx, "root", "admin-password")
	if e != nil {
		t.Fatal(e)
	}
	user, e := a.CreateUser(ctx, "alice", "alice-password", "user")
	if e != nil {
		t.Fatal(e)
	}
	ep := model.Endpoint{Name: "slow", URL: "http://example.com", Type: "http", IntervalSeconds: 60, FailThreshold: 3}
	if e = d.SaveEndpoint(db.WithOwner(ctx, user.ID), &ep, true); e != nil {
		t.Fatal(e)
	}
	m := New(d, notifier.NewDispatcher(d))
	checker := blockingChecker{make(chan struct{}), make(chan struct{})}
	done := make(chan struct{})
	go func() { m.check(ctx, ep, checker); close(done) }()
	<-checker.started
	ids, e := a.DeleteUser(ctx, user.ID, admin.ID)
	if e != nil {
		t.Fatal(e)
	}
	for _, id := range ids {
		m.Delete(id)
	}
	close(checker.release)
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("check did not finish")
	}
	if len(m.StateManager.GetDashboardStats()) != 0 || len(m.state) != 0 || len(m.results) != 0 {
		t.Fatal("late check recreated state")
	}
}
