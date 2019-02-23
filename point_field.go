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

func (f *PointField) ToField(ensureValidFieldName func(string, bool) (string, error)) (string, error) {
	name1, err := ensureValidFieldName(f.latitude, false)
	if err != nil {
		return "", err
	}
	name2, err := ensureValidFieldName(f.longitude, false)
	if err != nil {
		return "", err
	}
	return "spatial.point(" + name1 + ", " + name2 + ")", nil
}
