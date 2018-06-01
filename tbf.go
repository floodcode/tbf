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
	bot                   tgbot.TelegramBot
	routes                map[string]func(Request)
	sessions              map[string]chan Request
	sessionsMutex         sync.Mutex
	callbackQueryListener func(CallbackQueryRequest)
	cmdMatch              *regexp.Regexp
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
		bot:                   bot,
		routes:                map[string]func(Request){},
		sessions:              map[string]chan Request{},
		sessionsMutex:         sync.Mutex{},
		callbackQueryListener: func(CallbackQueryRequest) {},
		cmdMatch:              cmdMatch,
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
func (f *TelegramBotFramework) AddRoute(command string, action func(Request)) {
	f.routes[strings.ToLower(command)] = action
}

// OnCallbackQuery sets callback query listener
func (f *TelegramBotFramework) OnCallbackQuery(listener func(CallbackQueryRequest)) {
	f.callbackQueryListener = listener
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
	} else if update.CallbackQuery != nil {
		f.callbackQueryListener(f.buildCallbackQueryRequest(update.CallbackQuery))
	}
}

func (f *TelegramBotFramework) buildRequest(
	msg *tgbot.Message) Request {
	var command string
	var args string
	match := f.cmdMatch.FindStringSubmatch(msg.Text)
	if match != nil {
		command = strings.ToLower(strings.TrimSpace(match[1]))
		args = strings.TrimSpace(match[2])
	}

	return Request{
		BotFramework: f,
		Bot:          &f.bot,
		Message:      msg,
		Command:      command,
		Args:         args,
		Session:      fmt.Sprintf("%d:%d", msg.Chat.ID, msg.From.ID),
	}
}

func (f *TelegramBotFramework) buildCallbackQueryRequest(
	query *tgbot.CallbackQuery) CallbackQueryRequest {
	return CallbackQueryRequest{
		BotFramework:  f,
		Bot:           &f.bot,
		CallbackQuery: query,
	}
}

func (f *TelegramBotFramework) handleRequest(request Request) {
	f.sessionsMutex.Lock()
	if _, ok := f.sessions[request.Session]; ok {
		f.sessions[request.Session] <- request
		f.sessionsMutex.Unlock()
		return
	}

	if len(request.Command) == 0 {
		f.sessionsMutex.Unlock()
		return
	}

	f.sessions[request.Session] = make(chan Request, 10)
	f.sessions[request.Session] <- request
	f.sessionsMutex.Unlock()

	go f.runAction(request.Session)
}

func (f *TelegramBotFramework) runAction(session string) {
	for {
		select {
		case req := <-f.sessions[session]:
			if action, ok := f.routes[req.Command]; ok {
				action(req)
			} else {
				// No such command
			}
		default:
			f.sessionsMutex.Lock()
			close(f.sessions[session])
			delete(f.sessions, session)
			f.sessionsMutex.Unlock()
			return
		}
	}
}
