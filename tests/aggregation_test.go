package tests

import (
	"testing"
	"time"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func NewOrdersAll() *ravendb.AbstractIndexCreationTask {
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

func aggregation_canCorrectlyAggregate_Double(t *testing.T, driver *RavenTestDriver) {

	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewOrdersAll()
	err = index.Execute(store, nil, "")
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
		assert.NoError(t, err)

		session.Close()
	}

	err = driver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		q := session.QueryInIndex(index)
		f := ravendb.NewFacetBuilder()
		f.ByField("region").MaxOn("total").MinOn("total")
		q2, err := q.AggregateByFacet(f.GetFacet())
		assert.NoError(t, err)
		result, err := q2.Execute()
		assert.NoError(t, err)

		facetResult := result["region"]

		values := facetResult.Values
		val := values[0]
		assert.Equal(t, val.Count, 2)
		assert.Equal(t, *val.Min, float64(1))
		assert.Equal(t, *val.Max, float64(1.1))

		n := 0
		for _, x := range values {
			if x.Range == "1" {
				n++
			}
		}
		assert.Equal(t, n, 1)

		session.Close()
	}
}

func getFirstFacetValueOfRange(values []*ravendb.FacetValue, rang string) *ravendb.FacetValue {
	for _, x := range values {
		if x.Range == rang {
			return x
		}
	}
	return nil
}

func aggregationCanCorrectlyAggregateMultipleItems(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewOrdersAll()
	err = index.Execute(store, nil, "")
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
		assert.NoError(t, err)

		session.Close()
	}

	err = driver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		q := session.QueryInIndex(index)
		f := ravendb.NewFacetBuilder()
		f.ByField("product").SumOn("total")
		q2, err := q.AggregateByFacet(f.GetFacet())
		assert.NoError(t, err)
		f = ravendb.NewFacetBuilder()
		f.ByField("currency").SumOn("total")
		q2, err = q2.AndAggregateByFacet(f.GetFacet())
		assert.NoError(t, err)
		r, err := q2.Execute()
		assert.NoError(t, err)

		facetResult := r["product"]

		values := facetResult.Values
		assert.Equal(t, len(values), 2)

		x := getFirstFacetValueOfRange(values, "milk")
		assert.Equal(t, *x.Sum, float64(12))

		x = getFirstFacetValueOfRange(values, "iphone")
		assert.Equal(t, *x.Sum, float64(3333))

		facetResult = r["currency"]
		values = facetResult.Values
		assert.Equal(t, len(values), 2)

		x = getFirstFacetValueOfRange(values, "eur")
		assert.Equal(t, *x.Sum, float64(3336))

		x = getFirstFacetValueOfRange(values, "nis")
		assert.Equal(t, *x.Sum, float64(9))

		session.Close()
	}
}

func aggregationCanCorrectlyAggregateMultipleAggregations(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewOrdersAll()
	err = index.Execute(store, nil, "")
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
		assert.NoError(t, err)

		session.Close()
	}

	err = driver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		q := session.QueryInIndex(index)
		f := ravendb.NewFacetBuilder()
		f.ByField("product").MaxOn("total").MinOn("total")
		q2, err := q.AggregateByFacet(f.GetFacet())
		assert.NoError(t, err)
		r, err := q2.Execute()
		assert.NoError(t, err)

		facetResult := r["product"]
		values := facetResult.Values
		assert.Equal(t, len(values), 2)

		x := getFirstFacetValueOfRange(values, "milk")
		assert.Equal(t, *x.Max, float64(9))
		assert.Equal(t, *x.Min, float64(3))

		x = getFirstFacetValueOfRange(values, "iphone")
		assert.Equal(t, *x.Max, float64(3333))
		assert.Equal(t, *x.Min, float64(3333))

		session.Close()
	}
}

func aggregationCanCorrectlyAggregateDisplayName(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewOrdersAll()
	err = index.Execute(store, nil, "")
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
		assert.NoError(t, err)

		session.Close()
	}

	err = driver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		q := session.QueryInIndex(index)
		f := ravendb.NewFacetBuilder()
		f.ByField("product").WithDisplayName("productMax").MaxOn("total")
		q2, err := q.AggregateByFacet(f.GetFacet())
		assert.NoError(t, err)
		f = ravendb.NewFacetBuilder()
		f.ByField("product").WithDisplayName("productMin")
		q2, err = q2.AndAggregateByFacet(f.GetFacet())
		assert.NoError(t, err)
		r, err := q2.Execute()
		assert.NoError(t, err)

		assert.Equal(t, len(r), 2)
		assert.Equal(t, *r["productMax"].Values[0].Max, float64(3333))
		assert.Equal(t, r["productMin"].Values[1].Count, 2)

		session.Close()
	}
}

func aggregationCanCorrectlyAggregateRanges(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewOrdersAll()
	err = index.Execute(store, nil, "")
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
		assert.NoError(t, err)

		session.Close()
	}

	err = driver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		_range := ravendb.NewRangeBuilder("total")

		q := session.QueryInIndex(index)
		f := ravendb.NewFacetBuilder()
		f.ByField("product").SumOn("total")
		q2, err := q.AggregateByFacet(f.GetFacet())
		assert.NoError(t, err)
		f = ravendb.NewFacetBuilder()
		fop := f.ByRanges(
			_range.IsLessThan(100),
			_range.IsGreaterThanOrEqualTo(100).IsLessThan(500),
			_range.IsGreaterThanOrEqualTo(500).IsLessThan(1500),
			_range.IsGreaterThanOrEqualTo(1500))
		fop.SumOn("total")
		facet := f.GetFacet()
		q2, err = q2.AndAggregateByFacet(facet)
		assert.NoError(t, err)
		r, err := q2.Execute()
		assert.NoError(t, err)

		facetResult := r["product"]
		values := facetResult.Values
		assert.Equal(t, len(values), 2)

		x := getFirstFacetValueOfRange(values, "milk")
		assert.Equal(t, *x.Sum, float64(12))

		x = getFirstFacetValueOfRange(values, "iphone")
		assert.Equal(t, *x.Sum, float64(3333))

		facetResult = r["total"]
		values = facetResult.Values
		assert.Equal(t, len(values), 4)

		x = getFirstFacetValueOfRange(values, "total < 100")
		assert.Equal(t, *x.Sum, float64(12))

		x = getFirstFacetValueOfRange(values, "total >= 1500")
		assert.Equal(t, *x.Sum, float64(3333))

		session.Close()
	}
}

func now() ravendb.Time {
	return ravendb.Time(time.Now())
}

func setYears(t2 ravendb.Time, nYear int) ravendb.Time {
	t := time.Time(t2)
	diff := nYear - t.Year()
	t = t.AddDate(diff, 0, 0)
	return ravendb.Time(t)
}

func addYears(t ravendb.Time, nYears int) ravendb.Time {
	t2 := time.Time(t)
	t2 = t2.AddDate(nYears, 0, 0)
	return ravendb.Time(t2)
}

func addDays(t ravendb.Time, nDays int) ravendb.Time {
	t2 := time.Time(t)
	t2 = t2.Add(time.Hour * 24 * time.Duration(nDays))
	return ravendb.Time(t2)
}

func addHours(t ravendb.Time, nHours int) ravendb.Time {
	t2 := time.Time(t)
	t2 = t2.Add(time.Hour * time.Duration(nHours))
	return ravendb.Time(t2)
}

func addMinutes(t ravendb.Time, nMinutes int) ravendb.Time {
	t2 := time.Time(t)
	t2 = t2.Add(time.Minute * time.Duration(nMinutes))
	return ravendb.Time(t2)
}

func aggregationCanCorrectlyAggregateDateTimeDataTypeWithRangeCounts(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewItemsOrdersAll()
	err = index.Execute(store, nil, "")
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		item1 := &ItemsOrder{
			Items: []string{"first", "second"},
			At:    now(),
		}

		item2 := &ItemsOrder{
			Items: []string{"first", "second"},
			At:    addDays(now(), -1),
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

	minValue := setYears(now(), 1980)

	end0 := addDays(now(), -2)
	end1 := addDays(now(), -1)
	end2 := now()

	err = driver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	builder := ravendb.NewRangeBuilder("at")

	{
		session := openSessionMust(t, store)
		q := session.QueryInIndex(index)
		q = q.WhereGreaterThanOrEqual("at", end0)
		f := ravendb.NewFacetBuilder()
		r1 := builder.IsGreaterThanOrEqualTo(minValue)              // all - 4
		r2 := builder.IsGreaterThanOrEqualTo(end0).IsLessThan(end1) // 1
		r3 := builder.IsGreaterThanOrEqualTo(end1).IsLessThan(end2) // 3
		f.ByRanges(r1, r2, r3)
		q2, err := q.AggregateByFacet(f.GetFacet())
		assert.NoError(t, err)
		r, err := q2.Execute()
		assert.NoError(t, err)

		facetResults := r["at"].Values
		assert.Equal(t, facetResults[0].Count, 4)

		assert.Equal(t, facetResults[1].Count, 1)
		assert.Equal(t, facetResults[2].Count, 3)

		session.Close()
	}
}

// code coverage for RangeBuilder.IsLessThanOrEqualTo and RangeBuilder.IsGreaterThan
func goAggregationIsGreaterThanAndIsLessThanOrEqualTo(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewOrdersAll()
	err = index.Execute(store, nil, "")
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		data := []struct {
			currency Currency
			product  string
			total    float64
		}{
			{EUR, "Milk", 3},
			{NIS, "Milk", 9},
			{EUR, "iPhone", 3333},
		}
		for _, d := range data {
			obj := &AggOrder{
				Currency: d.currency,
				Product:  d.product,
				Total:    d.total,
			}
			err = session.Store(obj)
			assert.NoError(t, err)
		}

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	err = driver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		q := session.QueryInIndex(index)
		b := ravendb.NewFacetBuilder()
		b.ByField("product").SumOn("total")
		q2, err := q.AggregateByFacet(b.GetFacet())
		assert.NoError(t, err)
		b = ravendb.NewFacetBuilder()
		rng := ravendb.NewRangeBuilder("total")
		fop := b.ByRanges(
			rng.IsGreaterThan(1),
			rng.IsGreaterThanOrEqualTo(100).IsLessThanOrEqualTo(499),
			rng.IsGreaterThanOrEqualTo(500).IsLessThanOrEqualTo(1499),
			rng.IsGreaterThanOrEqualTo(1500))
		fop.SumOn("total")
		facet := b.GetFacet()
		q2, err = q2.AndAggregateByFacet(facet)
		assert.NoError(t, err)
		r, err := q2.Execute()
		assert.NoError(t, err)

		facetResult := r["product"]
		values := facetResult.Values
		assert.Equal(t, len(values), 2)

		x := getFirstFacetValueOfRange(values, "milk")
		assert.Equal(t, *x.Sum, float64(12))

		x = getFirstFacetValueOfRange(values, "iphone")
		assert.Equal(t, *x.Sum, float64(3333))

		facetResult = r["total"]
		values = facetResult.Values
		assert.Equal(t, len(values), 4)

		x = getFirstFacetValueOfRange(values, "total > 1")
		assert.Equal(t, *x.Sum, float64(3345))

		x = getFirstFacetValueOfRange(values, "total >= 1500")
		assert.Equal(t, *x.Sum, float64(3333))

		session.Close()
	}
}

type ItemsOrder struct {
	Items []string     `json:"items"`
	At    ravendb.Time `json:"at"`
}

func NewItemsOrdersAll() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("ItemsOrders_All")
	res.Map = "docs.ItemsOrders.Select(order => new { order.at,\n" +
		"                          order.items })"
	return res
}

func TestAggregation(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	aggregation_canCorrectlyAggregate_Double(t, driver)
	aggregationCanCorrectlyAggregateRanges(t, driver)
	aggregationCanCorrectlyAggregateMultipleItems(t, driver)
	aggregationCanCorrectlyAggregateMultipleAggregations(t, driver)
	aggregationCanCorrectlyAggregateDateTimeDataTypeWithRangeCounts(t, driver)
	aggregationCanCorrectlyAggregateDisplayName(t, driver)

	// tests unique to go
	goAggregationIsGreaterThanAndIsLessThanOrEqualTo(t, driver)
}
