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