package ravendb

import "strings"

var _ QueryToken = &OrderByToken{}

var (
	OrderByToken_random          = NewOrderByToken("random()", false, OrderingType_STRING)
	OrderByToken_scoreAscending  = NewOrderByToken("score()", false, OrderingType_STRING)
	OrderByToken_scoreDescending = NewOrderByToken("score()", true, OrderingType_STRING)
)

type OrderByToken struct {
	_fieldName  string
	_descending bool
	_ordering   OrderingType
}

func NewOrderByToken(fieldName string, descending bool, ordering OrderingType) *OrderByToken {
	return &OrderByToken{
		_fieldName:  fieldName,
		_descending: descending,
		_ordering:   ordering,
	}
}

func OrderByToken_createDistanceAscending(fieldName string, latitudeParameterName string, longitudeParameterName string) *OrderByToken {
	return NewOrderByToken("spatial.distance("+fieldName+", spatial.point($"+latitudeParameterName+", $"+longitudeParameterName+"))", false, OrderingType_STRING)
}

func OrderByToken_createDistanceAscending2(fieldName string, shapeWktParameterName string) *OrderByToken {
	return NewOrderByToken("spatial.distance("+fieldName+", spatial.wkt($"+shapeWktParameterName+"))", false, OrderingType_STRING)
}

func OrderByToken_createDistanceDescending(fieldName string, latitudeParameterName string, longitudeParameterName string) *OrderByToken {
	return NewOrderByToken("spatial.distance("+fieldName+", spatial.point($"+latitudeParameterName+", $"+longitudeParameterName+"))", true, OrderingType_STRING)
}

func OrderByToken_createDistanceDescending2(fieldName string, shapeWktParameterName string) *OrderByToken {
	return NewOrderByToken("spatial.distance("+fieldName+", spatial.wkt($"+shapeWktParameterName+"))", true, OrderingType_STRING)
}

func OrderByToken_createRandom(seed string) *OrderByToken {
	if seed == "" {
		panicIf(true, "seed cannot be null")
		// NewIllegalArgumentException("seed cannot be null");
	}
	seed = strings.Replace(seed, "'", "''", -1)
	return NewOrderByToken("random('"+seed+"')", false, OrderingType_STRING)
}

func OrderByToken_createAscending(fieldName string, ordering OrderingType) *OrderByToken {
	return NewOrderByToken(fieldName, false, ordering)
}

func OrderByToken_createDescending(fieldName string, ordering OrderingType) *OrderByToken {
	return NewOrderByToken(fieldName, true, ordering)
}

func (t *OrderByToken) WriteTo(writer *strings.Builder) {
	QueryToken_writeField(writer, t._fieldName)

	switch t._ordering {
	case OrderingType_LONG:
		writer.WriteString(" as long")
		break
	case OrderingType_DOUBLE:
		writer.WriteString(" as double")
		break
	case OrderingType_ALPHA_NUMERIC:
		writer.WriteString(" as alphaNumeric")
		break
	}

	if t._descending { // we only add this if we have to, ASC is the default and reads nicer
		writer.WriteString(" desc")
	}
}
