package main

// 合宙 luatos errDump 日志接收服务
// https://gitee.com/openLuat/luatos-devlog
import (
	"database/sql"
	"flag"
	"log"
	"net"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 命令行参数
	listenAddr := flag.String("a", ":9072", "UDP port to listen on")
	dbType := flag.String("dbtype", "sqlite3", "Database type: sqlite3 or mysql")
	dbConnStr := flag.String("dbconn", "dev_log.db", "Database connection string, "+
		"eg: logdb.sqlite3 ,  \"root:123@tcp(localhost:3306)/log\"")
	flag.Parse()

	// 创建UDP地址
	addr, err := net.ResolveUDPAddr("udp", *listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listen UDP on ", *listenAddr)

	// 创建UDP连接
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 创建数据库连接
	db, err := sql.Open(*dbType, *dbConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 创建表
	if *dbType == "sqlite3" {
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS dev_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			dev TEXT,
			proj TEXT,
			lodver TEXT,
			selfver TEXT,
			devsn TEXT,
			errlog TEXT,
			ipaddr TEXT,
			ct INTEGER
		)`)
	} else if *dbType == "mysql" {
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS dev_log (
			id INT AUTO_INCREMENT PRIMARY KEY,
			dev varchar(64),
			proj varchar(64),
			lodver varchar(64),
			selfver varchar(32),
			devsn varchar(64),
			errlog TEXT,
			ipaddr varchar(64),
			ct timestamp
		)`)
	}
	if err != nil {
		log.Fatal(err)
	}

	// 接收和处理UDP消息
	buffer := make([]byte, 4096)
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println(err)
			continue
		}
		if _, err = conn.WriteToUDP([]byte(`{"r": 1}`), addr); err != nil {
			log.Println("Ack msg error:", err)
		}

		/**
		pencpu-slc_lodverxxx,0.9.0,866250060829193,91937594125402,long error mesage
		解析以上实例为数据库字段：
		```
			dev: 866250060829193
			proj: opencpu-slc
			lodver: lodverxxx
			selfver: 0.9.0
			devsn: 91937594125402
			errlog: long error mesage
		```
		**/
		message := string(buffer[:n])
		log.Println("Receive msg: ", message)
		fields := strings.SplitN(message, ",", 5)
		if len(fields) != 5 {
			log.Println("Invalid message format:", message)
			continue
		}

		proj_lodver := strings.SplitN(fields[0], "_", 2)
		if len(proj_lodver) != 2 {
			log.Println("Invalid device format:", fields[0])
			continue
		}

		// 解析字段
		ct := time.Now().UTC()
		proj := proj_lodver[0]
		lodver := proj_lodver[1]
		selfver := fields[1]
		dev := fields[2]
		devsn := fields[3]
		errlog := fields[4]
		ipaddr := addr.IP.String()

		// 将字段存储到数据库
		_, err = db.Exec(`INSERT INTO dev_log (dev, proj, lodver, selfver, devsn, errlog, ipaddr, ct)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, dev, proj, lodver, selfver, devsn, errlog, ipaddr, ct)
		if err != nil {
			log.Println("Insert log error:", err)
		} else {
			log.Println("Message stored in the database.")
		}
	}
}
