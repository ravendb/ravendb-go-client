package ravendb

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrorWrapping(t *testing.T) {
	{
		origErr := errors.New("an error")
		err := newIllegalStateError("n: %d", 5, origErr)
		assert.Equal(t, err.Error(), "n: 5")
		assert.Equal(t, origErr, err.wrapped)
	}

	{
		err := newIllegalArgumentError("just text")
		assert.Equal(t, err.Error(), "just text")
		assert.Nil(t, err.wrapped)
	}

	{
		err := newIllegalArgumentError("%d, %s", 3, "hey")
		assert.Equal(t, err.Error(), "3, hey")
		assert.Nil(t, err.wrapped)
	}

	{
		err := newSubscriptionDoesNotExistError("")
		assert.True(t, isRavenError(err))
	}

	{
		err := makeRavenErrorFromName("IndexCompilationException", "message")
		_, ok := err.(*IndexCompilationError)
		assert.True(t, ok)
		assert.Equal(t, "message", err.Error())
	}

}
