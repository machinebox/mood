package textboxtally

import (
	"sort"
	"sync"

	"github.com/machinebox/sdk-go/textbox"
)

// Tally keeps a summary of textbox.Analysis objects.
type Tally struct {
	lock     sync.RWMutex // protects Tally
	keywords map[string]int
	entities map[string]map[string]int

	count int

	sentimentAverageSum   float64
	sentimentAverageTotal float64
}

// New makes a new Tally.
func New() *Tally {
	return &Tally{
		keywords: make(map[string]int),
		entities: make(map[string]map[string]int),
	}
}

// Add adds the textbox.Analysis to this tally.
func (t *Tally) Add(analysis *textbox.Analysis) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.count++
	for _, keyword := range analysis.Keywords {
		t.keywords[keyword.Keyword]++
	}
	for _, sentence := range analysis.Sentences {
		t.sentimentAverageSum += sentence.Sentiment
		t.sentimentAverageTotal++
		for _, entity := range sentence.Entities {
			if _, ok := t.entities[entity.Type]; !ok {
				t.entities[entity.Type] = make(map[string]int)
			}
			t.entities[entity.Type][entity.Text]++
		}
	}
}

// AllKeywords gets an unordered list of keywords.
//
// May be sorted with:
// 	sort.SliceStable(keywords, func(i, j int) bool {
// 		return keywords[j].Count < keywords[i].Count
// 	})
func (t *Tally) AllKeywords() []Keyword {
	t.lock.RLock()
	defer t.lock.RUnlock()
	keywords := make([]Keyword, len(t.keywords))
	i := 0
	for k, c := range t.keywords {
		keywords[i] = Keyword{
			Keyword: k,
			Count:   c,
		}
		i++
	}
	return keywords
}

// TopKeywords gets the top ten keywords sorted by most frequent.
func (t *Tally) TopKeywords() []Keyword {
	keywords := t.AllKeywords()
	sort.SliceStable(keywords, func(i, j int) bool {
		return keywords[j].Count < keywords[i].Count
	})
	if len(keywords) > 10 {
		return keywords[:10]
	}
	return keywords
}

// AllEntities gets all entities mapped by their type.
func (t *Tally) AllEntities() map[string][]Entity {
	t.lock.RLock()
	defer t.lock.RUnlock()
	ents := make(map[string][]Entity)
	for typ, entities := range t.entities {
		for entityText, count := range entities {
			ents[typ] = append(ents[typ], Entity{
				Text:  entityText,
				Count: count,
			})
		}
	}
	return ents
}

// TopEntities gets the top ten entities sorted by most frequent,
// mapped on their type.
func (t *Tally) TopEntities() map[string][]Entity {
	ents := t.AllEntities()
	for typ, entities := range ents {
		sort.SliceStable(entities, func(i, j int) bool {
			return entities[j].Count < entities[i].Count
		})
		if len(entities) > 10 {
			ents[typ] = entities[:10]
		}
	}
	return ents
}

// SentimentAverage is the average sentiment of all sentences seen so far.
func (t *Tally) SentimentAverage() float64 {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.sentimentAverageSum / t.sentimentAverageTotal
}

// Count gets the number of textbox.Analysis objects that have
// been included.
func (t *Tally) Count() int {
	var c int
	t.lock.RLock()
	c = t.count
	t.lock.RUnlock()
	return c
}

// Keyword represents a keyword with frequency.
type Keyword struct {
	Keyword string `json:"keyword"`
	Count   int    `json:"count"`
}

// Entity represents a single entity.
type Entity struct {
	Text  string `json:"text"`
	Count int    `json:"count"`
}
