package http

import (
	"fmt"
	"github.com/OneOfOne/xxhash"
)

func GetServerHash(url string) string {
	return fmt.Sprintf("%x", xxhash.ChecksumString64(url))
}

func GetServerHashWithSeed(url string, database string) string {
	return fmt.Sprintf("%x", xxhash.ChecksumString64S(url, xxhash.ChecksumString64(database)))
}
