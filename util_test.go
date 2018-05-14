package ravendb

import "testing"

type FooStruct struct {
	s string
}

func TestTypeName(t *testing.T) {
	v := FooStruct{}
	name := getTypeName(v)
	if name != "ravendb.FooStruct" {
		t.Fatalf("expected '%s', got '%s'", "ravendb.FooStruct", name)
	}
	name = getTypeName(&v)
	if name != "ravendb.FooStruct" {
		t.Fatalf("expected '%s', got '%s'", "ravendb.FooStruct", name)
	}
	name = getShortTypeName(v)
	if name != "FooStruct" {
		t.Fatalf("expected '%s', got '%s'", "FooStruct", name)
	}
	name = getShortTypeName(&v)
	if name != "FooStruct" {
		t.Fatalf("expected '%s', got '%s'", "FooStruct", name)
	}
}
