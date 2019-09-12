package ylf

import (
	"bufio"
	"errors"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type RtTables struct {
	TableId   string
	TableName string
}

func GetRtTableAll() ([]RtTables, error) {
	rtTablePath := "/etc/iproute2/rt_tables"
	var rtTable []RtTables
	var rtTableString string
	if rtTableBytes, err := ioutil.ReadFile(rtTablePath); err != nil {
		return nil, err
	} else {
		rtTableString = string(rtTableBytes)
	}

	reg := regexp.MustCompile(`[0-9]+ [a-zA-Z0-9\@\.\_]+`)
	splitArr := reg.FindAllString(rtTableString, -1)
	if splitArr != nil {
		for _, v := range splitArr {
			splitRow := strings.Split(v, " ")
			if len(splitRow) < 2 {
				continue
			}
			rtTable = append(rtTable, RtTables{
				TableId:   splitRow[0],
				TableName: splitRow[1],
			})
		}
	}
	return rtTable, nil
}

/*同步RT TABLE设置 存在则更新，不存在则新增*/
func SynRtTableItem(tableName ...string) ([]RtTables, error) {
	rtTables, err := GetRtTableAll()
	if err != nil {
		return nil, err
	}

	/*建立 rtTable的map映射*/
	rtMap := make(map[int]string)
	for i := 0; i <= 255; i++ {
		rtMap[i] = ""
	}

	/*将/etc/iproute2/rt_tables已有的数据写入map映射*/
	for _, v := range rtTables {
		tableId, _ := strconv.Atoi(v.TableId)
		rtMap[tableId] = v.TableName
	}

	/*申请可用地址*/
	for _, rows := range tableName {
		if rows == "" {
			continue
		}
		haveIndex := -1
		canUseIndex := -1
		for mk, mv := range rtMap {
			if mv == rows {
				haveIndex = mk
			} else if mv == "" {
				canUseIndex = mk
			}
		}
		if haveIndex != -1 {
			rtMap[haveIndex] = rows
		} else if canUseIndex != -1 {
			rtMap[canUseIndex] = rows
		} else {
			return nil, errors.New("rt tables已经没有足够的可使用地址")
		}
	}

	/*map转结构体*/
	rtTables = []RtTables{}
	for k, v := range rtMap {
		if v != "" {
			rtTables = append(rtTables, RtTables{
				TableId:   strconv.Itoa(k),
				TableName: v,
			})
		}
	}

	if err := WriteRtTable(rtTables); err != nil {
		return rtTables, err
	}

	return rtTables, nil
}

func DeleteRtTableItem(tableName ...string) ([]RtTables, error) {
	rtTables, err := GetRtTableAll()
	if err != nil {
		return nil, err
	}
	/*删除对应rtTable中的元素*/
	for _, tableItem := range tableName {
		for i, v := range rtTables {
			if v.TableName == tableItem {
				rtTables = append(rtTables[:i], rtTables[i+1:]...)
				break
			}
		}
	}
	if err := WriteRtTable(rtTables); err != nil {
		return rtTables, err
	}
	return rtTables, nil
}

func WriteRtTable(rtTables [] RtTables) error {
	strData := STRUCToString(&rtTables)

	rtTablePath := "/etc/iproute2/rt_tables"
	f, err := os.Create(rtTablePath)
	if err != nil {
		return errors.New("在打开rt_table表时发生错误:" + err.Error())
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	_, err = w.WriteString(strData)
	if err != nil {
		return errors.New("在写入rt_table表时发生错误:" + err.Error())
	}
	w.Flush()

	return nil
}