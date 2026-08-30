package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var baseDir = os.Getenv("HOME") + "/PixelBackup"

// parseUsers reads USERS="<telegram-id>:<folder>,<telegram-id>:<folder>".
func parseUsers(s string) map[int64]string {
	users := map[int64]string{}
	for _, pair := range strings.Split(s, ",") {
		idStr, folder, ok := strings.Cut(strings.TrimSpace(pair), ":")
		if !ok {
			continue
		}
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err != nil {
			continue
		}
		users[id] = filepath.Base(strings.TrimSpace(folder))
	}
	return users
}

func main() {
	users := parseUsers(os.Getenv("USERS"))
	if len(users) == 0 {
		log.Fatal(`USERS empty, expected USERS="123456789:anakin"`)
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("running as @%s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	for update := range bot.GetUpdatesChan(u) {
		msg := update.Message
		if msg == nil {
			continue
		}
		folder, ok := users[msg.From.ID]
		if !ok {
			continue
		}

		switch {
		case msg.Document != nil:
			name := msg.Document.FileName
			if name == "" {
				name = time.Now().Format("2006-01-02_15-04-05")
			}
			reply(bot, msg, save(bot, msg.Document.FileID, folder, name))
		case msg.Photo != nil:
			reply(bot, msg, "Telegram compressed this photo. Send it as a FILE (attach > File) to keep original quality.")
		case msg.Video != nil:
			reply(bot, msg, "Send videos as FILE too. Note: bots can only download up to 20 MB.")
		}
	}
}

func save(bot *tgbotapi.BotAPI, fileID, folder, name string) string {
	f, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return "error: " + err.Error()
	}
	resp, err := http.Get(f.Link(bot.Token))
	if err != nil {
		return "download failed: " + err.Error()
	}
	defer resp.Body.Close()

	dir := filepath.Join(baseDir, folder)
	os.MkdirAll(dir, 0755)
	dst := filepath.Join(dir, filepath.Base(name))
	out, err := os.Create(dst)
	if err != nil {
		return "write failed: " + err.Error()
	}
	defer out.Close()
	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return "write failed: " + err.Error()
	}
	return fmt.Sprintf("saved %s (%.1f MB), syncing to Google Photos", name, float64(n)/1e6)
}

func reply(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, text string) {
	m := tgbotapi.NewMessage(msg.Chat.ID, text)
	m.ReplyToMessageID = msg.MessageID
	bot.Send(m)
}
