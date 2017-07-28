package http

import (
	"github.com/OneOfOne/xxhash"
	"fmt"
)

func GetServerHash(url string) string{
	return fmt.Sprintf("%x", xxhash.ChecksumString64(url))
}

func GetServerHashS(url string, database string) string{
	return fmt.Sprintf("%x", xxhash.ChecksumString64S(url, xxhash.ChecksumString64(database)))
}
