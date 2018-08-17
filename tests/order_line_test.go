package tests

type OrderLine struct {
	Product      string  `json:"product"`
	ProductName  string  `json:"productName"`
	PricePerUnit float64 `json:"pricePerUnit"`
	Quantity     int     `json:"quantity"`
	Discount     float64 `json:"discount"`
}
