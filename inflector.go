package ravendb

import (
	"regexp"
	"strings"
)

// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/util/Inflector.java#L14
// https://github.com/ixmatus/inflector
// https://github.com/tangzero/inflector : Go version but could be faster / simpler
// https://github.com/gedex/inflector : better Go version
// https://github.com/c9s/inflect : another one in Go

type rule struct {
	pattern     string
	replacement string
	rx          *regexp.Regexp
}

func mkRule(pattern, replacement string) rule {
	return rule{
		pattern:     pattern,
		replacement: replacement,
	}
}

var (
	plurals = []rule{
		mkRule("(.*)$", "$1s"),
		mkRule("(.*)s$", "$1s"),
		mkRule("(ax|test)is$", "$1es"),
		mkRule("(octop|vir)us$", "$1i"),
		mkRule("(alias|status)$", "$1es"),
		mkRule("(bu)s$", "$1ses"),
		mkRule("(buffal|tomat)o$", "$1oes"),
		mkRule("([ti])um$", "$1a"),
		mkRule("(.*)sis$", "$1ses"),
		mkRule("(?:([^f])fe|([lr])f)$", "$1$2ves"),
		mkRule("(hive)$", "$1s"),
		mkRule("([^aeiouy]|qu)y$", "$1ies"),
		mkRule("(x|ch|ss|sh)$", "$1es"),
		mkRule("(matr|vert|ind)ix|ex$", "$1ices"),
		mkRule("([m|l])ouse$", "$1ice"),
		mkRule("^(ox)$", "$1en"),
		mkRule("(quiz)$", "$1zes"),
	}

	irregular = map[string]string{
		"person": "people",
		"man":    "men",
		"child":  "children",
		"sex":    "sexes",
		"move":   "moves",
	}

	uncountables = map[string]struct{}{
		"equipment":   struct{}{},
		"information": struct{}{},
		"rice":        struct{}{},
		"money":       struct{}{},
		"species":     struct{}{},
		"series":      struct{}{},
		"fish":        struct{}{},
		"sheep":       struct{}{},
	}
)

func init() {
	// cache compiled regular expressions
	for i := range plurals {
		r := &plurals[i]
		s := "(?i)" + r.pattern
		r.rx = regexp.MustCompile(s)
	}
}

/*
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	var res []rune
	firstChar := true
	for _, c := range s {
		if firstChar {
			if unicode.IsUpper(c) {
				// already capitalized
				return s
			}
			firstChar = false
			c = unicode.ToUpper(c)
		}
		res = append(res, c)
	}
	return string(res)
}
*/

func tryRules(sl string) (string, bool) {
	nRules := len(plurals)
	for i := nRules - 1; i >= 0; i-- {
		r := plurals[i]
		match := r.rx.FindAllStringSubmatchIndex(sl, -1)
		if len(match) == 0 {
			continue
		}
		// TODO: implement
	}
	return sl, false
}

func pluralize(s string) string {
	sl := strings.ToLower(s)
	if _, ok := uncountables[sl]; ok {
		return s
	}
	if res, ok := irregular[s]; ok {
		return res
	}
	res, ok := tryRules(sl)
	if ok {
		return res
	}
	// TODO: temporary, redundant if tryRules is implemented
	if strings.HasSuffix(s, "s") {
		return s
	}
	return s + "s"
}
