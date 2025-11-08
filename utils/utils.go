package utils

import (
	"math/rand"
	"time"
)

const CHARACTER_SET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456788"

func GenerateRandomString(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = CHARACTER_SET[seededRand.Intn(len(CHARACTER_SET))]
	}
	return string(b)
}
