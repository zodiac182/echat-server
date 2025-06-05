package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/zodiac182/echat/chat"
	"github.com/zodiac182/echat/global"
	"github.com/zodiac182/echat/message"
	"github.com/zodiac182/echat/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:generate go env -w GO111MODULE=on
//go:generate go env -w GOPROXY=https://goproxy.cn,direct
//go:generate go mod tidy

// TODO:   客户端断线重连机制
func main() {
	initDb()
	router := gin.Default()
	router.GET("/ws/client", handleWebSocket)
	router.GET("/ws/mgr", handleAdminWebSocket)

	// 添加 CORS 中间件
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 允许所有来源
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// go client.StartHeartbeatChecker()
	router.Static("/public", "./public")

	router.GET("/api/history", queryHistoryMessage)

	router.POST("/api/upload", uploadFile)
	router.GET("/api/tickets", chat.HandleServiceLogin)
	router.Run(":8080")
}

func initDb() {
	log.Println("数据库初始化...")
	dsn := "host=" + global.DB_HOST + " user=" + global.DB_USERNAME + " password=" + global.DB_PASSWORD + " dbname=" + global.DB_NAME + " port=" + global.DB_PORT + " sslmode=disable TimeZone=Asia/Shanghai"

	var err error

	global.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	global.DB.AutoMigrate(&model.Message{})
}

// 处理 WebSocket 连接
func handleWebSocket(c *gin.Context) {
	chat.HandleClientSocket(c)
}

func handleAdminWebSocket(c *gin.Context) {
	chat.HandleMgrClientSocket(c)
}

// 获取历史记录
func queryHistoryMessage(c *gin.Context) {
	ticketId := c.Query("ticketId")
	messages := message.GetHistoryMessage(ticketId)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": messages})
}

// 上传文件
var UploadDir string = "./public"

func uploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "文件获取失败"})
		return
	}

	// 确保目录存在
	err = os.MkdirAll(UploadDir, os.ModePerm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "无法创建上传目录"})
		return
	}

	// 防止文件重名
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
	filepath := fmt.Sprintf("%s/%s", UploadDir, filename)

	err = c.SaveUploadedFile(file, filepath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "保存失败"})
		return
	}

	// 可以通过URL访问（前提是用 static 公开了 uploads 路径）

	c.JSON(http.StatusOK, gin.H{
		"code":     0,
		"msg":      "上传成功",
		"fileName": filename,
	})
}
