package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/AshishBagdane/report-engine/internal/engine"
	"github.com/AshishBagdane/report-engine/internal/formatter"
	"github.com/AshishBagdane/report-engine/internal/output"
	"github.com/AshishBagdane/report-engine/internal/processor"
	"github.com/AshishBagdane/report-engine/internal/provider"
)

func main() {
	// Mock Provider that might simulate health checks
	prov := provider.NewMockProvider([]map[string]interface{}{})

	// Build Engine
	eng, err := engine.NewEngineBuilder().
		WithProvider(prov).
		WithProcessor(&processor.BaseProcessor{}).
		WithFormatter(formatter.NewJSONFormatter("")).
		WithOutput(&output.ConsoleOutput{}).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Setup HTTP Server
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		healthStatus := eng.Health(ctx)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(healthStatus); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	port := ":8080"
	fmt.Printf("Starting server on %s\n", port)
	fmt.Println("Try: curl http://localhost:8080/health")

	// Start server (blocking)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
