package ravendb

func DocumentQueryHelper_addSpaceIfNeeded(previousToken QueryToken, currentToken QueryToken, writer *StringBuilder) {
	if previousToken == nil {
		return
	}

	skip := false
	if _, ok := previousToken.(*OpenSubclauseToken); ok {
		skip = true
	} else if _, ok := currentToken.(*CloseSubclauseToken); ok {
		skip = true
	} else if _, ok := currentToken.(*IntersectMarkerToken); ok {
		skip = true
	}
	if skip {
		return
	}
	writer.append(" ")
}
