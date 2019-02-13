[![Linux build Status](https://travis-ci.org/ravendb/ravendb-go-client.svg?branch=master)](https://travis-ci.org/ravendb/ravendb-go-client) [![Windows build status](https://ci.appveyor.com/api/projects/status/rf326yoxl1uf444h/branch/master?svg=true)](https://ci.appveyor.com/project/ravendb/ravendb-go-client/branch/master)

This is information on how to use the library. For docs on working on the library itself see [readme-dev.md](readme-dev.md).

This library requires go 1.11 or later.

API reference: https://godoc.org/github.com/ravendb/ravendb-go-client

## Documentation

To learn basics of RavenDB, read [RavenDB Documentation](https://ravendb.net/docs/article-page/4.1/csharp).

## Getting started

Full source code of those examples is in `examples` directory.

1. Import the package
```go
import (
	ravendb "github.com/ravendb/ravendb-go-client"
)
```

2. Initialize document store (you should have one DocumentStore instance per application)
```go
func getDocumentStore(databaseName string) (*ravendb.DocumentStore, error) {
	serverNodes := []string{"http://live-test.ravendb.net"}
	store := ravendb.NewDocumentStore(serverNodes, databaseName)
	if err := store.Initialize(); err != nil {
		return nil, err
	}
	return store, nil
}
```

3. Open a session and close it when done
```go
session, err = store.OpenSession()
if err != nil {
	log.Fatalf("store.OpenSession() failed with %s", err)
}
// ... use session
session.Close()
```

4. Call `SaveChanges()` to persist changes in a session:
```go
var e *northwind.Employee
err = session.Load(&e, "employees/7-A")
if err != nil {
    log.Fatalf("session.Load() failed with %s\n", err)
}

origName := e.FirstName
e.FirstName = e.FirstName + "Changed"
err = session.Store(e)
if err != nil {
    log.Fatalf("session.Store() failed with %s\n", err)
}

err = session.SaveChanges()
if err != nil {
    log.Fatalf("session.SaveChanges() failed with %s\n", err)
}

var e2 *northwind.Employee
err = session.Load(&e2, "employees/7-A")
if err != nil {
    log.Fatalf("session.Load() failed with %s\n", err)
}
fmt.Printf("Updated Employee.FirstName from '%s' to '%s'\n", origName, e2.FirstName)
```
See `loadUpdateSave()` in [examples/main.go](examples/main.go) for full example.

## CRUD example

### Storing documents
```go
product := &northwind.Product{
    Name:         "iPhone X",
    PricePerUnit: 999.99,
    Category:     "electronis",
    ReorderLevel: 15,
}
err = session.Store(product)
if err != nil {
    log.Fatalf("session.Store() failed with %s\n", err)
}
```
See `crudStore()` in [examples/main.go](examples/main.go) for full example.


### Loading documents

```go
var e *northwind.Employee
err = session.Load(&e, "employees/7-A")
if err != nil {
    log.Fatalf("session.Load() failed with %s\n", err)
}
fmt.Printf("employee: %#v\n", e)
```
See `crudLoad()` in [examples/main.go](examples/main.go) for full example.

### Loading documents with includes

Some entities point to other entities via id. For example `Employee` has `ReportsTo` field which is an id of `Employee` that it reports to.

To improve performance by minimizing number of server requests, we can use includes functionality to load such linked entities.

```go
// load employee with id "employees/7-A" and entity whose id is ReportsTo
var e *northwind.Employee
err = session.Include("ReportsTo").Load(&e, "employees/5-A")
if err != nil {
    log.Fatalf("session.Load() failed with %s\n", err)
}
if e.ReportsTo == "" {
    fmt.Printf("Employee with id employees/5-A doesn't report to anyone\n")
    return
}

numRequests := session.GetNumberOfRequests()
var reportsTo *northwind.Employee
err = session.Load(&reportsTo, e.ReportsTo)
if err != nil {
    log.Fatalf("session.Load() failed with %s\n", err)
}
if numRequests != session.GetNumberOfRequests() {
    fmt.Printf("Something's wrong, this shouldn't send a request to the server\n")
} else {
    fmt.Printf("Loading e.ReportsTo employee didn't require a new request to the server because we've loaded it in original requests thanks to using Include functionality\n")
}
```
See `crudLoadWithInclude()` in [examples/main.go](examples/main.go) for full example.

### Updating documents

```go
// load entity from the server
var p *northwind.Product
err = session.Load(&p, productID)
if err != nil {
    log.Fatalf("session.Load() failed with %s\n", err)
}

// update price
origPrice = p.PricePerUnit
newPrice = origPrice + 10
p.PricePerUnit = newPrice
err = session.Store(p)
if err != nil {
    log.Fatalf("session.Store() failed with %s\n", err)
}

// persist changes on the server
err = session.SaveChanges()
if err != nil {
    log.Fatalf("session.SaveChanges() failed with %s\n", err)
}
```
See `crudUpdate()` in [examples/main.go](examples/main.go) for full example.

### Deleting documents

Deleting using entity

```go
// store a product and remember its id in productID

var p *northwind.Product
err = session.Load(&p, productID)
if err != nil {
    log.Fatalf("session.Load() failed with %s\n", err)
}

err = session.DeleteEntity(p)
if err != nil {
    log.Fatalf("session.Delete() failed with %s\n", err)
}

err = session.SaveChanges()
if err != nil {
    log.Fatalf("session.SaveChanges() failed with %s\n", err)
}

```

Deleting using id

```go
// store a product and remember its id in productID

err = session.Delete(productID)
if err != nil {
    log.Fatalf("session.Delete() failed with %s\n", err)
}

err = session.SaveChanges()
if err != nil {
    log.Fatalf("session.SaveChanges() failed with %s\n", err)
}
```
See `crudDeleteUsingID()` in [examples/main.go](examples/main.go) for full example.

## Querying documents

### Selecting what to query

First you need to decide what you you query.

RavenDB stores documents in collections. By default each type (struct) is stored in its own collection e.g. `Employee` struct is stored in collection `employees`.

You can query a collection given its name:

```go
q := session.QueryCollection("employees")
```

See `queryCollectionByName()` in [examples/main.go](examples/main.go) for full example.

You can query a collection for a given type:

```go
tp := reflect.TypeOf(&northwind.Employee{})
q := session.QueryCollectionForType(tp)
```
See `queryCollectionByType()` in [examples/main.go](examples/main.go) for full example.

You can query an index.

```go
q := session.QueryIndex("Orders/ByCompany")
```
See `queryIndex()` in [examples/main.go](examples/main.go) for full example.

### Limit what is returned

```go
tp := reflect.TypeOf(&northwind.Product{})
q := session.QueryCollectionForType(tp)

q = q.WaitForNonStaleResults(0)
q = q.WhereEquals("Name", "iPhone X")
q = q.OrderBy("PricePerUnit")
q = q.Take(2) // limit to 2 results
```
See `queryComplex()` in [examples/main.go](examples/main.go) for full example.

### Obtain the results

You can get all matching results:

```go
var products []*northwind.Product
err = q.GetResults(&products)
```
See `queryComplex()` in [examples/main.go](examples/main.go) for full example.

You can get just first one:
```go
var first *northwind.Employee
err = q.First(&first)
```
See `queryFirst()` in [examples/main.go](examples/main.go) for full example.

## Overview of [DocumentQuery](https://godoc.org/github.com/ravendb/ravendb-go-client#DocumentQuery) methods

### SelectFields() - projections using a single field

```go
// RQL equivalent: from employees select FirstName
q = q.SelectFields(reflect.TypeOf(""), "FirstName")

var names []string
err = q.GetResults(&names)
```
See `querySelectSingleField()` in [examples/main.go](examples/main.go) for full example.

### SelectFields() - projections using multiple fields

```go
type employeeNameTitle struct {
	FirstName string
	Title     string
}

// RQL equivalent: from employees select FirstName, Title
tp := reflect.TypeOf(&northwind.Employee{})
q := session.QueryCollectionForType(tp)
q = q.SelectFields(reflect.TypeOf(&employeeNameTitle{}), "FirstName", "Title")
```
See `querySelectFields()` in [examples/main.go](examples/main.go) for full example.

### Distinct()

```go
// RQL equivalent: from employees select distinct Title
tp := reflect.TypeOf(&northwind.Employee{})
q := session.QueryCollectionForType(tp)
q = q.SelectFields(reflect.TypeOf(""), "Title")
q = q.Distinct()
```
See `queryDistinct()` in [examples/main.go](examples/main.go) for full example.

### WhereEquals() / WhereNotEquals()

```go
// RQL equivalent: from employees where Title = 'Sales Representative'
tp := reflect.TypeOf(&northwind.Employee{})
q := session.QueryCollectionForType(tp)
q = q.WhereEquals("Title", "Sales Representative")
```
See `queryEquals()` in [examples/main.go](examples/main.go) for full example.

### WhereIn

```go
// RQL equivalent: from employees where Title in ['Sales Representative', 'Sales Manager']
tp := reflect.TypeOf(&northwind.Employee{})
q := session.QueryCollectionForType(tp)
q = q.WhereIn("Title", []interface{}{"Sales Representative", "Sales Manager"})
```
See `queryIn()` in [examples/main.go](examples/main.go) for full example.

### WhereStartsWith() / WhereEndsWith()

```go
// RQL equivalent:
// from employees where startsWith('Ro')
tp := reflect.TypeOf(&northwind.Employee{})
q := session.QueryCollectionForType(tp)
q = q.WhereStartsWith("FirstName", "Ro")
```
See `queryStartsWith()` and `queryEndsWith` in [examples/main.go](examples/main.go) for full example.

### WhereBetween()

```go
// RQL equivalent:
// from orders where Freight between 11 and 13
tp := reflect.TypeOf(&northwind.Order{})
q := session.QueryCollectionForType(tp)
q = q.WhereBetween("Freight", 11, 13)
```
See `queryBetween()` in [examples/main.go](examples/main.go) for full example.

### WhereGreaterThan() / WhereGreaterThanOrEqual() / WhereLessThan() / WhereLessThanOrEqual()

```go
// RQL equivalent:
// from orders where Freight Freight > 11
tp := reflect.TypeOf(&northwind.Order{})
q := session.QueryCollectionForType(tp)
// can also be WhereGreaterThanOrEqual(), WhereLessThan(), WhereLessThanOrEqual()
q = q.WhereGreaterThan("Freight", 11)
```
See `queryGreater()` in [examples/main.go](examples/main.go) for full example.

### WhereExists()

Checks if the field exists.

```go
// RQL equivalent:
// from employees where exists ("ReportsTo")
tp := reflect.TypeOf(&northwind.Employee{})
q := session.QueryCollectionForType(tp)
q = q.WhereExists("ReportsTo")
```
See `queryExists()` in [examples/main.go](examples/main.go) for full example.

### ContainsAny() / ContainsAll()

```go
// RQL equivalent:
// from employees where FirstName in ("Anne", "Nancy")
tp := reflect.TypeOf(&northwind.Employee{})
q := session.QueryCollectionForType(tp)
q = q.ContainsAny("FirstName", []interface{}{"Anne", "Nancy"})
```
See `queryContainsAny()` in [examples/main.go](examples/main.go) for full example.

### Search()

Performs full-text search.

```go
// RQL equivalent:
// from employees where search(FirstName, 'Anne Nancy')
tp := reflect.TypeOf(&northwind.Employee{})
q := session.QueryCollectionForType(tp)
q = q.Search("FirstName", "Anne Nancy")
```
See `querySearch()` in [examples/main.go](examples/main.go) for full example.

### OpenSubclause() / CloseSubclause()

```go
// RQL equivalent:
// from employees where (FirstName = 'Steven') or (Title = 'Sales Representative' and LastName = 'Davolio')
tp := reflect.TypeOf(&northwind.Employee{})
q := session.QueryCollectionForType(tp)
q = q.WhereEquals("FirstName", "Steven")
q = q.OrElse()
q = q.OpenSubclause()
q = q.WhereEquals("Title", "Sales Representative")
q = q.WhereEquals("LastName", "Davolio")
q = q.CloseSubclause()
```
See `querySubclause()` in [examples/main.go](examples/main.go) for full example.

### Not()

```go
// RQL equivalent:
// from employees where not FirstName = 'Steven'
tp := reflect.TypeOf(&northwind.Employee{})
q := session.QueryCollectionForType(tp)
q = q.Not()
q = q.WhereEquals("FirstName", "Steven")
```
See `queryNot()` in [examples/main.go](examples/main.go) for full example.

### AndAlso() / OrElse()

```go
// RQL equivalent:
// from employees where FirstName = 'Steven' or FirstName  = 'Nancy'
tp := reflect.TypeOf(&northwind.Employee{})
q := session.QueryCollectionForType(tp)
q = q.WhereEquals("FirstName", "Steven")
// can also be AndElse()
q = q.OrElse()
q = q.WhereEquals("FirstName", "Nancy")
```
See `queryOrElse()` in [examples/main.go](examples/main.go) for full example.

### UsingDefaultOperator()

Sets default operator (which will be used if no `AndAlso()` / `OrElse()` was called. Just after query instantiation, OR is used as default operator. Default operator can be changed only adding any conditions.

### OrderBy() / RandomOrdering()

```go
// RQL equivalent:
// from employees order by FirstName
tp := reflect.TypeOf(&northwind.Employee{})
q := session.QueryCollectionForType(tp)
// can also be RandomOrdering()
q = q.OrderBy("FirstName")
```
See `queryOrderBy()` in [examples/main.go](examples/main.go) for full example.

### Take()

```go
// RQL equivalent:
// from employees order by FirstName desc
tp := reflect.TypeOf(&northwind.Employee{})
q := session.QueryCollectionForType(tp)
q = q.OrderByDescending("FirstName")
q = q.Take(2)
```
See `queryTake()` in [examples/main.go](examples/main.go) for full example.

### Skip()

```go
// RQL equivalent:
// from employees order by FirstName desc
tp := reflect.TypeOf(&northwind.Employee{})
q := session.QueryCollectionForType(tp)
q = q.OrderByDescending("FirstName")
q = q.Take(2)
q = q.Skip(1)
```
See `querySkip()` in [examples/main.go](examples/main.go) for full example.

### Getting query statistics

To obtain query statistics use `Statistics()` method.

```go
var stats *ravendb.QueryStatistics
tp := reflect.TypeOf(&northwind.Employee{})
q := session.QueryCollectionForType(tp)
q = q.WhereGreaterThan("FirstName", "Bernard")
q = q.OrderByDescending("FirstName")
q.Statistics(&stats)
```
Statistics:
```
Statistics:
{IsStale:           false,
 DurationInMs:      0,
 TotalResults:      7,
 SkippedResults:    0,
 Timestamp:         2019-02-13 02:57:31.5226409 +0000 UTC,
 IndexName:         "Auto/employees/ByLastNameAndReportsToAndSearch(FirstName)AndTitle",
 IndexTimestamp:    2019-02-13 02:57:31.5226409 +0000 UTC,
 LastQueryTime:     2019-02-13 03:50:25.7602429 +0000 UTC,
 TimingsInMs:       {},
 ResultEtag:        7591488513381790088,
 ResultSize:        0,
 ScoreExplanations: {}}
 ```
See `queryStatistics()` in [examples/main.go](examples/main.go) for full example.

### GetResults() / First() / Single() / Count()

`GetResults()` - returns all results

`First()` - first result

`Single()` - first result, returns error if there's more entries

`Count()` - returns the number of the results (not affected by take())

See `queryFirst()`, `querySingle()` and `queryCount()` in [examples/main.go](examples/main.go) for full example.

## Attachments

### Store attachments

```go
fileStream, err := os.Open(path)
if err != nil {
    log.Fatalf("os.Open() failed with '%s'\n", err)
}
defer fileStream.Close()

fmt.Printf("new employee id: %s\n", e.ID)
err = session.Advanced().Attachments().StoreEntity(e, "photo.png", fileStream, "image/png")

// could also be done using document id
// err = session.Advanced().Attachments().Store(e.ID, "photo.png", fileStream, "image/png")

if err != nil {
    log.Fatalf("session.Advanced().Attachments().StoreEntity() failed with '%s'\n", err)
}

err = session.SaveChanges()
```

### Get attachments

```go
attachment, err := session.Advanced().Attachments().Get(docID, "photo.png")
if err != nil {
    log.Fatalf("session.Advanced().Attachments().Get() failed with '%s'\n", err)
}
defer attachment.Close()
fmt.Print("Attachment details:\n")
pretty.Print(attachment.Details)
// read attachment data
// attachment.Data is io.Reader
var attachmentData bytes.Buffer
n, err := io.Copy(&attachmentData, attachment.Data)
if err != nil {
    log.Fatalf("io.Copy() failed with '%s'\n", err)
}
fmt.Printf("Attachment size: %d bytes\n", n)
```

Attachment details:
```
{AttachmentName: {Name:        "photo.png",
                  Hash:        "MvUEcrFHSVDts5ZQv2bQ3r9RwtynqnyJzIbNYzu1ZXk=",
                  ContentType: "image/png",
                  Size:        4579},
 ChangeVector:   "A:4905-dMAeI9ANZ06DOxCRLnSmNw",
 DocumentID:     "employees/44-A"}
Attachment size: 4579 bytes
```

### Check if attachment exists

```go
name := "photo.png"
exists, err := session.Advanced().Attachments().Exists(docID, name)
if err != nil {
    log.Fatalf("session.Advanced().Attachments().Exists() failed with '%s'\n", err)
}
```

### Get attachment names

```go
names, err := session.Advanced().Attachments().GetNames(doc)
if err != nil {
    log.Fatalf("session.Advanced().Attachments().GetNames() failed with '%s'\n", err)
}
```

Attachment names:
```
[{Name:        "photo.png",
  Hash:        "MvUEcrFHSVDts5ZQv2bQ3r9RwtynqnyJzIbNYzu1ZXk=",
  ContentType: "image/png",
  Size:        4579}]
```
