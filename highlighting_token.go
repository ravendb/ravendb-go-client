package ravendb

import (
	"strconv"
	"strings"
)

var _ queryToken = &highlightingToken{}

type highlightingToken struct {
	_fieldName            string
	_fragmentLength       int
	_fragmentCount        int
	_optionsParameterName string
}

func (t *highlightingToken) writeTo(writer *strings.Builder) error {
	writer.WriteString("highlight(")
	writer.WriteString(t._fieldName)

	writer.WriteString(",")
	writer.WriteString(strconv.Itoa(t._fragmentLength))
	writer.WriteString(",")
	writer.WriteString(strconv.Itoa(t._fragmentCount))

	if t._optionsParameterName != "" {
		writer.WriteString(",$")
		writer.WriteString(t._optionsParameterName)
	}

	writer.WriteString(")")
	return nil
}

/*
   private HighlightingToken(String fieldName, int fragmentLength, int fragmentCount, String operationsParameterName) {
       _fieldName = fieldName;
       _fragmentLength = fragmentLength;
       _fragmentCount = fragmentCount;
       _optionsParameterName = operationsParameterName;
   }

   public static HighlightingToken create(String fieldName, int fragmentLength, int fragmentCount, String optionsParameterName) {
       return new HighlightingToken(fieldName, fragmentLength, fragmentCount, optionsParameterName);
   }
*/
