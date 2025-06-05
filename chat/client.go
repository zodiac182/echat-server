package chat

// 用户系统

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zodiac182/echat/global"
	"github.com/zodiac182/echat/message"
	"github.com/zodiac182/echat/model"
)

// 一个room对应一个session
// 每个room下有多个client，每个client对应一个用户
// 通过判断是否有client的连接来判断对方是否在线

// 定义聊天客户端
type Client struct {
	conn          *websocket.Conn // websocket连接
	msg           chan []byte     // 用于发送消息
	userId        string          // 用户ID
	userName      string          // 用户名
	toUserId      string          // 目标用户ID
	toUserName    string          // 目标用户名
	roomId        string          // 房间ID, 对应ticketID
	lastHeartbeat time.Time       // 最后心跳时间
}

func (c *Client) getHistoryMessage() {
	// 用于在客户端连接之后，获取历史消息
	messages := message.GetHistoryMessage(c.roomId)
	// 获取历史消息就意味着消息状态要变成已读, 所以需要更新数据库
	for _, chatMsg := range messages {
		if c.userId == chatMsg.ToUserId && c.userId != chatMsg.FromUserId { // 别人发送给我的消息
			tx := global.DB.Model(&model.Message{}).Where("id = ? AND read_status = ?", chatMsg.ID, false).Updates(map[string]interface{}{"read_status": true})
			if tx.RowsAffected > 0 {
				log.Println("状态更新，需要自己发送消息,更新消息已读状态")
				// 通知发送者
				chatMsg.ReadStatus = true
				retMsg, err := json.Marshal(chatMsg)
				if err != nil {
					log.Println("JSON 序列化失败:", err)
					continue
				}
				c.sendToUser(retMsg, chatMsg.FromUserId)
			}
		}

		msgContent, err := json.Marshal(chatMsg)
		if err != nil {
			log.Println("JSON 序列化失败:", err)
			continue
		}
		c.sendToSelf(msgContent) // 将历史消息发送到自己的客户端
	}
}

func (c *Client) readPump() {
	// 为每一个用户提供一个go rountine
	// 用于从客户端读取消息并处理
	// 用于处理用户的输入
	log.Println("readPump 启动", c.roomId)
	defer func() {
		chatMsg := model.Message{
			MessageType: model.SystemMsg,
			Content:     fmt.Sprintf("%s 已离线", c.userId),
		}

		_, err := json.Marshal(chatMsg)
		if err != nil {
			log.Println("JSON 序列化失败:", err)
			return
		}
		c.disconnect()
		log.Println("readPump关闭", c.roomId)
	}()

	for {
		_, msg, err := c.conn.ReadMessage()
		// log.Printf("receive message: %s\n", string(msg))
		if err != nil {
			break
		}
		// log.Printf("从客户端读取消息%s\n", string(msg))
		var chatMsg model.Message

		err = json.Unmarshal(msg, &chatMsg)
		if err != nil {
			log.Println("receive message.JSON 解析失败:", err)
			continue
		}
		switch chatMsg.MessageType {
		case model.PingMsg: // 心跳包
			c.lastHeartbeat = time.Now()
			continue
		case model.SystemMsg: // 系统消息
			// 忽略
			continue
		default:
			// 默认用户消息

			// 首先保存数据库, 就会有ID等信息
			chatMsg.ReadStatus = false
			global.DB.Create(&chatMsg)

			msgContent, err := json.Marshal(chatMsg)
			if err != nil {
				log.Println("数据库获取数据JSON 序列化失败:", err)
				continue
			}

			// 将消息发给目标用户,如果发送成功则为已读,对方不在线则为未读
			// 发送成功需要更新已读状态
			chatMsg.ReadStatus = c.sendToUser(msgContent, chatMsg.ToUserId)

			// 保存新聊天记录到数据库
			// global.DB.Create(&chatMsg)
			if chatMsg.ReadStatus {
				global.DB.Model(&model.Message{}).Where("id = ?", chatMsg.ID).Updates(map[string]interface{}{"read_status": true})

				msgContent, err = json.Marshal(chatMsg)
				if err != nil {
					log.Println("数据库获取数据JSON 序列化失败:", err)
					continue
				}
			}
			c.sendToSelf(msgContent)
		}

	}
}

func (c *Client) writePump() {
	// 用于向客户端发送消息
	log.Println("writePump 启动", c.roomId)
	for msg := range c.msg {
		// log.Printf("向websocket客户端: 房间号 %s 用户 %s 写入消息: %s\n", c.roomId, c.userId, string(msg))
		err := c.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			break
		}

		// 更新当前的已读和未读状态
		// 如果是发给自己的，则不更新状态
		// 如果是发送给目标用户,则更新状态
		// var chatMsg model.Message
		// err = json.Unmarshal(msg, &chatMsg)
		// if err != nil {
		// 	log.Println("JSON 解析失败:", err)
		// 	continue
		// }
		// // 非即时聊天消息, 忽略
		// if chatMsg.MessageType != model.UserMsg {
		// 	continue
		// }

		// if c.userId == chatMsg.ToUserId {
		// 	tx := global.DB.Model(&model.Message{}).Where("id = ? AND is_read = ?", chatMsg.ID, false).Updates(map[string]interface{}{"is_read": true})
		// 	if tx.RowsAffected > 0 {
		// 		log.Println("状态更新，需要自己发送消息,更新消息已读状态")
		// 		// 通知发送者
		// 		chatMsg.IsRead = true
		// 		retMsg, err := json.Marshal(chatMsg)
		// 		if err != nil {
		// 			log.Println("JSON 序列化失败:", err)
		// 			continue
		// 		}
		// 		c.sendToSelf(retMsg)
		// 	}
		// }

	}
	log.Println("writePump结束", c.roomId)
}

func (c *Client) broadcast(msg []byte) {
	// 向所有客户端发送消息，包括自己
	// 发送给自己是用于历史消息，以及判断是否发送成功

	roomIns, ok := rooms.Load(c.roomId)
	if !ok {
		log.Println("room not exist")
		return
	}

	room := roomIns.(*sync.Map)
	room.Range(func(key, value interface{}) bool {
		client, ok := value.(*Client)
		if !ok {
			return true // 忽略非 *Client 的 key
		}

		select {
		case client.msg <- msg:
			// 正常发送消息
		default:
			// 发送失败，关闭 channel 并移除 client
			close(client.msg)
			room.Delete(client)
		}

		return true // 继续遍历
	})
}

func (c *Client) sendToUser(msg []byte, toUserId string) bool {
	// 向指定用户发送消息, 主要用于向客服发送
	log.Println("当前rooms列表：")

	rooms.Range(func(key, value any) bool {
		roomId := key.(string)
		room := value.(*sync.Map)
		log.Printf("- roomId: %s, room: %p\n", roomId, room)
		room.Range(func(key, value any) bool {
			clientId := key.(string)
			client := value.(*Client)
			log.Printf("  - clientId: %s, client: %p\n", clientId, client)
			return true
		})
		return true
	})
	// log.Println("sendToUser", string(msg), toUserId)
	roomIns, ok := rooms.Load(c.roomId)
	room := roomIns.(*sync.Map)
	if !ok {
		log.Println("room not exist")
		return false
	}

	if client, exists := room.Load(toUserId); exists {
		// log.Printf("向用户 %s 发送消息: %s\n", toUserId, string(msg))
		client.(*Client).msg <- msg
	} else {
		log.Printf("用户 %s 不在线, 无法发送消息\n", toUserId)
		return false
	}
	return true
}

func (c *Client) sendToSelf(msg []byte) {
	// 向自己发送消息, 用于历史消息和新消息
	// log.Println("向自己发送消息", string(msg))
	select {
	case c.msg <- msg:
	default:
		// 防止阻塞
		RemoveClientFromRoom(c.roomId, c)
		close(c.msg)
	}
}

func (c *Client) disconnect() {
	log.Printf("客户端 %s-%s 断开连接\n", c.roomId, c.userId)
	RemoveClientFromRoom(c.roomId, c)
	close(c.msg)
	c.conn.Close()
}

// 客户开启心跳检测
func (c *Client) StartHeartbeatChecker() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C // 阻塞, 60秒执行一次
		now := time.Now()

		if now.Sub(c.lastHeartbeat) > 90*time.Second {
			log.Printf("chat client检测到无效连接，清理: ticketId=%s, userId=%s, userName=%s", c.roomId, c.userId, c.userName)
			RemoveClientFromRoom(c.roomId, c)
			c.conn.Close()
			return // 退出go routine
		}
	}
}
