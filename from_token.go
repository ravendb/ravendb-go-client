package ravendb

import "strings"

var _ queryToken = &fromToken{}

type fromToken struct {
	collectionName string
	indexName      string
	isDynamic      bool
	alias          string
}

func newFromToken(indexName string, collectionName string, alias string) *fromToken {
	//TODO: figure out why this triggers in queryQueryWithSelect
	//it's the same check as in writeTo()
	//panicIf(indexName == "" && collectionName == "", "Either indexName or collectionName must be specified")
	return &fromToken{
		collectionName: collectionName,
		indexName:      indexName,
		isDynamic:      collectionName != "",
		alias:          alias,
	}
}

func createFromToken(indexName string, collectionName string, alias string) *fromToken {
	return newFromToken(indexName, collectionName, alias)
}

func (t *fromToken) writeTo(writer *strings.Builder) {
	panicIf(t.indexName == "" && t.collectionName == "", "Either indexName or collectionName must be specified")
	// newIllegalStateError("Either indexName or collectionName must be specified");

	if t.isDynamic {
		writer.WriteString("from ")

		hasWhitespace := strings.ContainsAny(t.collectionName, " \t\r\n")
		if hasWhitespace {
			if strings.Contains(t.collectionName, "\"") {
				t.throwInvalidCollectionName()
			}
			writer.WriteString(`"`)
			writer.WriteString(t.collectionName)
			writer.WriteString(`"`)
		} else {
			writeQueryTokenField(writer, t.collectionName)
		}
	} else {
		writer.WriteString("from index '")
		writer.WriteString(t.indexName)
		writer.WriteString("'")
	}

	if t.alias != "" {
		writer.WriteString(" as ")
		writer.WriteString(t.alias)
	}
}

func (t *fromToken) throwInvalidCollectionName() {
	panicIf(true, "Collection name cannot contain a quote, but was: %s", t.collectionName)
	// newIllegalArgumentError("Collection name cannot contain a quote, but was: " + collectionName);
}
