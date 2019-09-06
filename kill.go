package ylf

import (
	"errors"
	"os/exec"
)

func KillPid(pid string) error {
	cli := exec.Command("bash", "-c", "kill -9 "+pid)
	if _, err := cli.Output(); err != nil {
		return errors.New(pid + "进程不存在或已经被杀掉了")
	}
	return nil
}
