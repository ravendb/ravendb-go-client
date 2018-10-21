package main

import (
	"fmt"
	"github.com/ravendb/ravendb-go-client"
)

var (
	serverURI = "http://live-test.ravendb.net"
	dbName    = "Demo"
)

type Employee struct {
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

type Address struct {
	Line1      string      `json:"Line1"`
	Line2      interface{} `json:"Line2"`
	City       string      `json:"City"`
	Region     interface{} `json:"Region"`
	PostalCode string      `json:"PostalCode"`
	Country    string      `json:"Country"`
	Location   interface{} `json:"Location"`
}

type Order struct {
	Company   string              `json:"Company"`
	Employee  *Employee           `json:"Employee"`
	Freight   float64             `json:"Freight"`
	Lines     []Line              `json:"Lines"`
	OrderedAt string              `json:"OrderedAt"`
	RequireAt string              `json:"RequireAt"`
	ShipTo    ShipTo              `json:"ShipTo"`
	ShipVia   string              `json:"ShipVia"`
	ShippedAt *ravendb.ServerTime `json:"ShippedAt"`
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

func panicIfErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	store := ravendb.NewDocumentStoreWithUrlAndDatabase(serverURI, dbName)
	err := store.Initialize()
	panicIfErr(err)

	{
		session, err := store.OpenSession()
		panicIfErr(err)
		var e *Employee
		err = session.Load(&e, "employees/7-A")
		panicIfErr(err)
		fmt.Printf("employee: %#v\n", e)
	}

	// TODO: not supported yet
	if false {
		session, err := store.OpenSession()
		panicIfErr(err)
		session.Include("employee")
		var o *Order
		err = session.Load(&o, "orders/827-A")
		panicIfErr(err)
		fmt.Printf("order: %#v\n", o)
	}

}
