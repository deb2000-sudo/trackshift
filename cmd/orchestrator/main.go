package main

import (
	"log"
	"net/http"
	"os"

	"github.com/deb2000-sudo/trackshift/internal/orchestrator"
)

func main() {
	addr := ":8000"
	if v := os.Getenv("ORCH_LISTEN_ADDR"); v != "" {
		addr = v
	}

	svc := orchestrator.NewService()
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux)

	log.Printf("Orchestrator listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("orchestrator server error: %v", err)
	}
}

