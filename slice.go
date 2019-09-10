package ylf

func SliceMap(slice []interface{}, cb func(interface{}) interface{}) []interface{} {
	dup := make([]interface{}, len(slice))

	copy(dup, slice)

	for index, val := range slice {
		dup[index] = cb(val)
	}

	return dup
}
