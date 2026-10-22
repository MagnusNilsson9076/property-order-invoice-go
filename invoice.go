package main

import (
	"fmt"
	"html"
	"strings"
	"time"
)

type Order struct {
	ID          string        `json:"id"`
	TenantEmail string        `json:"tenant_email"`
	Property    string        `json:"property"`
	Maintenance []Maintenance `json:"maintenance"`
	Documents   []Document    `json:"documents"`
	Inspections []Inspection  `json:"inspections"`
}

type Maintenance struct {
	Description string `json:"description"`
	Cents       int64  `json:"cents"`
}

type Document struct {
	Name     string `json:"name"`
	Received bool   `json:"received"`
}

type Inspection struct {
	Name      string `json:"name"`
	Due       string `json:"due"`
	Completed bool   `json:"completed"`
}

type Invoice struct {
	HTML       string
	TotalCents int64
	Reminders  []string
}

func buildInvoice(order Order, today time.Time) (Invoice, error) {
	if strings.TrimSpace(order.ID) == "" || strings.TrimSpace(order.TenantEmail) == "" || strings.TrimSpace(order.Property) == "" || len(order.Maintenance) == 0 {
		return Invoice{}, fmt.Errorf("id, tenant_email, property and maintenance are required")
	}
	var body strings.Builder
	fmt.Fprintf(&body, "<h1>Invoice %s</h1><p>Property: %s</p><table>", html.EscapeString(order.ID), html.EscapeString(order.Property))
	var out Invoice
	for _, item := range order.Maintenance {
		if strings.TrimSpace(item.Description) == "" || item.Cents < 0 || item.Cents > 100000000 || out.TotalCents > 100000000-item.Cents {
			return Invoice{}, fmt.Errorf("invalid maintenance line")
		}
		out.TotalCents += item.Cents
		fmt.Fprintf(&body, "<tr><td>%s</td><td>$%d.%02d</td></tr>", html.EscapeString(item.Description), item.Cents/100, item.Cents%100)
	}
	fmt.Fprintf(&body, "</table><p>Total: $%d.%02d</p>", out.TotalCents/100, out.TotalCents%100)
	for _, doc := range order.Documents {
		if !doc.Received && strings.TrimSpace(doc.Name) != "" {
			out.Reminders = append(out.Reminders, "Document needed: "+doc.Name)
		}
	}
	for _, inspection := range order.Inspections {
		due, err := time.Parse("2006-01-02", inspection.Due)
		if err != nil || strings.TrimSpace(inspection.Name) == "" {
			return Invoice{}, fmt.Errorf("invalid inspection name or due date")
		}
		if !inspection.Completed && !due.After(today.AddDate(0, 0, 7)) {
			out.Reminders = append(out.Reminders, "Inspection due: "+inspection.Name+" ("+inspection.Due+")")
		}
	}
	for _, reminder := range out.Reminders {
		fmt.Fprintf(&body, "<p>%s</p>", html.EscapeString(reminder))
	}
	out.HTML = "<!doctype html><html><body>" + body.String() + "</body></html>"
	return out, nil
}
