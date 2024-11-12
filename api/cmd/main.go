package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"stockTicker/pkg"
	"time"

	finnhub "github.com/Finnhub-Stock-API/finnhub-go/v2"
)

func main() {
	http.HandleFunc("/events", eventsHandler)
	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func eventsHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers to allow all origins. You may want to restrict this to specific origins in a production environment.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	cfg := finnhub.NewConfiguration()
	cfg.AddDefaultHeader("X-Finnhub-Token", "csmkfjhr01qn12jeu9v0csmkfjhr01qn12jeu9vg")
	finnhubClient := finnhub.NewAPIClient(cfg).DefaultApi

	ticker := time.NewTicker(2 * time.Second)

	defer ticker.Stop()

	ctx := r.Context()

	for {
		select {
		case <-ticker.C:
			// Fetch stock quote for a given symbol
			symbol := "GOOG"
			quote, _, err := finnhubClient.Quote(context.Background()).Symbol(symbol).Execute()

			if err != nil {
				log.Printf("Error fetching stock details: %v", err)
				return
			}

			stockQuote := pkg.StockQuote{
				Symbol:    symbol,
				Price:     *quote.C,
				Open:      *quote.O,
				High:      *quote.H,
				Low:       *quote.L,
				PrevClose: *quote.Pc,
			}

			stockQuoteJson, jErr := json.Marshal(stockQuote)

			if jErr != nil {
				fmt.Println(err)
				return
			}

			fmt.Fprintf(w, "data: %+v\n\n", string(stockQuoteJson))

			w.(http.Flusher).Flush() // Flush the data to the client

		case <-ctx.Done():
			// Exit if the client disconnects
			log.Println("Client disconnected")
			return
		}
	}
}
