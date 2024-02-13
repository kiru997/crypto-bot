package helper

import (
	"strconv"

	"github.com/jaevor/go-nanoid"
)

func RandomNumber(len int) int64 {
	decenaryID, _ := nanoid.CustomASCII("0123456789", len)
	i, _ := strconv.ParseInt(decenaryID(), 10, 64)
	return i
}
