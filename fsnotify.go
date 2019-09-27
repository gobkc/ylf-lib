package ylf

import (
	"errors"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"log"
	"path/filepath"
)

type FsNotifyInterface interface {
	WriteEven(fileName string)
	CreateEven(fileName string)
	DeleteEven(fileName string)
	RenameEven(fileName string)
	ChmodEven(fileName string)
}

func StartWatch(path string, fs FsNotifyInterface) {
	watch, err := fsnotify.NewWatcher();
	if err != nil {
		log.Fatal(err);
	}

	go func() {
		for {
			select {
			case ev := <-watch.Events:
				{
					//判断事件发生的类型，如下5种
					// Create 创建
					// Write 写入
					// Remove 删除
					// Rename 重命名
					// Chmod 修改权限
					if ev.Op&fsnotify.Create == fsnotify.Create {
						fs.CreateEven(ev.Name)
					}
					if ev.Op&fsnotify.Write == fsnotify.Write {
						fs.WriteEven(ev.Name)
					}
					if ev.Op&fsnotify.Remove == fsnotify.Remove {
						fs.DeleteEven(ev.Name)
					}
					if ev.Op&fsnotify.Rename == fsnotify.Rename {
						fs.RenameEven(ev.Name)
					}
					if ev.Op&fsnotify.Chmod == fsnotify.Chmod {
						fs.ChmodEven(ev.Name)
					}
				}
			case err := <-watch.Errors:
				{
					log.Println("error : ", err);
					return
				}
			}
		}
	}()

	go func(path string) {
		defer watch.Close()
		//添加要监控的对象，文件或文件夹
		err = watch.Add(path)
		if err != nil {
			log.Fatal(err)
		}
		select {}
	}(path)
}


type PppLogInfo struct {
	PppError string
	Ppp      string
	Local    string
	Remote   string
	Dns1     string
	Dns2     string
}

func GetLogIp(path string) (result PppLogInfo, err error) {
	fileC, err := ioutil.ReadFile(path)
	if err != nil {
		return result, err
	}

	f := string(fileC)

	result.Ppp = ExpFind(`interface (.*)\n`, f)
	result.Local = ExpFind(`local  IP address (.*)\n`, f)
	result.Remote = ExpFind(`remote IP address (.*)\n`, f)
	result.Dns1 = ExpFind(`primary   DNS address (.*)\n`, f)
	result.Dns2 = ExpFind(`secondary DNS address (.*)\n*`, f)
	result.PppError = ExpFind(`Connect.*\n([^\f]*)Connection`, f)
	if result.PppError != "" {
		err = errors.New(result.PppError)
	}
	return result, err
}

func GetLogIp2(f string) (result PppLogInfo, err error) {
	result.Ppp = ExpFindLast(`interface (.*)\n`, f)
	result.Local = ExpFindLast(`local  IP address (.*)\n`, f)
	result.Remote = ExpFindLast(`remote IP address (.*)\n`, f)
	result.Dns1 = ExpFindLast(`primary   DNS address (.*)\n`, f)
	result.Dns2 = ExpFindLast(`secondary DNS address (.*)\n*`, f)
	result.PppError = ExpFindLast(`Connect.*\n([^\f]*)Connection`, f)
	if result.PppError != "" {
		err = errors.New(result.PppError)
	}
	return result, err
}

func GetFileByPath(path string) string {
	fullFileName := filepath.Base(path)
	if len(fullFileName) > 4 {
		return string([]byte(fullFileName)[:len(fullFileName)-4])
	}
	return ""
}
