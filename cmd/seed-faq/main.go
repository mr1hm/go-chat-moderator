package main

import (
	"log"

	"github.com/mr1hm/go-chat-moderator/internal/faq"
	"github.com/mr1hm/go-chat-moderator/internal/moderation/mistralai"
	"github.com/mr1hm/go-chat-moderator/internal/shared/config"
	"github.com/mr1hm/go-chat-moderator/internal/shared/sqlite"
)

var seedFAQs = []struct {
	Question string
	Answer   string
}{
	{
		Question: "How do I create a new chat room?",
		Answer:   "Log in to your account, then click the 'Create Room' button and enter a name for your room.",
	},
	{
		Question: "How does message moderation work?",
		Answer:   "All messages are automatically scanned for inappropriate content using AI. Messages flagged as harmful may be blocked.",
	},
	{
		Question: "How do I join an existing room?",
		Answer:   "View all available rooms on the main page and click on any room to join and start chatting.",
	},
	{
		Question: "What are the chat rules?",
		Answer:   "Be respectful to other users. Hate speech, harassment, and explicit content are not allowed.",
	},
	{
		Question: "How do I log out?",
		Answer:   "Click the logout button in the top right corner of the page.",
	},
}

func main() {
	dbCfg := config.LoadDBConfig()
	sqlite.Init(dbCfg.DBPath)
	defer sqlite.Close()

	mistralCfg := config.LoadMistralAIConfig()
	client := mistralai.NewClient(mistralCfg.Key)

	repo := faq.NewFAQRepository()
	service := faq.NewRAGService(repo, client)

	for _, item := range seedFAQs {
		log.Printf("Seeding: %s", item.Question)
		_, err := service.Add(item.Question, item.Answer)
		if err != nil {
			log.Printf("Failed: %v", err)
			continue
		}
		log.Println("Done")
	}

	log.Println("Seeding complete")
}
