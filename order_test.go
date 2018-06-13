package ravendb

import "time"

type Order struct {
	ID        String
	Company   String
	Employee  String
	OrderedAt time.Time
	RequireAt time.Time
	ShippedAt time.Time
	ShipTo    *Address
	ShipVia   String
	Freight   float64
	Lines     []OrderLine
}

func (o *Order) getId() String {
	return o.ID
}

func (o *Order) setId(id String) {
	o.ID = id
}

func (o *Order) getCompany() String {
	return o.Company
}

func (o *Order) setCompany(company String) {
	o.Company = company
}

func (o *Order) getEmployee() String {
	return o.Employee
}

func (o *Order) setEmployee(employee String) {
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

func (o *Order) getShipVia() String {
	return o.ShipVia
}

func (o *Order) setShipVia(shipVia String) {
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
