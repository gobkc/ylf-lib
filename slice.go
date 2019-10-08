package ylf

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
)

func ArrayMap(data interface{}, fc func(i int, row interface{}) interface{}) error {
	dataElem := reflect.ValueOf(data).Elem()
	for i := 0; i < dataElem.Len(); i++ {
		currentRow := dataElem.Index(i)
		currentRowData := currentRow.Interface()
		callBackRow := fc(i, currentRowData)
		callBackRowData := reflect.ValueOf(callBackRow)
		currentRow.Set(callBackRowData)
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(dataElem.Interface()); err != nil {
		return err
	}
	gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(data)
	return nil
}

func SliceMap(slice []interface{}, cb func(interface{}) interface{}) []interface{} {
	dup := make([]interface{}, len(slice))

	copy(dup, slice)

	for index, val := range slice {
		dup[index] = cb(val)
	}

	return dup
}

func InSlice(slice []interface{}, value interface{}) bool {
	for _, row := range slice {
		realRow := fmt.Sprintf("%v", row)
		realValue := fmt.Sprintf("%v", value)
		if realRow == realValue {
			return true
		}
	}
	return false
}

func RemoveRepeatedElement(arr []string) (newArr []string) {
	newArr = make([]string, 0)
	for i := 0; i < len(arr); i++ {
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			if arr[i] == arr[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, arr[i])
		}
	}
	return
}

/*效率更高的去重函数 用法：RemoveRepeated(&noPassAccount)*/
func RemoveRepeated(arr *[]string){
	var newArr []string
	newArr = make([]string, 0)
	for i := 0; i < len(*arr); i++ {
		repeat := false
		for j := i + 1; j < len(*arr); j++ {
			if (*arr)[i] == (*arr)[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, (*arr)[i])
		}
	}
	if newArr!=nil{
		*arr = newArr
	}
	return
}