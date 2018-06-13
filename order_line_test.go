package ravendb

type OrderLine struct {
	Product      String
	ProductName  String
	PricePerUnit float64
	Quantity     int
	Discount     float64
}

func (o *OrderLine) getProduct() String {
	return o.Product
}

func (o *OrderLine) setProduct(product String) {
	o.Product = product
}

func (o *OrderLine) getProductName() String {
	return o.ProductName
}

func (o *OrderLine) setProductName(productName String) {
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
