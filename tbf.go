package tbf

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/floodcode/tgbot"
)

const (
	cmdMatchTemplate = `(?s)^\/([a-zA-Z_]+)(?:@%s)?(?:[\s\n]+(.+)|)$`
)

// TelegramBotFramework simplifies interraction with TelegramBot
type TelegramBotFramework struct {
	bot           tgbot.TelegramBot
	routes        map[string]func(BotRequest)
	sessions      map[string]chan BotRequest
	sessionsMutex sync.Mutex
	cmdMatch      *regexp.Regexp
}

// PollConfig represents bot's polling configuration
type PollConfig struct {
	Delay int
}

// ListenConfig represents bot's webhook configuration
type ListenConfig struct {
	Host           string
	Port           uint16
	KeyFilename    string
	CertFilename   string
	MaxConnections int
	AllowedUpdates []string
}

// BotRequest contains basic information about bot request
type BotRequest struct {
	BotFramework *TelegramBotFramework
	Bot          *tgbot.TelegramBot
	Message      *tgbot.Message
	Command      string
	Args         string
	session      string
}

// WaitRequest waits until new message sent in chat-user chain and returns request
func (r *BotRequest) WaitRequest() BotRequest {
	result := <-r.BotFramework.sessions[r.session]
	return result
}

// New returns new TelegramBotFramework instance
func New(apiKey string) (TelegramBotFramework, error) {
	bot, err := tgbot.New(apiKey)
	if err != nil {
		return TelegramBotFramework{}, err
	}

	botUser, err := bot.GetMe()
	if err != nil {
		return TelegramBotFramework{}, err
	}

	cmdMatch := regexp.MustCompile(fmt.Sprintf(cmdMatchTemplate, botUser.Username))
	return TelegramBotFramework{
		bot:           bot,
		routes:        map[string]func(BotRequest){},
		sessions:      map[string]chan BotRequest{},
		sessionsMutex: sync.Mutex{},
		cmdMatch:      cmdMatch,
	}, nil
}

// Poll starts updates polling
func (f *TelegramBotFramework) Poll(config PollConfig) error {
	return f.bot.Poll(tgbot.PollConfig{
		Callback: f.updatesCallback,
		Delay:    config.Delay,
	})
}

// Listen starts HTTPS server to receive updates
func (f *TelegramBotFramework) Listen(config ListenConfig) error {
	return f.bot.Listen(tgbot.ListenConfig{
		Callback:       f.updatesCallback,
		Host:           config.Host,
		Port:           config.Port,
		KeyFilename:    config.KeyFilename,
		CertFilename:   config.CertFilename,
		MaxConnections: config.MaxConnections,
		AllowedUpdates: config.AllowedUpdates,
	})
}

// AddRoute is used to register new command with callback
func (f *TelegramBotFramework) AddRoute(command string, action func(BotRequest)) {
	f.routes[strings.ToLower(command)] = action
}

func (f *TelegramBotFramework) updatesCallback(updates []tgbot.Update) {
	for _, update := range updates {
		f.processUpdate(update)
	}
}

func (f *TelegramBotFramework) processUpdate(update tgbot.Update) {
	if update.Message != nil {
		if update.Message.Text != "" {
			f.handleRequest(f.buildRequest(update.Message))
		}
	}
}

func (f *TelegramBotFramework) buildRequest(msg *tgbot.Message) BotRequest {
	var command string
	var args string
	match := f.cmdMatch.FindStringSubmatch(msg.Text)
	if match != nil {
		command = strings.ToLower(strings.TrimSpace(match[1]))
		args = strings.TrimSpace(match[2])
	}

	return BotRequest{
		BotFramework: f,
		Bot:          &f.bot,
		Message:      msg,
		Command:      command,
		Args:         args,
		session:      fmt.Sprintf("%d:%d", msg.Chat.ID, msg.From.ID),
	}
}

func (f *TelegramBotFramework) handleRequest(request BotRequest) {
	f.sessionsMutex.Lock()
	if _, ok := f.sessions[request.session]; ok {
		f.sessions[request.session] <- request
		f.sessionsMutex.Unlock()
		return
	}

	if len(request.Command) == 0 {
		f.sessionsMutex.Unlock()
		return
	}

	f.sessions[request.session] = make(chan BotRequest, 10)
	f.sessions[request.session] <- request
	f.sessionsMutex.Unlock()

	f.runAction(request.session)
}

func (f *TelegramBotFramework) runAction(session string) {
	for {
		f.sessionsMutex.Lock()
		select {
		case req := <-f.sessions[session]:
			f.sessionsMutex.Unlock()
			if action, ok := f.routes[req.Command]; ok {
				action(req)
			} else {
				// No such command
			}
		default:
			close(f.sessions[session])
			delete(f.sessions, session)
			f.sessionsMutex.Unlock()
			return
		}
	}
}
