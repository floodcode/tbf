package tbf

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/floodcode/tgbot"
)

const (
	cmdMatchTemplate = `(?s)^\/([a-zA-Z_]+)(?:@%s)?(?:[\s\n]+(.+)|)$`
)

// TelegramBotFramework simplifies interraction with TelegramBot
type TelegramBotFramework struct {
	bot      tgbot.TelegramBot
	routes   map[string]func(BotRequest)
	cmdMatch *regexp.Regexp
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
	Bot     *tgbot.TelegramBot
	Message *tgbot.Message
	Command string
	Args    string
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
		bot:      bot,
		routes:   map[string]func(BotRequest){},
		cmdMatch: cmdMatch,
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
			f.matchRoute(update.Message)
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
		Bot:     &f.bot,
		Message: msg,
		Command: command,
		Args:    args,
	}
}

func (f *TelegramBotFramework) matchRoute(msg *tgbot.Message) {
	request := f.buildRequest(msg)
	if len(request.Command) == 0 {
		return
	}

	if action, ok := f.routes[request.Command]; ok {
		action(request)
	}
}
