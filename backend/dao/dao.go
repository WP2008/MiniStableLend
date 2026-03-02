package dao

import (
	"github.com/gomodule/redigo/redis"
	"gorm.io/gorm"
)

var Mysql *gorm.DB
var RedisConn *redis.Pool

func NewDao() {
	// 初始化数据库连接和Redis连接
	Mysql = InitMysql()
	RedisConn = InitRedis()

	// 自动迁移数据库表结构
	Mysql.AutoMigrate(&Position{}, &ScannerProgress{})
}
