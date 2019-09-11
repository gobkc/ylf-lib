package ylf

import (
	"github.com/fsnotify/fsnotify"
	"log"
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
