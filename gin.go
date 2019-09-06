package ylf

import (
	"flag"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func Start(address string, debug string, route func(router *gin.Engine) gin.IRoutes) {
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

	/*默认debug模式,如果debug设置为false,则关闭debug模式*/
	if debug == "false" {
		gin.SetMode(gin.ReleaseMode)
	}

	app := gin.Default()
	router := route(app)
	router.Use(Cors())

	app.Run(address)
}

/*中间件 默认开启*/
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		/*放行所有OPTIONS方法*/
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		/*处理请求*/
		c.Next()
	}
}
