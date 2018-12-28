package tests

import (
	"reflect"
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func ravendb_8761_can_group_by_array_values(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	ravendb_8761_putDocs(t, store)

	{
		session := openSessionMust(t, store)

		var productCounts1 []*ProductCount
		q := session.Advanced().RawQuery("from Orders group by lines[].product\n" +
			"  order by count()\n" +
			"  select key() as productName, count() as count")
		q = q.WaitForNonStaleResults()
		err = q.ToList(&productCounts1)
		assert.NoError(t, err)

		q2 := session.Advanced().DocumentQueryOld(reflect.TypeOf(&Order{}))
		q3 := q2.GroupBy("lines[].product")
		q3 = q3.SelectKeyWithNameAndProjectedName("", "productName")
		q2 = q3.SelectCount()
		var productCounts2 []*ProductCount
		err = q2.ToList(&productCounts2)
		assert.NoError(t, err)

		combined := [][]*ProductCount{productCounts1, productCounts2}
		for _, products := range combined {
			assert.Equal(t, len(products), 2)

			product := products[0]
			assert.Equal(t, product.ProductName, "products/1")
			assert.Equal(t, product.Count, 1)

			product = products[1]
			assert.Equal(t, product.ProductName, "products/2")
			assert.Equal(t, product.Count, 2)
		}

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var productCounts1 []*ProductCount
		q := session.Advanced().RawQuery("from Orders\n" +
			" group by lines[].product, shipTo.country\n" +
			" order by count() \n" +
			" select lines[].product as productName, shipTo.country as country, count() as count")
		err = q.ToList(&productCounts1)
		assert.NoError(t, err)

		var productCounts2 []*ProductCount
		q2 := session.Advanced().DocumentQueryOld(reflect.TypeOf(&Order{}))
		q3 := q2.GroupBy("lines[].product", "shipTo.country")
		q3 = q3.SelectKeyWithNameAndProjectedName("lines[].product", "productName")
		q3 = q3.SelectKeyWithNameAndProjectedName("shipTo.country", "country")
		q2 = q3.SelectCount()
		err = q2.ToList(&productCounts2)
		assert.NoError(t, err)

		combined := [][]*ProductCount{productCounts1, productCounts2}
		for _, products := range combined {
			assert.Equal(t, len(products), 2)

			product := products[0]
			assert.Equal(t, product.ProductName, "products/1")
			assert.Equal(t, product.Count, 1)
			assert.Equal(t, product.Country, "USA")

			product = products[1]
			assert.Equal(t, product.ProductName, "products/2")
			assert.Equal(t, product.Count, 2)
			assert.Equal(t, product.Country, "USA")
		}

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var productCounts1 []*ProductCount
		q := session.Advanced().RawQuery("from Orders\n" +
			" group by lines[].product, lines[].quantity\n" +
			" order by lines[].quantity\n" +
			" select lines[].product as productName, lines[].quantity as quantity, count() as count")
		err = q.ToList(&productCounts1)
		assert.NoError(t, err)

		var productCounts2 []*ProductCount
		q2 := session.Advanced().DocumentQueryOld(reflect.TypeOf(&Order{}))
		q3 := q2.GroupBy("lines[].product", "lines[].quantity")
		q3 = q3.SelectKeyWithNameAndProjectedName("lines[].product", "productName")
		q3 = q3.SelectKeyWithNameAndProjectedName("lines[].quantity", "quantity")
		q2 = q3.SelectCount()
		err = q2.ToList(&productCounts2)
		assert.NoError(t, err)

		combined := [][]*ProductCount{productCounts1, productCounts2}
		for _, products := range combined {
			assert.Equal(t, len(products), 3)

			product := products[0]
			assert.Equal(t, product.ProductName, "products/1")

			assert.Equal(t, product.Count, 1)
			assert.Equal(t, product.Quantity, 1)

			product = products[1]
			assert.Equal(t, product.ProductName, "products/2")
			assert.Equal(t, product.Count, 1)
			assert.Equal(t, product.Quantity, 2)

			product = products[2]
			assert.Equal(t, product.ProductName, "products/2")
			assert.Equal(t, product.Count, 1)
			assert.Equal(t, product.Quantity, 3)
		}

		session.Close()
	}
}

func ravendb_8761_can_group_by_array_content(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	ravendb_8761_putDocs(t, store)

	{
		session := openSessionMust(t, store)

		orderLine1 := &OrderLine{
			Product:  "products/1",
			Quantity: 1,
		}

		orderLine2 := &OrderLine{
			Product:  "products/2",
			Quantity: 2,
		}

		address := &Address{
			Country: "USA",
		}

		order := &Order{
			ShipTo: address,
			Lines:  []*OrderLine{orderLine1, orderLine2},
		}

		err = session.Store(order)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var productCounts1 []*ProductCount
		q := session.Advanced().RawQuery("from Orders group by array(lines[].product)\n" +
			" order by count()\n" +
			" select key() as products, count() as count")
		q = q.WaitForNonStaleResults()
		err = q.ToList(&productCounts1)
		assert.NoError(t, err)

		q2 := session.Advanced().DocumentQueryOld(reflect.TypeOf(&Order{}))
		q3 := q2.GroupBy2(ravendb.GroupBy_array("lines[].product"))
		q3 = q3.SelectKeyWithNameAndProjectedName("", "products")
		q2 = q3.SelectCount()
		q2 = q2.OrderBy("count")
		var productCounts2 []*ProductCount
		err = q2.ToList(&productCounts2)
		assert.NoError(t, err)

		combined := [][]*ProductCount{productCounts1, productCounts2}
		for _, products := range combined {
			assert.Equal(t, len(products), 2)

			product := products[0]
			assert.Equal(t, product.Products, []string{"products/2"})
			assert.Equal(t, product.Count, 1)

			product = products[1]
			assert.Equal(t, product.Products, []string{"products/1", "products/2"})

			assert.Equal(t, product.Count, 2)
		}

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var productCounts1 []*ProductCount
		q := session.Advanced().RawQuery("from Orders\n" +
			" group by array(lines[].product), shipTo.country\n" +
			" order by count()\n" +
			" select lines[].product as products, shipTo.country as country, count() as count")
		q = q.WaitForNonStaleResults()
		err = q.ToList(&productCounts1)
		assert.NoError(t, err)

		q2 := session.Advanced().DocumentQueryOld(reflect.TypeOf(&Order{}))
		q3 := q2.GroupBy2(ravendb.GroupBy_array("lines[].product"), ravendb.GroupBy_field("shipTo.country"))
		q3 = q3.SelectKeyWithNameAndProjectedName("lines[].product", "products")
		q2 = q3.SelectCount()
		q2 = q2.OrderBy("count")
		var productCounts2 []*ProductCount
		err = q2.ToList(&productCounts2)
		assert.NoError(t, err)

		combined := [][]*ProductCount{productCounts1, productCounts2}
		for _, products := range combined {
			assert.Equal(t, len(products), 2)

			product := products[0]
			assert.Equal(t, product.Products, []string{"products/2"})
			assert.Equal(t, product.Count, 1)

			product = products[1]
			assert.Equal(t, product.Products, []string{"products/1", "products/2"})

			assert.Equal(t, product.Count, 2)
		}

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var productCounts1 []*ProductCount
		q := session.Advanced().RawQuery("from Orders\n" +
			" group by array(lines[].product), array(lines[].quantity)\n" +
			" order by lines[].quantity\n" +
			" select lines[].product as products, lines[].quantity as quantities, count() as count")
		q = q.WaitForNonStaleResults()
		err = q.ToList(&productCounts1)
		assert.NoError(t, err)

		q2 := session.Advanced().DocumentQueryOld(reflect.TypeOf(&Order{}))
		q3 := q2.GroupBy2(ravendb.GroupBy_array("lines[].product"), ravendb.GroupBy_array("lines[].quantity"))
		q3 = q3.SelectKeyWithNameAndProjectedName("lines[].product", "products")
		q3 = q3.SelectKeyWithNameAndProjectedName("lines[].quantity", "quantities")
		q2 = q3.SelectCount()
		q2 = q2.OrderBy("count")
		var productCounts2 []*ProductCount
		err = q2.ToList(&productCounts2)
		assert.NoError(t, err)

		combined := [][]*ProductCount{productCounts1, productCounts2}
		for _, products := range combined {
			assert.Equal(t, len(products), 2)

			product := products[0]
			assert.Equal(t, product.Products, []string{"products/2"})

			assert.Equal(t, product.Count, 1)
			assert.Equal(t, product.Quantities, []int{3})

			product = products[1]
			assert.Equal(t, product.Products, []string{"products/1", "products/2"})
			assert.Equal(t, product.Count, 2)
			assert.Equal(t, product.Quantities, []int{1, 2})

		}

		session.Close()
	}
}

type ProductCount struct {
	ProductName string   `json:"productName"`
	Count       int      `json:"count"`
	Country     string   `json:"country"`
	Quantity    int      `json:"quantity"`
	Products    []string `json:"products"`
	Quantities  []int    `json:"quantities"`
}

func ravendb_8761_putDocs(t *testing.T, store *ravendb.IDocumentStore) {
	var err error

	session := openSessionMust(t, store)
	order1 := &Order{}

	orderLine11 := &OrderLine{
		Product:  "products/1",
		Quantity: 1,
	}

	orderLine12 := &OrderLine{
		Product:  "products/2",
		Quantity: 2,
	}

	order1.Lines = []*OrderLine{orderLine11, orderLine12}

	address1 := &Address{
		Country: "USA",
	}

	order1.ShipTo = address1

	err = session.Store(order1)
	assert.NoError(t, err)

	orderLine21 := &OrderLine{
		Product:  "products/2",
		Quantity: 3,
	}

	address2 := &Address{
		Country: "USA",
	}
	order2 := &Order{
		Lines:  []*OrderLine{orderLine21},
		ShipTo: address2,
	}

	err = session.Store(order2)
	assert.NoError(t, err)

	err = session.SaveChanges()
	assert.NoError(t, err)

	session.Close()
}

func TestRavenDB8761(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches the order of Java tests
	ravendb_8761_can_group_by_array_content(t, driver)

	ravendb_8761_can_group_by_array_values(t, driver)
}
