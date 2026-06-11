package utils

import (
	"crypto/sha1"
	"fmt"
	"hash/maphash"
	"math/big"

	"github.com/google/uuid"
)

// charSet62 contains the characters used for base62 encoding
const charSet62 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// length defines the length of the generated short code
const length = 6

// base is the base for base62 encoding
var base *big.Int = big.NewInt(62)

// base62 encodes a byte slice into a base62 string
func base62(hash []byte) string {
	num := new(big.Int).SetBytes(hash)
	short := ""
	for num.Sign() > 0 {
		rem := new(big.Int)
		num.DivMod(num, base, rem)
		short = string(charSet62[rem.Int64()]) + short
	}
	return short
}

// UniqueId generates a unique identifier based on the input word and a UUID
func UniqueId(word string) string {
	uid := uuid.New().String()
	seed := maphash.MakeSeed()
	var h maphash.Hash
	h.SetSeed(seed)
	h.WriteString(word)
	return fmt.Sprintf("%d-%s", h.Sum64(), uid)
}

// GetShortUrl generates a short code for a given URL using sha1 and base62 encoding
func GetShortUrl(url string) string {
	hash := sha1.Sum([]byte(url))
	return base62(hash[:length])
}
