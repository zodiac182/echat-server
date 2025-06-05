package chat

// 客户系统

type CustomerService struct {
	UserId   string
	UserName string
}

var CSName = "管理员"

// 获取所有工单,并创建room
func (cs *CustomerService) GetTickets(userId string) []*Ticket {
	// 1. 获取所有工单
	// 2. 创建room

	var tickets = []*Ticket{
		{
			TicketId:    "ABCD-202506041312541256",
			Title:       "校网无法上网",
			Description: "请帮忙检查一下校园网是否正常，以及是否有流量费用。还有，请提供一下你所使用的设备型号和系统版本。另外，请提供一下你所遇到的具体问题。",
			Status:      "待处理",
			CreateBy:    "张三",
			CreaterId:   "zhangsan",
			OwnedBy:     CSName,
			OwnerId:     "admin",
		},
		{
			TicketId:    "EHELP-20250605152900694",
			Title:       "校网无法上网",
			Description: "请帮忙检查一下校园网是否正常，以及是否有流量费用。还有，请提供一下你所使用的设备型号和系统版本。另外，请提供一下你所遇到的具体问题。",
			Status:      "待处理",
			CreateBy:    "李四",
			CreaterId:   "18181818181",
			OwnedBy:     "田浩文",
			OwnerId:     "6",
		},
	}

	// 3. 遍历工单,创建room
	// for _, ticket := range tickets {
	// 	ticketId := ticket.TicketId
	// 	// 4. 创建room
	// 	customerAgent := &Client{
	// 		userId:         ticket.OwnerId,
	// 		userName:       CSName,
	// 		toUserId:   ticket.CreaterId,
	// 		toUserName: ticket.CreateBy,
	// 		roomId:         ticketId,
	// 	}

	// 	AddClientToRoom(ticketId, customerAgent)
	// }

	return tickets
}
