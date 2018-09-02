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

func ShapeToken_circle(radiusParameterName string, latitudeParameterName string, longitudeParameterName string, radiusUnits SpatialUnits) *shapeToken {
	if radiusUnits == "" {
		return NewShapeToken("spatial.circle($" + radiusParameterName + ", $" + latitudeParameterName + ", $" + longitudeParameterName + ")")
	}

	if radiusUnits == SpatialUnits_KILOMETERS {
		return NewShapeToken("spatial.circle($" + radiusParameterName + ", $" + latitudeParameterName + ", $" + longitudeParameterName + ", 'Kilometers')")
	}
	return NewShapeToken("spatial.circle($" + radiusParameterName + ", $" + latitudeParameterName + ", $" + longitudeParameterName + ", 'Miles')")
}

func ShapeToken_wkt(shapeWktParameterName string) *shapeToken {
	return NewShapeToken("spatial.wkt($" + shapeWktParameterName + ")")
}

func (t *shapeToken) writeTo(writer *strings.Builder) {
	writer.WriteString(t.shape)
}
