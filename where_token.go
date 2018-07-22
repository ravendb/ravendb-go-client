package ravendb

import "math"

var _ QueryToken = &WhereToken{}

type MethodType = string

const (
	MethodType_CMP_X_CHG = "CmpXChg"
)

type WhereMethodCall struct {
	methodType MethodType
	parameters []string
	property   string
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

func NewWhereOptions() *WhereOptions {
	return &WhereOptions{}
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
	res := NewWhereToken()
	res.whereOperator = op
	res.fieldName = fieldName
	res.parameterName = parameterName
	return res
}

func WhereToken_create2(op WhereOperator, fieldName string, parameterName string, options *WhereOptions) *WhereToken {
	token := NewWhereToken()
	token.fieldName = fieldName
	token.parameterName = parameterName
	token.whereOperator = op
	if options != nil {
		token.options = options
	} else {
		token.options = NewWhereOptions()
	}
	return token
}

func (t *WhereToken) getFieldName() string {
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

func (t *WhereToken) getOptions() *WhereOptions {
	return t.options
}

func (t *WhereToken) setOptions(options *WhereOptions) {
	t.options = options
}

func (t *WhereToken) addAlias(alias string) {
	if t.fieldName == "id()" {
		return
	}
	t.fieldName = alias + "." + t.fieldName
}

func (t *WhereToken) writeMethod(writer *StringBuilder) bool {
	if t.options.getMethod() != nil {
		switch t.options.getMethod().methodType {
		case MethodType_CMP_X_CHG:
			writer.append("cmpxchg(")
			break
		default:
			panicIf(true, "Unsupported method: %s", t.options.getMethod().methodType)
			// TODO: return as error?
			//return NewIllegalArgumentException("Unsupported method: %s", options.getMethod().methodType);
		}

		first := true
		for _, parameter := range t.options.getMethod().parameters {
			if !first {
				writer.append(",")
			}
			first = false
			writer.append("$")
			writer.append(parameter)
		}
		writer.append(")")

		if t.options.getMethod().property != "" {
			writer.append(".")
			writer.append(t.options.getMethod().property)
		}
		return true
	}

	return false
}

func (t *WhereToken) writeTo(writer *StringBuilder) {
	options := t.options
	if options.boost != 0 {
		writer.append("boost(")
	}

	if options.fuzzy != 0 {
		writer.append("fuzzy(")
	}

	if options.proximity != 0 {
		writer.append("proximity(")
	}

	if options.exact {
		writer.append("exact(")
	}

	switch t.whereOperator {
	case WhereOperator_SEARCH:
		writer.append("search(")
		break
	case WhereOperator_LUCENE:
		writer.append("lucene(")
		break
	case WhereOperator_STARTS_WITH:
		writer.append("startsWith(")
		break
	case WhereOperator_ENDS_WITH:
		writer.append("endsWith(")
		break
	case WhereOperator_EXISTS:
		writer.append("exists(")
		break
	case WhereOperator_SPATIAL_WITHIN:
		writer.append("spatial.within(")
		break
	case WhereOperator_SPATIAL_CONTAINS:
		writer.append("spatial.contains(")
		break
	case WhereOperator_SPATIAL_DISJOINT:
		writer.append("spatial.disjoint(")
		break
	case WhereOperator_SPATIAL_INTERSECTS:
		writer.append("spatial.intersects(")
		break
	case WhereOperator_REGEX:
		writer.append("regex(")
		break
	}

	t.writeInnerWhere(writer)

	if options.exact {
		writer.append(")")
	}

	if options.proximity != 0 {
		writer.append(", ")
		writer.append(options.proximity)
		writer.append(")")
	}

	if options.fuzzy != 0 {
		writer.append(", ")
		writer.append(options.fuzzy)
		writer.append(")")
	}

	if options.boost != 0 {
		writer.append(", ")
		writer.append(options.boost)
		writer.append(")")
	}
}

func (t *WhereToken) writeInnerWhere(writer *StringBuilder) {

	QueryToken_writeField(writer, t.fieldName)

	switch t.whereOperator {
	case WhereOperator_EQUALS:
		writer.append(" = ")
		break

	case WhereOperator_NOT_EQUALS:
		writer.append(" != ")
		break
	case WhereOperator_GREATER_THAN:
		writer.append(" > ")
		break
	case WhereOperator_GREATER_THAN_OR_EQUAL:
		writer.append(" >= ")
		break
	case WhereOperator_LESS_THAN:
		writer.append(" < ")
		break
	case WhereOperator_LESS_THAN_OR_EQUAL:
		writer.append(" <= ")
		break
	default:
		t.specialOperator(writer)
		return
	}

	if !t.writeMethod(writer) {
		writer.append("$").append(t.parameterName)
	}
}

func (t *WhereToken) specialOperator(writer *StringBuilder) {
	options := t.options
	parameterName := t.parameterName
	switch t.whereOperator {
	case WhereOperator_IN:
		writer.append(" in ($")
		writer.append(parameterName)
		writer.append(")")
	case WhereOperator_ALL_IN:
		writer.append(" all in ($")
		writer.append(parameterName)
		writer.append(")")
	case WhereOperator_BETWEEN:
		writer.append(" between $")
		writer.append(options.fromParameterName)
		writer.append(" and $")
		writer.append(options.toParameterName)
	case WhereOperator_SEARCH:
		writer.append(", $")
		writer.append(parameterName)
		if options.searchOperator == SearchOperator_AND {
			writer.append(", and")
		}
		writer.append(")")
	case WhereOperator_LUCENE, WhereOperator_STARTS_WITH, WhereOperator_ENDS_WITH, WhereOperator_REGEX:
		writer.append(", $")
		writer.append(parameterName)
		writer.append(")")
	case WhereOperator_EXISTS:
		writer.append(")")
	case WhereOperator_SPATIAL_WITHIN, WhereOperator_SPATIAL_CONTAINS, WhereOperator_SPATIAL_DISJOINT, WhereOperator_SPATIAL_INTERSECTS:
		writer.append(", ")
		options.whereShape.writeTo(writer)

		if math.Abs(options.distanceErrorPct-Constants_Documents_Indexing_Spatial_DEFAULT_DISTANCE_ERROR_PCT) > 1e-40 {
			writer.append(", ")
			writer.append(options.distanceErrorPct)
		}
		writer.append(")")
	default:
		panicIf(true, "unsupported operator %s", t.whereOperator)
	}
}
