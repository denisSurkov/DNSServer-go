package structures

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type QueryCache struct {
	mutex sync.Mutex
	items map[string]retrievedAnswer
}

func NewQueryCache() *QueryCache {
	return &QueryCache{
		items: make(map[string]retrievedAnswer),
	}
}

type retrievedAnswer struct {
	answers     []*DNSRecord
	retrievedAt time.Time
}

func (q *QueryCache) Get(question *DNSQuestion) ([]*DNSRecord, bool) {
	questionString := makeQuestionString(question)

	q.mutex.Lock()
	retrieved, ok := q.items[questionString]
	q.mutex.Unlock()

	if !ok {
		return nil, false
	}

	var totalAnswers []*DNSRecord
	currentTime := time.Now()
	for _, answer := range retrieved.answers {
		timeToLive := int(answer.TimeToLive)
		if retrieved.retrievedAt.Add(time.Second * time.Duration(timeToLive)).After(currentTime) {
			totalAnswers = append(totalAnswers, answer)
		}
	}

	if totalAnswers != nil {
		return totalAnswers, true
	}

	return nil, false
}

func (q *QueryCache) Set(question *DNSQuestion, answers []*DNSRecord) {
	retrivedAt := time.Now()
	item := retrievedAnswer{
		answers:     answers,
		retrievedAt: retrivedAt,
	}
	questionString := makeQuestionString(question)

	q.mutex.Lock()
	q.items[questionString] = item
	q.mutex.Unlock()
}

func makeQuestionString(question *DNSQuestion) string {
	str := fmt.Sprintf("%s-%b-%b", question.QName, question.QType, question.QClass)
	log.Println(str)
	return str
}
