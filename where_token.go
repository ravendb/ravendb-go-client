package ravendb

import (
	"math"
	"strings"
)

var _ queryToken = &whereToken{}

type MethodsType = string

const (
	MethodsTypeCmpXChg = "CmpXChg"
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
		case MethodsTypeCmpXChg:
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
	case WhereOperatorSearch:
		writer.WriteString("search(")
	case WhereOperatorLucene:
		writer.WriteString("lucene(")
	case WhereOperatorStartsWith:
		writer.WriteString("startsWith(")
	case WhereOperatorEndsWith:
		writer.WriteString("endsWith(")
	case WhereOperatorExists:
		writer.WriteString("exists(")
	case WhereOperatorSpatialWithin:
		writer.WriteString("spatial.within(")
	case WhereOperatorSpatialContains:
		writer.WriteString("spatial.contains(")
	case WhereOperatorSpatialDisjoint:
		writer.WriteString("spatial.disjoint(")
	case WhereOperatorSpatialIntersects:
		writer.WriteString("spatial.intersects(")
	case WhereOperatorRegex:
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
	case WhereOperatorEquals:
		writer.WriteString(" = ")
	case WhereOperatorNotEquals:
		writer.WriteString(" != ")
	case WhereOperatorGreaterThan:
		writer.WriteString(" > ")
	case WhereOperatorGreaterThanOrEqual:
		writer.WriteString(" >= ")
	case WhereOperatorLessThan:
		writer.WriteString(" < ")
	case WhereOperatorLessThanOrEqual:
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
	case WhereOperatorIn:
		writer.WriteString(" in ($")
		writer.WriteString(parameterName)
		writer.WriteString(")")
	case WhereOperatorAllIn:
		writer.WriteString(" all in ($")
		writer.WriteString(parameterName)
		writer.WriteString(")")
	case WhereOperatorBetween:
		writer.WriteString(" between $")
		writer.WriteString(options.fromParameterName)
		writer.WriteString(" and $")
		writer.WriteString(options.toParameterName)
	case WhereOperatorSearch:
		writer.WriteString(", $")
		writer.WriteString(parameterName)
		if options.searchOperator == SearchOperator_AND {
			writer.WriteString(", and")
		}
		writer.WriteString(")")
	case WhereOperatorLucene, WhereOperatorStartsWith, WhereOperatorEndsWith, WhereOperatorRegex:
		writer.WriteString(", $")
		writer.WriteString(parameterName)
		writer.WriteString(")")
	case WhereOperatorExists:
		writer.WriteString(")")
	case WhereOperatorSpatialWithin, WhereOperatorSpatialContains, WhereOperatorSpatialDisjoint, WhereOperatorSpatialIntersects:
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
