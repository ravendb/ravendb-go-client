# Imports

Since you cant build project module-by-module with relative imports - they are absolute. But because of that you need explicitly(for now) change vendor name in imports. Line i use for this purpose is:

`find . -type f -name "*.go" -exec sed -i'' -e 's#github.com/ravendb#foo/ravendb#g' {} +`

# Running tests
test http package:

`go test -c -i github.com/ravendb-go-client/http`

`./http.go`

# Usage

#####Load a single or several document\s from the store:
```go
package main

import (
	"morty/ravendb-go-client/store"
	"fmt"
)

func main() {
	urls := []string{"http://localhost:8080"}
	store, _ := store.NewDocumentStore(urls, "test", nil)
	session, _ := store.OpenSession("", nil)
	ids := []string{"foos/1"}
	foo, _ := session.Load(ids, nil, []string{}, []interface{}{})
	for _, b := range foo{
		fmt.Printf("%v", *b)
	}
}
```

##### Delete the document
```go
_ := session.Delete("foos/1", nil, nil)
	_ = session.SaveChanges()
```