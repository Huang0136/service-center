package constants

import (
	"database/sql"
	"fmt"
	"log"

	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/redis.v5"
	"github.com/garyburd/redigo/redis"
)

var ScDB *sql.DB
var RedisDB *redis.Client

func initDB() {
	fmt.Println(Configs)
	// 初始化MySQL
	db, err := sql.Open(Configs["database"], Configs["mysql.url"])
	if err != nil {
		log.Fatalln("初始化MySQL数据库出错", err)
	}

	ScDB = db
	log.Println(ScDB)

	// 初始化Redis
	redisDB, _ := strconv.Atoi(Configs["redis.database"])
	redisClient := redis.NewClient(&redis.Options{
		Addr:     Configs["redis.addr"],
		Password: Configs["redis.password"],
		DB:       redisDB,
	})

	RedisDB = redisClient
	log.Println("初始化redis数据库成功,", RedisDB)

}
