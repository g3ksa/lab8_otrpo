package main

import (
	"github.com/joho/godotenv"
	"log"
	"net/smtp"
	"os"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	userEmailMap = make(map[int64]string)
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file")
	}
}

func isValidEmail(email string) bool {
	regex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(regex)
	return re.MatchString(email)
}

func sendEmail(to string, message string) error {
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpEmail := os.Getenv("SMTP_EMAIL")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	auth := smtp.PlainAuth("", smtpEmail, smtpPassword, strings.Split(smtpServer, ":")[0])
	msg := "From: " + smtpEmail + "\n" +
		"To: " + to + "\n" +
		"Subject: Сообщение от Telegram-бота\n\n" +
		message
	return smtp.SendMail(smtpServer, auth, smtpEmail, []string{to}, []byte(msg))
}

func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is not set")
	}
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		panic(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := update.Message.Text

		if text == "/start" {
			bot.Send(tgbotapi.NewMessage(chatID, "Добро пожаловать! Пожалуйста, введите ваш email:"))
			continue
		}

		if email, exists := userEmailMap[chatID]; exists {
			err := sendEmail(email, text)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при отправке сообщения."))
				log.Printf("Failed to send email: %s", err)
			} else {
				bot.Send(tgbotapi.NewMessage(chatID, "Сообщение успешно отправлено!"))
			}
			delete(userEmailMap, chatID)
		} else {
			if isValidEmail(text) {
				userEmailMap[chatID] = text
				bot.Send(tgbotapi.NewMessage(chatID, "Email подтвержден. Теперь введите текст сообщения:"))
			} else {
				bot.Send(tgbotapi.NewMessage(chatID, "Некорректный email. Попробуйте еще раз:"))
			}
		}
	}
}
