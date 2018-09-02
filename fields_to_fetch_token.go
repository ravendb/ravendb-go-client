package ravendb

import "strings"

var _ QueryToken = &FieldsToFetchToken{}

type FieldsToFetchToken struct {
	fieldsToFetch  []string
	projections    []string
	customFunction bool
}

func NewFieldsToFetchToken(fieldsToFetch []string, projections []string, customFunction bool) *FieldsToFetchToken {
	return &FieldsToFetchToken{
		fieldsToFetch:  fieldsToFetch,
		projections:    projections,
		customFunction: customFunction,
	}
}

func FieldsToFetchToken_create(fieldsToFetch []string, projections []string, customFunction bool) *FieldsToFetchToken {
	if len(fieldsToFetch) == 0 {
		panicIf(true, "fieldToFetch cannot be null")
		//return NewIllegalArgumentException("fieldToFetch cannot be null");
	}

	if !customFunction && len(projections) != len(fieldsToFetch) {
		panicIf(true, "Length of projections must be the same as length of field to fetch")
		// return NewIllegalArgumentException("Length of projections must be the same as length of field to fetch");
	}

	return NewFieldsToFetchToken(fieldsToFetch, projections, customFunction)
}

func (t *FieldsToFetchToken) WriteTo(writer *strings.Builder) {
	for i, fieldToFetch := range t.fieldsToFetch {

		if i > 0 {
			writer.WriteString(", ")
		}

		QueryToken_writeField(writer, fieldToFetch)

		if t.customFunction {
			continue
		}

		// Note: Java code has seemingly unnecessary checks (conditions that would
		// be rejected in FieldsToFetchToken_create)
		projection := t.projections[i]

		if projection == "" || projection == fieldToFetch {
			continue
		}

		writer.WriteString(" as ")
		writer.WriteString(projection)
	}
}
