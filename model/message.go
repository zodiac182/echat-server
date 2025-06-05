package model

import (
	"gorm.io/gorm"
)

const (
	UserMsg   string = "user"
	SystemMsg string = "system"
	PingMsg   string = "ping"
)

// Message Type:
//	 websocket.TextMessage	1	文本消息（UTF-8 编码的字符串）。
//	 websocket.BinaryMessage	2	二进制消息（如图片、Protobuf）。
//	 websocket.CloseMessage	8	连接关闭信号。
//	 websocket.PingMessage	9	Ping 心跳（服务端自动回复）。
//	 websocket.PongMessage	10	Pong 响应（服务端发送）。

// 数据库存放的收发消息
type Message struct {
	gorm.Model
	MessageType  string `json:"messageType" gorm:"-"` // 'user'  'system'  'ping'
	FromUserName string `json:"fromUserName" gorm:"type:varchar(64)"`
	FromUserId   string `json:"fromUserId" gorm:"type:varchar(64)"`
	ToUserName   string `json:"toUserName" gorm:"type:varchar(64)"`
	ToUserId     string `json:"toUserId" gorm:"type:varchar(64)"`
	TicketId     string `json:"ticketId" gorm:"type:varchar(255)"`
	Content      string `json:"content" gorm:"type:text"`
	ContentType  string `json:"contentType" gorm:"type:varchar(10);default:'text'"` // 'text' or 'image'
	ReadStatus   bool   `json:"readStatus" gorm:"default:false"`
}

func (Message) TableName() string {
	return "messages"
}
