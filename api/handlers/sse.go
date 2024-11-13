package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"stockTicker/database"
	"stockTicker/model"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

func (s *SSEManager) StockTickerLiveUpdates(c *gin.Context) {
	flusher, ok := c.Writer.(http.Flusher)

	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming unsupported"})
		return
	}

	w, _, err := websocket.DefaultDialer.Dial("wss://ws.finnhub.io?token=csmkfjhr01qn12jeu9v0csmkfjhr01qn12jeu9vg", nil)
	if err != nil {
		panic(err)
	}

	conn, err := database.ConnectClickHouse()

	if err != nil {
		log.Fatalf("could not connect to ClickHouse: %v", err)
	}

	defer func() {
		w.Close()
		conn.Close()
	}()

	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Type")
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	// Get user favorite stocks
	symbols := []string{"AAPL", "AMZN"}

	for _, s := range symbols {
		msg, _ := json.Marshal(map[string]interface{}{"type": "subscribe", "symbol": s})
		w.WriteMessage(websocket.TextMessage, msg)
	}

	var msg interface{}

	for {
		if err := w.ReadJSON(&msg); err != nil {
			panic(err)
		}

		var dataContainer model.DataContainer

		stockQuoteJsonByte, jErr := json.Marshal(msg)

		if jErr != nil {
			log.Fatalf("Error on Json: %v", err)
			return
		}

		if err := json.Unmarshal(stockQuoteJsonByte, &dataContainer); err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
			return
		}

		for _, data := range dataContainer.Data {

			createdAt := time.Now().Format("2006-01-02 15:04:05")
			stockQuote := model.StockPrice{
				Symbol:    data.S,
				Price:     data.P,
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
		}

		fmt.Printf("Message from server: %+v\n\n", dataContainer.Data)
		fmt.Fprintf(c.Writer, "data: %+v\n\n", string(stockQuoteJsonByte))

		flusher.Flush()
	}
}
