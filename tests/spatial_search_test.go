package tests

import (
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func NewSpatialIdx() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("SpatialIdx")
	res.Map = "docs.Events.Select(e => new {\n" +
		"    capacity = e.capacity,\n" +
		"    venue = e.venue,\n" +
		"    date = e.date,\n" +
		"    coordinates = this.CreateSpatialField(((double ? ) e.latitude), ((double ? ) e.longitude))\n" +
		"})"

	res.Index("venue", ravendb.FieldIndexingSearch)
	return res
}

func spatialSearchCanDoSpatialSearchWithClientApi(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	err = NewSpatialIdx().Execute(store, nil, "")
	assert.NoError(t, err)

	now := now()

	{
		session := openSessionMust(t, store)

		err = session.Store(NewEventWithDate("a/1", 38.9579000, -77.3572000, now))
		assert.NoError(t, err)
		err = session.Store(NewEventWithDate("a/2", 38.9690000, -77.3862000, addDays(now, 1)))
		assert.NoError(t, err)
		err = session.Store(NewEventWithDate("b/2", 38.9690000, -77.3862000, addDays(now, 2)))
		assert.NoError(t, err)
		err = session.Store(NewEventWithDate("c/3", 38.9510000, -77.4107000, addYears(now, 3)))
		assert.NoError(t, err)
		err = session.Store(NewEventWithDate("d/1", 37.9510000, -77.4107000, addYears(now, 3)))
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	err = driver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		var events []*Event
		var statsRef *ravendb.QueryStatistics
		q, err := session.QueryIndex("SpatialIdx")
		assert.NoError(t, err)
		q = q.Statistics(&statsRef)
		q = q.WhereLessThanOrEqual("date", addYears(now, 1))
		q = q.WithinRadiusOf("coordinates", 6.0, 38.96939, -77.386398)
		q = q.OrderByDescending("date")
		err = q.GetResults(&events)
		assert.NoError(t, err)

		assert.True(t, len(events) > 0)

		session.Close()
	}
}

func spatialSearchCanDoSpatialSearchWithClientApi3(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewSpatialIdx()
	err = index.Execute(store, nil, "")
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		q, err := session.Advanced().QueryIndex(index.IndexName)
		assert.NoError(t, err)
		fn := func(f *ravendb.SpatialCriteriaFactory) ravendb.SpatialCriteria {
			return f.WithinRadius(5, 38.9103000, -77.3942)
		}
		q = q.Spatial3("coordinates", fn)
		matchingVenues := q.WaitForNonStaleResults(0)
		iq := matchingVenues.GetIndexQuery()

		assert.Equal(t, iq.GetQuery(), "from index 'SpatialIdx' where spatial.within(coordinates, spatial.circle($p0, $p1, $p2))")
		assert.Equal(t, iq.GetQueryParameters()["p0"], 5.0)
		assert.Equal(t, iq.GetQueryParameters()["p1"], 38.9103)
		assert.Equal(t, iq.GetQueryParameters()["p2"], -77.3942)

		session.Close()
	}
}

func spatialSearchCanDoSpatialSearchWithClientApiWithinGivenCapacity(t *testing.T, driver *RavenTestDriver) {

	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewSpatialIdx()
	err = index.Execute(store, nil, "")
	assert.NoError(t, err)

	now := now()
	{
		session := openSessionMust(t, store)

		err = session.Store(NewEventWithDateAndCapacity("a/1", 38.9579000, -77.3572000, now, 5000))
		assert.NoError(t, err)
		err = session.Store(NewEventWithDateAndCapacity("a/2", 38.9690000, -77.3862000, addDays(now, 1), 5000))
		assert.NoError(t, err)
		err = session.Store(NewEventWithDateAndCapacity("b/2", 38.9690000, -77.3862000, addDays(now, 2), 2000))
		assert.NoError(t, err)
		err = session.Store(NewEventWithDateAndCapacity("c/3", 38.9510000, -77.4107000, addYears(now, 3), 1500))
		assert.NoError(t, err)
		err = session.Store(NewEventWithDateAndCapacity("d/1", 37.9510000, -77.4107000, addYears(now, 3), 1500))
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	err = driver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		var queryStats *ravendb.QueryStatistics

		var events []*Event
		q, err := session.QueryIndex("SpatialIdx")
		assert.NoError(t, err)
		q = q.Statistics(&queryStats)
		q = q.OpenSubclause()
		q = q.WhereGreaterThanOrEqual("capacity", 0)
		q = q.AndAlso()
		q = q.WhereLessThanOrEqual("capacity", 2000)
		q = q.CloseSubclause()
		q = q.WithinRadiusOf("coordinates", 6.0, 38.96939, -77.386398)
		q = q.OrderByDescending("date")
		err = q.GetResults(&events)
		assert.NoError(t, err)

		assert.Equal(t, queryStats.TotalResults, 2)

		var a []string
		for _, event := range events {
			a = append(a, event.Venue)
		}

		assert.True(t, stringArrayContainsExactly(a, []string{"c/3", "b/2"}))

		session.Close()
	}
}

func spatialSearchCanDoSpatialSearchWithClientApiAddOrder(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewSpatialIdx()
	err = index.Execute(store, nil, "")
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		err = session.Store(NewEvent("a/1", 38.9579000, -77.3572000))
		assert.NoError(t, err)
		err = session.Store(NewEvent("b/1", 38.9579000, -77.3572000))
		assert.NoError(t, err)
		err = session.Store(NewEvent("c/1", 38.9579000, -77.3572000))
		assert.NoError(t, err)
		err = session.Store(NewEvent("a/2", 38.9690000, -77.3862000))
		assert.NoError(t, err)
		err = session.Store(NewEvent("b/2", 38.9690000, -77.3862000))
		assert.NoError(t, err)
		err = session.Store(NewEvent("c/2", 38.9690000, -77.3862000))
		assert.NoError(t, err)
		err = session.Store(NewEvent("a/3", 38.9510000, -77.4107000))
		assert.NoError(t, err)
		err = session.Store(NewEvent("b/3", 38.9510000, -77.4107000))
		assert.NoError(t, err)
		err = session.Store(NewEvent("c/3", 38.9510000, -77.4107000))
		assert.NoError(t, err)
		err = session.Store(NewEvent("d/1", 37.9510000, -77.4107000))
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	err = driver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		q, err := session.QueryIndex("spatialIdx")
		assert.NoError(t, err)
		q = q.WithinRadiusOf("coordinates", 6.0, 38.96939, -77.386398)
		q = q.OrderByDistanceLatLong("coordinates", 38.96939, -77.386398)
		q = q.AddOrder("venue", false)
		var events []*Event
		err = q.GetResults(&events)
		assert.NoError(t, err)

		var a []string
		for _, event := range events {
			a = append(a, event.Venue)
		}
		assert.True(t, stringArrayContainsExactly(a, []string{"a/2", "b/2", "c/2", "a/1", "b/1", "c/1", "a/3", "b/3", "c/3"}))

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		q, err := session.QueryIndex("spatialIdx")
		assert.NoError(t, err)
		q = q.WithinRadiusOf("coordinates", 6.0, 38.96939, -77.386398)
		q = q.AddOrder("venue", false)
		q = q.OrderByDistanceLatLong("coordinates", 38.96939, -77.386398)
		var events []*Event
		err = q.GetResults(&events)
		assert.NoError(t, err)

		var a []string
		for _, event := range events {
			a = append(a, event.Venue)
		}
		assert.True(t, stringArrayContainsExactly(a, []string{"a/1", "a/2", "a/3", "b/1", "b/2", "b/3", "c/1", "c/2", "c/3"}))

		session.Close()
	}
}

type Event struct {
	Venue     string       `json:"venue"`
	Latitude  float64      `json:"latitude"`
	Longitude float64      `json:"longitude"`
	Date      ravendb.Time `json:"date"`
	Capacity  int          `json:"capacity"`
}

func NewEvent(venue string, latitude float64, longitude float64) *Event {
	return &Event{
		Venue:     venue,
		Latitude:  latitude,
		Longitude: longitude,
	}
}

func NewEventWithDate(venue string, latitude float64, longitude float64, date ravendb.Time) *Event {
	return &Event{
		Venue:     venue,
		Latitude:  latitude,
		Longitude: longitude,
		Date:      date,
	}
}

func NewEventWithDateAndCapacity(venue string, latitude float64, longitude float64, date ravendb.Time, capacity int) *Event {
	return &Event{
		Venue:     venue,
		Latitude:  latitude,
		Longitude: longitude,
		Date:      date,
		Capacity:  capacity,
	}
}

func TestSpatialSearch(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	spatialSearchCanDoSpatialSearchWithClientApi3(t, driver)
	spatialSearchCanDoSpatialSearchWithClientApiWithinGivenCapacity(t, driver)
	spatialSearchCanDoSpatialSearchWithClientApiAddOrder(t, driver)
	spatialSearchCanDoSpatialSearchWithClientApi(t, driver)
}
