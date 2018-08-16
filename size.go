package ravendb

// Size describes size of entity on disk
type Size struct {
	SizeInBytes int    `json:"SizeInBytes"`
	HumaneSize  string `json:"HumaneSize"`
}

func (s *Size) GetSizeInBytes() int {
	return s.SizeInBytes
}

func (s *Size) GetHumaneSize() string {
	return s.HumaneSize
}

/*
public void setSizeInBytes(long sizeInBytes) {
	this.sizeInBytes = sizeInBytes;
}

public void setHumaneSize(string humaneSize) {
	this.humaneSize = humaneSize;
}
*/
