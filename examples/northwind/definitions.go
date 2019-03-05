package northwind

import "github.com/ravendb/ravendb-go-client"

// definitions for Northwind test database as hosted at https://live-test.ravendb.net
// in database "Demo"
// see https://ravendb.net/docs/article-page/4.1/csharp/start/about-examples

// Category describes a product category
type Category struct {
	ID          string
	Name        string `json:"Name"`
	Description string `json:"Description"`
}

// Contact describes a contact
type Contact struct {
	Name  string
	Title string
}

// Company describes a company
type Company struct {
	ID         string
	Name       string   `json:"Name"`
	ExternalID string   `json:"ExternalId"`
	Phone      string   `json:"Phone,omitempty"`
	Fax        string   `json:"Fax,omitempty"`
	Contact    *Contact `json:"Contact"`
	Address    *Address `json:"Address"`
}

// Employee describes an employee
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
	ReportsTo   string       `json:"ReportsTo"` // id of Employee struct
	Notes       []string     `json:"Notes"`
	Territories []string     `json:"Territories"`
}

// Order describes an order
type Order struct {
	ID        string
	Company   string        `json:"Company"`  // id of Company struct
	Employee  string        `json:"Employee"` // id of Employee struct
	OrderedAt ravendb.Time  `json:"OrderedAt"`
	RequireAt ravendb.Time  `json:"RequireAt"`
	ShippedAt *ravendb.Time `json:"ShippedAt"`
	ShipTo    *Address      `json:"ShipTo"`
	ShipVia   string        `json:"ShipVia"` // id of Shipper struct
	Freight   float64       `json:"Freight"`
	Lines     []*OrderLine  `json:"Lines"`
}

// Product describes a product
type Product struct {
	ID              string
	Name            string  `json:"Name"`
	Supplier        string  `json:"Supplier"` // id of Supplier struct
	Category        string  `json:"Category"` // id of Category struct
	QuantityPerUnit string  `json:"QuantityPerUnit"`
	PricePerUnit    float64 `json:"PricePerUnit"`
	UnitsInStock    int     `json:"UnitsInStock"`
	UnitsOnOrder    int     `json:"UnitsOnOrder"`
	Discontinued    bool    `json:"Discontinued"`
	ReorderLevel    int     `json:"ReorderLevel"`
}

// Region describes a region
type Region struct {
	ID          string
	Name        string      `json:"Name"`
	Territories []Territory `json:"Territories,omitempty"`
}

// Territory describes a territory
type Territory struct {
	Code string `json:"Code"`
	Name string `json:"Name"`
}

// Shipper describes a shipper
type Shipper struct {
	ID     string
	Name   string `json:"Name"`
	Phoene string `json:"Phone"`
}

// Supplier describes a supplier
type Supplier struct {
	ID       string
	Name     string   `json:"Name"`
	Phone    string   `json:"Phone"`
	Fax      string   `json:"Fax,omitempty"`
	HomePage string   `json:"HomePage,omitempty"`
	Contact  *Contact `json:"Contact"`
	Address  *Address `json:"Address"`
}

// Address describes an address
type Address struct {
	Line1      string    `json:"Line1"`
	Line2      string    `json:"Line2,omitempty"`
	City       string    `json:"City"`
	Region     string    `json:"Region,omitempty"`
	PostalCode string    `json:"PostalCode"`
	Country    string    `json:"Country"`
	Location   *Location `json:"Location"`
}

// Location describes a location
type Location struct {
	Latitude  float64
	Longitude float64
}

// OrderLine describes an order line
type OrderLine struct {
	Product      string  `json:"Product"` // id of Product string
	ProductName  string  `json:"ProductName"`
	PricePerUnit float64 `json:"PricePerUnit"`
	Quantity     int     `json:"Quantity"`
	Discount     float64 `json:"Discount"`
}
