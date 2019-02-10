package ravendb

import "strings"

var _ queryToken = &fromToken{}

type fromToken struct {
	collectionName string
	indexName      string
	isDynamic      bool
	alias          string
}

func createFromToken(indexName string, collectionName string, alias string) *fromToken {
	return &fromToken{
		collectionName: collectionName,
		indexName:      indexName,
		isDynamic:      collectionName != "",
		alias:          alias,
	}
}

func (t *fromToken) writeTo(writer *strings.Builder) error {
	if t.indexName == "" && t.collectionName == "" {
		return newIllegalStateError("Either indexName or collectionName must be specified")
	}

	if t.isDynamic {
		writer.WriteString("from ")

		hasWhitespace := strings.ContainsAny(t.collectionName, " \t\r\n")
		if hasWhitespace {
			err := throwIfInvalidCollectionName(t.collectionName)
			if err != nil {
				return err
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
	return nil
}

func throwIfInvalidCollectionName(collectionName string) error {
	if strings.Contains(collectionName, "\"") {
		return newIllegalArgumentError("Collection name cannot contain a quote, but was: " + collectionName)
	}
	return nil
}
