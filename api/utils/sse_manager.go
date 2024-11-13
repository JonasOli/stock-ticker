package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"stockTicker/database"
	"stockTicker/model"
	"time"

	"github.com/Finnhub-Stock-API/finnhub-go/v2"
	"github.com/gin-gonic/gin"
)

type SSEManager struct {
	clients map[chan string]bool
}

var SSEManagerInstance = NewSSEManager()

func NewSSEManager() *SSEManager {
	return &SSEManager{
		clients: make(map[chan string]bool),
	}
}

func (s *SSEManager) StockTickerEventsHandler(c *gin.Context) {
	flusher, ok := c.Writer.(http.Flusher)

	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming unsupported"})
		return
	}

	cfg := finnhub.NewConfiguration()
	cfg.AddDefaultHeader("X-Finnhub-Token", "csmkfjhr01qn12jeu9v0csmkfjhr01qn12jeu9vg")
	finnhubClient := finnhub.NewAPIClient(cfg).DefaultApi

	conn, err := database.ConnectClickHouse()
	if err != nil {
		log.Fatalf("could not connect to ClickHouse: %v", err)
	}

	ticker := time.NewTicker(2 * time.Second)

	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Type")
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

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

			stockQuote := model.HistoricalPrices{
				Symbol:    symbol,
				Price:     *quote.C,
				Open:      *quote.O,
				High:      *quote.H,
				Low:       *quote.L,
				PrevClose: *quote.Pc,
			}

			insertHistoricalPrices := `
				INSERT INTO stock_ticker.HistoricalPrices
				(
					ticker_symbol
					, date
					, open_price
					, high_price
					, low_price
					, close_price
					, adjusted_close
					, volume
					, created_at
				)
				VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?);`

			if _, err := conn.Exec(insertHistoricalPrices, symbol, time.Now().Format("2006-01-02"), *quote.O, *quote.H, *quote.L, *quote.Pc, 0, 0, time.Now().Format("2006-01-02 15:04:05")); err != nil {
				log.Fatalf("failed to insert data: %v", err)
			}

			stockQuoteJson, jErr := json.Marshal(stockQuote)

			if jErr != nil {
				fmt.Println(err)
				return
			}

			fmt.Fprintf(c.Writer, "data: %+v\n\n", string(stockQuoteJson))

			flusher.Flush()
		}
	}
}
