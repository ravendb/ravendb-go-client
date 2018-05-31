package ravendb

// Size describes size of entity on disk
type Size struct {
	SizeInBytes int64  `json:"SizeInBytes"`
	HumaneSize  string `json:"HumaneSize"`
}

/*
public long getSizeInBytes() {
	return sizeInBytes;
}

public void setSizeInBytes(long sizeInBytes) {
	this.sizeInBytes = sizeInBytes;
}

public String getHumaneSize() {
	return humaneSize;
}

public void setHumaneSize(String humaneSize) {
	this.humaneSize = humaneSize;
}
*/
