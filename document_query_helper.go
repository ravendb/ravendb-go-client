package ravendb

import "strings"

func DocumentQueryHelper_addSpaceIfNeeded(previousToken queryToken, currentToken queryToken, writer *strings.Builder) {
	if previousToken == nil {
		return
	}

	skip := false
	if _, ok := previousToken.(*openSubclauseToken); ok {
		skip = true
	} else if _, ok := currentToken.(*closeSubclauseToken); ok {
		skip = true
	} else if _, ok := currentToken.(*intersectMarkerToken); ok {
		skip = true
	}
	if skip {
		return
	}
	writer.WriteString(" ")
}
