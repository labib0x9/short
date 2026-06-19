package utils

import (
	"encoding/base64"

	"github.com/google/uuid"
)

var (
	length = 8
)

// seed = 16bytes, h = 8bytes
func GetShortUrl(url string) string {
	seed := uuid.New()
	h := hash([]byte(url))
	for i := range length {
		pos := uint64(i * 8)
		seed[i] = seed[i] ^ byte(h>>pos)
	}

	urlEncoded := base64.RawURLEncoding.EncodeToString(seed[:])
	return urlEncoded[:length]
}

// fnv64 hashing
func hash(b []byte) uint64 {
	h := uint64(14695981039346656037) // initial
	for _, c := range b {
		h ^= uint64(c)
		h *= 1099511628211 // fnv prime
	}
	return h
}
