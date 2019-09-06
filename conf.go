package ylf

import (
	"bufio"
	"io"
	"os"
	"reflect"
	"strings"
)

/* demo
type Config struct {
	Server struct {
		Port string
	}
	Mysql struct {
		Host     string
		Port     string
		User     string
		Password string
		Db       string
	}
}

newConf := Config{}
if err:= ylf.NewConfig("./config.conf").Get(&newConf);err!=nil{
	fmt.Println(err.Error())
}

*/

const middle = "========="

type Config struct {
	MyMap    map[string]string
	MyStr    string
	FileName string
}

func NewConfig(fileName string) *Config {
	return &Config{FileName: fileName}
}

func (c *Config) Get(data interface{}) error {
	c.MyMap = make(map[string]string)
	f, err := os.Open(c.FileName)
	if err != nil {
		return err
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	for {
		b, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		s := strings.TrimSpace(string(b))
		if strings.Index(s, "#") == 0 {
			continue
		}

		n1 := strings.Index(s, "[")
		n2 := strings.LastIndex(s, "]")
		if n1 > -1 && n2 > -1 && n2 > n1+1 {
			c.MyStr = strings.TrimSpace(s[n1+1 : n2])
			continue
		}

		if len(c.MyStr) == 0 {
			continue
		}
		index := strings.Index(s, "=")
		if index < 0 {
			continue
		}

		first := strings.TrimSpace(s[:index])
		if len(first) == 0 {
			continue
		}
		second := strings.TrimSpace(s[index+1:])

		pos := strings.Index(second, "\t#")
		if pos > -1 {
			second = second[0:pos]
		}

		pos = strings.Index(second, " #")
		if pos > -1 {
			second = second[0:pos]
		}

		pos = strings.Index(second, "\t//")
		if pos > -1 {
			second = second[0:pos]
		}

		pos = strings.Index(second, " //")
		if pos > -1 {
			second = second[0:pos]
		}

		if len(second) == 0 {
			continue
		}

		key := c.MyStr + middle + first
		c.MyMap[key] = strings.TrimSpace(second)
	}

	/*将c.myMap写入 data interface*/
	dataElem := reflect.ValueOf(data).Elem()
	dataType := dataElem.Type()

	for i := 0; i < dataElem.NumField(); i++ {
		f := dataElem.Field(i)
		ft := f.Kind().String()
		filedName := dataType.Field(i).Name
		if ft=="struct"{
			sDataElem := f
			for si:=0; si<sDataElem.NumField();si++{
				//sf := sDataElem.Field(si)
				//sft := sf.Kind().String()
				sFileName := sDataElem.Type().Field(si).Name
				/*设置值*/
				skey := strings.ToLower(filedName + middle + sFileName)
				dataElem.Field(i).Field(si).SetString(c.MyMap[skey])
			}
		}
	}
	return nil
}

func (c *Config) InitConfig(path string) {
	c.MyMap = make(map[string]string)

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		b, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		s := strings.TrimSpace(string(b))
		//fmt.Println(s)
		if strings.Index(s, "#") == 0 {
			continue
		}

		n1 := strings.Index(s, "[")
		n2 := strings.LastIndex(s, "]")
		if n1 > -1 && n2 > -1 && n2 > n1+1 {
			c.MyStr = strings.TrimSpace(s[n1+1 : n2])
			continue
		}

		if len(c.MyStr) == 0 {
			continue
		}
		index := strings.Index(s, "=")
		if index < 0 {
			continue
		}

		first := strings.TrimSpace(s[:index])
		if len(first) == 0 {
			continue
		}
		second := strings.TrimSpace(s[index+1:])

		pos := strings.Index(second, "\t#")
		if pos > -1 {
			second = second[0:pos]
		}

		pos = strings.Index(second, " #")
		if pos > -1 {
			second = second[0:pos]
		}

		pos = strings.Index(second, "\t//")
		if pos > -1 {
			second = second[0:pos]
		}

		pos = strings.Index(second, " //")
		if pos > -1 {
			second = second[0:pos]
		}

		if len(second) == 0 {
			continue
		}

		key := c.MyStr + middle + first
		c.MyMap[key] = strings.TrimSpace(second)
	}
}

func (c Config) Read(node, key string) string {
	key = node + middle + key
	v, found := c.MyMap[key]
	if !found {
		return ""
	}
	return v
}

func (c Config) ReadList(node, key string) map[string]string {
	returnList := make(map[string]string)
	key = node + middle + key
	nodeLen := len(node)
	stringLen := len(node + middle)
	var tmpNode string
	for k, v := range c.MyMap {
		tmpNode = string([]rune(k)[:nodeLen])
		if tmpNode == node {
			returnList[string([]rune(k)[stringLen:])] = v
		}
	}
	return returnList
}
