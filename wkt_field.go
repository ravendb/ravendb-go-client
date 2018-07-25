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

func (f *WktField) toField(ensureValidFieldName func(string, bool) string) string {
	return "spatial.wkt(" + ensureValidFieldName(f.wkt, false) + ")"
}
