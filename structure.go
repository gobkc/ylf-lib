package ylf

import "reflect"

func STRUCToString(iFace interface{}) string {
	m := reflect.ValueOf(iFace).Elem()
	var rStr string
	for i := 0; i < m.Len(); i++ {
		rowKeyLen := m.Index(i).NumField()
		for rowKey := 0; rowKey < rowKeyLen; rowKey++ {
			rowV := m.Index(i).Field(rowKey).String()
			rStr = rStr + " " + rowV
		}
		rStr = rStr + "\n"
	}
	return rStr
}
