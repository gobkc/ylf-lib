链接MYSQL

    var Mysql *MysqlStructure

	mysqlConf := &Conf.Mysql
	Mysql = NewMysql(mysqlConf.Host, mysqlConf.Port, mysqlConf.User, mysqlConf.Password, mysqlConf.Db)

查询单条记录

	adslAccount := model.AdslAccount{}
	Mysql.Sql("SELECT * FROM `adsl_account`").Find(&adslAccount)

查询多条记录

	adslAccountArray := []model.AdslAccount{}
	Mysql.Sql("SELECT * FROM `adsl_account`").Select(&adslAccountArray)

插入记录

	Mysql.Sql("INSERT INTO `adsl_account` SET `eth`='eth1',`macvlan`='macvlan1',`account`='account111',`password`='password1',`dialing_account`='d111',`running_threads`='r111'").Insert()

批量插入/修改记录

    //数据库中存在ID为1,2数据就修改，不存在则新增。 
    //用来判断新增或保存的 有主键，或则唯一索引(联合唯一索引也可以)
	run1 := []model.AdslRun{
		{
			Id: 1,
			Account:"ttt2",
		},
		{
			Id: 2,
			Account:"ttt3",
		},
	}
	ylf.Mysql.SaveAll(&run1, "id", "account")


事务

	Mysql.Begin()
	Mysql.Sql("UPDATE `adsl_account` SET  eth = 'eth4747' WHERE id =1").Update()
	if _,err:=Mysql.Sql("UPDATE `adsl_account` SET  eth = '8789' WHERE id =2").Update();err!=nil{
		Mysql.RollBack()
	}else{
		Mysql.Commit()
	}

数据迁移

	ylf.Mysql.AutoMigrate(&model.AdslRun{}, &model.AdslAccount{})

