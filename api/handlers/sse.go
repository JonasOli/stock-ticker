package handlers

import (
    "stockTicker/utils"

    "github.com/gin-gonic/gin"
)

func HandleSSE(c *gin.Context) {
    utils.SSEManagerInstance.StockTickerEventsHandler(c)
}
