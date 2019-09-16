package ylf

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"text/template"
	"time"
)

type WhiteListStructure struct {
	ClientPass string
}

type BlackListStructure struct {
	ClientBlock string
}

type DanteStructure struct {
	TemplatePath     string `description:模板路径`
	SavePath         string `description:设置的配置文件路径`
	Logoutput        string
	Privileged       string
	Unprivileged     string
	Interal          string
	Port             string
	External         string
	SocksMethod      string
	ClientMethod     string
	ClientPassSlice  []WhiteListStructure `description:白名单`
	ClientBlockSlice []BlackListStructure `description:黑名单`
}

/*初始化dante类,可以选择传递templatePath,也可以选择不传递*/
func NewDante(templatePath ...string) *DanteStructure {
	var realTempPath string

	if templatePath != nil && templatePath[0] != "" {
		realTempPath = templatePath[0]
	} else {
		currentDirectory := filepath.Dir(CurrentFile())
		realTempPath = currentDirectory + "/dante-template/dante.conf"
	}

	/*初始化的时候提供默认值*/
	return &DanteStructure{
		realTempPath,
		"./account-conf",
		"",
		"root",
		"nobody",
		"",
		"1080",
		"ppp0",
		"none",
		"none",
		nil,
		nil,
	}
}

/*返回配置文件的所有字符串*/
func (d *DanteStructure) GetContent() (string, error) {
	buf, err := ioutil.ReadFile(d.TemplatePath)

	if err != nil {
		return "", errors.New("在读取dante配置文件时发生错误" + err.Error())
	}

	content := string(buf)
	return content, nil
}

/*保存设置*/
func (d *DanteStructure) Save(account ...string) error {
	/*0.获取用户账户*/
	var userAccount string

	if account != nil && account[0] != "" {
		userAccount = account[0]
	} else {
		return errors.New("保存dante设置时没有指定账户名称")
	}

	/*1.获取dante模板文件*/
	danteTemplate, err := d.GetContent()
	if err != nil {
		return err
	}

	/*2.替换变量为结构体内容*/
	tempParse, _ := template.New("dante template").Parse(danteTemplate)
	var buf bytes.Buffer
	tempParse.Execute(&buf, d)
	result := buf.String()

	/*保存*/
	savePath := fmt.Sprintf("%s/%s.conf", d.SavePath, userAccount)
	if err := ioutil.WriteFile(savePath, []byte(result), 0); err != nil {
		return errors.New("在保存用户配置文件时发生错误:" + err.Error())
	}

	return nil
}

/*设置用户配置文件保存路径*/
func (d *DanteStructure) SetSavePath(path string) *DanteStructure {
	d.SavePath = path
	return d
}

/*设置注销参数*/
func (d *DanteStructure) SetLogoutput(logoutput string) *DanteStructure {
	d.Logoutput = logoutput
	return d
}

/*设置特权账户,如:root*/
func (d *DanteStructure) SetPrivileged(privileged string) *DanteStructure {
	d.Privileged = privileged
	return d
}

/*设置没有特权的账户,如:nobody*/
func (d *DanteStructure) SetUnprivileged(unPrivileged string) *DanteStructure {
	d.Unprivileged = unPrivileged
	return d
}

/*设置拨号成功后的内部IP*/
func (d *DanteStructure) SetInteral(interal string) *DanteStructure {
	d.Interal = interal
	return d
}

/*设置拨号成功后的Port*/
func (d *DanteStructure) SetPort(port string) *DanteStructure {
	d.Port = port
	return d
}

/*设置外部IP或则对外的网卡*/
func (d *DanteStructure) SetExternal(external string) *DanteStructure {
	d.External = external
	return d
}

/*设置socket对外访问限制方法*/
func (d *DanteStructure) SetSocksMethod(socksMethod string) *DanteStructure {
	d.SocksMethod = socksMethod
	return d
}

/*设置客户端访问方法*/
func (d *DanteStructure) SetClientMethod(clientMethod string) *DanteStructure {
	d.ClientMethod = clientMethod
	return d
}

/*设置黑名单*/
func (d *DanteStructure) SetBlackIp(ip ...string) *DanteStructure {
	if ip != nil {
		for _, v := range ip {
			d.ClientBlockSlice = append(d.ClientBlockSlice, BlackListStructure{
				fmt.Sprintf("from: %s to: 0.0.0.0/0", v),
			})
		}
	}
	return d
}

/*限制用户访问特定的IP*/
func (d *DanteStructure) SetBlackLimit(fromIp string, toIp string) *DanteStructure {
	d.ClientBlockSlice = append(d.ClientBlockSlice, BlackListStructure{
		fmt.Sprintf("from: %s to: %s", fromIp, toIp),
	})
	return d
}

/*设置白名单*/
func (d *DanteStructure) SetWhiteIp(ip ...string) *DanteStructure {
	if ip != nil {
		for _, v := range ip {
			d.ClientPassSlice = append(d.ClientPassSlice, WhiteListStructure{
				fmt.Sprintf("from: %s to: 0.0.0.0/0", v),
			})
		}
	}
	return d
}

/*允许用户访问特定的IP*/
func (d *DanteStructure) SetWhiteLimit(fromIp string, toIp string) *DanteStructure {
	d.ClientPassSlice = append(d.ClientPassSlice, WhiteListStructure{
		fmt.Sprintf("from: %s to: %s", fromIp, toIp),
	})
	return d
}

/*替换文件内容 暂时不用  预留*/
func replaceFileContent(filePath string, findStr string, replaceStr string) {
	buf, _ := ioutil.ReadFile(filePath)
	content := string(buf)

	//替换
	newContent := strings.Replace(content, findStr, replaceStr, -1)

	//重新写入
	ioutil.WriteFile(filePath, []byte(newContent), 0)
}

/*启动dante服务*/
func (d *DanteStructure) StartDante(fileName interface{}, times float64, fc func(userId interface{}, status interface{}, resetUserId interface{}) error) *DanteStructure {
	filePath := fmt.Sprintf("%s/%s/%v.conf", d.getCurrentDirectory(), d.SavePath, fileName)
	cmdString := fmt.Sprintf("sockd -f %s", filePath)
	log.Println("开始启动dante:", cmdString)

	ctx, cancel := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, "bash", "-c", cmdString)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	var buf bytes.Buffer

	/*标准输出绑定到buf*/
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	cmd.Start()
	fmt.Println("获取到当前子线程的PID:", cmd.Process.Pid)

	time.Sleep(time.Duration(times) * time.Minute)

	/*停止进程*/
	cancel()

	/*停止子线程*/
	d.StopDante(fileName)

	/*释放数据库中的资源*/
	if err := fc(fileName, 1, 0); err != nil {
		log.Println(err.Error())
	}

	cmd.Wait()
	return d
}

func (d *DanteStructure) StopDante(configFileName interface{}) error {
	cli := exec.Command("bash", "-c", fmt.Sprintf("kill -9 $(ps -ef|grep sockd|grep  %v.conf|awk '{print $2}')", configFileName))
	if _, err := cli.Output(); err != nil {
		return errors.New("服务已经被停止了")
	}
	log.Println("停止了进程")
	return nil
}

func (d *DanteStructure) getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func CurrentFile() string {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		log.Println("Can not get current file info")
	}
	return file
}
