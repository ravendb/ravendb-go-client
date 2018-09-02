package ravendb

import "strings"

var _ QueryToken = &MoreLikeThisToken{}

type MoreLikeThisToken struct {
	documentParameterName string
	optionsParameterName  string
	whereTokens           []QueryToken
}

func NewMoreLikeThisToken() *MoreLikeThisToken {
	return &MoreLikeThisToken{}
}

func (t *MoreLikeThisToken) WriteTo(writer *strings.Builder) {
	writer.WriteString("moreLikeThis(")

	if t.documentParameterName == "" {
		var prevToken QueryToken
		for _, whereToken := range t.whereTokens {
			DocumentQueryHelper_addSpaceIfNeeded(prevToken, whereToken, writer)
			whereToken.WriteTo(writer)
			prevToken = whereToken
		}
	} else {
		writer.WriteString("$")
		writer.WriteString(t.documentParameterName)
	}

	if t.optionsParameterName == "" {
		writer.WriteString(")")
		return
	}

	writer.WriteString(", $")
	writer.WriteString(t.optionsParameterName)
	writer.WriteString(")")

}
