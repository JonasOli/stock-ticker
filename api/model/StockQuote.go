package model


type StockData struct {
	C []string   `json:"c"` // assuming 'c' is a list of integers
	P float32 `json:"p"` // price
	S string  `json:"s"` // stock symbol
	T int64   `json:"t"` // timestamp (using Unix timestamp in ms here)
	V int     `json:"v"` // volume
}

type DataContainer struct {
	Data []StockData `json:"data"`
	Type string      `json:"type"`
}

type StockPrice struct {
	Symbol    string  `json:"symbol"`
	Price     float32 `json:"price"`
	CreatedAt string  `json:"createdAt"`
}

type HistoricalPrices struct {
	Symbol    string  `json:"symbol"`
	Price     float32 `json:"price"`
	Open      float32 `json:"open"`
	High      float32 `json:"high"`
	Low       float32 `json:"low"`
	PrevClose float32 `json:"prevClose"`
}
