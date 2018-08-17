package tests

import "time"

type Order struct {
	ID        string
	Company   string       `json:"company"`
	Employee  string       `json:"employee"`
	OrderedAt time.Time    `json:"orderedAt"`
	RequireAt time.Time    `json:"requiredAt"`
	ShippedAt time.Time    `json:"shippedAt"`
	ShipTo    *Address     `json:"shipTo"`
	ShipVia   string       `json:"shipVia"`
	Freight   float64      `json:"freight"`
	Lines     []*OrderLine `json:"lines"`
}
