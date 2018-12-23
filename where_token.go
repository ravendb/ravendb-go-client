package ravendb

import (
	"math"
	"strings"
)

var _ queryToken = &whereToken{}

type MethodsType = string

const (
	MethodsType_CMP_X_CHG = "CmpXChg"
)

type whereMethodCall struct {
	methodType MethodsType
	parameters []string
	property   string
}

func newWhereMethodCall() *whereMethodCall {
	return &whereMethodCall{}
}

type whereOptions struct {
	searchOperator    SearchOperator
	fromParameterName string
	toParameterName   string
	// TODO: does it have to be *float64 to indicate 'no value provided' ?
	boost float64
	// TODO: does it have to be *float64 to indicate 'no value provided' ?
	fuzzy float64
	// TODO: does it have to be *int to indicate 'no value provided' ?
	proximity        int
	exact            bool
	method           *whereMethodCall
	whereShape       *shapeToken
	distanceErrorPct float64
}

func WhereOptions_defaultOptions() *whereOptions {
	return newWhereOptions()
}

func newWhereOptions() *whereOptions {
	return &whereOptions{}
}

func NewWhereOptionsWithExact(exact bool) *whereOptions {
	return &whereOptions{
		exact: exact,
	}
}

func NewWhereOptionsWithOperator(search SearchOperator) *whereOptions {
	return &whereOptions{
		searchOperator: search,
	}
}

func NewWhereOptionsWithTokenAndDistance(shape *shapeToken, distance float64) *whereOptions {
	return &whereOptions{
		whereShape:       shape,
		distanceErrorPct: distance,
	}
}

func NewWhereOptionsWithMethod(methodType MethodsType, parameters []string, property string, exact bool) *whereOptions {
	method := newWhereMethodCall()
	method.methodType = methodType
	method.parameters = parameters
	method.property = property

	return &whereOptions{
		method: method,
		exact:  exact,
	}
}

func NewWhereOptionsWithFromTo(exact bool, from string, to string) *whereOptions {
	return &whereOptions{
		exact:             exact,
		fromParameterName: from,
		toParameterName:   to,
	}
}

type whereToken struct {
	fieldName     string
	whereOperator WhereOperator
	parameterName string
	options       *whereOptions
}

func newWhereToken() *whereToken {
	return &whereToken{}
}

func createWhereToken(op WhereOperator, fieldName string, parameterName string) *whereToken {
	return createWhereTokenWithOptions(op, fieldName, parameterName, nil)
}

func createWhereTokenWithOptions(op WhereOperator, fieldName string, parameterName string, options *whereOptions) *whereToken {
	token := newWhereToken()
	token.fieldName = fieldName
	token.parameterName = parameterName
	token.whereOperator = op
	if options != nil {
		token.options = options
	} else {
		token.options = WhereOptions_defaultOptions()
	}
	return token
}

func (t *whereToken) addAlias(alias string) {
	if t.fieldName == "id()" {
		return
	}
	t.fieldName = alias + "." + t.fieldName
}

func (t *whereToken) writeMethod(writer *strings.Builder) bool {
	if t.options.method != nil {
		switch t.options.method.methodType {
		case MethodsType_CMP_X_CHG:
			writer.WriteString("cmpxchg(")
		default:
			panicIf(true, "Unsupported method: %s", t.options.method.methodType)
			// TODO: return as error?
			//return newIllegalArgumentError("Unsupported method: %s", options.method.methodType);
		}

		first := true
		for _, parameter := range t.options.method.parameters {
			if !first {
				writer.WriteString(",")
			}
			first = false
			writer.WriteString("$")
			writer.WriteString(parameter)
		}
		writer.WriteString(")")

		if t.options.method.property != "" {
			writer.WriteString(".")
			writer.WriteString(t.options.method.property)
		}
		return true
	}

	return false
}

func (t *whereToken) writeTo(writer *strings.Builder) {
	options := t.options
	if options.boost != 0 {
		writer.WriteString("boost(")
	}

	if options.fuzzy != 0 {
		writer.WriteString("fuzzy(")
	}

	if options.proximity != 0 {
		writer.WriteString("proximity(")
	}

	if options.exact {
		writer.WriteString("exact(")
	}

	switch t.whereOperator {
	case WhereOperator_SEARCH:
		writer.WriteString("search(")
	case WhereOperator_LUCENE:
		writer.WriteString("lucene(")
	case WhereOperator_STARTS_WITH:
		writer.WriteString("startsWith(")
	case WhereOperator_ENDS_WITH:
		writer.WriteString("endsWith(")
	case WhereOperator_EXISTS:
		writer.WriteString("exists(")
	case WhereOperator_SPATIAL_WITHIN:
		writer.WriteString("spatial.within(")
	case WhereOperator_SPATIAL_CONTAINS:
		writer.WriteString("spatial.contains(")
	case WhereOperator_SPATIAL_DISJOINT:
		writer.WriteString("spatial.disjoint(")
	case WhereOperator_SPATIAL_INTERSECTS:
		writer.WriteString("spatial.intersects(")
	case WhereOperator_REGEX:
		writer.WriteString("regex(")
	}

	t.writeInnerWhere(writer)

	if options.exact {
		writer.WriteString(")")
	}

	if options.proximity != 0 {
		writer.WriteString(", ")
		builderWriteInt(writer, options.proximity)
		writer.WriteString(")")
	}

	if options.fuzzy != 0 {
		writer.WriteString(", ")
		builderWriteFloat64(writer, options.fuzzy)
		writer.WriteString(")")
	}

	if options.boost != 0 {
		writer.WriteString(", ")
		builderWriteFloat64(writer, options.boost)
		writer.WriteString(")")
	}
}

func (t *whereToken) writeInnerWhere(writer *strings.Builder) {

	writeQueryTokenField(writer, t.fieldName)

	switch t.whereOperator {
	case WhereOperator_EQUALS:
		writer.WriteString(" = ")
	case WhereOperator_NOT_EQUALS:
		writer.WriteString(" != ")
	case WhereOperator_GREATER_THAN:
		writer.WriteString(" > ")
	case WhereOperator_GREATER_THAN_OR_EQUAL:
		writer.WriteString(" >= ")
	case WhereOperator_LESS_THAN:
		writer.WriteString(" < ")
	case WhereOperator_LESS_THAN_OR_EQUAL:
		writer.WriteString(" <= ")
	default:
		t.specialOperator(writer)
		return
	}

	if !t.writeMethod(writer) {
		writer.WriteString("$")
		writer.WriteString(t.parameterName)
	}
}

func (t *whereToken) specialOperator(writer *strings.Builder) {
	options := t.options
	parameterName := t.parameterName
	switch t.whereOperator {
	case WhereOperator_IN:
		writer.WriteString(" in ($")
		writer.WriteString(parameterName)
		writer.WriteString(")")
	case WhereOperator_ALL_IN:
		writer.WriteString(" all in ($")
		writer.WriteString(parameterName)
		writer.WriteString(")")
	case WhereOperator_BETWEEN:
		writer.WriteString(" between $")
		writer.WriteString(options.fromParameterName)
		writer.WriteString(" and $")
		writer.WriteString(options.toParameterName)
	case WhereOperator_SEARCH:
		writer.WriteString(", $")
		writer.WriteString(parameterName)
		if options.searchOperator == SearchOperator_AND {
			writer.WriteString(", and")
		}
		writer.WriteString(")")
	case WhereOperator_LUCENE, WhereOperator_STARTS_WITH, WhereOperator_ENDS_WITH, WhereOperator_REGEX:
		writer.WriteString(", $")
		writer.WriteString(parameterName)
		writer.WriteString(")")
	case WhereOperator_EXISTS:
		writer.WriteString(")")
	case WhereOperator_SPATIAL_WITHIN, WhereOperator_SPATIAL_CONTAINS, WhereOperator_SPATIAL_DISJOINT, WhereOperator_SPATIAL_INTERSECTS:
		writer.WriteString(", ")
		options.whereShape.writeTo(writer)

		if math.Abs(options.distanceErrorPct-IndexingSpatialDefaultDistnaceErrorPct) > 1e-40 {
			writer.WriteString(", ")
			builderWriteFloat64(writer, options.distanceErrorPct)
		}
		writer.WriteString(")")
	default:
		panicIf(true, "unsupported operator %d", t.whereOperator)
	}
}
