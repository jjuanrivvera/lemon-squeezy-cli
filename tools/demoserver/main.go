// Command demoserver supplies invented API responses for an account-free VHS recording.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	addr := "127.0.0.1:8643"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	server := &http.Server{
		Addr: addr, Handler: demoHandler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	fmt.Fprintln(os.Stderr, "demo API listening on", addr)
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func record(kind, id string, attributes map[string]any) map[string]any {
	return map[string]any{"type": kind, "id": id, "attributes": attributes}
}

func demoHandler() http.Handler {
	var mu sync.Mutex
	discounts := []any{}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch r.Method + " " + r.URL.Path {
		case "GET /v1/products":
			writeJSON(w, map[string]any{"data": []any{
				record("products", "101", map[string]any{"store_id": 1, "name": "Demo Field Notes", "status": "published", "price": 1900, "price_formatted": "$19.00"}),
				record("products", "102", map[string]any{"store_id": 1, "name": "Demo Icon Pack", "status": "published", "price": 2900, "price_formatted": "$29.00"}),
			}})
		case "GET /v1/orders":
			writeJSON(w, map[string]any{"data": []any{
				record("orders", "201", map[string]any{"order_number": 1001, "user_name": "Demo Buyer 01", "user_email": "buyer01@example.invalid", "currency": "USD", "total": 1900, "total_formatted": "$19.00", "status": "paid"}),
				record("orders", "202", map[string]any{"order_number": 1002, "user_name": "Demo Buyer 02", "user_email": "buyer02@example.invalid", "currency": "USD", "total": 2900, "total_formatted": "$29.00", "status": "paid"}),
			}})
		case "GET /v1/discounts":
			writeJSON(w, map[string]any{"data": discounts})
		case "POST /v1/discounts":
			var body struct {
				Data struct {
					Attributes    map[string]any `json:"attributes"`
					Relationships map[string]any `json:"relationships"`
				} `json:"data"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Data.Attributes == nil {
				http.Error(w, "expected JSON:API attributes", http.StatusBadRequest)
				return
			}
			attributes := body.Data.Attributes
			attributes["status"] = "published"
			discount := record("discounts", "301", attributes)
			discount["relationships"] = body.Data.Relationships
			discounts = append(discounts, discount)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			writeJSON(w, map[string]any{"data": discount})
		default:
			http.NotFound(w, r)
		}
	})
}
