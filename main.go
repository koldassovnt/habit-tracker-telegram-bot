package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/koldassovnt/habit-tracker-telegram-bot/config"
	"github.com/koldassovnt/habit-tracker-telegram-bot/db"
	"github.com/koldassovnt/habit-tracker-telegram-bot/inputs"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	pool, err := db.Connect(ctx, cfg.Database.PgxDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	store := db.NewStore(pool)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutting down...")
		cancel()
		pool.Close()
		os.Exit(0)
	}()

	b, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		log.Fatal(err)
	}

	setCommands(b)

	go inputs.StartScheduler(ctx, b, store)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.GetUpdatesChan(u)

	log.Printf("Bot @%s is running...", b.Self.UserName)

	for update := range updates {
		go inputs.HandleUpdate(ctx, b, store, update)
	}
}

func setCommands(b *tgbotapi.BotAPI) {
	commands := []tgbotapi.BotCommand{
		{Command: "managecategory", Description: "Add, edit, delete category"},
		{Command: "managehabit", Description: "Add, edit, delete habit"},
		{Command: "trackhabit", Description: "Track a habit"},
		{Command: "trackpast", Description: "Track a habit for a past day"},
		{Command: "untrack", Description: "Remove a tracked habit"},
		{Command: "todaystatus", Description: "Status of tracked habits for today"},
		{Command: "paststatus", Description: "Status of tracked habits for a past day"},
		{Command: "help", Description: "Show available commands"},
	}
	if _, err := b.Request(tgbotapi.NewSetMyCommands(commands...)); err != nil {
		log.Printf("Failed to set commands: %v", err)
	}
}
