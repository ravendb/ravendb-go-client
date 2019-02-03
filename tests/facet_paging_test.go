package tests

import (
	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
	"sort"
	"strings"
	"testing"
)

const (
	numCameras = 1000
)

var (
	_data []*Camera
)

func facetPaging_canPerformFacetedPagingSearchWithNoPageSizeNoMaxResults_HitsDesc(t *testing.T, driver *RavenTestDriver) {
	facetOptions := ravendb.NewFacetOptions()
	facetOptions.Start = 2
	facetOptions.TermSortMode = ravendb.FacetTermSortModeCountDesc
	facetOptions.IncludeRemainingTerms = true

	facet := ravendb.NewFacet()
	facet.FieldName = "manufacturer"
	facet.Options = facetOptions

	facets := []*ravendb.Facet{facet}

	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	facetPaging_setup(t, store)
	{
		session := openSessionMust(t, store)

		facetSetup := &ravendb.FacetSetup{}
		facetSetup.ID = "facets/CameraFacets"
		facetSetup.Facets = facets

		err = session.Store(facetSetup)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		q := session.QueryInIndexNamed("CameraCost")
		ag := q.AggregateUsing("facets/CameraFacets")
		facetResults, err := ag.Execute()
		assert.NoError(t, err)

		cameraCounts := map[string]int{}
		for _, camera := range _data {
			cameraCounts[camera.Manufacturer]++
		}

		var camerasByHits []string
		for c := range cameraCounts {
			camerasByHits = append(camerasByHits, c)
		}
		sort.Slice(camerasByHits, func(i, j int) bool {
			namei := camerasByHits[i]
			namej := camerasByHits[j]
			ci := cameraCounts[namei]
			cj := cameraCounts[namej]
			if ci != cj {
				return cj < ci // reverse order
			}
			return namej > namei
		})
		camerasByHits = camerasByHits[2:]
		for i, s := range camerasByHits {
			camerasByHits[i] = strings.ToLower(s)
		}

		vals := facetResults["manufacturer"].Values
		assert.Equal(t, len(vals), 3)

		assert.Equal(t, vals[0].Range, camerasByHits[0])
		assert.Equal(t, vals[1].Range, camerasByHits[1])
		assert.Equal(t, vals[2].Range, camerasByHits[2])

		for _, f := range vals {
			fM := strings.ToLower(f.Range)
			inMemoryCount := 0
			for _, camera := range _data {
				camM := strings.ToLower(camera.Manufacturer)
				if camM == fM {
					inMemoryCount++
				}
			}
			assert.Equal(t, f.Count, inMemoryCount)
		}

		fr := facetResults["manufacturer"]
		assert.Equal(t, fr.RemainingTermsCount, 0)
		assert.Equal(t, len(fr.RemainingTerms), 0)
		assert.Equal(t, fr.RemainingHits, 0)

		session.Close()
	}
}

func facetPaging_canPerformFacetedPagingSearchWithNoPageSizeWithMaxResults_HitsDesc(t *testing.T, driver *RavenTestDriver) {
	facetOptions := ravendb.NewFacetOptions()
	facetOptions.Start = 2
	facetOptions.PageSize = 2
	facetOptions.TermSortMode = ravendb.FacetTermSortModeCountDesc
	facetOptions.IncludeRemainingTerms = true

	facet := ravendb.NewFacet()
	facet.FieldName = "manufacturer"
	facet.Options = facetOptions

	facets := []*ravendb.Facet{facet}

	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	facetPaging_setup(t, store)
	{
		session := openSessionMust(t, store)

		facetSetup := &ravendb.FacetSetup{}
		facetSetup.ID = "facets/CameraFacets"
		facetSetup.Facets = facets

		err = session.Store(facetSetup)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		q := session.QueryInIndexNamed("CameraCost")
		ag := q.AggregateUsing("facets/CameraFacets")
		facetResults, err := ag.Execute()
		assert.NoError(t, err)

		cameraCounts := map[string]int{}
		for _, camera := range _data {
			cameraCounts[camera.Manufacturer]++
		}

		var camerasByHits []string
		for c := range cameraCounts {
			camerasByHits = append(camerasByHits, c)
		}
		sort.Slice(camerasByHits, func(i, j int) bool {
			namei := camerasByHits[i]
			namej := camerasByHits[j]
			ci := cameraCounts[namei]
			cj := cameraCounts[namej]
			if ci != cj {
				return cj < ci // reverse order
			}
			return namej > namei
		})
		camerasByHits = camerasByHits[2:]
		if len(camerasByHits) > 2 {
			camerasByHits = camerasByHits[:2]
		}

		for i, s := range camerasByHits {
			camerasByHits[i] = strings.ToLower(s)
		}

		vals := facetResults["manufacturer"].Values
		assert.Equal(t, len(vals), 2)

		assert.Equal(t, vals[0].Range, camerasByHits[0])
		assert.Equal(t, vals[1].Range, camerasByHits[1])

		for _, f := range vals {
			fM := strings.ToLower(f.Range)
			inMemoryCount := 0
			for _, camera := range _data {
				camM := strings.ToLower(camera.Manufacturer)
				if camM == fM {
					inMemoryCount++
				}
			}
			assert.Equal(t, f.Count, inMemoryCount)

		}

		// Note: Java does it inside the above loop, for no reason
		fr := facetResults["manufacturer"]
		assert.Equal(t, fr.RemainingTermsCount, 1)
		assert.Equal(t, len(fr.RemainingTerms), 1)

		var counts []int
		for _, count := range cameraCounts {
			counts = append(counts, count)
		}
		sort.Ints(counts)
		assert.Equal(t, counts[0], fr.RemainingHits)
		// fmt.Printf("Remaining hits: %d, first: %d, last: %d\n", fr.RemainingHits, counts[0], counts[len(counts)-1])

		session.Close()
	}
}

func facetPaging_setup(t *testing.T, store *ravendb.DocumentStore) {

	s, err := store.OpenSession()
	assert.NoError(t, err)
	defer s.Close()

	indexDefinition := &ravendb.IndexDefinition{}
	indexDefinition.Name = "CameraCost"
	indexDefinition.Maps = []string{
		"from camera in docs select new { camera.manufacturer, camera.model, camera.cost, camera.dateOfListing, camera.megapixels } ",
	}

	op := ravendb.NewPutIndexesOperation(indexDefinition)
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)

	counter := 0
	for _, camera := range _data {
		err = s.Store(camera)
		assert.NoError(t, err)
		counter++

		if counter%(numCameras/25) == 0 {
			err = s.SaveChanges()
			assert.NoError(t, err)
		}
	}

	err = s.SaveChanges()
	assert.NoError(t, err)

	err = waitForIndexing(store, "", 0)
	assert.NoError(t, err)
}

func TestFacetPaging(t *testing.T) {
	// t.Parallel()

	_data = facetTestBaseGetCameras(numCameras)

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	facetPaging_canPerformFacetedPagingSearchWithNoPageSizeNoMaxResults_HitsDesc(t, driver)
	facetPaging_canPerformFacetedPagingSearchWithNoPageSizeWithMaxResults_HitsDesc(t, driver)
}
