package main

// Note: this file is just to make sure that the code for the book examples
// compile. This code is not supposed to be run

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/kylelemons/godebug/pretty"
	"github.com/ravendb/ravendb-go-client"
	"github.com/ravendb/ravendb-go-client/examples/northwind"
)

var (
	// change to false to run the examples against a
	// local server running on  port 8080
	usePublicTestServer = false

	// if running RavenDB locally on port 8080
	serverLocalURL = "http://localhost:8080"

	// if using public RavenDB test instance
	serverPublicTestURL = "http://live-test.ravendb.net"

	testDatabaseName string
)

var (
	serverURL    = "http://localhost:8080"
	databaseName = "YourDatabaseName"
)

var (
	globalDocumentStore *ravendb.DocumentStore
)

func createDocumentStore() (*ravendb.DocumentStore, error) {
	if globalDocumentStore != nil {
		return globalDocumentStore, nil
	}
	urls := []string{serverURL}
	store := ravendb.NewDocumentStore(urls, databaseName)
	err := store.Initialize()
	if err != nil {
		return nil, err
	}
	globalDocumentStore = store
	return globalDocumentStore, nil
}

func genUID() string {
	var u [16]byte
	io.ReadFull(rand.Reader, u[:])
	return hex.EncodeToString(u[:])
}

func genRandomDatabaseName() string {
	return "demo-" + genUID()
}

func waitForIndexing(store *ravendb.DocumentStore, database string, timeout time.Duration) error {
	admin := store.Maintenance().ForDatabase(database)
	if timeout == 0 {
		timeout = time.Minute
	}

	sp := time.Now()
	for time.Since(sp) < timeout {
		op := ravendb.NewGetStatisticsOperation("")
		err := admin.Send(op)
		if err != nil {
			return err
		}
		databaseStatistics := op.Command.Result
		isDone := true
		hasError := false
		for _, index := range databaseStatistics.Indexes {
			if index.State == ravendb.IndexStateDisabled {
				continue
			}
			if index.IsStale || strings.HasPrefix(index.Name, ravendb.IndexingSideBySideIndexNamePrefix) {
				isDone = false
			}
			if index.State == ravendb.IndexStateError {
				hasError = true
			}
		}
		if isDone {
			return nil
		}
		if hasError {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	op := ravendb.NewGetIndexErrorsOperation(nil)
	err := admin.Send(op)
	if err != nil {
		return err
	}
	return ravendb.NewTimeoutError("The indexes stayed stale for more than %s", timeout)
}

func createSampleNorthwindDatabase(store *ravendb.DocumentStore) error {
	sampleData := ravendb.NewCreateSampleDataOperation()
	err := store.Maintenance().Send(sampleData)
	if err != nil {
		fmt.Printf("createSampleNorthwindDatabase: store.Maintance().Send() failed with '%s'\n", err)
		return err
	}
	err = waitForIndexing(store, store.GetDatabase(), 0)
	if err != nil {
		fmt.Printf("watiForIndexing() failed with '%s'\n", err)
		return err
	}
	return nil
}

// create a new, randomly named database for just tests and populate with
// sample Northwind data.
// if usePublicTestServer we'll use public RavenDB instance. Otherwise
// we'll talk to local server on port 8080
func createTestDocumentStore() (*ravendb.DocumentStore, error) {
	if globalDocumentStore != nil {
		return globalDocumentStore, nil
	}
	urls := []string{serverLocalURL}
	if usePublicTestServer {
		urls[0] = serverPublicTestURL
	}

	testDatabaseName = genRandomDatabaseName()
	databaseName = testDatabaseName

	// "test.manager" is a dummy database
	// we need a store, even if it points to a dummy database,
	// to create a new database and then create a store out of thtat
	storeManager := ravendb.NewDocumentStore(urls, "test.manager")

	err := storeManager.Initialize()
	if err != nil {
		fmt.Printf("createTestDocumentStore: storeManager.Initialize() failed with '%s'\n", err)
		return nil, err
	}

	databaseRecord := ravendb.NewDatabaseRecord()
	databaseRecord.DatabaseName = testDatabaseName

	// replicationFactor seems to be a minimum number of nodes with the data
	// so it must be less than 3 (we have 3 nodes and might kill one, leaving
	// only 2)
	createDatabaseOperation := ravendb.NewCreateDatabaseOperation(databaseRecord, 1)
	err = storeManager.Maintenance().Server().Send(createDatabaseOperation)
	if err != nil {
		fmt.Printf("d.store.Maintenance().Server().Send(createDatabaseOperation) failed with %s\n", err)
		return nil, err
	}

	store := ravendb.NewDocumentStore(urls, testDatabaseName)
	err = store.Initialize()
	if err != nil {
		fmt.Printf("createTestDocumentStore: store.Initialize() failed with '%s'\n", err)
		return nil, err
	}

	fmt.Printf("Created a test database '%s'\n", testDatabaseName)
	globalDocumentStore = store
	err = createSampleNorthwindDatabase(store)
	if err != nil {
		fmt.Printf("createTestDocumentStore: createSampleNorthwindDatabase() failed with '%s'\n", err)
		return nil, err
	}

	return globalDocumentStore, nil
}

func deleteTestDatabase() error {
	if globalDocumentStore == nil || testDatabaseName == "" {
		return nil
	}
	fmt.Printf("Deleting test database '%s'\n", testDatabaseName)
	op := ravendb.NewDeleteDatabasesOperation(testDatabaseName, true)
	return globalDocumentStore.Maintenance().Server().Send(op)
}

func createDocument(companyName, companyPhone, contactName, contactTitle string) error {
	newCompany := &northwind.Company{
		Name:  companyName,
		Phone: companyPhone,
		Contact: &northwind.Contact{
			Name:  contactName,
			Title: contactTitle,
		},
	}

	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	err = session.Store(newCompany)
	if err != nil {
		return err
	}

	theNewDocumentID := newCompany.ID

	err = session.SaveChanges()
	if err != nil {
		return err
	}

	fmt.Printf("Document id: %s\n", theNewDocumentID)
	return nil
}

func sessionChapter() error {
	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	//   Run your business logic:
	//
	//   Store documents
	//   Load and Modify documents
	//   Query indexes & collections
	//   Delete documents
	//   .... etc.

	err = session.SaveChanges()
	if err != nil {
		return err
	}

	// make sure to close the sessoin
	session.Close()

	return nil
}

func editDocumentChapter(companyName string) error {
	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	var company *northwind.Company
	err = session.Load(&company, "companies/5-A")
	if err != nil {
		return err
	}
	// if not found, company is not changed and will remain nil
	if company == nil {
		return nil
	}
	company.Name = companyName

	err = session.SaveChanges()
	if err != nil {
		return err
	}

	return nil
}

func deleteDocumentChapter(documentID string) error {
	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	err = session.DeleteByID(documentID, "")
	if err != nil {
		return err
	}

	err = session.SaveChanges()
	if err != nil {
		return err
	}

	return nil
}

func createRelatedDocuments(productName, supplierName, supplierPhone string) error {
	supplier := &northwind.Supplier{
		Name:  supplierName,
		Phone: supplierPhone,
	}

	category := &northwind.Category{
		Name:        "NoSQL Databases",
		Description: "Non-relational databases",
	}

	product := &northwind.Product{
		Name: productName,
	}

	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	err = session.Store(supplier)
	if err != nil {
		return err
	}
	err = session.Store(category)
	if err != nil {
		return err
	}

	product.Supplier = supplier.ID
	product.Category = category.ID

	err = session.Store(product)
	if err != nil {
		return err
	}

	err = session.SaveChanges()
	if err != nil {
		return err
	}

	return nil
}

func loadRelatedDocuments(pricePerUnit float64, phone string) error {
	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	var product *northwind.Product
	err = session.Include("supplier").Load(&product, "products/34-A")
	if err != nil {
		return err
	}
	if product == nil {
		// not found
		return nil
	}

	var supplier *northwind.Supplier
	err = session.Load(&supplier, product.Supplier)
	if err != nil || supplier == nil {
		return err
	}

	product.PricePerUnit = pricePerUnit
	supplier.Phone = phone

	err = session.SaveChanges()
	if err != nil {
		return err
	}

	return nil
}

func queryRelatedDocuments() error {
	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	session.Advanced().SetMaxNumberOfRequestsPerSession(128)

	tp := reflect.TypeOf(&northwind.Order{})
	q := session.QueryCollectionForType(tp)
	q = q.Include("Lines.Product")
	q = q.WhereNotEquals("ShippedAt", nil)
	var shippedOrders []*northwind.Order
	err = q.GetResults(&shippedOrders)
	if err != nil {
		return err
	}

	fmt.Printf("got %d shipped orders\n", len(shippedOrders))

	for _, shippedOrder := range shippedOrders {
		var productIDs []string
		for _, line := range shippedOrder.Lines {
			productIDs = append(productIDs, line.Product)
		}

		for i, productID := range productIDs {
			var product *northwind.Product
			err = session.Load(&product, productID)
			if err != nil {
				return err
			}
			product.UnitsOnOrder += shippedOrder.Lines[i].Quantity
		}
	}

	err = session.SaveChanges()
	if err != nil {
		return err
	}

	return nil
}

func indexRelatedDocuments(categoryName string) error {
	index := ravendb.NewIndexCreationTask("Products/ByCategoryName")
	index.Map = `docs.Products.Select(product => new {
		CategoryName = (this.LoadDocument(product.Category, "Categories")).Name
	})
`
	err := globalDocumentStore.ExecuteIndex(index, "")
	if err != nil {
		return err
	}
	err = waitForIndexing(globalDocumentStore, "", 0)
	if err != nil {
		return err
	}

	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	var productsWithCategoryName []*northwind.Product
	q := session.QueryIndex(index.IndexName)
	q = q.WhereEquals("CategoryName", categoryName)
	err = q.GetResults(&productsWithCategoryName)
	if err != nil {
		return err
	}
	pretty.Print(productsWithCategoryName)
	return nil
}

func storeAttachement(documentID string, attachmentPath string) error {
	stream, err := os.Open(attachmentPath)
	if err != nil {
		return err
	}
	defer stream.Close()

	contentType := mime.TypeByExtension(filepath.Ext(attachmentPath))
	attachmentName := filepath.Base(attachmentPath)

	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	err = session.Advanced().Attachments().Store(documentID, attachmentName, stream, contentType)
	if err != nil {
		return err
	}

	err = session.SaveChanges()
	if err != nil {
		return err
	}

	return nil
}

func enableRevisions(collection1, collection2 string) error {
	dur := ravendb.Duration(time.Hour * 24 * 14)
	defaultConfig := &ravendb.RevisionsCollectionConfiguration{
		Disabled:                 false,
		PurgeOnDelete:            false,
		MinimumRevisionsToKeep:   5,
		MinimumRevisionAgeToKeep: &dur,
	}
	revisiionConfiguration1 := &ravendb.RevisionsCollectionConfiguration{
		Disabled: true,
	}
	revisiionConfiguration2 := &ravendb.RevisionsCollectionConfiguration{
		PurgeOnDelete: true,
	}
	collections := map[string]*ravendb.RevisionsCollectionConfiguration{
		collection1: revisiionConfiguration1,
		collection2: revisiionConfiguration2,
	}

	myRevisionsConfiguration := &ravendb.RevisionsConfiguration{
		DefaultConfig: defaultConfig,
		Collections:   collections,
	}

	revisionsConfigurationOperation := ravendb.NewConfigureRevisionsOperation(myRevisionsConfiguration)
	return globalDocumentStore.Maintenance().Send(revisionsConfigurationOperation)
}

func getRevisions() error {
	myRevisionsConfiguration := &ravendb.RevisionsConfiguration{
		DefaultConfig: &ravendb.RevisionsCollectionConfiguration{
			Disabled: false,
		},
	}

	revisionsConfigurationOperation := ravendb.NewConfigureRevisionsOperation(myRevisionsConfiguration)
	err := globalDocumentStore.Maintenance().Send(revisionsConfigurationOperation)
	if err != nil {
		return nil
	}

	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	var company *northwind.Company
	err = session.Load(&company, "companies/7-A")
	if err != nil {
		return err
	}
	company.Name = "Name 1"
	err = session.SaveChanges()
	if err != nil {
		return err
	}

	company.Name = "Name 2"
	company.Phone = "052-1234-567"
	err = session.SaveChanges()
	if err != nil {
		return err
	}

	var revisions []*northwind.Company
	err = session.Advanced().Revisions().GetFor(&revisions, "companies/7-A")
	if err != nil {
		return err
	}
	pretty.Print(revisions)

	return nil
}

func getRevisionsTest() {
	err := getRevisions()
	if err != nil {
		fmt.Printf("getRevisionsTest() failed with '%s'\n", err)
	}
}

func indexRelatedDocumentsTest() {
	err := indexRelatedDocuments("Produce")
	if err != nil {
		fmt.Printf("indexRelatedDocuments() failed with '%s'\n", err)
	}
}

func queryRelatedDocumentsTest() {
	err := queryRelatedDocuments()
	if err != nil {
		fmt.Printf("queryRelatedDocuments() failed with '%s'\n", err)
	}
}

var (
	testFunctions = map[string]func(){
		"indexRelatedDocuments": indexRelatedDocumentsTest,
		"queryRelatedDocuments": queryRelatedDocumentsTest,
		"getRevisions":          getRevisionsTest,
	}
)

func usageAndExit() {
	fmt.Print(`To run:
go run dive-into-raven/main.go <testName>
e.g.
go run dive-into-raven/main.go indexRelatedDocuments
	`)
	os.Exit(1)
}

func must(err error, format string, args ...interface{}) {
	if err != nil {
		fmt.Printf(format, args...)
		panic(err)
	}
}

func main() {
	if len(os.Args) != 2 {
		usageAndExit()
	}
	testNameArg := os.Args[1]
	testName := strings.TrimSuffix(testNameArg, "Test")
	testFn, ok := testFunctions[testName]
	if !ok {
		fmt.Printf("'%s' is not a known test function\n", testNameArg)
		usageAndExit()
	}

	_, err := createTestDocumentStore()
	defer deleteTestDatabase()
	must(err, "createTestDocumentStore() failed with %s\n", err)
	fmt.Printf("Running %s\n", testNameArg)
	testFn()
}
