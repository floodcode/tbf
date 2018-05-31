package tbf

import (
	"errors"
	"time"

	"github.com/floodcode/tgbot"
)

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

// WaitNextTimeout waits with timeout until new message sent in chat-user chain
// and returns request if new request was made or error if timeout was exceeded
func (r *BotRequest) WaitNextTimeout(d time.Duration) (BotRequest, error) {
	select {
	case result := <-r.BotFramework.sessions[r.session]:
		return result, nil
	case <-time.After(d):
		return BotRequest{}, errors.New("request wait timeout was exceeded")
	}
}

// QuickMessage sends text message to the origin chat
func (r *BotRequest) QuickMessage(text string) (tgbot.Message, error) {
	return r.Bot.SendMessage(tgbot.SendMessageConfig{
		ChatID: tgbot.ChatID(r.Message.Chat.ID),
		Text:   text,
	})
}

// QuickMessageMD sends text message with markdown parse mode to the origin chat
func (r *BotRequest) QuickMessageMD(text string) (tgbot.Message, error) {
	return r.Bot.SendMessage(tgbot.SendMessageConfig{
		ChatID:    tgbot.ChatID(r.Message.Chat.ID),
		Text:      text,
		ParseMode: tgbot.ParseModeMarkdown(),
	})
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
