package ravendb

func QueryFieldUtil_escapeIfNecessary(name string) string {
	if stringIsEmpty(name) ||
		Constants_Documents_Indexing_Fields_DOCUMENT_ID_FIELD_NAME == name ||
		Constants_Documents_Indexing_Fields_REDUCE_KEY_HASH_FIELD_NAME == name ||
		Constants_Documents_Indexing_Fields_REDUCE_KEY_KEY_VALUE_FIELD_NAME == name ||
		Constants_Documents_Indexing_Fields_SPATIAL_SHAPE_FIELD_NAME == name {
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
