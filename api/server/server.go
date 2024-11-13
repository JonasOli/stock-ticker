package server

import (
    "stockTicker/handlers"

    "github.com/gin-gonic/gin"
)

type Server struct {
    engine *gin.Engine
}

func NewServer() *Server {
    engine := gin.Default()
    sseManager := handlers.NewSSEManager()
    
    // engine.GET("/api/messages", handlers.GetMessages)
    // engine.POST("/api/messages", handlers.CreateMessage)
    engine.GET("/stock-events", sseManager.StockTickerLiveUpdates)

    return &Server{engine: engine}
}

func (s *Server) Run(addr string) error {
    return s.engine.Run(addr)
}
