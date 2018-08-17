package ravendb

var _ DynamicSpatialField = &PointField{}

type PointField struct {
	latitude  string
	longitude string
}

func NewPointField(latitude string, longitude string) *PointField {
	return &PointField{
		latitude:  latitude,
		longitude: longitude,
	}
}

func (f *PointField) ToField(ensureValidFieldName func(string, bool) string) string {
	return "spatial.point(" + ensureValidFieldName(f.latitude, false) + ", " + ensureValidFieldName(f.longitude, false) + ")"
}
