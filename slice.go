package ylf

import "fmt"

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