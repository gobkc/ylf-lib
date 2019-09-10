package ylf

import (
	"github.com/satori/go.uuid"
	"math/rand"
	"strconv"
	"time"
)

func Uuid() string {
	return uuid.NewV4().String()
}

func  GetRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func Rand() string {
	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(10000)
	return GetRandomString(4) + strconv.Itoa(num)
}
