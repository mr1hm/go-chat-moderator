package faq

import (
	"sort"
	"sync"
	"sync/atomic"

	"github.com/viterin/vek"
)

type EmbeddingCache struct {
	mtx        sync.RWMutex
	embeddings map[string][]float64
	faqIDs     []string
	ready      atomic.Bool
}

func NewEmbeddingCache() *EmbeddingCache {
	return &EmbeddingCache{}
}

func (c *EmbeddingCache) IsReady() bool {
	return c.ready.Load()
}

func (c *EmbeddingCache) Load(faqs []*FAQ) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.embeddings = make(map[string][]float64, len(faqs))
	c.faqIDs = make([]string, 0, len(faqs))

	for _, faq := range faqs {
		if len(faq.Embedding) > 0 {
			c.embeddings[faq.ID] = faq.Embedding
			c.faqIDs = append(c.faqIDs, faq.ID)
		}
	}
	c.ready.Store(true)
}

func (c *EmbeddingCache) FindTopK(queryEmbedding []float64, k int) []string {
	c.mtx.RLock()
	defer c.mtx.RUnlock()

	type scored struct {
		id    string
		score float64
	}

	scores := make([]scored, 0, len(c.faqIDs))
	for _, id := range c.faqIDs {
		emb := c.embeddings[id]
		dot := vek.Dot(queryEmbedding, emb)
		normQ := vek.Norm(queryEmbedding)
		normE := vek.Norm(emb)
		if normQ > 0 && normE > 0 {
			score := dot / (normQ * normE)
			scores = append(scores, scored{id, score})
		}
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	result := make([]string, 0, k)
	for i := 0; i < k && i < len(scores); i++ {
		result = append(result, scores[i].id)
	}

	return result
}
