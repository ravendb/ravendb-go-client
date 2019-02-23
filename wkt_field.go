package ravendb

var _ DynamicSpatialField = &WktField{}

type WktField struct {
	wkt string
}

func NewWktField(wkt string) *WktField {
	return &WktField{
		wkt: wkt,
	}
}

func (f *WktField) ToField(ensureValidFieldName func(string, bool) (string, error)) (string, error) {
	name, err := ensureValidFieldName(f.wkt, false)
	if err != nil {
		return "", err
	}
	return "spatial.wkt(" + name + ")", nil
}
