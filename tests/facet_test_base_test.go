package tests

import (
	"github.com/ravendb/ravendb-go-client"
	"math/rand"
	"time"
)

var (
	FEATURES      = []string{"Image Stabilizer", "Tripod", "Low Light Compatible", "Fixed Lens", "LCD"}
	MANUFACTURERS = []string{"Sony", "Nikon", "Phillips", "Canon", "Jessops"}
	MODELS        = []string{"Model1", "Model2", "Model3", "Model4", "Model5"}
	RANDOM        = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func NewCameraCostIndex() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("CameraCost")
	m := `from camera in docs.Cameras select new  { camera.manufacturer,
                            camera.model,
                            camera.cost,
                            camera.dateOfListing,
                            camera.megapixels }`
	res.Maps = []string{m}
	return res
}

func facetTestBaseCreateCameraTestIndex(store *ravendb.DocumentStore) error {
	index := NewCameraCostIndex()
	indexDef := index.CreateIndexDefinition()
	op := ravendb.NewPutIndexesOperation(indexDef)
	return store.Maintenance().Send(op)
}

func facetTestBaseInsertCameraData(store *ravendb.DocumentStore, cameras []*Camera, shouldWaitForIndexing bool) error {
	session, err := store.OpenSession()
	if err != nil {
		return err
	}
	defer session.Close()
	for _, camera := range cameras {
		err = session.Store(camera)
		if err != nil {
			return err
		}
	}

	err = session.SaveChanges()
	if err != nil {
		return err
	}

	if !shouldWaitForIndexing {
		return nil
	}

	return waitForIndexing(store, "", 0)
}

func facetTestBaseGetFacets() []ravendb.FacetBase {
	facet1 := ravendb.NewFacet()
	facet1.FieldName = "manufacturer"

	costRangeFacet := ravendb.NewRangeFacet(nil)
	costRangeFacet.Ranges = []string{
		"cost <= 200",
		"cost >= 200 and cost <= 400",
		"cost >= 400 and cost <= 600",
		"cost >= 600 and cost <= 800",
		"cost >= 800",
	}

	megaPixelsRangeFacet := ravendb.NewRangeFacet(nil)
	megaPixelsRangeFacet.Ranges = []string{
		"megapixels <= 3",
		"megapixels >= 3 and megapixels <= 7",
		"megapixels >= 7 and megapixels <= 10",
		"megapixels >= 10",
	}
	return []ravendb.FacetBase{facet1, costRangeFacet, megaPixelsRangeFacet}
}

func facetTestBaseGetCameras(numCameras int) []*Camera {
	var cameraList []*Camera
	for i := 1; i <= numCameras; i++ {
		camera := &Camera{}
		y := 1980 + RANDOM.Intn(30)
		m := time.Month(RANDOM.Intn(12))
		d := RANDOM.Intn(27)
		t := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
		camera.DateOfListing = ravendb.Time(t)
		camera.Manufacturer = MANUFACTURERS[RANDOM.Intn(len(MANUFACTURERS))]
		camera.Model = MODELS[RANDOM.Intn(len(MODELS))]
		camera.Cost = RANDOM.Float64()*900 + 100
		camera.Zoom = RANDOM.Intn(10) + 1
		camera.Megapixels = RANDOM.Float64()*10 + 1.0
		camera.ImageStabilizer = RANDOM.Float64() > 0.6
		camera.AdvancedFeatures = []string{"??"}

		cameraList = append(cameraList, camera)
	}

	return cameraList
}

// json tags to match what Java sends
type Camera struct {
	ID string

	DateOfListing ravendb.Time `json:"dateOfListing"`
	Manufacturer  string       `json:"manufacturer"`
	Model         string       `json:"model"`
	Cost          float64      `json:"cost"`

	Zoom             int      `json:"zoom"`
	Megapixels       float64  `json:"megapixels"`
	ImageStabilizer  bool     `json:"imageStabilizer"`
	AdvancedFeatures []string `json:"advancedFeatures"`
}
