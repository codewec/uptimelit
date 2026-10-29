package model

import (
	"encoding/json"
	"testing"
)

func TestRecentCheckIncludesZeroResponseTime(t *testing.T) {
	encoded, err := json.Marshal(MinimalCheck{IsUp: true, ResponseTimeMs: 0})
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	if string(fields["response_time_ms"]) != "0" {
		t.Fatalf("zero-millisecond checks must expose their duration: %s", encoded)
	}
}
