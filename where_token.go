package ravendb

import (
	"math"
	"strings"
)

var _ QueryToken = &WhereToken{}

type MethodsType = string

const (
	MethodsType_CMP_X_CHG = "CmpXChg"
)

type WhereMethodCall struct {
	methodType MethodsType
	parameters []string
	property   string
}

func NewWhereMethodCall() *WhereMethodCall {
	return &WhereMethodCall{}
}

type WhereOptions struct {
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
	method           *WhereMethodCall
	whereShape       *ShapeToken
	distanceErrorPct float64
}

func WhereOptions_defaultOptions() *WhereOptions {
	return NewWhereOptions()
}

func NewWhereOptions() *WhereOptions {
	return &WhereOptions{}
}

func NewWhereOptionsWithExact(exact bool) *WhereOptions {
	return &WhereOptions{
		exact: exact,
	}
}

func NewWhereOptionsWithOperator(search SearchOperator) *WhereOptions {
	return &WhereOptions{
		searchOperator: search,
	}
}

func NewWhereOptionsWithTokenAndDistance(shape *ShapeToken, distance float64) *WhereOptions {
	return &WhereOptions{
		whereShape:       shape,
		distanceErrorPct: distance,
	}
}

func NewWhereOptionsWithMethod(methodType MethodsType, parameters []string, property string, exact bool) *WhereOptions {
	method := NewWhereMethodCall()
	method.methodType = methodType
	method.parameters = parameters
	method.property = property

	return &WhereOptions{
		method: method,
		exact:  exact,
	}
}

func NewWhereOptionsWithFromTo(exact bool, from string, to string) *WhereOptions {
	return &WhereOptions{
		exact:             exact,
		fromParameterName: from,
		toParameterName:   to,
	}
}

func (o *WhereOptions) getSearchOperator() SearchOperator {
	return o.searchOperator
}

func (o *WhereOptions) setSearchOperator(searchOperator SearchOperator) {
	o.searchOperator = searchOperator
}

func (o *WhereOptions) getFromParameterName() string {
	return o.fromParameterName
}

func (o *WhereOptions) setFromParameterName(fromParameterName string) {
	o.fromParameterName = fromParameterName
}

func (o *WhereOptions) getToParameterName() string {
	return o.toParameterName
}

func (o *WhereOptions) setToParameterName(toParameterName string) {
	o.toParameterName = toParameterName
}

func (o *WhereOptions) getBoost() float64 {
	return o.boost
}

func (o *WhereOptions) setBoost(boost float64) {
	o.boost = boost
}

func (o *WhereOptions) getFuzzy() float64 {
	return o.fuzzy
}

func (o *WhereOptions) setFuzzy(fuzzy float64) {
	o.fuzzy = fuzzy
}

func (o *WhereOptions) getProximity() int {
	return o.proximity
}

func (o *WhereOptions) setProximity(proximity int) {
	o.proximity = proximity
}

func (o *WhereOptions) isExact() bool {
	return o.exact
}

func (o *WhereOptions) setExact(exact bool) {
	o.exact = exact
}

func (o *WhereOptions) getMethod() *WhereMethodCall {
	return o.method
}

func (o *WhereOptions) setMethod(method *WhereMethodCall) {
	o.method = method
}

func (o *WhereOptions) getWhereShape() *ShapeToken {
	return o.whereShape
}

func (o *WhereOptions) setWhereShape(whereShape *ShapeToken) {
	o.whereShape = whereShape
}

func (o *WhereOptions) getDistanceErrorPct() float64 {
	return o.distanceErrorPct
}

func (o *WhereOptions) setDistanceErrorPct(distanceErrorPct float64) {
	o.distanceErrorPct = distanceErrorPct
}

type WhereToken struct {
	fieldName     string
	whereOperator WhereOperator
	parameterName string
	options       *WhereOptions
}

func NewWhereToken() *WhereToken {
	return &WhereToken{}
}

func WhereToken_create(op WhereOperator, fieldName string, parameterName string) *WhereToken {
	return WhereToken_createWithOptions(op, fieldName, parameterName, nil)
}

func WhereToken_createWithOptions(op WhereOperator, fieldName string, parameterName string, options *WhereOptions) *WhereToken {
	token := NewWhereToken()
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

func (t *WhereToken) GetFieldName() string {
	return t.fieldName
}

func (t *WhereToken) setFieldName(fieldName string) {
	t.fieldName = fieldName
}

func (t *WhereToken) getWhereOperator() WhereOperator {
	return t.whereOperator
}

func (t *WhereToken) setWhereOperator(whereOperator WhereOperator) {
	t.whereOperator = whereOperator
}

func (t *WhereToken) getParameterName() string {
	return t.parameterName
}

func (t *WhereToken) setParameterName(parameterName string) {
	t.parameterName = parameterName
}

func (t *WhereToken) GetOptions() *WhereOptions {
	return t.options
}

func (t *WhereToken) SetOptions(options *WhereOptions) {
	t.options = options
}

func (t *WhereToken) addAlias(alias string) {
	if t.fieldName == "id()" {
		return
	}
	t.fieldName = alias + "." + t.fieldName
}

func (t *WhereToken) writeMethod(writer *strings.Builder) bool {
	if t.options.getMethod() != nil {
		switch t.options.getMethod().methodType {
		case MethodsType_CMP_X_CHG:
			writer.WriteString("cmpxchg(")
			break
		default:
			panicIf(true, "Unsupported method: %s", t.options.getMethod().methodType)
			// TODO: return as error?
			//return NewIllegalArgumentException("Unsupported method: %s", options.getMethod().methodType);
		}

		first := true
		for _, parameter := range t.options.getMethod().parameters {
			if !first {
				writer.WriteString(",")
			}
			first = false
			writer.WriteString("$")
			writer.WriteString(parameter)
		}
		writer.WriteString(")")

		if t.options.getMethod().property != "" {
			writer.WriteString(".")
			writer.WriteString(t.options.getMethod().property)
		}
		return true
	}

	return false
}

func (t *WhereToken) WriteTo(writer *strings.Builder) {
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
		break
	case WhereOperator_LUCENE:
		writer.WriteString("lucene(")
		break
	case WhereOperator_STARTS_WITH:
		writer.WriteString("startsWith(")
		break
	case WhereOperator_ENDS_WITH:
		writer.WriteString("endsWith(")
		break
	case WhereOperator_EXISTS:
		writer.WriteString("exists(")
		break
	case WhereOperator_SPATIAL_WITHIN:
		writer.WriteString("spatial.within(")
		break
	case WhereOperator_SPATIAL_CONTAINS:
		writer.WriteString("spatial.contains(")
		break
	case WhereOperator_SPATIAL_DISJOINT:
		writer.WriteString("spatial.disjoint(")
		break
	case WhereOperator_SPATIAL_INTERSECTS:
		writer.WriteString("spatial.intersects(")
		break
	case WhereOperator_REGEX:
		writer.WriteString("regex(")
		break
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

func (t *WhereToken) writeInnerWhere(writer *strings.Builder) {

	QueryToken_writeField(writer, t.fieldName)

	switch t.whereOperator {
	case WhereOperator_EQUALS:
		writer.WriteString(" = ")
		break

	case WhereOperator_NOT_EQUALS:
		writer.WriteString(" != ")
		break
	case WhereOperator_GREATER_THAN:
		writer.WriteString(" > ")
		break
	case WhereOperator_GREATER_THAN_OR_EQUAL:
		writer.WriteString(" >= ")
		break
	case WhereOperator_LESS_THAN:
		writer.WriteString(" < ")
		break
	case WhereOperator_LESS_THAN_OR_EQUAL:
		writer.WriteString(" <= ")
		break
	default:
		t.specialOperator(writer)
		return
	}

	if !t.writeMethod(writer) {
		writer.WriteString("$")
		writer.WriteString(t.parameterName)
	}
}

func (t *WhereToken) specialOperator(writer *strings.Builder) {
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
		options.whereShape.WriteTo(writer)

		if math.Abs(options.distanceErrorPct-Constants_Documents_Indexing_Spatial_DEFAULT_DISTANCE_ERROR_PCT) > 1e-40 {
			writer.WriteString(", ")
			builderWriteFloat64(writer, options.distanceErrorPct)
		}
		writer.WriteString(")")
	default:
		panicIf(true, "unsupported operator %d", t.whereOperator)
	}
}
