package model

type HistoricalPrices struct {
	Symbol    string  `json:"symbol"`
	Price     float32 `json:"price"`
	Open      float32 `json:"open"`
	High      float32 `json:"high"`
	Low       float32 `json:"low"`
	PrevClose float32 `json:"prev_close"`
}
