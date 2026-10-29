package monitor

import (
	"context"
	"testing"

	"github.com/codewec/uptimelit/backend/internal/db"
	"github.com/codewec/uptimelit/backend/internal/model"
)

type resultChecker bool

func (c resultChecker) Check(context.Context) (*int, int, bool, string) {
	code := 503
	if c {
		code = 200
	}
	return &code, 1, bool(c), ""
}

type recordedAlerts struct{ titles []string }

func (a *recordedAlerts) SendAlert(_ int64, _, title, _ string) { a.titles = append(a.titles, title) }

func TestNotificationsFollowConfirmedOutages(t *testing.T) {
	const down = "🚨 Service Down"
	const recovered = "✅ Service Recovered"
	for _, tt := range []struct {
		name          string
		threshold     int
		checks        []bool
		notifications []string
	}{
		{"healthy startup", 3, []bool{true, true}, []string{"", ""}},
		{"first failure then success", 3, []bool{false, true}, []string{"", ""}},
		{"two failures then success", 3, []bool{true, false, false, true}, []string{"", "", "", ""}},
		{"third failure confirms outage", 3, []bool{false, false, false, true, true}, []string{"", "", down, recovered, ""}},
		{"ongoing outage sends no duplicates", 3, []bool{true, false, false, false, false, false, true, true}, []string{"", "", "", down, "", "", recovered, ""}},
		{"success resets consecutive failures", 3, []bool{false, false, true, false, false, true}, []string{"", "", "", "", "", ""}},
		{"confirmed and brief outages remain independent", 3, []bool{false, false, false, true, false, true, false, false, false, true}, []string{"", "", down, recovered, "", "", "", "", down, recovered}},
		{"threshold one includes first check", 1, []bool{false, false, true, true, false, true}, []string{down, "", recovered, "", down, recovered}},
		{"custom threshold", 2, []bool{false, true, false, false, true}, []string{"", "", "", down, recovered}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			database, err := db.New(":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer database.Close()
			if _, err = database.SQL().Exec(`INSERT INTO users(id,username,role) VALUES('owner','owner','admin')`); err != nil {
				t.Fatal(err)
			}
			ep := model.Endpoint{Name: "Service", Type: "http", URL: "https://example.com", IntervalSeconds: 60, FailThreshold: tt.threshold}
			if err = database.SaveEndpoint(db.WithOwner(context.Background(), "owner"), &ep, true); err != nil {
				t.Fatal(err)
			}
			alerts := &recordedAlerts{}
			manager := New(database, nil)
			manager.dispatcher = alerts
			for i, up := range tt.checks {
				before := len(alerts.titles)
				manager.check(context.Background(), ep, resultChecker(up))
				want := tt.notifications[i]
				if want == "" {
					if len(alerts.titles) != before {
						t.Fatalf("check %d: unexpected notification %q", i+1, alerts.titles[before:])
					}
				} else {
					if len(alerts.titles) != before+1 || alerts.titles[before] != want {
						t.Fatalf("check %d: notifications=%v, want one %q", i+1, alerts.titles[before:], want)
					}
				}
				stats := manager.StateManager.GetDashboardStats()[ep.ID]
				if stats == nil || stats.LastIsUp == nil || *stats.LastIsUp != up {
					t.Fatalf("check %d: visible status must reflect raw result immediately", i+1)
				}
			}
		})
	}
}
