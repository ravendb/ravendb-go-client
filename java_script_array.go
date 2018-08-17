package ravendb

import (
	"fmt"
	"strings"
)

// JavaScriptArray builds arguments for patch operation on array fields
type JavaScriptArray struct {
	suffix     int
	argCounter int

	pathToArray string

	scriptLines []string
	Parameters  map[string]Object
}

// NewJavaScriptArray creates a new JavaScriptArray
func NewJavaScriptArray(suffix int, pathToArray string) *JavaScriptArray {
	return &JavaScriptArray{
		suffix:      suffix,
		pathToArray: pathToArray,
		Parameters:  map[string]Object{},
	}
}

// Add builds expression that adds an elements to array
func (a *JavaScriptArray) Add(args ...interface{}) *JavaScriptArray {
	// TODO: more efficient if len(args) > 1
	for _, u := range args {
		argumentName := a.getNextArgumentName()

		s := "this." + a.pathToArray + ".push(args." + argumentName + ")"
		a.scriptLines = append(a.scriptLines, s)
		a.Parameters[argumentName] = u
	}

	return a
}

// RemoveAt builds expression that removes an element at index from array
func (a *JavaScriptArray) RemoveAt(index int) *JavaScriptArray {
	argumentName := a.getNextArgumentName()

	s := "this." + a.pathToArray + ".splice(args." + argumentName + ", 1)"
	a.scriptLines = append(a.scriptLines, s)
	a.Parameters[argumentName] = index

	return a
}

func (a *JavaScriptArray) getNextArgumentName() string {
	s := fmt.Sprintf("val_%d_%d", a.argCounter, a.suffix)
	a.argCounter++
	return s
}

func (a *JavaScriptArray) getScript() string {
	return strings.Join(a.scriptLines, "\r")
}
