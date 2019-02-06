package main

import ravendb "github.com/ravendb/ravendb-go-client"

type Category struct {
	ID          string
	Name        string `json:"Name"`
	Description string `json:"Description"`
}

type Contact struct {
	Name  string
	Title string
}

type Company struct {
	ID         string
	Name       string   `json:"Name"`
	ExternalID string   `json:"ExternalId"`
	Phone      string   `json:"Phone,omitempty"`
	Fax        string   `json:"Fax,omitempty"`
	Contact    *Contact `json:"Contact"`
}

type Employee struct {
	ID          string
	LastName    string   `json:"LastName"`
	FirstName   string   `json:"FirstName"`
	Title       string   `json:"Title"`
	Address     *Address `json:"Address"`
	HiredAt     string   `json:"HiredAt"`
	Birthday    string   `json:"Birthday"`
	HomePhone   string   `json:"HomePhone"`
	Extension   string   `json:"Extension"`
	ReportsTo   string   `json:"ReportsTo"`
	Notes       []string `json:"Notes"`
	Territories []string `json:"Territories"`
}

type Order struct {
	ID        string
	Company   string        `json:"Company"`
	Employee  *Employee     `json:"Employee"`
	Freight   float64       `json:"Freight"`
	Lines     []Line        `json:"Lines"`
	OrderedAt string        `json:"OrderedAt"`
	RequireAt string        `json:"RequireAt"`
	ShipTo    ShipTo        `json:"ShipTo"`
	ShipVia   string        `json:"ShipVia"`
	ShippedAt *ravendb.Time `json:"ShippedAt"`
}

type Product struct {
	// TODO: add json tags
	ID              string
	Name            string
	Supplier        *Supplier
	Category        *Category
	QuantityPerUnit string
	PricePerUnit    float64
	UnitsInStock    int
	UnistsOnOrder   int
	Discontinued    bool
	ReorderLevel    int
}

type Region struct {
	// TODO: define me
	ID string
}

type Shipper struct {
	ID string
}

type Supplier struct {
	ID       string
	Contact  *Company `json:"Contact"`
	Name     string   `json:"Name"`
	Address  *Address `json:"Address"`
	Phone    string   `json:"Phone"`
	Fax      string   `json:"Fax,omitempty"`
	HomePage string   `json:"HomePage,omitempty"`
}

type Address struct {
	Line1      string      `json:"Line1"`
	Line2      interface{} `json:"Line2"`
	City       string      `json:"City"`
	Region     interface{} `json:"Region"`
	PostalCode string      `json:"PostalCode"`
	Country    string      `json:"Country"`
	Location   interface{} `json:"Location"`
}

type Line struct {
	Discount     float64 `json:"Discount"`
	PricePerUnit float64 `json:"PricePerUnit"`
	Product      string  `json:"Product"`
	ProductName  string  `json:"ProductName"`
	Quantity     int64   `json:"Quantity"`
}

type ShipTo struct {
	City       string      `json:"City"`
	Country    string      `json:"Country"`
	Line1      string      `json:"Line1"`
	Line2      interface{} `json:"Line2"`
	Location   interface{} `json:"Location"`
	PostalCode string      `json:"PostalCode"`
	Region     interface{} `json:"Region"`
}
