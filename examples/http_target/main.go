package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"answer":    "Agent Sail mock target is healthy.",
			"citations": []string{"demo-policy"},
		})
	})
	log.Println("Agent Sail HTTP target listening on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
