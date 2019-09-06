package ylf

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

/*开启守护进程*/
func Daemon() int {
	cmd := exec.Command(os.Args[0])
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	err := cmd.Start()
	if err == nil {
		cmd.Process.Release()
		os.Exit(0)
	}

	return 0
}

/*结束守护进程*/
func StopDaemon() error {
	args := os.Args
	filePath := args[0]
	_, fileName := filepath.Split(filePath)
	command := "ps -ef|pgrep " + fileName
	cmd := exec.Command("bash", "-c", "kill -9 $("+command+")")
	if _, err := cmd.Output(); err != nil {
		return errors.New("服务已经被停止了")
	}
	return nil
}