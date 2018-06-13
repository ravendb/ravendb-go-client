package ravendb

type OrderLine struct {
	Product      string
	ProductName  string
	PricePerUnit float64
	Quantity     int
	Discount     float64
}

func (o *OrderLine) getProduct() string {
	return o.Product
}

func (o *OrderLine) setProduct(product string) {
	o.Product = product
}

func (o *OrderLine) getProductName() string {
	return o.ProductName
}

func (o *OrderLine) setProductName(productName string) {
	o.ProductName = productName
}

func (o *OrderLine) getPricePerUnit() float64 {
	return o.PricePerUnit
}

func (o *OrderLine) setPricePerUnit(pricePerUnit float64) {
	o.PricePerUnit = pricePerUnit
}

func (o *OrderLine) getQuantity() int {
	return o.Quantity
}

func (o *OrderLine) setQuantity(quantity int) {
	o.Quantity = quantity
}

func (o *OrderLine) getDiscount() float64 {
	return o.Discount
}

func (o *OrderLine) setDiscount(discount float64) {
	o.Discount = discount
}
