package ylf

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"reflect"
	"strings"
	"time"
)

type MysqlStructure struct {
	*sql.DB
	Host         string
	Port         string
	User         string
	Password     string
	Database     string
	SqlString    string
	Param        []interface{}
	Rows         *sql.Rows
	Err          error
	Transaction  *sql.Tx
	Fields       MigrateStructure
	TableAttr    TableAttr
	PrimaryIndex string
	UniqueIndex  []UniqueIndex
	Index        []Index
	Migrate      []MigrateStructure
}

func NewMysql(host string, port string, user string, password string, db string) *MysqlStructure {
	connString := fmt.Sprintf("%s:%s@tcp(%s:%s)/", user, password, host, port)
	conn, err := GetConn(connString)
	if err != nil {
		log.Fatalln(err.Error())
	} else {
		SyncDb(conn, db)
		/*同步数据库结束，重新链接*/
		conn, _ = GetConn(connString + db)
	}
	return &MysqlStructure{
		DB:       conn,
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: db,
	}
}

func GetConn(config string) (conn *sql.DB, err error) {
	if conn, err = sql.Open("mysql", config); err != nil {
		log.Fatalln(err)
		return conn, err
	}
	conn.SetConnMaxLifetime(100)
	conn.SetMaxOpenConns(100)
	conn.SetMaxIdleConns(10)
	return conn, err
}

func SyncDb(conn *sql.DB, dbName string) {
	if err := HasDb(conn, dbName); err != nil {
		if newErr := CreateDb(conn, dbName); newErr != nil {
			log.Fatalln(newErr)
		}
	} else {
		log.Println("已经存在数据库:", dbName, "跳过同步")
	}
}

func CreateDb(conn *sql.DB, dbName string) error {
	createDb, err := conn.Prepare("CREATE DATABASE " + dbName)
	if err != nil {
		log.Println("创建数据库失败！")
		return err
	}
	res, err := createDb.Exec()
	if err != nil {
		log.Println("执行CREATE DATABASE " + dbName + "失败！")
		return err
	}
	log.Println("创建数据库", dbName, "成功！")
	aff, _ := res.RowsAffected()
	log.Println("影响行数：", aff)
	return nil
}

func HasDb(conn *sql.DB, dbName string) error {
	var hasDbString string
	sql := "SHOW DATABASES LIKE '" + dbName + "'"
	if err := conn.QueryRow(sql).Scan(&hasDbString); err != nil {
		log.Println("检测到不存在数据库：" + dbName)
		return err
	}
	return nil
}

func (mysql *MysqlStructure) Sql(sql string, param ...interface{}) *MysqlStructure {
	mysql.SqlString = sql
	mysql.Param = param
	return mysql
}

/*开启事务*/
func (mysql *MysqlStructure) Begin() (err error) {
	mysql.Transaction, err = mysql.DB.Begin()
	if err != nil {
		return err
	}
	return nil
}

/*提交事务*/
func (mysql *MysqlStructure) Commit() error {
	if err := mysql.Transaction.Commit(); err != nil {
		return err
	}
	mysql.Transaction = nil
	return nil
}

/*回滚事务*/
func (mysql *MysqlStructure) RollBack() error {
	if err := mysql.Transaction.Rollback(); err != nil {
		return err
	}
	mysql.Transaction = nil
	return nil
}

func (mysql *MysqlStructure) SaveAll(data interface{}, fields ...interface{}) error {
	var rowValue []string
	var allRows []string

	var fieldsArr []string
	var updateFieldsArr []string
	dataElem := reflect.ValueOf(data).Elem()
	var tableName string
	for i := 0; i < dataElem.Len(); i++ {
		rowKeyLen := dataElem.Index(i).NumField()
		rowValue = []string{}
		for rowKey := 0; rowKey < rowKeyLen; rowKey++ {
			dataType := reflect.TypeOf(dataElem.Index(i).Interface())
			defaultTag := dataType.Field(rowKey).Tag.Get("default")
			fName := mysql.snakeString(dataType.Field(rowKey).Name)
			rowV := dataElem.Index(i).Field(rowKey).Interface()

			/*如果结构体tag中找到now或update_now,或则此字段不在fields中,都不会对此字段做任何操作*/
			if defaultTag == "now()" || defaultTag == "update_now()" || InSlice(fields, fName) == false {
				continue
			}

			/*只有第一行数据，用来取出字段名，表名等信息*/
			if i == 0 {
				fieldsArr = append(fieldsArr, fmt.Sprintf("`%s`", fName))
				updateFieldsArr = append(updateFieldsArr, fmt.Sprintf("%s = VALUES(%s)", fName, fName))
				tableName = mysql.snakeString(reflect.TypeOf(dataElem.Index(i).Interface()).Name())
			}
			rowValue = append(rowValue, fmt.Sprintf("'%v'", rowV))
		}
		allRows = append(allRows, fmt.Sprintf("(%s)", strings.Join(rowValue, ",")))
	}

	if tableName == "" {
		return errors.New("数据格式不正确")
	}

	sql := fmt.Sprintf("INSERT INTO `%s`(%s) VALUES %s ON DUPLICATE KEY UPDATE %s",
		tableName,
		strings.Join(fieldsArr, ","),
		strings.Join(allRows, ","),
		strings.Join(updateFieldsArr, ","),
	)

	if _, err := mysql.Sql(sql).Insert(); err != nil {
		return err
	}

	return nil
}

func (mysql *MysqlStructure) Insert() (lastId int64, err error) {
	var prepare *sql.Stmt
	if mysql.Transaction != nil {
		prepare, err = mysql.Transaction.Prepare(mysql.SqlString)
	} else {
		prepare, err = mysql.DB.Prepare(mysql.SqlString)
	}

	defer prepare.Close()

	if err != nil {
		log.Println("预执行SQL失败：", mysql.SqlString)
		return lastId, err
	}

	res, err := prepare.Exec()
	if err != nil {
		log.Println("执行SQL失败：", mysql.SqlString)
		return lastId, err
	}
	log.Println("执行SQL成功：", mysql.SqlString)
	aff, _ := res.RowsAffected()
	log.Println("影响行数：", aff)
	lastId, _ = res.LastInsertId()
	return lastId, nil
}

func (mysql *MysqlStructure) Exec() (result int64, err error) {
	return mysql.Update()
}

func (mysql *MysqlStructure) Update() (affected int64, err error) {
	var prepare *sql.Stmt
	if mysql.Transaction != nil {
		prepare, err = mysql.Transaction.Prepare(mysql.SqlString)
	} else {
		prepare, err = mysql.DB.Prepare(mysql.SqlString)
	}
	defer prepare.Close()

	if err != nil {
		log.Println("预执行SQL失败：", mysql.SqlString)
		return affected, err
	}
	res, err := prepare.Exec()
	if err != nil {
		log.Println("执行SQL失败：", mysql.SqlString)
		return affected, err
	}
	log.Println("执行SQL成功：", mysql.SqlString)
	affected, _ = res.RowsAffected()
	log.Println("影响行数：", affected)
	return affected, nil
}

func (mysql *MysqlStructure) Delete() (affected int64, err error) {
	var prepare *sql.Stmt
	if mysql.Transaction != nil {
		prepare, err = mysql.Transaction.Prepare(mysql.SqlString)
	} else {
		prepare, err = mysql.DB.Prepare(mysql.SqlString)
	}
	defer prepare.Close()

	if err != nil {
		log.Println("预执行SQL失败：", mysql.SqlString)
		return affected, err
	}
	res, err := prepare.Exec()
	if err != nil {
		log.Println("执行SQL失败：", mysql.SqlString)
		return affected, err
	}
	log.Println("执行删除SQL成功：", mysql.SqlString)
	affected, _ = res.RowsAffected()
	log.Println("影响行数：", affected)
	return affected, nil
}

func (mysql *MysqlStructure) Find(data interface{}) error {
	dataElem := reflect.ValueOf(data).Elem()
	//dataType := dataElem.Kind().String()
	row := make([]interface{}, dataElem.NumField())
	for i := 0; i < dataElem.NumField(); i++ {
		row[i] = dataElem.Field(i).Addr().Interface()
	}

	rows, err := mysql.DB.Query(mysql.SqlString, mysql.Param...)
	if err != nil {
		return err
	}

	for rows.Next() {
		if err := rows.Scan(row...); err != nil {
			log.Fatal(err)
		}
		for i := 0; i < dataElem.NumField(); i++ {
			switch dataElem.Type().Field(i).Type.String() {
			case "time.Time":
				date := dataElem.Field(i).Interface().(time.Time)
				fmt.Println(date.Format("2006-01-02"))
			default:
				fmt.Println("do nothing")
			}
		}
	}
	rows.Close()
	return nil
}

func (mysql *MysqlStructure) Select(data interface{}) error {
	val := reflect.Indirect(reflect.ValueOf(data))
	typ := val.Type()
	mVal := reflect.Indirect(reflect.New(typ.Elem())).Addr()
	f := mVal.Interface()
	dataElem := reflect.ValueOf(f).Elem()
	//dataType := dataElem.Kind().String()
	row := make([]interface{}, dataElem.NumField())
	for i := 0; i < dataElem.NumField(); i++ {
		row[i] = dataElem.Field(i).Addr().Interface()
	}

	rows, err := mysql.DB.Query(mysql.SqlString, mysql.Param...)
	if err != nil {
		return err
	}

	for rows.Next() {
		if err := rows.Scan(row...); err != nil {
			log.Fatal(err)
		}
		for i := 0; i < dataElem.NumField(); i++ {
			switch dataElem.Type().Field(i).Type.String() {
			case "time.Time":
				date := dataElem.Field(i).Interface().(time.Time)
				fmt.Println(date.Format("2006-01-02"))
			default:
				fmt.Println("do nothing")
			}
		}
		val = reflect.Append(val, dataElem)
	}
	rows.Close()

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(val.Interface()); err != nil {
		return err
	}
	gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(data)
	return nil
}

var Mysql *MysqlStructure

/*数据迁移 数据行结构体*/
type MigrateStructure struct {
	Field string
	FieldsAttr
}

/*数据迁移 字段属性*/
type FieldsAttr struct {
	Type        string
	Json        string
	Auto        string
	Null        string
	Default     string
	Hidden      string
	UniqueIndex string
	Index       string
	Comment     string
}

/*数据迁移 表属性*/
type TableAttr struct {
	ENGINE         string
	AUTO_INCREMENT string
	DEFAULT        string
}

/*主索引*/
var PrimaryIndex string

/*唯一索引*/
type UniqueIndex struct {
	Name  string
	Index []string
}

/*普通索引*/
type Index struct {
	Name  string
	Index []string
}

/*批量迁移数据表*/
func (mysql *MysqlStructure) AutoMigrate(data ...interface{}) error {
	for _, v := range data {
		if err := mysql.MigrateOne(v); err != nil {
			return err
		}
	}
	return nil
}

/*迁移单个数据表*/
func (mysql *MysqlStructure) MigrateOne(data interface{}) error {
	tableName := mysql.snakeString(reflect.TypeOf(data).Elem().Name())
	if err := mysql.GetStructureFields(data); err != nil {
		return err
	}

	/*遍历Migrate,生成SQL string*/
	var sqlSlice []string
	var sql string
	for _, v := range mysql.Migrate {
		sql = fmt.Sprintf("`%s`", v.Field)
		if v.Type != "" {
			sql += " " + v.Type
		}

		if v.Null != "true" {
			sql += " NOT NULL"
		}

		if v.Type == "timestamp" && v.Default == "now()" {
			sql += " DEFAULT CURRENT_TIMESTAMP"
		} else if v.Type == "timestamp" && v.Default == "update_now()" {
			sql += " DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"
		} else if v.Default != "" {
			sql += " DEFAULT " + v.Default
		}

		if v.Auto == "true" {
			sql += " AUTO_INCREMENT"
		}

		if v.Comment != "" {
			sql += fmt.Sprintf(" COMMENT '%s'", v.Comment)
		}
		sqlSlice = append(sqlSlice, sql)
	}

	/*补充主键*/
	if mysql.PrimaryIndex != "" {
		sqlSlice = append(sqlSlice, fmt.Sprintf("PRIMARY KEY (`%s`)", mysql.PrimaryIndex))
	}

	/*补充唯一索引*/
	uniqueIndex := mysql.GetUniqueIndexMap()
	if mysql.UniqueIndex != nil {
		for i, v := range uniqueIndex {
			sql = fmt.Sprintf("UNIQUE KEY `%s` (%s)", i, strings.Join(v, ","))
			sqlSlice = append(sqlSlice, sql)
		}
	}

	/*补充普通索引*/
	index := mysql.GetIndexMap()
	if mysql.UniqueIndex != nil {
		for i, v := range index {
			sql = fmt.Sprintf("KEY `%s` (%s)", i, strings.Join(v, ","))
			sqlSlice = append(sqlSlice, sql)
		}
	}

	/*表属性，先写死，后期根据需要修改*/
	tableAttr := &TableAttr{
		ENGINE:         "InnoDB",
		AUTO_INCREMENT: "0",
		DEFAULT:        "CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci",
	}
	taValue := reflect.ValueOf(tableAttr).Elem()

	var tabString string
	for i := 0; i < taValue.NumField(); i++ {
		tv := taValue.Field(i).String()
		tn := reflect.TypeOf(tableAttr).Elem().Field(i).Name
		if tn == "DEFAULT" {
			tabString += fmt.Sprintf("%s %s ", tn, tv)
		} else {
			tabString += fmt.Sprintf("%s=%s ", tn, tv)
		}
	}

	sql = fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` ( %s ) %s", tableName, strings.Join(sqlSlice, ","), tabString)
	if _, err := mysql.Sql(sql).Exec(); err != nil {
		return err
	}
	mysql.ClearData()
	return nil
}

func (mysql *MysqlStructure) ClearData() {
	//SqlString    string
	//Param        []interface{}
	//Rows         *sql.Rows
	//Err          error
	//Transaction  *sql.Tx
	//Fields       MigrateStructure
	//TableAttr    TableAttr
	//PrimaryIndex string
	//UniqueIndex  []UniqueIndex
	//Index        []Index
	//Migrate      []MigrateStructure

	mysql.SqlString = ""
	mysql.Param = nil
	mysql.Rows = nil
	mysql.Err = nil
	mysql.Transaction = nil
	mysql.Fields = MigrateStructure{}
	mysql.TableAttr = TableAttr{}
	mysql.PrimaryIndex = ""
	mysql.UniqueIndex = nil
	mysql.Index = nil
	mysql.Migrate = nil
}

/*获取索引MAP*/
func (mysql *MysqlStructure) GetUniqueIndexMap() map[string][]string {
	uniqueIndex := make(map[string][]string)

	for _, v := range mysql.UniqueIndex {
		if uniqueIndex[v.Index[0]] != nil {
			uniqueIndex[v.Index[0]] = append(uniqueIndex[v.Index[0]], fmt.Sprintf("`%s`", v.Name))
		} else {
			uniqueIndex[v.Index[0]] = []string{fmt.Sprintf("`%s`", v.Name)}
		}
	}
	return uniqueIndex
}

/*获取索引MAP*/
func (mysql *MysqlStructure) GetIndexMap() map[string][]string {
	uniqueIndex := make(map[string][]string)

	for _, v := range mysql.Index {
		if uniqueIndex[v.Index[0]] != nil {
			uniqueIndex[v.Index[0]] = append(uniqueIndex[v.Index[0]], fmt.Sprintf("`%s`", v.Name))
		} else {
			uniqueIndex[v.Index[0]] = []string{fmt.Sprintf("`%s`", v.Name)}
		}
	}
	return uniqueIndex
}

/*处理结构体的字段 属性 TAG*/
func (mysql *MysqlStructure) GetStructureFields(data interface{}) error {
	t := reflect.TypeOf(data).Elem()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return errors.New("Check type error not Struct")
	}
	fieldNum := t.NumField()

	var migrate []MigrateStructure
	for i := 0; i < fieldNum; i++ {
		tag := t.Field(i).Tag
		fieldName := mysql.snakeString(t.Field(i).Name)
		mRow := MigrateStructure{
			Field: fieldName,
			FieldsAttr: FieldsAttr{
				Type:        tag.Get("type"),
				Json:        tag.Get("json"),
				Auto:        tag.Get("auto"),
				Null:        tag.Get("null"),
				Default:     tag.Get("default"),
				Hidden:      tag.Get("hidden"),
				UniqueIndex: tag.Get("unique_index"),
				Index:       tag.Get("index"),
				Comment:     tag.Get("comment"),
			},
		}
		if mRow.FieldsAttr.Auto == "true" {
			mysql.PrimaryIndex = fieldName
		}

		if mRow.FieldsAttr.UniqueIndex != "" {
			mysql.UniqueIndex = append(mysql.UniqueIndex, UniqueIndex{
				Name:  fieldName,
				Index: []string{mRow.FieldsAttr.UniqueIndex},
			})
		}

		if mRow.FieldsAttr.Index != "" {
			mysql.Index = append(mysql.Index, Index{
				Name:  fieldName,
				Index: []string{mRow.FieldsAttr.Index},
			})
		}

		migrate = append(migrate, mRow)
	}
	mysql.Migrate = migrate
	return nil
}

/*驼峰命名转蛇形*/
func (mysql *MysqlStructure) snakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data[:]))
}
