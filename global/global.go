package global

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

var (
	DB_HOST         string
	DB_USERNAME     string
	DB_PASSWORD     string
	DB_PORT         string
	DB_NAME         string
	ITSM_SERVER_URL string
)

var DB *gorm.DB

func init() {
	log.Println("系统初始化...")
	if strings.ToLower(os.Getenv("APP_ENV")) != "production" {
		_ = godotenv.Load() // 忽略错误
	}
	DB_HOST = os.Getenv("DB_HOST")
	if DB_HOST == "" {
		DB_HOST = "localhost" // 默认路径
	}

	DB_USERNAME = os.Getenv("DB_USERNAME")
	if DB_USERNAME == "" {
		DB_USERNAME = "postgres" // 默认用户名
	}

	DB_PASSWORD = os.Getenv("DB_PASSWORD")
	if DB_PASSWORD == "" {
		DB_PASSWORD = "password" // 默认密码
	}

	DB_PORT = os.Getenv("DB_PORT")
	if DB_PORT == "" {
		DB_PORT = "5432" // 默认端口
	}

	DB_NAME = os.Getenv("DB_NAME")
	if DB_NAME == "" {
		DB_NAME = "echat" // 默认数据库名
	}

	ITSM_SERVER_URL = os.Getenv("ITSM_SERVER_URL")
}
