package ravendb

func QueryFieldUtil_escapeIfNecessary(name string) string {
	if stringIsEmpty(name) ||
		IndexingFieldNameDocumentID == name ||
		IndexingFieldNameReduceKeyHash == name ||
		IndexingFieldNameReduceKeyValue == name ||
		IndexingFieldsNameSpatialShare == name {
		return name
	}

	escape := false
	insideEscaped := false

	for i, c := range name {

		if c == '\'' || c == '"' {
			insideEscaped = !insideEscaped
			continue
		}

		if i == 0 {
			if !Character_isLetter(c) && c != '_' && c != '@' && !insideEscaped {
				escape = true
				break
			}
		} else {
			if !Character_isLetterOrDigit(c) && c != '_' && c != '-' && c != '@' && c != '.' && c != '[' && c != ']' && !insideEscaped {
				escape = true
				break
			}
		}
	}

	if escape || insideEscaped {
		return "'" + name + "'"
	}

	return name
}
