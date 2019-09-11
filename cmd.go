package ylf

import (
	"errors"
	"fmt"
	"os/exec"
)

func BatchSetMacVlanAndPromisc(cmdString string) error {
	cmd := exec.Command("bash", "-c", "modprobe macvlan && "+cmdString)
	if _, err := cmd.Output(); err != nil {
		return errors.New("启用macvlan和promisc失败,详细原因:" + err.Error())
	}
	return nil
}

func CheckPid(pid string) bool {
	cmdString := fmt.Sprintf("ps -ef|grep %s|grep -v grep|awk '{print $2}'", pid)
	cmd := exec.Command("bash", "-c", cmdString)

	if result, _ := cmd.Output(); string(result) == "" {
		return false
	}

	return true
}
