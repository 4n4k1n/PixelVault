package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var users = map[int64]string{
	8969454987: "anakin",
}

var baseDir = os.Getenv("HOME") + "/PixelBackup"

func main() {
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
	dst := filepath.Join(dir, name)
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
