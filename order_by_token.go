package ravendb

import "strings"

var _ queryToken = &orderByToken{}

var (
	OrderByToken_random          = newOrderByToken("random()", false, OrderingType_STRING)
	OrderByToken_scoreAscending  = newOrderByToken("score()", false, OrderingType_STRING)
	OrderByToken_scoreDescending = newOrderByToken("score()", true, OrderingType_STRING)
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

func OrderByToken_createDistanceAscending(fieldName string, latitudeParameterName string, longitudeParameterName string) *orderByToken {
	return newOrderByToken("spatial.distance("+fieldName+", spatial.point($"+latitudeParameterName+", $"+longitudeParameterName+"))", false, OrderingType_STRING)
}

func OrderByToken_createDistanceAscending2(fieldName string, shapeWktParameterName string) *orderByToken {
	return newOrderByToken("spatial.distance("+fieldName+", spatial.wkt($"+shapeWktParameterName+"))", false, OrderingType_STRING)
}

func OrderByToken_createDistanceDescending(fieldName string, latitudeParameterName string, longitudeParameterName string) *orderByToken {
	return newOrderByToken("spatial.distance("+fieldName+", spatial.point($"+latitudeParameterName+", $"+longitudeParameterName+"))", true, OrderingType_STRING)
}

func OrderByToken_createDistanceDescending2(fieldName string, shapeWktParameterName string) *orderByToken {
	return newOrderByToken("spatial.distance("+fieldName+", spatial.wkt($"+shapeWktParameterName+"))", true, OrderingType_STRING)
}

func OrderByToken_createRandom(seed string) *orderByToken {
	if seed == "" {
		panicIf(true, "seed cannot be null")
		// newIllegalArgumentError("seed cannot be null");
	}
	seed = strings.Replace(seed, "'", "''", -1)
	return newOrderByToken("random('"+seed+"')", false, OrderingType_STRING)
}

func OrderByToken_createAscending(fieldName string, ordering OrderingType) *orderByToken {
	return newOrderByToken(fieldName, false, ordering)
}

func OrderByToken_createDescending(fieldName string, ordering OrderingType) *orderByToken {
	return newOrderByToken(fieldName, true, ordering)
}

func (t *orderByToken) writeTo(writer *strings.Builder) {
	writeQueryTokenField(writer, t.fieldName)

	switch t.ordering {
	case OrderingType_LONG:
		writer.WriteString(" as long")
	case OrderingType_DOUBLE:
		writer.WriteString(" as double")
	case OrderingType_ALPHA_NUMERIC:
		writer.WriteString(" as alphaNumeric")
	}

	if t.descending { // we only add this if we have to, ASC is the default and reads nicer
		writer.WriteString(" desc")
	}
}
