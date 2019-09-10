package ylf

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"reflect"
	"strings"
	"time"
)

type MysqlStructure struct {
	*sql.DB
	Host        string
	Port        string
	User        string
	Password    string
	Database    string
	SqlString   string
	Param       []interface{}
	Rows        *sql.Rows
	Err         error
	Transaction *sql.Tx
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

func (mysql *MysqlStructure) Insert() (lastId int64, err error) {
	var prepare *sql.Stmt
	if mysql.Transaction != nil {
		prepare, err = mysql.Transaction.Prepare(mysql.SqlString)
	} else {
		prepare, err = mysql.DB.Prepare(mysql.SqlString)
	}

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

func (mysql *MysqlStructure) Update() (affected int64, err error) {
	var prepare *sql.Stmt
	if mysql.Transaction != nil {
		prepare, err = mysql.Transaction.Prepare(mysql.SqlString)
	} else {
		prepare, err = mysql.DB.Prepare(mysql.SqlString)
	}

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
	Type string
	Json string
	Auto string
	Null string
	Default string
	Hidden string
}

/*数据迁移 表属性*/
type TableAttr struct {
	Engine string
	AutoIncrement string
	DEFAULT string
}

/*主索引*/
var PrimaryIndex string

/*唯一索引*/
type UniqueIndex struct {
	Name string
	Index []string
}

/*普通索引*/
type Index struct {
	Name string
	Index []string
}


func (mysql *MysqlStructure)AutoMigrate(data interface{}) {
	tableName := mysql.snakeString(reflect.TypeOf(data).Elem().Name())
	fields := mysql.GetStructureFields(data)
	fmt.Println(tableName, fields)
}

func (mysql *MysqlStructure)GetStructureFields(data interface{}) []MigrateStructure {
	t := reflect.TypeOf(data).Elem()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		log.Println("Check type error not Struct")
	}
	fieldNum := t.NumField()

	var migrate []MigrateStructure
	for i := 0; i < fieldNum; i++ {
		tag := t.Field(i).Tag
		fieldName := mysql.snakeString(t.Field(i).Name)
		mRow := MigrateStructure{
			Field:fieldName,
			FieldsAttr: FieldsAttr{
				Type:    tag.Get("type"),
				Json:    tag.Get("json"),
				Auto:    tag.Get("auto"),
				Null:    tag.Get("null"),
				Default: tag.Get("default"),
				Hidden:  tag.Get("hidden"),
			},
		}
		migrate = append(migrate,mRow)
	}
	return migrate
}

/*驼峰命名转蛇形*/
func (mysql *MysqlStructure)snakeString(s string) string {
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
