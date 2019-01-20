package ravendb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

type WithIntID struct {
	N  int
	B  bool
	ID int
}

func TestTryGetSetIDFromInstance(t *testing.T) {
	{
		// verify can get/set field name ID of type string
		exp := "hello"
		s := WithID{ID: exp}
		got, ok := TryGetIDFromInstance(s)
		assert.True(t, ok)
		assert.Equal(t, exp, got)

		got, ok = TryGetIDFromInstance(&s)
		assert.True(t, ok)
		assert.Equal(t, exp, got)

		exp = "new"
		ok = TrySetIDOnEntity(s, exp)

		// can't set on structs, only on pointer to structs
		assert.False(t, ok)
		assert.NotEqual(t, exp, s.ID)

		ok = TrySetIDOnEntity(&s, exp)
		assert.True(t, ok)
		assert.Equal(t, exp, s.ID)
	}

	{
		// id that is empty string is not valid
		s := WithID{}
		got, ok := TryGetIDFromInstance(s)
		assert.False(t, ok)
		assert.Equal(t, "", got)

		exp := "new"
		ok = TrySetIDOnEntity(&s, exp)
		assert.True(t, ok)
		assert.Equal(t, exp, s.ID)
	}

	{
		// "Id" is not valid name for id field, must be "ID"
		exp := "hello"
		s := WithId{Id: exp}
		got, ok := TryGetIDFromInstance(s)
		assert.False(t, ok)
		assert.Equal(t, "", got)

		got, ok = TryGetIDFromInstance(&s)
		assert.False(t, ok)
		assert.Equal(t, "", got)

		exp = "new"
		ok = TrySetIDOnEntity(s, exp)
		// can't set on structs, only on pointer to structs
		assert.False(t, ok)
		assert.NotEqual(t, exp, s.Id)

		ok = TrySetIDOnEntity(&s, exp)
		assert.False(t, ok)
	}

	{
		// verify doesn't get/set unexported field
		exp := "hello"
		s := Withid{id: exp}
		got, ok := TryGetIDFromInstance(s)
		assert.False(t, ok)
		assert.Equal(t, "", got)

		exp = "new"
		ok = TrySetIDOnEntity(s, exp)
		assert.False(t, ok)
		assert.Equal(t, "hello", s.id)
	}

	{
		// verify doesn't get/set if there's no ID field
		exp := "new"
		s := NoID{}
		got, ok := TryGetIDFromInstance(s)
		assert.False(t, ok)
		assert.Equal(t, "", got)
		ok = TrySetIDOnEntity(s, exp)
		assert.False(t, ok)
	}

	{
		// verify doesn't get/set if ID is not string
		exp := "new"
		s := WithIntID{ID: 5}
		got, ok := TryGetIDFromInstance(s)
		assert.False(t, ok)
		assert.Equal(t, "", got)

		ok = TrySetIDOnEntity(s, exp)
		assert.False(t, ok)
	}
}
