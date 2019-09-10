package ylf

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
)

/*此包用于读取/设置宽带帐号密码*/
/** demo
myConfig := new(lib.ConfFile)
myConfig.InitFile("./config.conf")
account :=myConfig.ReadAccount("07557550897705@163.gd")
myConfig.WriteAccount("07557550897705@163.gd","testPP")
fmt.Println(account)
*/
type ConfFileMap map[string]string
type ConfFile struct {
	FileName      string
	ConfFileArray []ConfFileMap
}

/*初始化并把配置文件读取到结构体内*/
func (c *ConfFile) InitFile(path string) {
	/*初始化的时候保存文件名到结构体*/
	c.FileName = path

	/*打开文件,打开不了就创建*/
	f, err := os.Open(path)
	if err != nil {
		/*重新创建文件*/
		f, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			panic(err)
		}
	}
	defer f.Close()

	r := bufio.NewReader(f)

	keys := []string{"account", "password"}
	for {
		b, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println(err)
			return
		}
		s := strings.TrimSpace(string(b))
		hideLine := string([]byte(s)[:1])
		if hideLine == "#" || s == "" {
			continue
		}
		reg := regexp.MustCompile(`[a-zA-Z0-9\@\.]+`)
		splitArr := reg.FindAllString(s, -1)
		if splitArr != nil && keys != nil {
			newMap := c.sliceToMap(keys, splitArr)
			c.ConfFileArray = append(c.ConfFileArray, newMap)
		}
	}
}

/*切片转MAP,第一个参数指定keys*/
func (c *ConfFile) sliceToMap(sKey, sValue []string) (map[string]string) {
	mapObj := map[string]string{}
	for sKeyIndex, _ := range sKey {
		if sValue != nil && sKey != nil && len(sValue) > sKeyIndex {
			mapObj[sKey[sKeyIndex]] = sValue[sKeyIndex]
		}
	}
	return mapObj
}

/*读取某个指定账户*/
func (c *ConfFile) ReadAccount(account string) map[string]string {
	for _, v := range c.ConfFileArray {
		if v["account"] == account {
			return v
		}
	}
	return nil
}

/*写入某个值到指定账户*/
func (c *ConfFile) WriteAccount(account string, password string) bool {
	newAccount := ConfFileMap{
		"account":  account,
		"password": password,
	}
	/*1.遍历账户列表,如果存在此账户,则替换*/
	hasAccount := false
	for i, v := range c.ConfFileArray {
		if v["account"] == account {
			c.ConfFileArray[i]["password"] = password
			hasAccount = true
			break
		}
	}

	/*2.如果不存在此账户,则将此账户添加到map数组*/
	if hasAccount == false {
		c.ConfFileArray = append(c.ConfFileArray, newAccount)
	}

	/*生成配置文件指定数据*/
	var confText string
	for _, v := range c.ConfFileArray {
		confText = confText + fmt.Sprintf("\"%s\" * \"%s\"\n", v["account"], v["password"])
	}

	f, err := os.Create(c.FileName)

	/*如果无法打开,则重新创建*/
	if err != nil {
		return false
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	_, err = w.WriteString(confText)
	if err != nil {
		return false
	} else {
		w.Flush()
		f.Close()
		return true
	}
}

/*批量写入账户*/
func (c *ConfFile) BatchWriteAccount(account interface{}) bool {
	m := reflect.ValueOf(account).Elem()
	for i := 0; i < m.Len(); i++ {
		account := m.Index(i).FieldByName("Account").String()
		password := m.Index(i).FieldByName("Password").String()
		hasAccount := false

		for si, sv := range c.ConfFileArray {
			if sv["account"] == account {
				c.ConfFileArray[si]["password"] = password
				hasAccount = true
				break
			}
		}

		if hasAccount == false {
			c.ConfFileArray = append(c.ConfFileArray, ConfFileMap{
				"account":  account,
				"password": password,
			})
		}

	}

	var confText string
	for _, v := range c.ConfFileArray {
		confText = confText + fmt.Sprintf("\"%s\" * \"%s\"\n", v["account"], v["password"])
	}

	f, err := os.Create(c.FileName)

	/*如果无法打开,则重新创建*/
	if err != nil {
		return false
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	_, err = w.WriteString(confText)
	if err != nil {
		return false
	} else {
		w.Flush()
		f.Close()
		return true
	}
}