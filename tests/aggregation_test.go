package tests

import (
	"reflect"
	"testing"
	"time"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func NewOrders_All() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("Orders_All")
	res.Map = "docs.AggOrders.Select(order => new { order.currency,\n" +
		"                          order.product,\n" +
		"                          order.total,\n" +
		"                          order.quantity,\n" +
		"                          order.region,\n" +
		"                          order.at,\n" +
		"                          order.tax })"
	return res
}

type Currency = string

// Note: must rename as it conflicts with Order in order_test.go
type AggOrder struct {
	Currency Currency `json:"currency"`
	Product  string   `json:"product"`
	Total    float64  `json:"total"`
	Region   int      `json:"region"`
}

const (
	EUR = "EUR"
	PLN = "PLN"
	NIS = "NIS"
)

func aggregation_canCorrectlyAggregate_Double(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewOrders_All()
	err = index.Execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		obj := &AggOrder{
			Currency: EUR,
			Product:  "Milk",
			Total:    1.1,
			Region:   1,
		}

		obj2 := &AggOrder{
			Currency: EUR,
			Product:  "Milk",
			Total:    1,
			Region:   1,
		}
		err = session.Store(obj)
		assert.NoError(t, err)
		err = session.Store(obj2)
		assert.NoError(t, err)

		err = session.SaveChanges()

		session.Close()
	}

	err = gRavenTestDriver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		q := session.QueryInIndexOld(reflect.TypeOf(&Order{}), index)
		builder := func(f ravendb.IFacetBuilder) {
			f.ByField("region").MaxOn("total").MinOn("total")
		}
		q2 := q.AggregateBy(builder)
		result, err := q2.Execute()
		assert.NoError(t, err)

		facetResult := result["region"]

		values := facetResult.Values
		val := values[0]
		assert.Equal(t, val.GetCount(), 2)
		assert.Equal(t, *val.GetMin(), float64(1))
		assert.Equal(t, *val.GetMax(), float64(1.1))

		n := 0
		for _, x := range values {
			if x.GetRange() == "1" {
				n++
			}
		}
		assert.Equal(t, n, 1)

		session.Close()
	}
}

func getFirstFacetValueOfRange(values []*ravendb.FacetValue, rang string) *ravendb.FacetValue {
	for _, x := range values {
		if x.GetRange() == rang {
			return x
		}
	}
	return nil
}

func aggregation_canCorrectlyAggregate_MultipleItems(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewOrders_All()
	err = index.Execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		obj := &AggOrder{
			Currency: EUR,
			Product:  "Milk",
			Total:    3,
		}

		obj2 := &AggOrder{
			Currency: NIS,
			Product:  "Milk",
			Total:    9,
		}

		obj3 := &AggOrder{
			Currency: EUR,
			Product:  "iPhone",
			Total:    3333,
		}

		err = session.Store(obj)
		assert.NoError(t, err)
		err = session.Store(obj2)
		assert.NoError(t, err)
		err = session.Store(obj3)
		assert.NoError(t, err)

		err = session.SaveChanges()

		session.Close()
	}

	err = gRavenTestDriver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		q := session.QueryInIndexOld(reflect.TypeOf(&AggOrder{}), index)
		builder := func(f ravendb.IFacetBuilder) {
			f.ByField("product").SumOn("total")
		}
		q2 := q.AggregateBy(builder)
		builder2 := func(f ravendb.IFacetBuilder) {
			f.ByField("currency").SumOn("total")
		}
		q2 = q2.AndAggregateBy(builder2)
		r, err := q2.Execute()
		assert.NoError(t, err)

		facetResult := r["product"]

		values := facetResult.Values
		assert.Equal(t, len(values), 2)

		x := getFirstFacetValueOfRange(values, "milk")
		assert.Equal(t, *x.GetSum(), float64(12))

		x = getFirstFacetValueOfRange(values, "iphone")
		assert.Equal(t, *x.GetSum(), float64(3333))

		facetResult = r["currency"]
		values = facetResult.Values
		assert.Equal(t, len(values), 2)

		x = getFirstFacetValueOfRange(values, "eur")
		assert.Equal(t, *x.GetSum(), float64(3336))

		x = getFirstFacetValueOfRange(values, "nis")
		assert.Equal(t, *x.GetSum(), float64(9))

		session.Close()
	}
}

func aggregation_canCorrectlyAggregate_MultipleAggregations(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewOrders_All()
	err = index.Execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		obj := &AggOrder{
			Currency: EUR,
			Product:  "Milk",
			Total:    3,
		}

		obj2 := &AggOrder{
			Currency: NIS,
			Product:  "Milk",
			Total:    9,
		}

		obj3 := &AggOrder{
			Currency: EUR,
			Product:  "iPhone",
			Total:    3333,
		}

		err = session.Store(obj)
		assert.NoError(t, err)
		err = session.Store(obj2)
		assert.NoError(t, err)
		err = session.Store(obj3)
		assert.NoError(t, err)

		err = session.SaveChanges()

		session.Close()
	}

	err = gRavenTestDriver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		q := session.QueryInIndexOld(reflect.TypeOf(&AggOrder{}), index)
		builder := func(f ravendb.IFacetBuilder) {
			f.ByField("product").MaxOn("total").MinOn("total")
		}
		q2 := q.AggregateBy(builder)
		r, err := q2.Execute()
		assert.NoError(t, err)

		facetResult := r["product"]
		values := facetResult.Values
		assert.Equal(t, len(values), 2)

		x := getFirstFacetValueOfRange(values, "milk")
		assert.Equal(t, *x.GetMax(), float64(9))
		assert.Equal(t, *x.GetMin(), float64(3))

		x = getFirstFacetValueOfRange(values, "iphone")
		assert.Equal(t, *x.GetMax(), float64(3333))
		assert.Equal(t, *x.GetMin(), float64(3333))

		session.Close()
	}
}

func aggregation_canCorrectlyAggregate_DisplayName(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewOrders_All()
	err = index.Execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		obj := &AggOrder{
			Currency: EUR,
			Product:  "Milk",
			Total:    3,
		}

		obj2 := &AggOrder{
			Currency: NIS,
			Product:  "Milk",
			Total:    9,
		}

		obj3 := &AggOrder{
			Currency: EUR,
			Product:  "iPhone",
			Total:    3333,
		}

		err = session.Store(obj)
		assert.NoError(t, err)
		err = session.Store(obj2)
		assert.NoError(t, err)
		err = session.Store(obj3)
		assert.NoError(t, err)

		err = session.SaveChanges()

		session.Close()
	}

	err = gRavenTestDriver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		q := session.QueryInIndexOld(reflect.TypeOf(&AggOrder{}), index)
		builder := func(f ravendb.IFacetBuilder) {
			f.ByField("product").WithDisplayName("productMax").MaxOn("total")
		}
		q2 := q.AggregateBy(builder)
		builder2 := func(f ravendb.IFacetBuilder) {
			f.ByField("product").WithDisplayName("productMin")
		}
		q2 = q2.AndAggregateBy(builder2)
		r, err := q2.Execute()
		assert.NoError(t, err)

		assert.Equal(t, len(r), 2)
		assert.Equal(t, *r["productMax"].Values[0].GetMax(), float64(3333))
		assert.Equal(t, r["productMin"].Values[1].GetCount(), 2)

		session.Close()
	}
}

func aggregation_canCorrectlyAggregate_Ranges(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewOrders_All()
	err = index.Execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		obj := &AggOrder{
			Currency: EUR,
			Product:  "Milk",
			Total:    3,
		}

		obj2 := &AggOrder{
			Currency: NIS,
			Product:  "Milk",
			Total:    9,
		}

		obj3 := &AggOrder{
			Currency: EUR,
			Product:  "iPhone",
			Total:    3333,
		}

		err = session.Store(obj)
		assert.NoError(t, err)
		err = session.Store(obj2)
		assert.NoError(t, err)
		err = session.Store(obj3)
		assert.NoError(t, err)

		err = session.SaveChanges()

		session.Close()
	}

	err = gRavenTestDriver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		_range := ravendb.RangeBuilder_forPath("total")

		q := session.QueryInIndexOld(reflect.TypeOf(&Order{}), index)
		builder := func(f ravendb.IFacetBuilder) {
			f.ByField("product").SumOn("total")
		}

		q2 := q.AggregateBy(builder)
		builder2 := func(f ravendb.IFacetBuilder) {
			fop := f.ByRanges(
				_range.IsLessThan(100),
				_range.IsGreaterThanOrEqualTo(100).IsLessThan(500),
				_range.IsGreaterThanOrEqualTo(500).IsLessThan(1500),
				_range.IsGreaterThanOrEqualTo(1500))
			fop.SumOn("total")

		}
		q2 = q2.AndAggregateBy(builder2)
		r, err := q2.Execute()
		assert.NoError(t, err)

		// Map<String, FacetResult> r = session
		facetResult := r["product"]
		values := facetResult.Values
		assert.Equal(t, len(values), 2)

		x := getFirstFacetValueOfRange(values, "milk")
		assert.Equal(t, *x.GetSum(), float64(12))

		x = getFirstFacetValueOfRange(values, "iphone")
		assert.Equal(t, *x.GetSum(), float64(3333))

		facetResult = r["total"]
		values = facetResult.Values
		assert.Equal(t, len(values), 4)

		x = getFirstFacetValueOfRange(values, "total < 100")
		assert.Equal(t, *x.GetSum(), float64(12))

		x = getFirstFacetValueOfRange(values, "total >= 1500")
		assert.Equal(t, *x.GetSum(), float64(3333))

		session.Close()
	}
}

func now() time.Time {
	return time.Now()
}

func aggregation_canCorrectlyAggregate_DateTimeDataType_WithRangeCounts(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewItemsOrders_All()
	err = index.Execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		item1 := &ItemsOrder{
			Items: []string{"first", "second"},
			At:    now(),
		}

		item2 := &ItemsOrder{
			Items: []string{"first", "second"},
			At:    ravendb.DateUtils_addDays(now(), -1),
		}

		item3 := &ItemsOrder{
			Items: []string{"first", "second"},
			At:    now(),
		}

		item4 := &ItemsOrder{
			Items: []string{"first"},
			At:    now(),
		}

		err = session.Store(item1)
		assert.NoError(t, err)
		err = session.Store(item2)
		assert.NoError(t, err)
		err = session.Store(item3)
		assert.NoError(t, err)
		err = session.Store(item4)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	minValue := ravendb.DateUtils_setYears(now(), 1980)

	end0 := ravendb.DateUtils_addDays(now(), -2)
	end1 := ravendb.DateUtils_addDays(now(), -1)
	end2 := now()

	err = gRavenTestDriver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	builder := ravendb.RangeBuilder_forPath("at")

	{
		session := openSessionMust(t, store)
		q := session.QueryInIndexOld(reflect.TypeOf(&ItemsOrder{}), index)
		q = q.WhereGreaterThanOrEqual("at", end0)
		fn := func(f ravendb.IFacetBuilder) {
			r1 := builder.IsGreaterThanOrEqualTo(minValue)              // all - 4
			r2 := builder.IsGreaterThanOrEqualTo(end0).IsLessThan(end1) // 0
			r3 := builder.IsGreaterThanOrEqualTo(end1).IsLessThan(end2) // 1
			f.ByRanges(r1, r2, r3)
		}
		q2 := q.AggregateBy(fn)
		r, err := q2.Execute()
		assert.NoError(t, err)

		facetResults := r["at"].Values
		assert.Equal(t, facetResults[0].GetCount(), 4)

		// TODO: comments in java code don't match the results
		// The times are serialized differently.
		// Go:   "2018-08-12T13:35:05.575851-07:00"
		// Java: "2018-08-13T19:32:16.7240000Z"
		assert.Equal(t, facetResults[1].GetCount(), 1) // we get 0
		assert.Equal(t, facetResults[2].GetCount(), 3) // we get 1

		session.Close()
	}
}

type ItemsOrder struct {
	Items []string  `json:"items"`
	At    time.Time `json:"at"`
}

func NewItemsOrders_All() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("ItemsOrders_All")
	res.Map = "docs.ItemsOrders.Select(order => new { order.at,\n" +
		"                          order.items })"
	return res
}

func TestAggregation(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	aggregation_canCorrectlyAggregate_Double(t)
	aggregation_canCorrectlyAggregate_Ranges(t)
	aggregation_canCorrectlyAggregate_MultipleItems(t)
	aggregation_canCorrectlyAggregate_MultipleAggregations(t)
	if ravendb.EnableFailingTests {
		aggregation_canCorrectlyAggregate_DateTimeDataType_WithRangeCounts(t)
	}
	aggregation_canCorrectlyAggregate_DisplayName(t)
}
