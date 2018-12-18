package tests

import (
	ravendb "github.com/ravendb/ravendb-go-client"
)

type Order struct {
	ID        string
	Company   string       `json:"company"`
	Employee  string       `json:"employee"`
	OrderedAt ravendb.Time `json:"orderedAt"`
	RequireAt ravendb.Time `json:"requireAt"`
	ShippedAt ravendb.Time `json:"shippedAt"`
	ShipTo    *Address     `json:"shipTo"`
	ShipVia   string       `json:"shipVia"`
	Freight   float64      `json:"freight"`
	Lines     []*OrderLine `json:"lines"`
}
