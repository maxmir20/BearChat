package api

import (
	"database/sql"
	"log"

	//MySQL driver
	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDB() *sql.DB {

	log.Println("attempting connections")

	var err error
	DB, err = sql.Open("mysql", "root:root@tcp(172.28.1.2:3306)/postsDB?parseTime=true")

	if err != nil {
		log.Println("couldnt connect")
		panic(err.Error())
	}

	err = DB.Ping()
	if err != nil {
		log.Println("couldnt ping")
		panic(err.Error())
	}

	return DB
}
