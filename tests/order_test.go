package tests

import "time"

type Order struct {
	ID        string
	Company   string
	Employee  string
	OrderedAt time.Time
	RequireAt time.Time
	ShippedAt time.Time
	ShipTo    *Address
	ShipVia   string
	Freight   float64
	Lines     []OrderLine
}

func NewOrder() *Order {
	return &Order{}
}

func (o *Order) getId() string {
	return o.ID
}

func (o *Order) setId(id string) {
	o.ID = id
}

func (o *Order) getCompany() string {
	return o.Company
}

func (o *Order) setCompany(company string) {
	o.Company = company
}

func (o *Order) getEmployee() string {
	return o.Employee
}

func (o *Order) setEmployee(employee string) {
	o.Employee = employee
}

func (o *Order) getOrderedAt() time.Time {
	return o.OrderedAt
}

func (o *Order) setOrderedAt(orderedAt time.Time) {
	o.OrderedAt = orderedAt
}

func (o *Order) getRequireAt() time.Time {
	return o.RequireAt
}

func (o *Order) setRequireAt(requireAt time.Time) {
	o.RequireAt = requireAt
}

func (o *Order) getShippedAt() time.Time {
	return o.ShippedAt
}

func (o *Order) setShippedAt(shippedAt time.Time) {
	o.ShippedAt = shippedAt
}

func (o *Order) getShipTo() *Address {
	return o.ShipTo
}

func (o *Order) setShipTo(shipTo *Address) {
	o.ShipTo = shipTo
}

func (o *Order) getShipVia() string {
	return o.ShipVia
}

func (o *Order) setShipVia(shipVia string) {
	o.ShipVia = shipVia
}

func (o *Order) getFreight() float64 {
	return o.Freight
}

func (o *Order) setFreight(freight float64) {
	o.Freight = freight
}

func (o *Order) getLines() []OrderLine {
	return o.Lines
}

func (o *Order) setLines(lines []OrderLine) {
	o.Lines = lines
}
