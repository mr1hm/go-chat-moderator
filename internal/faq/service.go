package faq

import (
	"fmt"
	"log"
	"strings"

	"github.com/mr1hm/go-chat-moderator/internal/moderation/mistralai"
)

type RAGService struct {
	repo          FAQRepository
	mistralClient *mistralai.Client
	cache         *EmbeddingCache
}

func NewRAGService(repo FAQRepository, client *mistralai.Client) *RAGService {
	return &RAGService{
		repo:          repo,
		mistralClient: client,
		cache:         NewEmbeddingCache(),
	}
}

func (s *RAGService) LoadEmbeddings() error {
	faqs, err := s.repo.List()
	if err != nil {
		return fmt.Errorf("failed to load FAQs: %w", err)
	}

	s.cache.Load(faqs)
	return nil
}

func (s *RAGService) IsReady() bool {
	return s.cache.IsReady()
}

func (s *RAGService) Ask(question string) (*AskResponse, error) {
	// Embed question
	embeddings, err := s.mistralClient.Embed([]string{question})
	if err != nil {
		return nil, fmt.Errorf("failed to embed question: %w", err)
	}
	queryEmbedding := embeddings[0]

	// Get top 3 similar results
	topIDs := s.cache.FindTopK(queryEmbedding, 3)

	var contextParts []string
	var sources []string
	for _, id := range topIDs {
		faq, err := s.repo.FindByID(id)
		if err != nil {
			log.Printf("error finding by ID: %v - Continuing to next ID...", err)
			continue
		}

		contextParts = append(contextParts, fmt.Sprintf("Q: %s\nA: %s", faq.Question, faq.Answer))
		sources = append(sources, faq.Question)
	}

	// Generate answer with RAG prompt
	systemPrompt := `You are a helpful FAQ assistant for a chat application called Go Chat Moderator.
		Answer questions based ONLY on the provided FAQ context. If the context doesn't contain
		relevant information, say "I don't have information about that in my FAQ database."
		Keep answers concise and helpful.`

	userPrompt := fmt.Sprintf(`
		FAQ Context: 
		---
		%s
		---
		User Question: %s
		Please answer based on the FAQ context above.`, strings.Join(contextParts, "\n\n"), question)

	answer, err := s.mistralClient.Chat([]mistralai.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate answer: %w", err)
	}

	return &AskResponse{
		Answer:  answer,
		Sources: sources,
	}, nil
}

func (s *RAGService) Add(question, answer string) (*FAQ, error) {
	embeddings, err := s.mistralClient.Embed([]string{question})
	if err != nil {
		return nil, fmt.Errorf("failed to embed FAQ: %w", err)
	}

	faq := &FAQ{
		Question:  question,
		Answer:    answer,
		Embedding: embeddings[0],
	}

	if err := s.repo.Create(faq); err != nil {
		return nil, err
	}

	s.LoadEmbeddings()
	return faq, nil
}

func (s *RAGService) DeleteFAQ(id string) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	s.LoadEmbeddings()
	return nil
}

func (s *RAGService) ListFAQs() ([]*FAQ, error) {
	return s.repo.List()
}
