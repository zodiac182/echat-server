package chat

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	rooms sync.Map // key: string(ticketId), value: *sync.Map
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有跨域
	},
}

// 定义ticket
type Ticket struct {
	TicketId    string `json:"ticketId"`
	SN          string `json:"sn"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreateBy    string `json:"createBy"`
	CreaterId   string `json:"createrId"`
	OwnedBy     string `json:"ownedBy"`
	OwnerId     string `json:"ownerId"`
}

func AddClientToRoom(roomID string, client *Client) {
	// 获取或创建房间
	roomIns, _ := rooms.LoadOrStore(roomID, &sync.Map{})

	room := roomIns.(*sync.Map) // 将room断言称sync.Map
	// 将客户端添加到房间
	room.Store(client.userId, client)
}

// func GetClientsInRoom(roomID string) []*Client {
// 	var clients []*Client

// 	if roomClients, ok := rooms.Load(roomID); ok {
// 		roomClients.(*sync.Map).Range(func(key, value interface{}) bool {
// 			clients = append(clients, key.(*Client))
// 			return true // 继续遍历
// 		})
// 	}

// 	return clients
// }

func RemoveClientFromRoom(roomID string, client *Client) {
	// 从房间移除
	if roomIns, ok := rooms.Load(roomID); ok {
		roomIns.(*sync.Map).Delete(client.userId)

		// 如果房间为空，删除整个房间
		// isEmpty := true
		// roomIns.(*sync.Map).Range(func(_, _ interface{}) bool {
		// 	isEmpty = false
		// 	return false // 终止遍历
		// })
		// if isEmpty {
		// 	rooms.Delete(roomID)
		// }
	}

	// 从客服的房间列表中移除
	// if clientRooms, ok := clients.Load(client); ok {
	// 	newRooms := []string{}
	// 	for _, id := range clientRooms.([]string) {
	// 		if id != roomID {
	// 			newRooms = append(newRooms, id)
	// 		}
	// 	}
	// 	clients.Store(client, newRooms)
	// }
}

// HandleWebSocket 处理 用户的WebSocket 请求
func HandleClientSocket(c *gin.Context) {
	ticketId := c.Query("ticketId")
	userId := c.Query("userId")
	userName := c.Query("userName")
	toUserId := c.Query("toUserId")
	toUserName := c.Query("toUserName")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("WebSocket 升级失败:", err)
		return
	}

	// 创建客户的websocket对象,用于处理客户端的消息
	client := &Client{
		conn:       conn,
		msg:        make(chan []byte, 1024),
		userId:     userId,
		userName:   userName,
		toUserId:   toUserId,
		toUserName: toUserName,
		roomId:     ticketId,
	}

	// 获取或创建房间
	roomIns, _ := rooms.LoadOrStore(ticketId, &sync.Map{})
	room := roomIns.(*sync.Map) // 将room断言称sync.Map

	// 将客服添加到房间
	if _, loaded := room.LoadOrStore(userId, client); loaded {
		fmt.Println("用户已经在房间中")
	}

	ctx, cancel := context.WithCancel(context.Background())

	// 启动读写 goroutine
	go client.writePump()
	go func() {
		client.readPump()
		cancel() // 停掉 SyncTicketsInfo
	}()

	go client.StartHeartbeatChecker(ctx) // 启动心跳检测

	client.getHistoryMessage() // 加载历史消息

	// // 返回登陆成功消息
	// chatMsg := model.Message{
	// 	Type:      "system",
	// 	SessionId: roomId,
	// 	Content:   fmt.Sprintf("%s 已上线", fromUser),
	// }

	// // onlineMsg, err := json.Marshal(chatMsg)
	// // if err != nil {
	// // 	fmt.Println("JSON 序列化失败:", err)
	// // 	return
	// // }

	// client.sendToOthers(onlineMsg) // 公告进入房间消息
}

// 客服登陆
// handle customer servier Socket
// 更新状态为在线
func HandleServiceLogin(c *gin.Context) {
	UserId := c.Query("userId")
	UserName := c.Query("userName")

	cs := CustomerService{
		UserId:   UserId,
		UserName: UserName,
	}

	c.JSON(http.StatusOK, cs.GetTickets(UserName))
}
