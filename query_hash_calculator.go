package ravendb

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"sort"
)

type QueryHashCalculator struct {
	_buffer bytes.Buffer
}

func NewQueryHashCalculator() *QueryHashCalculator {
	return &QueryHashCalculator{}
}

func (h *QueryHashCalculator) getHash() string {
	data := h._buffer.Bytes()
	return fmt.Sprintf("%x", md5.Sum(data))
}

func (h *QueryHashCalculator) write(v interface{}) {
	switch v2 := v.(type) {
	case string:
		io.WriteString(&h._buffer, v2)
	case []string:
		if len(v2) == 0 {
			io.WriteString(&h._buffer, "null-list-str")
			return
		}
		binary.Write(&h._buffer, binary.LittleEndian, len(v2))
		for _, s := range v2 {
			io.WriteString(&h._buffer, s)
		}
	case map[string]string:
		if len(v2) == 0 {
			io.WriteString(&h._buffer, "null-dic<string,string>")
			return
		}
		binary.Write(&h._buffer, binary.LittleEndian, len(v2))
		// in Go iteration over map is not stable, so need to manually sort keys
		var keys []string
		for k := range v2 {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := v2[k]
			io.WriteString(&h._buffer, k)
			io.WriteString(&h._buffer, v)
		}
	case map[string]Object:
		// this is Parameters
		if len(v2) == 0 {
			io.WriteString(&h._buffer, "null-dic<string,object>")
			return
		}
		binary.Write(&h._buffer, binary.LittleEndian, len(v2))
		// in Go iteration over map is not stable, so need to manually sort keys
		var keys []string
		for k := range v2 {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := v2[k]
			io.WriteString(&h._buffer, k)
			h.write(v)
		}
	default:
		// binary.Write handles all primitive types, except string
		binary.Write(&h._buffer, binary.LittleEndian, v)
	}
}
