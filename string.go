package ylf

import (
	"github.com/satori/go.uuid"
	"math/rand"
	"regexp"
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

func ExpFind(rule string, srcString string) string {
	re := regexp.MustCompile(rule)
	find := re.FindStringSubmatch(srcString)
	if find != nil && len(find) > 1 {
		return find[1]
	}
	return ""
}

func ExpFindLast(rule string, srcString string) string {
	re,_ := regexp.CompilePOSIX(rule)
	find := re.FindAllStringSubmatch(srcString,-1)
	if findLen := len(find);findLen>=1{
		last := find[findLen-1]
		if lastLen := len(last);lastLen>=1{
			return last[lastLen-1]
		}
	}
	return ""
}