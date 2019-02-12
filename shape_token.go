package ravendb

import "strings"

var _ queryToken = &shapeToken{}

type shapeToken struct {
	shape string
}

func shapeTokenCircle(radiusParameterName string, latitudeParameterName string, longitudeParameterName string, radiusUnits SpatialUnits) *shapeToken {
	if radiusUnits == "" {
		shape := "spatial.circle($" + radiusParameterName + ", $" + latitudeParameterName + ", $" + longitudeParameterName + ")"
		return &shapeToken{shape: shape}
	}

	if radiusUnits == SpatialUnitsKilometers {
		shape := "spatial.circle($" + radiusParameterName + ", $" + latitudeParameterName + ", $" + longitudeParameterName + ", 'Kilometers')"
		return &shapeToken{shape: shape}
	}
	shape := "spatial.circle($" + radiusParameterName + ", $" + latitudeParameterName + ", $" + longitudeParameterName + ", 'Miles')"
	return &shapeToken{shape: shape}
}

func shapeTokenWkt(shapeWktParameterName string) *shapeToken {
	shape := "spatial.wkt($" + shapeWktParameterName + ")"
	return &shapeToken{shape: shape}
}

func (t *shapeToken) writeTo(writer *strings.Builder) error {
	writer.WriteString(t.shape)
	return nil
}
