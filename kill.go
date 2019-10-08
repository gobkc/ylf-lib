package ylf

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
)

func KillPid(pid string) error {
	cli := exec.Command("bash", "-c", "kill -9 "+pid)
	if _, err := cli.Output(); err != nil {
		return errors.New(pid + "进程不存在或已经被杀掉了")
	}
	return nil
}


func CtrlC(fc func(signalTag string)) {
	var wg sync.WaitGroup
	/*waitGroup内部计数器+1,为0都时wait阻塞会释放，必须在goroutine执行之前执行你该*/
	wg.Add(1)
	/*将键盘输入信号转发到osSignal*/
	osSignal := make(chan os.Signal)
	signal.Notify(osSignal)
	/*goroutine开始*/
	/*goroutine结束*/
	select {
	case sig := <-osSignal:
		fc(fmt.Sprintf("%s", sig))
		os.Exit(1)
	}
	wg.Wait()
}
