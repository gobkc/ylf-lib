package ylf

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"syscall"
	"time"
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


type RunFunc func(pid int, content string)

func RunCmd(cmdString string, runFc RunFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "bash", "-c", cmdString)
	cmd.SysProcAttr = &syscall.SysProcAttr{}

	/*标准输出绑定到buf*/
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	cmd.Start()

	go readBuf(&buf, cmd.Process.Pid, &cancel, runFc)
	cmd.Wait()
	if !checkPid(cmd.Process.Pid){
		runFc(cmd.Process.Pid,buf.String())
	}
}

func readBuf(buff *bytes.Buffer, pid int, cancelFunc *context.CancelFunc, runFc RunFunc)  {
	tick := time.NewTicker(1000 * time.Millisecond)
Loop:
	for {
		select {
		case <-tick.C:
			if checkPid(pid) {
				fmt.Println("Pid:", pid)
				fmt.Println(buff.String())
				runFc(pid, buff.String())
				buff.Reset()
			} else {
				break Loop
			}
		}
	}
}

func checkPid(pid int) bool {
	cmdString := fmt.Sprintf("ps -ef|grep %v|grep -v grep|awk '{print $2}'", pid)
	cmd := exec.Command("bash", "-c", cmdString)

	if result, _ := cmd.Output(); string(result) == "" {
		return false
	}

	return true
}
