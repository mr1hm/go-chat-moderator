package faq

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mr1hm/go-chat-moderator/internal/moderation/mistralai"
)

// Mock repository implementation
type mockFAQRepo struct {
	faqs      map[string]*FAQ
	createErr error
	deleteErr error
	listErr   error
}

func newMockRepo() *mockFAQRepo {
	return &mockFAQRepo{
		faqs: make(map[string]*FAQ),
	}
}

func (m *mockFAQRepo) Create(faq *FAQ) error {
	if m.createErr != nil {
		return m.createErr
	}
	if faq.ID == "" {
		faq.ID = "generated-id"
	}
	faq.CreatedAt = time.Now()
	m.faqs[faq.ID] = faq
	return nil
}

func (m *mockFAQRepo) List() ([]*FAQ, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	result := make([]*FAQ, 0, len(m.faqs))
	for _, f := range m.faqs {
		result = append(result, f)
	}
	return result, nil
}

func (m *mockFAQRepo) FindByID(id string) (*FAQ, error) {
	if f, ok := m.faqs[id]; ok {
		return f, nil
	}
	return nil, ErrFAQNotFound
}

func (m *mockFAQRepo) Delete(id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.faqs[id]; !ok {
		return ErrFAQNotFound
	}
	delete(m.faqs, id)
	return nil
}

// Helper to create mock Mistral server that handles both embed and chat
func createMockMistralServer(t *testing.T, embedResponse [][]float64, chatResponse string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check which endpoint is being called based on request body
		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)

		if _, hasInput := reqBody["input"]; hasInput {
			// Embed request
			response := struct {
				Data []struct {
					Embedding []float64 `json:"embedding"`
					Index     int       `json:"index"`
				} `json:"data"`
			}{}

			for i, emb := range embedResponse {
				response.Data = append(response.Data, struct {
					Embedding []float64 `json:"embedding"`
					Index     int       `json:"index"`
				}{Embedding: emb, Index: i})
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}

		if _, hasMessages := reqBody["messages"]; hasMessages {
			// Chat request
			response := struct {
				Choices []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				} `json:"choices"`
			}{
				Choices: []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				}{
					{Message: struct {
						Content string `json:"content"`
					}{Content: chatResponse}},
				},
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
	}))
}

func TestRAGService_LoadEmbeddings(t *testing.T) {
	repo := newMockRepo()
	repo.faqs["1"] = &FAQ{ID: "1", Question: "Q1", Embedding: []float64{0.1, 0.2, 0.3}}
	repo.faqs["2"] = &FAQ{ID: "2", Question: "Q2", Embedding: []float64{0.4, 0.5, 0.6}}

	client := mistralai.NewClient("fake-key")
	service := NewRAGService(repo, client)

	err := service.LoadEmbeddings()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !service.IsReady() {
		t.Error("expected service to be ready after LoadEmbeddings")
	}
}

func TestRAGService_IsReady(t *testing.T) {
	repo := newMockRepo()
	client := mistralai.NewClient("fake-key")
	service := NewRAGService(repo, client)

	if service.IsReady() {
		t.Error("expected service to not be ready before LoadEmbeddings")
	}

	service.LoadEmbeddings()

	if !service.IsReady() {
		t.Error("expected service to be ready after LoadEmbeddings")
	}
}

func TestRAGService_ListFAQs(t *testing.T) {
	repo := newMockRepo()
	repo.faqs["1"] = &FAQ{ID: "1", Question: "Q1", Answer: "A1"}
	repo.faqs["2"] = &FAQ{ID: "2", Question: "Q2", Answer: "A2"}

	client := mistralai.NewClient("fake-key")
	service := NewRAGService(repo, client)

	faqs, err := service.ListFAQs()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(faqs) != 2 {
		t.Errorf("expected 2 FAQs, got %d", len(faqs))
	}
}

func TestRAGService_DeleteFAQ_Success(t *testing.T) {
	repo := newMockRepo()
	repo.faqs["1"] = &FAQ{ID: "1", Question: "Q1", Embedding: []float64{0.1, 0.2}}

	client := mistralai.NewClient("fake-key")
	service := NewRAGService(repo, client)
	service.LoadEmbeddings()

	err := service.DeleteFAQ("1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(repo.faqs) != 0 {
		t.Error("expected FAQ to be deleted")
	}
}

func TestRAGService_DeleteFAQ_NotFound(t *testing.T) {
	repo := newMockRepo()
	client := mistralai.NewClient("fake-key")
	service := NewRAGService(repo, client)

	err := service.DeleteFAQ("nonexistent")
	if err != ErrFAQNotFound {
		t.Errorf("expected ErrFAQNotFound, got %v", err)
	}
}
