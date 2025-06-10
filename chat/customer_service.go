package chat

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/zodiac182/echat/global"
)

// 客户系统

type CustomerService struct {
	UserId   string
	UserName string
}

var CSName = "管理员"

// 获取所有工单,并创建room
func (cs *CustomerService) GetTickets(userName string) []*Ticket {
	// 1. 获取所有工单

	// var tickets = []*Ticket{
	// 	{
	// 		TicketId:    "ABCD-202506041312541256",
	// 		Title:       "校网无法上网",
	// 		Description: "请帮忙检查一下校园网是否正常，以及是否有流量费用。还有，请提供一下你所使用的设备型号和系统版本。另外，请提供一下你所遇到的具体问题。",
	// 		Status:      "待处理",
	// 		CreateBy:    "张三",
	// 		CreaterId:   "zhangsan",
	// 		OwnedBy:     CSName,
	// 		OwnerId:     "admin",
	// 	},
	// 	{
	// 		TicketId:    "EHELP-20250605152900694",
	// 		Title:       "校网无法上网",
	// 		Description: "请帮忙检查一下校园网是否正常，以及是否有流量费用。还有，请提供一下你所使用的设备型号和系统版本。另外，请提供一下你所遇到的具体问题。",
	// 		Status:      "待处理",
	// 		CreateBy:    "李四",
	// 		CreaterId:   "18181818181",
	// 		OwnedBy:     "田浩文",
	// 		OwnerId:     "6",
	// 	},
	// }
	// 解析基础 URL
	itsmBaseUrl, err := url.Parse(global.ITSM_SERVER_URL)
	if err != nil {
		return nil
	}

	// 创建查询参数
	queryParams := itsmBaseUrl.Query()
	queryParams.Add("handlerName", userName)

	// 将参数附加到 URL 上
	itsmBaseUrl.RawQuery = queryParams.Encode()

	log.Printf("查询参数: %s", itsmBaseUrl.String())

	ticketsResponse, err := http.Get(itsmBaseUrl.String())
	if err != nil {
		log.Println("获取工单列表失败")
		return nil
	}
	defer ticketsResponse.Body.Close()

	var response map[string]interface{}
	err = json.NewDecoder(ticketsResponse.Body).Decode(&response)
	if err != nil {
		log.Printf("解析工单列表失败:  %+v\n", err)
		return nil
	}

	data, ok := response["data"].([]interface{})
	if !ok {
		log.Printf("解析工单列表失败:  %+v\n", "data 字段不是数组")
		return nil
	}

	// 2. 遍历工单
	var tickets []*Ticket
	for _, item := range data {
		detailData, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		ticket := ParseTicket(detailData)
		if ticket == nil {
			continue
		}
		tickets = append(tickets, ticket)
	}

	return tickets
}

func ParseTicket(ticketData map[string]interface{}) *Ticket {
	var ticket Ticket
	var ok bool
	ticketId, ok := ticketData["ticketId"].(float64)
	if !ok {
		return nil
	}
	ticket.TicketId = strconv.Itoa(int(ticketId))

	ticket.SN, ok = ticketData["sn"].(string)
	if !ok {
		log.Println("SN解析错误")
		return nil
	}
	ticket.Title, ok = ticketData["title"].(string)
	if !ok {
		log.Println("Title解析错误")
		return nil
	}
	ticket.Description, ok = ticketData["description"].(string)
	if !ok {
		log.Println("Description解析错误")
		formDataJson, _ := json.MarshalIndent(ticketData, "", "  ") // 缩进格式化
		log.Printf("formData:\n%s\n", string(formDataJson))
		ticket.Description = ""
		// return nil
	}
	ticket.Status, ok = ticketData["status"].(string)
	if !ok {
		log.Println("Status解析错误")
		return nil
	}

	owners, ok := ticketData["current_handlers"].([]interface{})
	if !ok {
		log.Println("current_handlers")
		return nil
	}
	if len(owners) > 0 {
		owner, ok := owners[0].(map[string]interface{})
		if !ok {
			log.Println("owner解析错误")
			return nil
		}
		ticket.OwnedBy, ok = owner["name"].(string)
		if !ok {
			log.Println("owner name解析错误")
			return nil
		}
		ownerId, ok := owner["id"].(float64)
		if !ok {
			log.Println("owner id解析错误")
			return nil
		}
		ticket.OwnerId = strconv.Itoa(int(ownerId))
	}

	// 处理学生ID及学生姓名
	formData, ok := ticketData["form_data"].(map[string]interface{})
	if !ok {
		log.Println("form_data解析错误")
		return nil
	}
	ticket.CreateBy, ok = formData["builtin_xingming"].(string)
	if !ok {
		log.Println("builtin_xingming解析错误")
		return nil
	}
	ticket.CreaterId, ok = formData["xuegonghao"].(string)
	if !ok {
		log.Println("xuegonghao解析错误")
		// formDataJson, _ := json.MarshalIndent(formData, "", "  ") // 缩进格式化
		// log.Printf("formData:\n%s\n", string(formDataJson))
		return nil
	}

	return &ticket
}
