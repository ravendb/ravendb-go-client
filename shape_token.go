package ravendb

import "strings"

var _ queryToken = &shapeToken{}

type shapeToken struct {
	shape string
}

func NewShapeToken(shape string) *shapeToken {
	return &shapeToken{
		shape: shape,
	}
}

func ShapeTokenCircle(radiusParameterName string, latitudeParameterName string, longitudeParameterName string, radiusUnits SpatialUnits) *shapeToken {
	if radiusUnits == "" {
		return NewShapeToken("spatial.circle($" + radiusParameterName + ", $" + latitudeParameterName + ", $" + longitudeParameterName + ")")
	}

	if radiusUnits == SpatialUnitsKilometers {
		return NewShapeToken("spatial.circle($" + radiusParameterName + ", $" + latitudeParameterName + ", $" + longitudeParameterName + ", 'Kilometers')")
	}
	return NewShapeToken("spatial.circle($" + radiusParameterName + ", $" + latitudeParameterName + ", $" + longitudeParameterName + ", 'Miles')")
}

func ShapeTokenWkt(shapeWktParameterName string) *shapeToken {
	return NewShapeToken("spatial.wkt($" + shapeWktParameterName + ")")
}

func (t *shapeToken) writeTo(writer *strings.Builder) {
	writer.WriteString(t.shape)
}
