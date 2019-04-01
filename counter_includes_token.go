package ravendb

import "strings"

var _ queryToken = &counterIncludesToken{}

type counterIncludesToken struct {
	_sourcePath    string
	_parameterName string
	_all           bool
}

func (t *counterIncludesToken) writeTo(writer *strings.Builder) error {
	writer.WriteString("counters(")
	if stringIsNotEmpty(t._sourcePath) {
		writer.WriteString(t._sourcePath)
		if !t._all {
			writer.WriteString(", ")
		}
		if !t._all {
			writer.WriteString("$")
			writer.WriteString(t._parameterName)
		}
	}
	writer.WriteString(")")
	return nil
}

func (t *counterIncludesToken) addAliasToPath(alias string) {
	if stringIsEmpty(t._sourcePath) {
		t._sourcePath = alias
	} else {
		t._sourcePath = alias + "." + t._sourcePath
	}
}

func newCounterIncludesToken(sourcePath string, parameterName string, all bool) *counterIncludesToken {
	return &counterIncludesToken{
		_sourcePath:    sourcePath,
		_parameterName: parameterName,
		_all:           all,
	}
}

/*
   public static CounterIncludesToken create(String sourcePath, String parameterName) {
       return new CounterIncludesToken(sourcePath, parameterName, false);
   }

   public static CounterIncludesToken all(String sourcePath) {
       return new CounterIncludesToken(sourcePath, null, true);
   }
*/
