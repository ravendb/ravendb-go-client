package ravendb

import "strings"

var _ queryToken = &orderByToken{}

var (
	orderByTokenRandom          = newOrderByToken("random()", false, OrderingTypeString)
	orderByTokenScoreAscending  = newOrderByToken("score()", false, OrderingTypeString)
	orderByTokenScoreDescending = newOrderByToken("score()", true, OrderingTypeString)
)

type orderByToken struct {
	fieldName  string
	descending bool
	ordering   OrderingType
}

func newOrderByToken(fieldName string, descending bool, ordering OrderingType) *orderByToken {
	return &orderByToken{
		fieldName:  fieldName,
		descending: descending,
		ordering:   ordering,
	}
}

func orderByTokenCreateDistanceAscending(fieldName string, latitudeParameterName string, longitudeParameterName string) *orderByToken {
	return newOrderByToken("spatial.distance("+fieldName+", spatial.point($"+latitudeParameterName+", $"+longitudeParameterName+"))", false, OrderingTypeString)
}

func orderByTokenCreateDistanceAscending2(fieldName string, shapeWktParameterName string) *orderByToken {
	return newOrderByToken("spatial.distance("+fieldName+", spatial.wkt($"+shapeWktParameterName+"))", false, OrderingTypeString)
}

func orderByTokenCreateDistanceDescending(fieldName string, latitudeParameterName string, longitudeParameterName string) *orderByToken {
	return newOrderByToken("spatial.distance("+fieldName+", spatial.point($"+latitudeParameterName+", $"+longitudeParameterName+"))", true, OrderingTypeString)
}

func orderByTokenCreateDistanceDescending2(fieldName string, shapeWktParameterName string) *orderByToken {
	return newOrderByToken("spatial.distance("+fieldName+", spatial.wkt($"+shapeWktParameterName+"))", true, OrderingTypeString)
}

func orderByTokenCreateRandom(seed string) *orderByToken {
	if seed == "" {
		panicIf(true, "seed cannot be null")
		// newIllegalArgumentError("seed cannot be null");
	}
	seed = strings.Replace(seed, "'", "''", -1)
	return newOrderByToken("random('"+seed+"')", false, OrderingTypeString)
}

func orderByTokenCreateAscending(fieldName string, ordering OrderingType) *orderByToken {
	return newOrderByToken(fieldName, false, ordering)
}

func orderByTokenCreateDescending(fieldName string, ordering OrderingType) *orderByToken {
	return newOrderByToken(fieldName, true, ordering)
}

func (t *orderByToken) writeTo(writer *strings.Builder) error {
	writeQueryTokenField(writer, t.fieldName)

	switch t.ordering {
	case OrderingTypeLong:
		writer.WriteString(" as long")
	case OrderingTypeDouble:
		writer.WriteString(" as double")
	case OrderingTypeAlphaNumeric:
		writer.WriteString(" as alphaNumeric")
	}

	if t.descending { // we only add this if we have to, ASC is the default and reads nicer
		writer.WriteString(" desc")
	}
	return nil
}
