package ravendb

import "testing"

type WithID struct {
	N  int
	B  bool
	ID string
}

type WithId struct {
	N  int
	B  bool
	Id string
}

type Withid struct {
	N  int
	B  bool
	id string
}

type NoID struct {
	N int
	B bool
}

func TestTryGetSetIDFromInstance(t *testing.T) {
	{
		exp := "hello"
		s := WithID{ID: exp}
		got, ok := tryGetIDFromInstance(s)
		if !ok {
			t.Fatalf("tryGetIDFromInstance on %#v failed", s)
		}
		if got != exp {
			t.Fatalf("got %v expected %v", got, exp)
		}
		got, ok = tryGetIDFromInstance(&s)
		if !ok {
			t.Fatalf("tryGetIDFromInstance on %#v failed", &s)
		}
		if got != exp {
			t.Fatalf("got %v expected %v", got, exp)
		}

		exp = "new"
		ok = trySetIDOnEntity(s, exp)
		// can't set on structs, only on pointer to structs
		if ok || s.ID == exp {
			t.Fatalf("trySetIDOnEntity should not succeed on %#v", s)
		}
		exp = "new2"
		ok = trySetIDOnEntity(&s, exp)
		if !ok {
			t.Fatalf("trySetIDOnEntity failed on %#v", s)
		}
		if s.ID != exp {
			t.Fatalf("trySetIDOnEntity didn't set ID field to %v on %#v", exp, s)
		}
	}

	{
		exp := "hello"
		s := WithId{Id: exp}
		got, ok := tryGetIDFromInstance(s)
		if !ok {
			t.Fatalf("tryGetIDFromInstance on %#v failed", s)
		}
		if got != exp {
			t.Fatalf("got %v expected %v", got, exp)
		}
		got, ok = tryGetIDFromInstance(&s)
		if !ok {
			t.Fatalf("tryGetIDFromInstance on %#v failed", &s)
		}
		if got != exp {
			t.Fatalf("got %v expected %v", got, exp)
		}
		exp = "new"
		ok = trySetIDOnEntity(s, exp)
		// can't set on structs, only on pointer to structs
		if ok || s.Id == exp {
			t.Fatalf("trySetIDOnEntity should not succeed on %#v", s)
		}
		exp = "new2"
		ok = trySetIDOnEntity(&s, exp)
		if !ok {
			t.Fatalf("trySetIDOnEntity failed on %#v", s)
		}
		if s.Id != exp {
			t.Fatalf("trySetIDOnEntity didn't set ID field to %v on %#v", exp, s)
		}
	}

	{
		// verify doesn't get/set unexported field
		exp := "hello"
		s := Withid{id: exp}
		got, ok := tryGetIDFromInstance(s)
		if ok || got != "" {
			t.Fatalf("got %v expected %v, ok: %v", got, exp, ok)
		}
		exp = "new"
		ok = trySetIDOnEntity(s, exp)
		if ok {
			t.Fatalf("trySetIDOnEntity should fail on %#v", s)
		}
	}

	{
		exp := "new"
		// verify doesn't get/set if there's no ID field
		s := NoID{}
		got, ok := tryGetIDFromInstance(s)
		if ok || got != "" {
			t.Fatalf("got %v expected %v, ok: %v", got, exp, ok)
		}
		ok = trySetIDOnEntity(s, exp)
		if ok {
			t.Fatalf("trySetIDOnEntity should fail on %#v", s)
		}
	}

}
