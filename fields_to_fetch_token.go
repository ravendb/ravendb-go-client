package ravendb

import "strings"

var _ queryToken = &fieldsToFetchToken{}

type fieldsToFetchToken struct {
	fieldsToFetch  []string
	projections    []string
	customFunction bool
}

func newFieldsToFetchToken(fieldsToFetch []string, projections []string, customFunction bool) *fieldsToFetchToken {
	return &fieldsToFetchToken{
		fieldsToFetch:  fieldsToFetch,
		projections:    projections,
		customFunction: customFunction,
	}
}

func FieldsToFetchToken_create(fieldsToFetch []string, projections []string, customFunction bool) *fieldsToFetchToken {
	if len(fieldsToFetch) == 0 {
		panicIf(true, "fieldToFetch cannot be null")
		//return newIllegalArgumentError("fieldToFetch cannot be null");
	}

	if !customFunction && len(projections) != len(fieldsToFetch) {
		panicIf(true, "Length of projections must be the same as length of field to fetch")
		// return newIllegalArgumentError("Length of projections must be the same as length of field to fetch");
	}

	return newFieldsToFetchToken(fieldsToFetch, projections, customFunction)
}

func (t *fieldsToFetchToken) writeTo(writer *strings.Builder) {
	for i, fieldToFetch := range t.fieldsToFetch {

		if i > 0 {
			writer.WriteString(", ")
		}

		writeQueryTokenField(writer, fieldToFetch)

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
