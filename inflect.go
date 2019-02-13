package ravendb

// this is inflect.go from https://github.com/kjk/inflect
// included directly to minimize dependencies
// under MIT license: https://github.com/kjk/inflect/blob/master/LICENSE

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var irregularRules = [][]string{
	// Pronouns.
	{"I", "we"},
	{"me", "us"},
	{"he", "they"},
	{"she", "they"},
	{"them", "them"},
	{"myself", "ourselves"},
	{"yourself", "yourselves"},
	{"itself", "themselves"},
	{"herself", "themselves"},
	{"himself", "themselves"},
	{"themself", "themselves"},
	{"is", "are"},
	{"was", "were"},
	{"has", "have"},
	{"this", "these"},
	{"that", "those"},
	// Words ending in with a consonant and `o`.
	{"echo", "echoes"},
	{"dingo", "dingoes"},
	{"volcano", "volcanoes"},
	{"tornado", "tornadoes"},
	{"torpedo", "torpedoes"},
	// Ends with `us`.
	{"genus", "genera"},
	{"viscus", "viscera"},
	// Ends with `ma`.
	{"stigma", "stigmata"},
	{"stoma", "stomata"},
	{"dogma", "dogmata"},
	{"lemma", "lemmata"},
	{"schema", "schemata"},
	{"anathema", "anathemata"},
	// Other irregular rules.
	{"ox", "oxen"},
	{"axe", "axes"},
	{"die", "dice"},
	{"yes", "yeses"},
	{"foot", "feet"},
	{"eave", "eaves"},
	{"goose", "geese"},
	{"tooth", "teeth"},
	{"quiz", "quizzes"},
	{"human", "humans"},
	{"proof", "proofs"},
	{"carve", "carves"},
	{"valve", "valves"},
	{"looey", "looies"},
	{"thief", "thieves"},
	{"groove", "grooves"},
	{"pickaxe", "pickaxes"},
	{"whiskey", "whiskies"},
}

var pluralizationRules = [][]string{
	{`/s?$/i`, `s`},
	{`/[^\u0000-\u007F]$/i`, `$0`},
	{`/([^aeiou]ese)$/i`, `$1`},
	{`/(ax|test)is$/i`, `$1es`},
	{`/(alias|[^aou]us|t[lm]as|gas|ris)$/i`, `$1es`},
	{`/(e[mn]u)s?$/i`, `$1s`},
	{`/([^l]ias|[aeiou]las|[ejzr]as|[iu]am)$/i`, `$1`},
	{`/(alumn|syllab|octop|vir|radi|nucle|fung|cact|stimul|termin|bacill|foc|uter|loc|strat)(?:us|i)$/i`, `$1i`},
	{`/(alumn|alg|vertebr)(?:a|ae)$/i`, `$1ae`},
	{`/(seraph|cherub)(?:im)?$/i`, `$1im`},
	{`/(her|at|gr)o$/i`, `$1oes`},
	{`/(agend|addend|millenni|dat|extrem|bacteri|desiderat|strat|candelabr|errat|ov|symposi|curricul|automat|quor)(?:a|um)$/i`, `$1a`},
	{`/(apheli|hyperbat|periheli|asyndet|noumen|phenomen|criteri|organ|prolegomen|hedr|automat)(?:a|on)$/i`, `$1a`},
	{`/sis$/i`, `ses`},
	{`/(?:(kni|wi|li)fe|(ar|l|ea|eo|oa|hoo)f)$/i`, `$1$2ves`},
	{`/([^aeiouy]|qu)y$/i`, `$1ies`},
	{`/([^ch][ieo][ln])ey$/i`, `$1ies`},
	{`/(x|ch|ss|sh|zz)$/i`, `$1es`},
	{`/(matr|cod|mur|sil|vert|ind|append)(?:ix|ex)$/i`, `$1ices`},
	{`/\b((?:tit)?m|l)(?:ice|ouse)$/i`, `$1ice`},
	{`/(pe)(?:rson|ople)$/i`, `$1ople`},
	{`/(child)(?:ren)?$/i`, `$1ren`},
	{`/eaux$/i`, `$0`},
	{`/m[ae]n$/i`, `men`},
	{`thou`, `you`},
}

var singularizationRules = [][]string{
	{`/s$/i`, ``},
	{`/(ss)$/i`, `$1`},
	{`/(wi|kni|(?:after|half|high|low|mid|non|night|[^\w]|^)li)ves$/i`, `$1fe`},
	{`/(ar|(?:wo|[ae])l|[eo][ao])ves$/i`, `$1f`},
	{`/ies$/i`, `y`},
	{`/\b([pl]|zomb|(?:neck|cross)?t|coll|faer|food|gen|goon|group|lass|talk|goal|cut)ies$/i`, `$1ie`},
	{`/\b(mon|smil)ies$/i`, `$1ey`},
	{`/\b((?:tit)?m|l)ice$/i`, `$1ouse`},
	{`/(seraph|cherub)im$/i`, `$1`},
	{`/(x|ch|ss|sh|zz|tto|go|cho|alias|[^aou]us|t[lm]as|gas|(?:her|at|gr)o|ris)(?:es)?$/i`, `$1`},
	{`/(analy|ba|diagno|parenthe|progno|synop|the|empha|cri)(?:sis|ses)$/i`, `$1sis`},
	{`/(movie|twelve|abuse|e[mn]u)s$/i`, `$1`},
	{`/(test)(?:is|es)$/i`, `$1is`},
	{`/(alumn|syllab|octop|vir|radi|nucle|fung|cact|stimul|termin|bacill|foc|uter|loc|strat)(?:us|i)$/i`, `$1us`},
	{`/(agend|addend|millenni|dat|extrem|bacteri|desiderat|strat|candelabr|errat|ov|symposi|curricul|quor)a$/i`, `$1um`},
	{`/(apheli|hyperbat|periheli|asyndet|noumen|phenomen|criteri|organ|prolegomen|hedr|automat)a$/i`, `$1on`},
	{`/(alumn|alg|vertebr)ae$/i`, `$1a`},
	{`/(cod|mur|sil|vert|ind)ices$/i`, `$1ex`},
	{`/(matr|append)ices$/i`, `$1ix`},
	{`/(pe)(rson|ople)$/i`, `$1rson`},
	{`/(child)ren$/i`, `$1`},
	{`/(eau)x?$/i`, `$1`},
	{`/men$/i`, `man`},
}

//Uncountable rules.
var uncountableRules = []string{
	// singular words with no plurals.
	"adulthood",
	"advice",
	"agenda",
	"aid",
	"alcohol",
	"ammo",
	"anime",
	"athletics",
	"audio",
	"bison",
	"blood",
	"bream",
	"buffalo",
	"butter",
	"carp",
	"cash",
	"chassis",
	"chess",
	"clothing",
	"cod",
	"commerce",
	"cooperation",
	"corps",
	"debris",
	"diabetes",
	"digestion",
	"elk",
	"energy",
	"equipment",
	"excretion",
	"expertise",
	"flounder",
	"fun",
	"gallows",
	"garbage",
	"graffiti",
	"headquarters",
	"health",
	"herpes",
	"highjinks",
	"homework",
	"housework",
	"information",
	"jeans",
	"justice",
	"kudos",
	"labour",
	"literature",
	"machinery",
	"mackerel",
	"mail",
	"media",
	"mews",
	"moose",
	"music",
	"mud",
	"manga",
	"news",
	"pike",
	"plankton",
	"pliers",
	"police",
	"pollution",
	"premises",
	"rain",
	"research",
	"rice",
	"salmon",
	"scissors",
	"series",
	"sewage",
	"shambles",
	"shrimp",
	"species",
	"staff",
	"swine",
	"tennis",
	"traffic",
	"transportation",
	"trout",
	"tuna",
	"wealth",
	"welfare",
	"whiting",
	"wildebeest",
	"wildlife",
	"you",
	// Regexes.
	`/[^aeiou]ese$/i`, // "chinese", "japanese"
	`/deer$/i`,        // "deer", "reindeer"
	`/fish$/i`,        // "fish", "blowfish", "angelfish"
	`/measles$/i`,
	`/o[iu]s$/i`, // "carnivorous"
	`/pox$/i`,    // "chickpox", "smallpox"
	`/sheep$/i`,
}

type rxRule struct {
	// TODO: for debugging, maybe remove when working
	rxStrJs string
	rxStrGo string

	rx          *regexp.Regexp
	replacement string
}

// Rule storage - pluralize and singularize need to be run sequentially,
// while other rules can be optimized using an object for instant lookups.
var pluralRules []rxRule
var singularRules []rxRule
var irregularPlurals = map[string]string{}
var irregularSingles = map[string]string{}
var uncountables = map[string]string{}

func init() {
	// order is important
	addIrregularRules()
	addPluralizationRules()
	addSingularizationRules()
	addUncountableRules()
}

// Add a pluralization rule to the collection.
func addPluralRule(rule string, replacement string) {
	rx, rxStrGo := sanitizeRule(rule)
	r := rxRule{
		rxStrJs:     rule,
		rxStrGo:     rxStrGo,
		rx:          rx,
		replacement: jsReplaceSyntaxToGo(replacement),
	}
	pluralRules = append(pluralRules, r)
}

var (
	unicodeSyntaxRx = regexp.MustCompile(`\\u([[:xdigit:]]{4})`)
)

// best-effort of converting javascript regex syntax to equivalent go syntax
func jsRxSyntaxToGo(rx string) string {
	s := rx
	caseInsensitive := false
	panicIf(s[0] != '/', "expected '%s' to start with '/'", rx)
	s = s[1:]
	n := len(s)
	if s[n-1] == 'i' {
		n--
		caseInsensitive = true
		s = s[:n]
	}
	panicIf(s[n-1] != '/', "expected '%s' to end with '/'", rx)
	s = s[:n-1]
	// \uNNNN syntax for unicode code points to \x{NNNN} syntax for hex character code
	s = unicodeSyntaxRx.ReplaceAllString(s, "\\x{$1}")
	if caseInsensitive {
		s = "(?i)" + s
	}
	return s
}

func jsReplaceSyntaxToGo(s string) string {
	s = strings.Replace(s, "$0", "${0}", -1)
	s = strings.Replace(s, "$1", "${1}", -1)
	s = strings.Replace(s, "$2", "${2}", -1)
	return s
}

// Sanitize a pluralization rule to a usable regular expression.
func sanitizeRule(rule string) (*regexp.Regexp, string) {
	// in JavaScript, regexpes start with /
	// others are just regular strings
	var s string
	if rule[0] != '/' {
		// a plain string match is converted to regexp that:
		// ^ ... $ : does exact match (matches at the beginning and end)
		// (?i) : is case-insensitive
		s = `(?i)^` + rule + `$`
	} else {
		s = jsRxSyntaxToGo(rule)
	}
	return regexp.MustCompile(s), s
}

// Add a singularization rule to the collection.
func addSingularRule(rule, replacement string) {
	rx, rxGo := sanitizeRule(rule)
	r := rxRule{
		rxStrJs:     rule,
		rxStrGo:     rxGo,
		rx:          rx,
		replacement: jsReplaceSyntaxToGo(replacement),
	}
	singularRules = append(singularRules, r)
}

// copied from strings.ToUpper
// returns true if s is uppercase
func isUpper(s string) bool {
	isASCII, hasLower := true, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= utf8.RuneSelf {
			isASCII = false
			break
		}
		hasLower = hasLower || (c >= 'a' && c <= 'z')
	}
	if isASCII {
		return !hasLower
	}
	for r := range s {
		if !unicode.IsUpper(rune(r)) {
			return false
		}
	}
	return true
}

// Pass in a word token to produce a function that can replicate the case on
// another word.
func restoreCase(word string, token string) string {
	// Tokens are an exact match.
	if word == token {
		return token
	}

	// Upper cased words. E.g. "HELLO".
	if isUpper(word) {
		return strings.ToUpper(token)
	}

	// Title cased words. E.g. "Title".
	prefix := word[:1]
	if isUpper(prefix) {
		return strings.ToUpper(token[:1]) + strings.ToLower(token[1:])
	}

	// Lower cased words. E.g. "test".
	return strings.ToLower(token)
}

// Replace a word using a rule.
func replace(word string, rule rxRule) string {
	// TODO: not sure if this covers all possibilities
	repl := rule.replacement
	if isUpper(word) {
		repl = strings.ToUpper(repl)
	}
	return rule.rx.ReplaceAllString(word, repl)
}

// Sanitize a word by passing in the word and sanitization rules.
func sanitizeWord(token string, word string, rules []rxRule) string {
	// Empty string or doesn't need fixing.
	if len(token) == 0 {
		return word
	}
	if _, ok := uncountables[token]; ok {
		return word
	}

	// Iterate over the sanitization rules and use the first one to match.
	// important that we iterate from the end
	n := len(rules)
	for i := n - 1; i >= 0; i-- {
		rule := rules[i]
		if rule.rx.MatchString(word) {
			return replace(word, rule)
		}
	}
	return word
}

// Replace a word with the updated word.
func replaceWord(word string, replaceMap map[string]string, keepMap map[string]string, rules []rxRule) string {
	// Get the correct token and case restoration functions.
	token := strings.ToLower(word)

	// Check against the keep object map.
	if _, ok := keepMap[token]; ok {
		return restoreCase(word, token)
	}

	// Check against the replacement map for a direct word replacement.
	if s, ok := replaceMap[token]; ok {
		return restoreCase(word, s)
	}

	// Run all the rules against the word.
	return sanitizeWord(token, word, rules)
}

// Check if a word is part of the map.
func checkWord(word string, replaceMap map[string]string, keepMap map[string]string, rules []rxRule) bool {
	token := strings.ToLower(word)

	if _, ok := keepMap[token]; ok {
		return true
	}

	if _, ok := replaceMap[token]; ok {
		return false
	}

	return sanitizeWord(token, token, rules) == token
}

// Add an irregular word definition.
func addIrregularRules() {
	for _, rule := range irregularRules {
		single := strings.ToLower(rule[0])
		plural := strings.ToLower(rule[1])

		irregularSingles[single] = plural
		irregularPlurals[plural] = single
	}
}

func addSingularizationRules() {
	for _, r := range singularizationRules {
		addSingularRule(r[0], r[1])
	}
}

func addUncountableRules() {
	for _, word := range uncountableRules {
		if word[0] != '/' {
			word = strings.ToLower(word)
			uncountables[word] = word
			continue
		}
		// Set singular and plural references for the word.
		addPluralRule(word, "$0")
		addSingularRule(word, "$0")
	}
}

func addPluralizationRules() {
	for _, rule := range pluralizationRules {
		addPluralRule(rule[0], rule[1])
	}
}

// Pluralize or singularize a word based on the passed in count.
func Pluralize(word string, count int, inclusive bool) string {
	var res string
	if count == 1 {
		res = ToSingular(word)
	} else {
		res = ToPlural(word)
	}

	if inclusive {
		return strconv.Itoa(count) + " " + res
	}
	return res
}

// IsPlural retruns true if word is plural
func IsPlural(word string) bool {
	return checkWord(word, irregularSingles, irregularPlurals, pluralRules)
}

// ToSingular singularizes a word.
func ToSingular(word string) string {
	return replaceWord(word, irregularPlurals, irregularSingles, singularRules)
}

// IsSingular returns true if a word is singular
func IsSingular(word string) bool {
	return checkWord(word, irregularPlurals, irregularSingles, singularRules)
}

// ToPlural makes a pluralized version of a word
func ToPlural(word string) string {
	return replaceWord(word, irregularSingles, irregularPlurals, pluralRules)
}
