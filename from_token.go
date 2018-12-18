package ravendb

import "strings"

var _ queryToken = &fromToken{}

type fromToken struct {
	collectionName string
	indexName      string
	dynamic        bool
	alias          string
}

func (t *fromToken) getCollectionName() string {
	return t.collectionName
}

func (t *fromToken) getIndexName() string {
	return t.indexName
}

func (t *fromToken) isDynamic() bool {
	return t.dynamic
}

func (t *fromToken) getAlias() string {
	return t.alias
}

func newFromToken(indexName string, collectionName string, alias string) *fromToken {
	return &fromToken{
		collectionName: collectionName,
		indexName:      indexName,
		dynamic:        collectionName != "",
		alias:          alias,
	}
}

func createFromToken(indexName string, collectionName string, alias string) *fromToken {
	return newFromToken(indexName, collectionName, alias)
}

func (t *fromToken) writeTo(writer *strings.Builder) {
	if t.indexName == "" && t.collectionName == "" {
		panicIf(true, "Either indexName or collectionName must be specified")
		// newIllegalStateError("Either indexName or collectionName must be specified");
	}

	if t.dynamic {
		writer.WriteString("from ")

		hasWhitespace := (strings.IndexAny(t.collectionName, " \t\r\n") != -1)
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
