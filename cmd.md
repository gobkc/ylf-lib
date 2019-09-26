	RunCmd(fmt.Sprintf("ping -w 3 %s", "www.baidu.com"), func(pid int, content string) {
		fmt.Printf("获取到的pid%v,获取到的content%s\n", pid, content)
	})
