package ylf

import (
	"errors"
	"os/exec"
)

func BatchSetMacVlanAndPromisc(cmdString string) error {
	cmd := exec.Command("bash", "-c", "modprobe macvlan && "+cmdString)
	if _, err := cmd.Output(); err != nil {
		return errors.New("启用macvlan和promisc失败,详细原因:" + err.Error())
	}
	return nil
}
