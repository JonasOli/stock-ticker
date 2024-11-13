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

	ticker := time.NewTicker(5 * time.Second)
	loc, timeErr := time.LoadLocation("America/New_York")

	if timeErr != nil {
		fmt.Println("Error loading timezone:", timeErr)
		return
	}

	now := time.Now().In(loc)
	marketOpen := time.Date(now.Year(), now.Month(), now.Day(), 9, 30, 0, 0, loc)
	marketClose := time.Date(now.Year(), now.Month(), now.Day(), 16, 0, 0, 0, loc)

	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Type")
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	for range ticker.C {
		if now.Before(marketOpen) || now.After(marketClose) {
			return
		}

		// Fetch stock quote for a given symbol
		symbol := "GOOG"
		quote, _, quoteErr := finnhubClient.Quote(context.Background()).Symbol(symbol).Execute()

		if quoteErr != nil {
			log.Printf("Error fetching stock details: %v", quoteErr)
			return
		}

		createdAt := time.Now().Format("2006-01-02 15:04:05")

		stockQuote := model.StockPrice{
			Symbol:    symbol,
			Price:     *quote.C,
			CreatedAt: createdAt,
		}

		insertStockPrices := `
				INSERT INTO stock_ticker.StockQuotes
					(ticker_symbol, price, created_at)
				VALUES
					(?, ?, ?);`

		if _, err = conn.Exec(
			insertStockPrices,
			stockQuote.Symbol,
			stockQuote.Price,
			stockQuote.CreatedAt,
		); err != nil {
			log.Fatalf("failed to insert data: %v", err)
		}

		stockQuoteJson, jErr := json.Marshal(stockQuote)

		if jErr != nil {
			log.Fatalf("Error on Json: %v", err)
			return
		}

		fmt.Fprintf(c.Writer, "data: %+v\n\n", string(stockQuoteJson))

		flusher.Flush()
	}
}
