package message

import (
	"github.com/zodiac182/echat/global"
	"github.com/zodiac182/echat/model"
)

// 获取历史记录
func GetHistoryMessage(roomId string) []model.Message {
	var messages []model.Message
	global.DB.Where("ticket_id =?", roomId).Order("created_at").Find(&messages)
	return messages
}
