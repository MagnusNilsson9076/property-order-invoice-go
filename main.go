package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("set INFRAI_API_KEY")
	}
	client := apiClient{key: key, http: &http.Client{Timeout: 30 * time.Second}}
	http.HandleFunc("POST /invoices", func(w http.ResponseWriter, r *http.Request) {
		var order Order
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&order); err != nil {
			http.Error(w, "invalid order JSON", http.StatusBadRequest)
			return
		}
		invoice, err := buildInvoice(order, time.Now().UTC())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// One key serves the PDF renderer and the mailer on the same base URL.
		pdf, err := client.post(r.Context(), "/v1/pdf/generate", order.ID+":pdf", map[string]any{"html": invoice.HTML, "store": false})
		if err != nil {
			writeAPIError(w, err)
			return
		}
		mail, err := client.post(r.Context(), "/v1/email/send", order.ID+":email", map[string]any{
			"to": order.TenantEmail, "subject": "Invoice " + order.ID, "html": invoice.HTML,
		})
		if err != nil {
			writeAPIError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			PDF        json.RawMessage `json:"pdf"`
			Email      json.RawMessage `json:"email"`
			TotalCents int64           `json:"total_cents"`
			Reminders  []string        `json:"reminders"`
		}{pdf, mail, invoice.TotalCents, invoice.Reminders})
	})
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	fmt.Printf("listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func writeAPIError(w http.ResponseWriter, err error) {
	var rejected *apiError
	if errors.As(err, &rejected) && rejected.Status >= 400 && rejected.Status < 500 {
		http.Error(w, rejected.Error(), rejected.Status)
		return
	}
	log.Printf("upstream request: %v", err)
	http.Error(w, "upstream request failed", http.StatusBadGateway)
}
