package cache

import (
	"strings"
	"sync"
	"time"
)

// Subtopics contains a list of subtopics in map form with addition to mutex locking
// The subtopicsMap is used to keep a record of subtopics to be later used to generate the subtopics id `query` for a topic
// and to check if the subtopic id given by a user exists
type Subtopics struct {
	mutex        *sync.RWMutex
	subtopicsMap map[string]Subtopic
}

// Subtopic represents the data which is cached for a subtopic to be used by the dp-search-data-extractor
type Subtopic struct {
	ID              string
	LocaliseKeyName string
	Slug            string
	ReleaseDate     *time.Time
}

// GetNewSubTopicsMap creates a new subtopics id map to store subtopic ids with addition to mutex locking
func NewSubTopicsMap() *Subtopics {
	return &Subtopics{
		mutex:        &sync.RWMutex{},
		subtopicsMap: make(map[string]Subtopic),
	}
}

// Get returns subtopic for given key (id)
func (sts *Subtopics) Get(key string) Subtopic {
	sts.mutex.RLock()
	defer sts.mutex.RUnlock()

	return sts.subtopicsMap[key]
}

// GetSubtopics returns an array of subtopics
func (sts *Subtopics) GetSubtopics() (subtopics []Subtopic) {
	if sts.subtopicsMap == nil {
		return
	}

	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	for _, subtopic := range sts.subtopicsMap {
		subtopics = append(subtopics, subtopic)
	}

	return subtopics
}

// CheckTopicIDExists returns subtopic for given key (id)
func (sts *Subtopics) CheckTopicIDExists(key string) bool {
	sts.mutex.RLock()
	defer sts.mutex.RUnlock()

	if _, ok := sts.subtopicsMap[key]; !ok {
		return false
	}

	return true
}

// GetSubtopicsIDsQuery gets the subtopics ID query for a topic
func (sts *Subtopics) GetSubtopicsIDsQuery() string {
	sts.mutex.RLock()
	defer sts.mutex.RUnlock()

	ids := make([]string, 0, len(sts.subtopicsMap))

	for id := range sts.subtopicsMap {
		ids = append(ids, id)
	}

	return strings.Join(ids, ",")
}

// AppendSubtopicID appends the subtopic id to the map stored in SubtopicsIDs with consideration to mutex locking
func (sts *Subtopics) AppendSubtopicID(id string, subtopic Subtopic) {
	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	if sts.subtopicsMap == nil {
		sts.subtopicsMap = make(map[string]Subtopic)
	}

	sts.subtopicsMap[id] = subtopic
}
