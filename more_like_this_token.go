package ravendb

import "strings"

var _ queryToken = &moreLikeThisToken{}

type moreLikeThisToken struct {
	documentParameterName string
	optionsParameterName  string
	whereTokens           []queryToken
}

func newMoreLikeThisToken() *moreLikeThisToken {
	return &moreLikeThisToken{}
}

func (t *moreLikeThisToken) writeTo(writer *strings.Builder) error {
	writer.WriteString("moreLikeThis(")

	if t.documentParameterName == "" {
		var prevToken queryToken
		for _, whereToken := range t.whereTokens {
			documentQueryHelperAddSpaceIfNeeded(prevToken, whereToken, writer)
			whereToken.writeTo(writer)
			prevToken = whereToken
		}
	} else {
		writer.WriteString("$")
		writer.WriteString(t.documentParameterName)
	}

	if t.optionsParameterName == "" {
		writer.WriteString(")")
		return nil
	}

	writer.WriteString(", $")
	writer.WriteString(t.optionsParameterName)
	writer.WriteString(")")

	return nil
}
