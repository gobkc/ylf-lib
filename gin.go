package ylf

import (
	"flag"
	"github.com/gin-gonic/gin"
	"log"
)

func start(route func(router *gin.Engine)) {
	isDaemon := flag.Bool("d", false, "守护进程模式运行程序,例:yunlifang-dial -d");
	stopDaemon := flag.Bool("stop", false, "退出守护模式程序,例:yunlifang-dial -stop");
	flag.Parse()

	/*如果启用了-d参数*/
	if *isDaemon == true {
		Daemon()
	}

	/*如果启用了-stop参数*/
	if *stopDaemon == true {
		if err := StopDaemon(); err != nil {
			log.Fatalln(err.Error())
		} else {
			log.Println("成功退出程序")
			return
		}
	}

	app := gin.Default()
	route(app)

	app.Run()
}
