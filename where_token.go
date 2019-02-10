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
	boost             float64
	fuzzy             float64
	proximity         int
	exact             bool
	method            *whereMethodCall
	whereShape        *shapeToken
	distanceErrorPct  float64
}

func defaultWhereOptions() *whereOptions {
	return newWhereOptions()
}

func newWhereOptions() *whereOptions {
	return &whereOptions{}
}

func newWhereOptionsWithExact(exact bool) *whereOptions {
	return &whereOptions{
		exact: exact,
	}
}

func newWhereOptionsWithOperator(search SearchOperator) *whereOptions {
	return &whereOptions{
		searchOperator: search,
	}
}

func newWhereOptionsWithTokenAndDistance(shape *shapeToken, distance float64) *whereOptions {
	return &whereOptions{
		whereShape:       shape,
		distanceErrorPct: distance,
	}
}

func newWhereOptionsWithMethod(methodType MethodsType, parameters []string, property string, exact bool) *whereOptions {
	method := newWhereMethodCall()
	method.methodType = methodType
	method.parameters = parameters
	method.property = property

	return &whereOptions{
		method: method,
		exact:  exact,
	}
}

func newWhereOptionsWithFromTo(exact bool, from string, to string) *whereOptions {
	return &whereOptions{
		exact:             exact,
		fromParameterName: from,
		toParameterName:   to,
	}
}

type whereToken struct {
	fieldName     string
	whereOperator whereOperator
	parameterName string
	options       *whereOptions
}

func createWhereToken(op whereOperator, fieldName string, parameterName string) *whereToken {
	return createWhereTokenWithOptions(op, fieldName, parameterName, nil)
}

func createWhereTokenWithOptions(op whereOperator, fieldName string, parameterName string, options *whereOptions) *whereToken {
	token := &whereToken{
		fieldName:     fieldName,
		parameterName: parameterName,
		whereOperator: op,
	}
	if options != nil {
		token.options = options
	} else {
		token.options = defaultWhereOptions()
	}
	return token
}

func (t *whereToken) addAlias(alias string) {
	if t.fieldName == "id()" {
		return
	}
	t.fieldName = alias + "." + t.fieldName
}

func (t *whereToken) writeMethod(writer *strings.Builder) (bool, error) {
	if t.options.method != nil {
		switch t.options.method.methodType {
		case MethodsTypeCmpXChg:
			writer.WriteString("cmpxchg(")
		default:
			return false, newIllegalArgumentError("Unsupported method: %s", t.options.method.methodType)
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
		return true, nil
	}

	return false, nil
}

func (t *whereToken) writeTo(writer *strings.Builder) error {
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
	case whereOperatorSearch:
		writer.WriteString("search(")
	case whereOperatorLucene:
		writer.WriteString("lucene(")
	case whereOperatorStartsWith:
		writer.WriteString("startsWith(")
	case whereOperatorEndsWith:
		writer.WriteString("endsWith(")
	case whereOperatorExists:
		writer.WriteString("exists(")
	case whereOperatorSpatialWithin:
		writer.WriteString("spatial.within(")
	case whereOperatorSpatialContains:
		writer.WriteString("spatial.contains(")
	case whereOperatorSpatialDisjoint:
		writer.WriteString("spatial.disjoint(")
	case whereOperatorSpatialIntersects:
		writer.WriteString("spatial.intersects(")
	case whereOperatorRegex:
		writer.WriteString("regex(")
	}

	// TODO: propagate error
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
	return nil
}

func (t *whereToken) writeInnerWhere(writer *strings.Builder) error {

	writeQueryTokenField(writer, t.fieldName)

	switch t.whereOperator {
	case whereOperatorEquals:
		writer.WriteString(" = ")
	case whereOperatorNotEquals:
		writer.WriteString(" != ")
	case whereOperatorGreaterThan:
		writer.WriteString(" > ")
	case whereOperatorGreaterThanOrEqual:
		writer.WriteString(" >= ")
	case whereOperatorLessThan:
		writer.WriteString(" < ")
	case whereOperatorLessThanOrEqual:
		writer.WriteString(" <= ")
	default:
		t.specialOperator(writer)
		return nil
	}

	ok, err := t.writeMethod(writer)
	if err != nil {
		return err
	}
	if !ok {
		writer.WriteString("$")
		writer.WriteString(t.parameterName)
	}
	return nil
}

func (t *whereToken) specialOperator(writer *strings.Builder) {
	options := t.options
	parameterName := t.parameterName
	switch t.whereOperator {
	case whereOperatorIn:
		writer.WriteString(" in ($")
		writer.WriteString(parameterName)
		writer.WriteString(")")
	case whereOperatorAllIn:
		writer.WriteString(" all in ($")
		writer.WriteString(parameterName)
		writer.WriteString(")")
	case whereOperatorBetween:
		writer.WriteString(" between $")
		writer.WriteString(options.fromParameterName)
		writer.WriteString(" and $")
		writer.WriteString(options.toParameterName)
	case whereOperatorSearch:
		writer.WriteString(", $")
		writer.WriteString(parameterName)
		if options.searchOperator == SearchOperatorAnd {
			writer.WriteString(", and")
		}
		writer.WriteString(")")
	case whereOperatorLucene, whereOperatorStartsWith, whereOperatorEndsWith, whereOperatorRegex:
		writer.WriteString(", $")
		writer.WriteString(parameterName)
		writer.WriteString(")")
	case whereOperatorExists:
		writer.WriteString(")")
	case whereOperatorSpatialWithin, whereOperatorSpatialContains, whereOperatorSpatialDisjoint, whereOperatorSpatialIntersects:
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
