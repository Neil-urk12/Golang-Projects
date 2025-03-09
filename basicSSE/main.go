package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func sseHandler(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Send events in a loop
	for i := 0; ; i++ {
		// Create event data
		eventData := fmt.Sprintf("Current time: %s - Event #%d", time.Now().Format(time.RFC3339), i+1)

		// Format event data according to SSE spec
		fmt.Fprintf(w, "data: %s\n\n", eventData)

		// Flush the response to ensure data is sent immediately
		flusher.Flush()

		log.Printf("Sent event: %s", eventData)

		// Sleep for a second before sending the next event
		time.Sleep(1 * time.Second)
	}
}

func main() {
	http.HandleFunc("/events", sseHandler)

	fmt.Println("SSE Server started at http://localhost:8080/events")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

