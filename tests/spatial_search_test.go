package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

func NewSpatialIdx() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("SpatialIdx")
	res.Map = "docs.Events.Select(e => new {\n" +
		"    capacity = e.capacity,\n" +
		"    venue = e.venue,\n" +
		"    date = e.date,\n" +
		"    coordinates = this.CreateSpatialField(((double ? ) e.latitude), ((double ? ) e.longitude))\n" +
		"})"

	res.Index("venue", ravendb.FieldIndexing_SEARCH)
	return res
}

func spatialSearch_can_do_spatial_search_with_client_api(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	err = NewSpatialIdx().Execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		err = session.Store(NewEventWithDate("a/1", 38.9579000, -77.3572000, time.Now()))
		assert.NoError(t, err)
		err = session.Store(NewEventWithDate("a/2", 38.9690000, -77.3862000, ravendb.DateUtils_addDays(time.Now(), 1)))
		assert.NoError(t, err)
		err = session.Store(NewEventWithDate("b/2", 38.9690000, -77.3862000, ravendb.DateUtils_addDays(time.Now(), 2)))
		assert.NoError(t, err)
		err = session.Store(NewEventWithDate("c/3", 38.9510000, -77.4107000, ravendb.DateUtils_addYears(time.Now(), 3)))
		assert.NoError(t, err)
		err = session.Store(NewEventWithDate("d/1", 37.9510000, -77.4107000, ravendb.DateUtils_addYears(time.Now(), 3)))
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	err = gRavenTestDriver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		var statsRef *ravendb.QueryStatistics
		q := session.QueryWithQuery(ravendb.GetTypeOf(&Event{}), ravendb.Query_index("SpatialIdx"))
		q = q.Statistics(&statsRef)
		q = q.WhereLessThanOrEqual("date", ravendb.DateUtils_addYears(time.Now(), 1))
		q = q.WithinRadiusOf("coordinates", 6.0, 38.96939, -77.386398)
		q = q.OrderByDescending("date")
		events, err := q.ToList()
		assert.NoError(t, err)

		assert.True(t, len(events) > 0)

		session.Close()
	}
}

func spatialSearch_can_do_spatial_search_with_client_api3(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewSpatialIdx()
	err = index.Execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		q := session.Advanced().DocumentQueryInIndex(ravendb.GetTypeOf(&Event{}), index)
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

func spatialSearch_can_do_spatial_search_with_client_api_within_given_capacity(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewSpatialIdx()
	err = index.Execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		err = session.Store(NewEventWithDateAndCapacity("a/1", 38.9579000, -77.3572000, time.Now(), 5000))
		assert.NoError(t, err)
		err = session.Store(NewEventWithDateAndCapacity("a/2", 38.9690000, -77.3862000, ravendb.DateUtils_addDays(time.Now(), 1), 5000))
		assert.NoError(t, err)
		err = session.Store(NewEventWithDateAndCapacity("b/2", 38.9690000, -77.3862000, ravendb.DateUtils_addDays(time.Now(), 2), 2000))
		assert.NoError(t, err)
		err = session.Store(NewEventWithDateAndCapacity("c/3", 38.9510000, -77.4107000, ravendb.DateUtils_addYears(time.Now(), 3), 1500))
		assert.NoError(t, err)
		err = session.Store(NewEventWithDateAndCapacity("d/1", 37.9510000, -77.4107000, ravendb.DateUtils_addYears(time.Now(), 3), 1500))
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	err = gRavenTestDriver.waitForIndexing(store, "", 0)

	{
		session := openSessionMust(t, store)

		var queryStats *ravendb.QueryStatistics

		q := session.QueryWithQuery(ravendb.GetTypeOf(&Event{}), ravendb.Query_index("SpatialIdx"))
		q = q.Statistics(&queryStats)
		q = q.OpenSubclause()
		q = q.WhereGreaterThanOrEqual("capacity", 0)
		q = q.AndAlso()
		q = q.WhereLessThanOrEqual("capacity", 2000)
		q = q.CloseSubclause()
		q = q.WithinRadiusOf("coordinates", 6.0, 38.96939, -77.386398)
		q = q.OrderByDescending("date")
		events, err := q.ToList()
		assert.NoError(t, err)

		assert.Equal(t, queryStats.GetTotalResults(), 2)

		var a []string
		for _, event := range events {
			a = append(a, event.(*Event).getVenue())
		}

		assert.True(t, ravendb.StringArrayContainsExactly(a, []string{"c/3", "b/2"}))

		session.Close()
	}
}

func spatialSearch_can_do_spatial_search_with_client_api_add_order(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewSpatialIdx()
	err = index.Execute(store)
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

	err = gRavenTestDriver.waitForIndexing(store, "", 0)

	{
		session := openSessionMust(t, store)

		q := session.QueryWithQuery(ravendb.GetTypeOf(&Event{}), ravendb.Query_index("spatialIdx"))
		q = q.WithinRadiusOf("coordinates", 6.0, 38.96939, -77.386398)
		q = q.OrderByDistanceLatLong("coordinates", 38.96939, -77.386398)
		q = q.AddOrder("venue", false)
		events, err := q.ToList()
		assert.NoError(t, err)

		var a []string
		for _, event := range events {
			a = append(a, event.(*Event).getVenue())
		}
		assert.True(t, ravendb.StringArrayContainsExactly(a, []string{"a/2", "b/2", "c/2", "a/1", "b/1", "c/1", "a/3", "b/3", "c/3"}))

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		q := session.QueryWithQuery(ravendb.GetTypeOf(&Event{}), ravendb.Query_index("spatialIdx"))
		q = q.WithinRadiusOf("coordinates", 6.0, 38.96939, -77.386398)
		q = q.AddOrder("venue", false)
		q = q.OrderByDistanceLatLong("coordinates", 38.96939, -77.386398)
		events, err := q.ToList()
		assert.NoError(t, err)

		var a []string
		for _, event := range events {
			a = append(a, event.(*Event).getVenue())
		}
		assert.True(t, ravendb.StringArrayContainsExactly(a, []string{"a/1", "a/2", "a/3", "b/1", "b/2", "b/3", "c/1", "c/2", "c/3"}))

		session.Close()
	}
}

type Event struct {
	Venue     string    `json:"venue"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Date      time.Time `json:"date"`
	Capacity  int       `json:"capacity"`
}

func NewEvent(venue string, latitude float64, longitude float64) *Event {
	return &Event{
		Venue:     venue,
		Latitude:  latitude,
		Longitude: longitude,
	}
}

func NewEventWithDate(venue string, latitude float64, longitude float64, date time.Time) *Event {
	return &Event{
		Venue:     venue,
		Latitude:  latitude,
		Longitude: longitude,
		Date:      date,
	}
}

func NewEventWithDateAndCapacity(venue string, latitude float64, longitude float64, date time.Time, capacity int) *Event {
	return &Event{
		Venue:     venue,
		Latitude:  latitude,
		Longitude: longitude,
		Date:      date,
		Capacity:  capacity,
	}
}

func (e *Event) getVenue() string {
	return e.Venue
}

func (e *Event) setVenue(venue string) {
	e.Venue = venue
}

func (e *Event) getLatitude() float64 {
	return e.Latitude
}

func (e *Event) setLatitude(latitude float64) {
	e.Latitude = latitude
}

func (e *Event) getLongitude() float64 {
	return e.Longitude
}

func (e *Event) setLongitude(longitude float64) {
	e.Longitude = longitude
}

func (e *Event) getDate() time.Time {
	return e.Date
}

func (e *Event) setDate(date time.Time) {
	e.Date = date
}

func (e *Event) getCapacity() int {
	return e.Capacity
}

func (e *Event) setCapacity(capacity int) {
	e.Capacity = capacity
}

func TestSpatialSearch(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	spatialSearch_can_do_spatial_search_with_client_api3(t)
	spatialSearch_can_do_spatial_search_with_client_api_within_given_capacity(t)
	spatialSearch_can_do_spatial_search_with_client_api_add_order(t)
	spatialSearch_can_do_spatial_search_with_client_api(t)
}
