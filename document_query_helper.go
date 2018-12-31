package ravendb

import "strings"

func documentQueryHelperAddSpaceIfNeeded(previousToken queryToken, currentToken queryToken, writer *strings.Builder) {
	if previousToken == nil {
		return
	}

	skip := false
	if _, ok := previousToken.(*openSubclauseToken); ok {
		skip = true
	} else if _, ok := currentToken.(*closeSubclauseToken); ok {
		skip = true
	} else if currentToken == intersectMarkerTokenInstance {
		skip = true
	}
	if skip {
		return
	}
	writer.WriteString(" ")
}
