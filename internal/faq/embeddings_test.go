package faq

import (
	"testing"
)

func TestEmbeddingCache_IsReady_BeforeLoad(t *testing.T) {
	cache := NewEmbeddingCache()

	if cache.IsReady() {
		t.Error("expected cache to not be ready before Load")
	}
}

func TestEmbeddingCache_IsReady_AfterLoad(t *testing.T) {
	cache := NewEmbeddingCache()

	faqs := []*FAQ{
		{ID: "1", Question: "test", Embedding: []float64{0.1, 0.2, 0.3}},
	}

	cache.Load(faqs)

	if !cache.IsReady() {
		t.Error("expected cache to be ready after Load")
	}
}

func TestEmbeddingCache_Load(t *testing.T) {
	cache := NewEmbeddingCache()

	faqs := []*FAQ{
		{ID: "1", Question: "Q1", Embedding: []float64{0.1, 0.2, 0.3}},
		{ID: "2", Question: "Q2", Embedding: []float64{0.4, 0.5, 0.6}},
		{ID: "3", Question: "Q3", Embedding: nil}, // No embedding - should be skipped
	}

	cache.Load(faqs)

	if len(cache.faqIDs) != 2 {
		t.Errorf("expected 2 FAQs loaded, got %d", len(cache.faqIDs))
	}

	if len(cache.embeddings) != 2 {
		t.Errorf("expected 2 embeddings, got %d", len(cache.embeddings))
	}
}

func TestEmbeddingCache_FindTopK(t *testing.T) {
	cache := NewEmbeddingCache()

	// Create FAQs with embeddings that have known similarity to query
	// Query will be [1, 0, 0] - unit vector in x direction
	// FAQ1: [1, 0, 0] - identical, similarity = 1.0
	// FAQ2: [0, 1, 0] - orthogonal, similarity = 0.0
	// FAQ3: [0.707, 0.707, 0] - 45 degrees, similarity ≈ 0.707
	faqs := []*FAQ{
		{ID: "faq1", Question: "Q1", Embedding: []float64{1, 0, 0}},
		{ID: "faq2", Question: "Q2", Embedding: []float64{0, 1, 0}},
		{ID: "faq3", Question: "Q3", Embedding: []float64{0.707, 0.707, 0}},
	}

	cache.Load(faqs)

	query := []float64{1, 0, 0}
	results := cache.FindTopK(query, 2)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// First result should be faq1 (most similar)
	if results[0] != "faq1" {
		t.Errorf("expected first result to be faq1, got %s", results[0])
	}

	// Second result should be faq3 (second most similar)
	if results[1] != "faq3" {
		t.Errorf("expected second result to be faq3, got %s", results[1])
	}
}

func TestEmbeddingCache_FindTopK_EmptyCache(t *testing.T) {
	cache := NewEmbeddingCache()
	cache.Load([]*FAQ{}) // Empty load

	query := []float64{1, 0, 0}
	results := cache.FindTopK(query, 3)

	if len(results) != 0 {
		t.Errorf("expected 0 results for empty cache, got %d", len(results))
	}
}

func TestEmbeddingCache_FindTopK_LessThanK(t *testing.T) {
	cache := NewEmbeddingCache()

	faqs := []*FAQ{
		{ID: "1", Question: "Q1", Embedding: []float64{1, 0, 0}},
		{ID: "2", Question: "Q2", Embedding: []float64{0, 1, 0}},
	}

	cache.Load(faqs)

	query := []float64{1, 0, 0}
	results := cache.FindTopK(query, 5) // Ask for 5, but only 2 exist

	if len(results) != 2 {
		t.Errorf("expected 2 results (all available), got %d", len(results))
	}
}

func TestEmbeddingCache_FindTopK_AllSameSimilarity(t *testing.T) {
	cache := NewEmbeddingCache()

	// All embeddings are the same - all have same similarity to query
	faqs := []*FAQ{
		{ID: "1", Question: "Q1", Embedding: []float64{1, 1, 1}},
		{ID: "2", Question: "Q2", Embedding: []float64{1, 1, 1}},
		{ID: "3", Question: "Q3", Embedding: []float64{1, 1, 1}},
	}

	cache.Load(faqs)

	query := []float64{1, 1, 1}
	results := cache.FindTopK(query, 2)

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}
