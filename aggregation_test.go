package ravendb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func NewOrders_All() *AbstractIndexCreationTask {
	res := NewAbstractIndexCreationTask("Orders_All")
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
	err = index.execute(store)
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

		q := session.queryInIndex(getTypeOf(&Order{}), index)
		builder := func(f IFacetBuilder) {
			f.byField("region").maxOn("total").minOn("total")
		}
		q2 := q.aggregateBy(builder)
		result, err := q2.execute()
		assert.NoError(t, err)

		facetResult := result["region"]

		values := facetResult.getValues()
		val := values[0]
		assert.Equal(t, val.getCount(), 2)
		assert.Equal(t, *val.getMin(), float64(1))
		assert.Equal(t, *val.getMax(), float64(1.1))

		n := 0
		for _, x := range values {
			if x.getRange() == "1" {
				n++
			}
		}
		assert.Equal(t, n, 1)

		session.Close()
	}
}

func getFirstFacetValueOfRange(values []*FacetValue, rang string) *FacetValue {
	for _, x := range values {
		if x.getRange() == rang {
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
	err = index.execute(store)
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

		q := session.queryInIndex(getTypeOf(&AggOrder{}), index)
		builder := func(f IFacetBuilder) {
			f.byField("product").sumOn("total")
		}
		q2 := q.aggregateBy(builder)
		builder2 := func(f IFacetBuilder) {
			f.byField("currency").sumOn("total")
		}
		q2 = q2.andAggregateBy(builder2)
		r, err := q2.execute()
		assert.NoError(t, err)

		facetResult := r["product"]

		values := facetResult.getValues()
		assert.Equal(t, len(values), 2)

		x := getFirstFacetValueOfRange(values, "milk")
		assert.Equal(t, *x.getSum(), float64(12))

		x = getFirstFacetValueOfRange(values, "iphone")
		assert.Equal(t, *x.getSum(), float64(3333))

		facetResult = r["currency"]
		values = facetResult.getValues()
		assert.Equal(t, len(values), 2)

		x = getFirstFacetValueOfRange(values, "eur")
		assert.Equal(t, *x.getSum(), float64(3336))

		x = getFirstFacetValueOfRange(values, "nis")
		assert.Equal(t, *x.getSum(), float64(9))

		session.Close()
	}
}

func aggregation_canCorrectlyAggregate_MultipleAggregations(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewOrders_All()
	err = index.execute(store)
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

		q := session.queryInIndex(getTypeOf(&AggOrder{}), index)
		builder := func(f IFacetBuilder) {
			f.byField("product").maxOn("total").minOn("total")
		}
		q2 := q.aggregateBy(builder)
		r, err := q2.execute()
		assert.NoError(t, err)

		facetResult := r["product"]
		values := facetResult.getValues()
		assert.Equal(t, len(values), 2)

		x := getFirstFacetValueOfRange(values, "milk")
		assert.Equal(t, *x.getMax(), float64(9))
		assert.Equal(t, *x.getMin(), float64(3))

		x = getFirstFacetValueOfRange(values, "iphone")
		assert.Equal(t, *x.getMax(), float64(3333))
		assert.Equal(t, *x.getMin(), float64(3333))

		session.Close()
	}
}

func aggregation_canCorrectlyAggregate_DisplayName(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewOrders_All()
	err = index.execute(store)
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

		q := session.queryInIndex(getTypeOf(&AggOrder{}), index)
		builder := func(f IFacetBuilder) {
			f.byField("product").withDisplayName("productMax").maxOn("total")
		}
		q2 := q.aggregateBy(builder)
		builder2 := func(f IFacetBuilder) {
			f.byField("product").withDisplayName("productMin")
		}
		q2 = q2.andAggregateBy(builder2)
		r, err := q2.execute()
		assert.NoError(t, err)

		assert.Equal(t, len(r), 2)
		assert.Equal(t, *r["productMax"].getValues()[0].getMax(), float64(3333))
		assert.Equal(t, r["productMin"].getValues()[1].getCount(), 2)

		session.Close()
	}
}

func aggregation_canCorrectlyAggregate_Ranges(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewOrders_All()
	err = index.execute(store)
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
		_range := RangeBuilder_forPath("total")

		q := session.queryInIndex(getTypeOf(&Order{}), index)
		builder := func(f IFacetBuilder) {
			f.byField("product").sumOn("total")
		}

		q2 := q.aggregateBy(builder)
		builder2 := func(f IFacetBuilder) {
			fop := f.byRanges(
				_range.isLessThan(100),
				_range.isGreaterThanOrEqualTo(100).isLessThan(500),
				_range.isGreaterThanOrEqualTo(500).isLessThan(1500),
				_range.isGreaterThanOrEqualTo(1500))
			fop.sumOn("total")

		}
		q2 = q2.andAggregateBy(builder2)
		r, err := q2.execute()
		assert.NoError(t, err)

		// Map<String, FacetResult> r = session
		facetResult := r["product"]
		values := facetResult.getValues()
		assert.Equal(t, len(values), 2)

		x := getFirstFacetValueOfRange(values, "milk")
		assert.Equal(t, *x.getSum(), float64(12))

		x = getFirstFacetValueOfRange(values, "iphone")
		assert.Equal(t, *x.getSum(), float64(3333))

		facetResult = r["total"]
		values = facetResult.getValues()
		assert.Equal(t, len(values), 4)

		x = getFirstFacetValueOfRange(values, "total < 100")
		assert.Equal(t, *x.getSum(), float64(12))

		x = getFirstFacetValueOfRange(values, "total >= 1500")
		assert.Equal(t, *x.getSum(), float64(3333))

		session.Close()
	}
}
func aggregation_canCorrectlyAggregate_DateTimeDataType_WithRangeCounts(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewItemsOrders_All()
	err = index.execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		item1 := &ItemsOrder{
			Items: []string{"first", "second"},
			At:    time.Now(),
		}

		item2 := &ItemsOrder{
			Items: []string{"first", "second"},
			At:    DateUtils_addDays(time.Now(), -1),
		}

		item3 := &ItemsOrder{
			Items: []string{"first", "second"},
			At:    time.Now(),
		}

		item4 := &ItemsOrder{
			Items: []string{"first"},
			At:    time.Now(),
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

	// items := []string{"second"}

	minValue := DateUtils_setYears(time.Now(), 1980)

	end0 := DateUtils_addDays(time.Now(), -2)
	end1 := DateUtils_addDays(time.Now(), -1)
	end2 := time.Now()

	err = gRavenTestDriver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	builder := RangeBuilder_forPath("at")

	{
		session := openSessionMust(t, store)
		q := session.queryInIndex(getTypeOf(&ItemsOrder{}), index)
		q = q.whereGreaterThanOrEqual("at", end0)
		fn := func(f IFacetBuilder) {
			r1 := builder.isGreaterThanOrEqualTo(minValue)              // all - 4
			r2 := builder.isGreaterThanOrEqualTo(end0).isLessThan(end1) // 0
			r3 := builder.isGreaterThanOrEqualTo(end1).isLessThan(end2) // 1
			f.byRanges(r1, r2, r3)
		}
		q2 := q.aggregateBy(fn)
		r, err := q2.execute()
		assert.NoError(t, err)

		facetResults := r["at"].getValues()
		assert.Equal(t, facetResults[0].getCount(), 4)

		// TODO: comments in java code don't match the results
		// The times are serialized differently.
		// Go:   "2018-08-12T13:35:05.575851-07:00"
		// Java: "2018-08-13T19:32:16.7240000Z"
		assert.Equal(t, facetResults[1].getCount(), 1) // we get 0
		assert.Equal(t, facetResults[2].getCount(), 3) // we get 1

		session.Close()
	}
}

type ItemsOrder struct {
	Items []string  `json:"items"`
	At    time.Time `json:"at"`
}

func NewItemsOrders_All() *AbstractIndexCreationTask {
	res := NewAbstractIndexCreationTask("ItemsOrders_All")
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
	if gEnableFailingTests {
		aggregation_canCorrectlyAggregate_DateTimeDataType_WithRangeCounts(t)
	}
	aggregation_canCorrectlyAggregate_DisplayName(t)
}
