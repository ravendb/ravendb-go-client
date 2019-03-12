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

	queriedType := reflect.TypeOf(&northwind.Order{})
	q := session.QueryCollectionForType(queriedType)
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

func queryOverview() error {
	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	queriedType := reflect.TypeOf(&northwind.Employee{})
	queryDefinition := session.QueryCollectionForType(queriedType)

	// Define actions such as:
	// Filter documents by documents fields
	// Filter documents by text criteria
	// Include related documents
	// Get the query stats
	// Sort results
	// Customise the returned entity fields (Projections)
	// Control results paging

	var queryResults []*northwind.Employee
	err = queryDefinition.GetResults(&queryResults)
	if err != nil {
		return err
	}
	fmt.Printf("Got %d results\n", len(queryResults))
	return nil
}

// EmployeeDetails describes details of an employee
type EmployeeDetails struct {
	FullName  string       `json:"FullName"`
	FirstName string       `json:"FirstName"`
	Title     string       `json:"Title"`
	HiredAt   ravendb.Time `json:"HiredAt"`
}

func queryExample() error {
	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	var queryResults []*EmployeeDetails

	queriedType := reflect.TypeOf(&northwind.Employee{})
	query := session.QueryCollectionForType(queriedType)
	{
		query = query.OpenSubclause()
		query = query.WhereEquals("FirstName", "Steven")
		query = query.OrElse()
		query = query.WhereEquals("Title", "Sales Representative")
		query = query.CloseSubclause()
	}
	query = query.Include("ReportsTo")

	var statistics *ravendb.QueryStatistics
	query = query.Statistics(&statistics)

	query = query.OrderByDescending("HiredAt")

	projectedType := reflect.TypeOf(&EmployeeDetails{})
	fields := []string{
		"FirstName",
		"Title",
		"HiredAt",
	}
	query = query.SelectFields(projectedType, fields...)
	query = query.Take(5)
	err = query.GetResults(&queryResults)
	if err != nil {
		return err
	}
	fmt.Printf("Got %d results\n", len(queryResults))
	if len(queryResults) > 0 {
		pretty.Print(queryResults[0])
	}
	return nil
}

func fullCollectionQuery() error {
	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	queriedType := reflect.TypeOf(&northwind.Company{})
	fullCollectionQuery := session.QueryCollectionForType(queriedType)

	var queryResults []*northwind.Company
	err = fullCollectionQuery.GetResults(&queryResults)
	if err != nil {
		return err
	}
	fmt.Printf("Got %d results\n", len(queryResults))

	return nil
}

func queryByDocumentID(employeeDocumentID string) error {
	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	queriedType := reflect.TypeOf(&northwind.Employee{})
	queryByDocumentID := session.QueryCollectionForType(queriedType)
	queryByDocumentID = queryByDocumentID.Where("ID", "==", employeeDocumentID)

	var employee *northwind.Employee
	err = queryByDocumentID.Single(&employee)
	if err != nil {
		return err
	}
	pretty.Print(employee)

	return nil
}

func queryFilterResultsBasic() error {
	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	queriedType := reflect.TypeOf(&northwind.Employee{})
	filteredQuery := session.QueryCollectionForType(queriedType)
	filteredQuery = filteredQuery.Where("FirstName", "==", "Anne")

	var filteredEmployees []*northwind.Employee
	err = filteredQuery.GetResults(&filteredEmployees)
	if err != nil {
		return err
	}
	fmt.Printf("Got %d results\n", len(filteredEmployees))
	if len(filteredEmployees) > 0 {
		pretty.Print(filteredEmployees[0])

	}

	return nil
}

func queryFilterResultsMultipleConditions(country string) error {
	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	queriedType := reflect.TypeOf(&northwind.Employee{})
	filteredQuery := session.QueryCollectionForType(queriedType)
	filteredQuery = filteredQuery.Where("FirstName", "==", "Anne")
	filteredQuery = filteredQuery.OrElse()
	{
		filteredQuery = filteredQuery.OpenSubclause()
		filteredQuery = filteredQuery.WhereEquals("Address.Country", country)
		filteredQuery = filteredQuery.Where("Territories.Count", ">", 2)
		filteredQuery = filteredQuery.WhereStartsWith("Title", "Sales")
		filteredQuery = filteredQuery.CloseSubclause()
	}

	var filteredEmployees []*northwind.Employee
	err = filteredQuery.GetResults(&filteredEmployees)
	if err != nil {
		return err
	}
	fmt.Printf("Got %d results\n", len(filteredEmployees))
	if len(filteredEmployees) > 0 {
		pretty.Print(filteredEmployees[0])
	}

	return nil
}

// CompanyDetails describes details about a company
type CompanyDetails struct {
	CompanyName string `json:"CompanyName"`
	City        string `json:"City"`
	Country     string `json:"Country"`
}

func queryProjectingIndividualFields() error {
	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	queriedType := reflect.TypeOf(&northwind.Company{})
	projectedQuery := session.QueryCollectionForType(queriedType)
	projectedType := reflect.TypeOf(&CompanyDetails{})
	fields := []string{"Name", "Address.City", "Address.Country"}
	projections := []string{"CompanyName", "City", "Country"}
	queryData := &ravendb.QueryData{
		Fields:      fields,
		Projections: projections,
	}
	projectedQuery = projectedQuery.SelectFieldsWithQueryData(projectedType, queryData)
	var projectedResults []*CompanyDetails
	err = projectedQuery.GetResults(&projectedResults)
	if err != nil {
		return err
	}
	fmt.Printf("Got %d results\n", len(projectedResults))
	if len(projectedResults) > 0 {
		pretty.Print(projectedResults[0])
	}
	return nil
}

func queryProjectingUsingFunctions() error {
	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	rawQuery := `declare function output(e) {
	    var format = function(p){ return p.FirstName + " " + p.LastName; };
	    return { FullName : format(e), Title: e.Title, HiredAt: e.HiredAt };
	}
	from Employees as e select output(e)
	`
	query := session.Advanced().RawQuery(rawQuery)
	var projectedResults []*EmployeeDetails
	err = query.GetResults(&projectedResults)
	if err != nil {
		return err
	}
	fmt.Printf("Got %d results\n", len(projectedResults))
	if len(projectedResults) > 0 {
		pretty.Print(projectedResults[0])
	}
	return nil
}

func staticIndexesOverview() error {
	indexName := "Employees/ByLastName"
	index := ravendb.NewIndexCreationTask(indexName)
	// Define:
	//    Map(s) functions
	//    Reduce function
	//    Additional indexing options per field
	index.Map = "from e in docs.Employees select new { e.LastName }"

	err := index.Execute(globalDocumentStore, nil, "")
	if err != nil {
		return err
	}

	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	queryOnIndex := session.QueryIndex(indexName)
	queryOnIndex = queryOnIndex.Where("LastName", "==", "SomeName")
	var queryResults []*northwind.Employee
	err = queryOnIndex.GetResults(&queryResults)
	if err != nil {
		return err
	}
	fmt.Printf("Got %d results\n", len(queryResults))
	if len(queryResults) > 0 {
		pretty.Print(queryResults[0])
	}
	return nil
}

func staticIndexesOverviewTest() {
	err := staticIndexesOverview()
	if err != nil {
		fmt.Printf("staticIndexesOverview() failed with '%s'\n", err)
	}
}

func queryProjectingUsingFunctionsTest() {
	err := queryProjectingUsingFunctions()
	if err != nil {
		fmt.Printf("queryProjectingUsingFunctions() failed with '%s'\n", err)
	}
}

func queryProjectingIndividualFieldsTest() {
	err := queryProjectingIndividualFields()
	if err != nil {
		fmt.Printf("queryProjectingIndividualFields() failed with '%s'\n", err)
	}
}

func queryFilterResultsMultipleConditionsTest() {
	err := queryFilterResultsMultipleConditions("USA")
	if err != nil {
		fmt.Printf("queryFilterResultsMultipleConditions() failed with '%s'\n", err)
	}
}

func queryFilterResultsBasicTest() {
	err := queryFilterResultsBasic()
	if err != nil {
		fmt.Printf("queryFilterResultsBasic() failed with '%s'\n", err)
	}
}

func queryByDocumentIDTest() {
	err := queryByDocumentID("employees/1-A")
	if err != nil {
		fmt.Printf("queryByDocumentID() failed with '%s'\n", err)
	}
}

func fullCollectionQueryTest() {
	err := fullCollectionQuery()
	if err != nil {
		fmt.Printf("fullCollectionQuery() failed with '%s'\n", err)
	}
}

func queryExampleTest() {
	err := queryExample()
	if err != nil {
		fmt.Printf("queryExample() failed with '%s'\n", err)
	}
}

func queryOverviewTest() {
	err := queryOverview()
	if err != nil {
		fmt.Printf("queryOverview() failed with '%s'\n", err)
	}
}

func getRevisionsTest() {
	err := getRevisions()
	if err != nil {
		fmt.Printf("getRevisions() failed with '%s'\n", err)
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
		"indexRelatedDocuments":                indexRelatedDocumentsTest,
		"queryRelatedDocuments":                queryRelatedDocumentsTest,
		"getRevisions":                         getRevisionsTest,
		"queryOverview":                        queryOverviewTest,
		"queryExample":                         queryExampleTest,
		"fullCollectionQuery":                  fullCollectionQueryTest,
		"queryByDocumentID":                    queryByDocumentIDTest,
		"queryFilterResultsBasic":              queryFilterResultsBasicTest,
		"queryFilterResultsMultipleConditions": queryFilterResultsMultipleConditionsTest,
		"queryProjectingIndividualFields":      queryProjectingIndividualFieldsTest,
		"queryProjectingUsingFunctions":        queryProjectingUsingFunctionsTest,
		"staticIndexesOverview":                staticIndexesOverviewTest,
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
