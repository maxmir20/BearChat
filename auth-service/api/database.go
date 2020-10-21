package api

import (
	"database/sql"
	"log"

	//MySQL driver
	_ "github.com/go-sql-driver/mysql"
)

//DB represents the connection to the MySQL database
var (
	DB *sql.DB
)

//InitDB creates the MySQL database connection
func InitDB() *sql.DB {

	log.Println("attempting connections")

	var err error
	
	// Open a SQL connection to the docker container hosting the database server
	// Assign the connection to the "DB" variable
	// Look at how it's done in the other microservices!
	dbType := "mysql"
	username := "root"
	password := "root"
	ipAddress := "tcp(172.28.1.2:3306)"
	dbName := "/auth"
	// "YOUR CODE HERE"
	
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