package ravendb

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"sort"
	"time"
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
	if v == nil {
		io.WriteString(&h._buffer, "null")
		return
	}
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
	case []interface{}:
		for _, v := range v2 {
			h.write(v)
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
	case map[string]interface{}:
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
			v = v2[k]
			io.WriteString(&h._buffer, k)
			if v == nil {
				io.WriteString(&h._buffer, "null")
				return
			}
			tp := reflect.TypeOf(v)
			if isPtrStruct(tp) || tp.Kind() == reflect.Struct {
				// when value of parameter is a struct or pointer to struct
				// it could be our param like SuggestionOptions or
				// param that is custom type used by the user
				s := fmt.Sprintf("%#v", v)
				io.WriteString(&h._buffer, s)
				return
			}
			h.write(v)
		}
	case bool:
		var toWrite int32 = 1
		if v2 {
			toWrite = 2
		}
		must(binary.Write(&h._buffer, binary.LittleEndian, toWrite))
	case time.Time:
		t := v2.UTC().Unix()
		must(binary.Write(&h._buffer, binary.LittleEndian, t))
	case int:
		must(binary.Write(&h._buffer, binary.LittleEndian, int64(v2)))
	default:
		//fmt.Printf("Writing value '%v' of type %T\n", v, v)
		// binary.Write handles all primitive types, except string and int
		must(binary.Write(&h._buffer, binary.LittleEndian, v))
	}
}
