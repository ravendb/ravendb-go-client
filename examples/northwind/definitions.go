package northwind

import "github.com/ravendb/ravendb-go-client"

// definitions for Northwind test database as hosted at https://live-test.ravendb.net
// in database "Demo"
// see https://ravendb.net/docs/article-page/4.1/csharp/start/about-examples

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
	Address    *Address `json:"Address"`
}

type Employee struct {
	ID          string
	LastName    string       `json:"LastName"`
	FirstName   string       `json:"FirstName"`
	Title       string       `json:"Title"`
	Address     *Address     `json:"Address"`
	HiredAt     ravendb.Time `json:"HiredAt"`
	Birthday    ravendb.Time `json:"Birthday"`
	HomePhone   string       `json:"HomePhone"`
	Extension   string       `json:"Extension"`
	ReportsTo   *Employee    `json:"ReportsTo"`
	Notes       []string     `json:"Notes"`
	Territories []string     `json:"Territories"`
}

type Order struct {
	ID        string
	Company   *Company      `json:"Company"`
	Employee  *Employee     `json:"Employee"`
	OrderedAt ravendb.Time  `json:"OrderedAt"`
	RequireAt ravendb.Time  `json:"RequireAt"`
	ShippedAt *ravendb.Time `json:"ShippedAt"`
	ShipTo    *Address      `json:"ShipTo"`
	ShipVia   *Shipper      `json:"ShipVia"`
	Freight   float64       `json:"Freight"`
	Lines     []*OrderLine  `json:"Lines"`
}

type Product struct {
	ID              string
	Name            string    `json:"Name"`
	Supplier        *Supplier `json:"Supplier"`
	Category        *Category `json:"Category"`
	QuantityPerUnit string    `json:"QuantityPerUnit"`
	PricePerUnit    float64   `json:"PricePerUnit"`
	UnitsInStock    int       `json:"UnitsInStock"`
	UnistsOnOrder   int       `json:"UnistsOnOrder"`
	Discontinued    bool      `json:"Discontinued"`
	ReorderLevel    int       `json:"ReorderLevel"`
}

type Region struct {
	ID          string
	Name        string      `json:"Name"`
	Territories []Territory `json:"Territories,omitempty"`
}

type Territory struct {
	Code string `json:"Code"`
	Name string `json:"Name"`
}

type Shipper struct {
	ID     string
	Name   string `json:"Name"`
	Phoene string `json:"Phone"`
}

type Supplier struct {
	ID       string
	Name     string   `json:"Name"`
	Phone    string   `json:"Phone"`
	Fax      string   `json:"Fax,omitempty"`
	HomePage string   `json:"HomePage,omitempty"`
	Contact  *Contact `json:"Contact"`
	Address  *Address `json:"Address"`
}

type Address struct {
	Line1      string    `json:"Line1"`
	Line2      string    `json:"Line2,omitempty"`
	City       string    `json:"City"`
	Region     string    `json:"Region,omitempty"`
	PostalCode string    `json:"PostalCode"`
	Country    string    `json:"Country"`
	Location   *Location `json:"Location"`
}

type Location struct {
	Latitude  float64
	Longitude float64
}

type OrderLine struct {
	Product      *Product `json:"Product"`
	ProductName  string   `json:"ProductName"`
	PricePerUnit float64  `json:"PricePerUnit"`
	Quantity     int      `json:"Quantity"`
	Discount     float64  `json:"Discount"`
}
