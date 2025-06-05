package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/zodiac182/echat/global"
	"github.com/zodiac182/echat/model"
)

// 定义管理客户端
type MgrClient struct {
	conn          *websocket.Conn // websocket连接
	msg           chan []byte     // 用于发送消息
	userId        string          // 用户ID
	mu            sync.Mutex
	lastHeartbeat time.Time // 最后心跳时间
}

var mgrClients sync.Map // userId -> *MgrClient

// HandleWebSocket 处理 用户的WebSocket 请求
func HandleMgrClientSocket(c *gin.Context) {

	userId := c.Query("userId")
	// 检查是否已存在连接
	if oldClient, ok := mgrClients.Load(userId); ok {
		log.Printf("已有连接，关闭旧连接: %s", userId)
		oldClient.(*MgrClient).disconnect()
		mgrClients.Delete(userId)
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("WebSocket 升级失败:", err)
		return
	}

	// 创建客户的websocket对象,用于处理客户端的消息
	client := &MgrClient{
		conn:   conn,
		msg:    make(chan []byte, 1024),
		userId: userId,
	}
	mgrClients.Store(userId, client)

	// 创建 context 并绑定 cancel
	ctx, cancel := context.WithCancel(context.Background())

	// 启动读写 goroutine
	go client.writePump()
	go func() {
		client.readPump()
		cancel() // 停掉 SyncTicketsInfo
	}()
	go client.SyncTicketsInfo(ctx)

	go client.StartHeartbeatChecker(ctx) // 启动心跳检测

}

// 获取当前用户的在线状态
func (m *MgrClient) GetTicketsInfo() []*Ticket {
	UserId := m.userId

	cs := CustomerService{
		UserId: UserId,
	}

	tickets := cs.GetTickets(UserId)

	for _, ticket := range tickets {
		// 获取未读消息数量
		var count int64
		global.DB.Model(&model.Message{}).Where("to_user_id =? and ticket_id =? and read_status =?", m.userId, ticket.TicketId, false).Count(&count)
		ticket.UnreadMsgCount = int(count)
		// log.Printf("查询： to_user_id = %s, ticket_id = %s, read_status = %t, count = %d\n", m.userId, ticket.TicketId, false, count)

		// 获取对方的在线状态
		roomIns, ok := rooms.Load(ticket.TicketId)
		if !ok {
			// log.Println("room not exist:", ticket.TicketId)
			continue
		}

		room := roomIns.(*sync.Map)
		room.Range(func(key, value any) bool {
			client := value.(*Client)
			// 获取对方的在线状态
			//TODO：写一个维持状态的方法
			if client.userId == ticket.CreaterId {
				ticket.OnlineStatus = true
			}
			return true
		})
	}
	return tickets
}

func (c *MgrClient) readPump() {
	// 为每一个用户提供一个go rountine
	// 用于从客户端读取消息并处理
	// 用于处理用户的输入
	log.Println("readPump 启动", c.userId)
	defer func() {
		c.disconnect()
		log.Println("readPump关闭", c.userId)
	}()

	for {
		_, msg, err := c.conn.ReadMessage()
		// log.Printf("receive message: %s\n", string(msg))
		if err != nil {
			break
		}

		if string(msg) != "ping" {
			return
		}

		//  收到ping消息，
		c.lastHeartbeat = time.Now()

	}
}

func (c *MgrClient) writePump() {
	// 用于向客户端发送消息
	log.Println("writePump 启动", c.userId)
	for msg := range c.msg {

		err := c.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			break
		}
	}
	log.Println("writePump结束", c.userId)
}

func (c *MgrClient) disconnect() {
	log.Printf("客户端 %s 断开连接\n", c.userId)
	defer func() {
		mgrClients.Delete(c.userId) // 删除映射
	}()

	// 避免重复关闭
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		close(c.msg)
		c.conn.Close()
		c.conn = nil
	}
}

// 客户开启心跳检测
func (c *MgrClient) StartHeartbeatChecker(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C: // 阻塞, 60秒执行一次
			now := time.Now()

			if now.Sub(c.lastHeartbeat) > 90*time.Second {
				log.Printf("MgrClient检测到无效连接，清理: userId = %s\n", c.userId)
				c.disconnect()
				return // 退出go routine
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *MgrClient) sendTicketsInfo(msg []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 判断连接是否已关闭（假设 conn 为 *websocket.Conn）
	if c.conn == nil {
		log.Println("Mgr WebSocket 连接不存在", c.userId)
		return
	}

	// 判断 channel 是否已关
	select {
	case c.msg <- msg:
		// 成功发送
	default:
		log.Println("channel 已满或已关闭，关闭连接")
		c.disconnect()
		return
	}

}

// 同步状态信息
func (c *MgrClient) SyncTicketsInfo(ctx context.Context) {
	log.Println("SyncTicketsInfo 启动", c.userId)
	defer func() {
		c.disconnect()
		log.Println("SyncTicketsInfo 结束", c.userId)
	}()

	send := func() {
		ticketsInfo := c.GetTicketsInfo()
		msg, err := json.Marshal(ticketsInfo)
		if err != nil {
			log.Println("Marshal ticketsInfo 失败:", err)
			return
		}
		c.sendTicketsInfo(msg)
	}

	send() // 立即执行一次

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			send()
		}
	}
}
