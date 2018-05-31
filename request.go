package tbf

import "github.com/floodcode/tgbot"

// BotRequest contains basic information about bot request
type BotRequest struct {
	BotFramework *TelegramBotFramework
	Bot          *tgbot.TelegramBot
	Message      *tgbot.Message
	Command      string
	Args         string
	session      string
}

// WaitNext waits until new message sent in chat-user chain and returns request
func (r *BotRequest) WaitNext() BotRequest {
	result := <-r.BotFramework.sessions[r.session]
	return result
}

// QuickReply sends text message in reply to the origin message
func (r *BotRequest) QuickReply(text string) (tgbot.Message, error) {
	return r.Bot.SendMessage(tgbot.SendMessageConfig{
		ChatID:           tgbot.ChatID(r.Message.Chat.ID),
		Text:             text,
		ReplyToMessageID: r.Message.MessageID,
	})
}

// QuickReplyMD sends text message with markdown parse mode in reply to the origin message
func (r *BotRequest) QuickReplyMD(text string) (tgbot.Message, error) {
	return r.Bot.SendMessage(tgbot.SendMessageConfig{
		ChatID:           tgbot.ChatID(r.Message.Chat.ID),
		Text:             text,
		ReplyToMessageID: r.Message.MessageID,
		ParseMode:        tgbot.ParseModeMarkdown(),
	})
}

// SendTyping sends chat action "typing" to the origin chat
func (r *BotRequest) SendTyping() (bool, error) {
	return r.Bot.SendChatAction(tgbot.SendChatActionConfig{
		ChatID: tgbot.ChatID(r.Message.Chat.ID),
		Action: tgbot.ChatActionTyping(),
	})
}
