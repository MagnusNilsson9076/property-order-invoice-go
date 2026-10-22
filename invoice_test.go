package main

import (
	"strings"
	"testing"
	"time"
)

func TestInvoiceDecisions(t *testing.T) {
	today := time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name          string
		due           string
		completed     bool
		wantReminders int
	}{
		{"upcoming inspection", "2026-09-25", false, 2},
		{"future inspection", "2026-10-01", false, 1},
		{"completed inspection", "2026-09-25", true, 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			order := Order{ID: "PM-42", TenantEmail: "tenant@example.com", Property: "Unit <7>",
				Maintenance: []Maintenance{{Description: "Sink repair", Cents: 12550}},
				Documents:   []Document{{Name: "Access form", Received: false}},
				Inspections: []Inspection{{Name: "Fire alarm", Due: tc.due, Completed: tc.completed}}}
			got, err := buildInvoice(order, today)
			if err != nil {
				t.Fatal(err)
			}
			if got.TotalCents != 12550 || len(got.Reminders) != tc.wantReminders {
				t.Fatalf("total=%d reminders=%v", got.TotalCents, got.Reminders)
			}
			if !strings.Contains(got.HTML, "Unit &lt;7&gt;") || !strings.Contains(got.HTML, "$125.50") {
				t.Fatalf("unexpected invoice HTML: %s", got.HTML)
			}
		})
	}
}
