package tbf

import (
	"errors"
	"time"

	"github.com/floodcode/tgbot"
)

// Request contains basic information about bot request
type Request struct {
	BotFramework *TelegramBotFramework
	Bot          *tgbot.TelegramBot
	Message      *tgbot.Message
	Command      string
	Args         string
	Session      string
}

// WaitNext waits until new message sent in chat-user chain and returns request
func (r Request) WaitNext() Request {
	result := <-r.BotFramework.sessions[r.Session]
	return result
}

// WaitNextTimeout waits with timeout until new message sent in chat-user chain
// and returns request if new request was made or error if timeout was exceeded
func (r Request) WaitNextTimeout(d time.Duration) (Request, error) {
	select {
	case result := <-r.BotFramework.sessions[r.Session]:
		return result, nil
	case <-time.After(d):
		return Request{}, errors.New("request wait timeout was exceeded")
	}
}

// SendMessage sends text message with additional parameters to the origin chat
func (r Request) SendMessage(config tgbot.SendMessageConfig) (tgbot.Message, error) {
	config.ChatID = tgbot.ChatID(r.Message.Chat.ID)
	return r.Bot.SendMessage(config)
}

// SendReply sends text message with additional parameters in reply to the origin message
func (r Request) SendReply(config tgbot.SendMessageConfig) (tgbot.Message, error) {
	config.ChatID = tgbot.ChatID(r.Message.Chat.ID)
	config.ReplyToMessageID = r.Message.MessageID
	return r.Bot.SendMessage(config)
}

// QuickMessage sends text message to the origin chat
func (r Request) QuickMessage(text string) (tgbot.Message, error) {
	return r.Bot.SendMessage(tgbot.SendMessageConfig{
		ChatID: tgbot.ChatID(r.Message.Chat.ID),
		Text:   text,
	})
}

// QuickMessageMD sends text message with markdown parse mode to the origin chat
func (r Request) QuickMessageMD(text string) (tgbot.Message, error) {
	return r.Bot.SendMessage(tgbot.SendMessageConfig{
		ChatID:    tgbot.ChatID(r.Message.Chat.ID),
		Text:      text,
		ParseMode: tgbot.ParseModeMarkdown(),
	})
}

// QuickReply sends text message in reply to the origin message
func (r Request) QuickReply(text string) (tgbot.Message, error) {
	return r.Bot.SendMessage(tgbot.SendMessageConfig{
		ChatID:           tgbot.ChatID(r.Message.Chat.ID),
		Text:             text,
		ReplyToMessageID: r.Message.MessageID,
	})
}

// QuickReplyMD sends text message with markdown parse mode in reply to the origin message
func (r Request) QuickReplyMD(text string) (tgbot.Message, error) {
	return r.Bot.SendMessage(tgbot.SendMessageConfig{
		ChatID:           tgbot.ChatID(r.Message.Chat.ID),
		Text:             text,
		ReplyToMessageID: r.Message.MessageID,
		ParseMode:        tgbot.ParseModeMarkdown(),
	})
}

// SendTyping sends chat action "typing" to the origin chat
func (r Request) SendTyping() (bool, error) {
	return r.Bot.SendChatAction(tgbot.SendChatActionConfig{
		ChatID: tgbot.ChatID(r.Message.Chat.ID),
		Action: tgbot.ChatActionTyping(),
	})
}

// CallbackQueryRequest contains basic information about bot callback query request
type CallbackQueryRequest struct {
	BotFramework  *TelegramBotFramework
	Bot           *tgbot.TelegramBot
	CallbackQuery *tgbot.CallbackQuery
}

// NoAnswer sends empty answer to the origin callback query
func (r CallbackQueryRequest) NoAnswer() (bool, error) {
	return r.Bot.AnswerCallbackQuery(tgbot.AnswerCallbackQueryConfig{})
}

// Answer sends answer to the origin callback query
func (r CallbackQueryRequest) Answer(config tgbot.AnswerCallbackQueryConfig) (bool, error) {
	config.CallbackQueryID = r.CallbackQuery.ID
	return r.Bot.AnswerCallbackQuery(config)
}
