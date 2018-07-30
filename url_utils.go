package ravendb

import "net/url"

// TODO: this is more complicated in Java code. Not sure if warranted
func UrlUtils_escapeDataString(stringToEscape string) string {
	if len(stringToEscape) == 0 {
		return stringToEscape
	}
	return url.QueryEscape(stringToEscape)
	/*
		var position int
		char[] dest = escapeString(stringToEscape, 0, stringToEscape.length(), null, position, false);
		if (dest == null) {
			return stringToEscape;
		}
		return new string(dest, 0, position.value);

		panic("NYI")
		return stringToEscape
	*/
}
