package server

import (
	"fmt"
	"sync"
)

var ErrorOffsetNotFound = fmt.Errorf("offset not found")

type Record struct {
	Value  string `json:"value"`
	Offset int    `json:"offset"`
}

type Log struct {
	mu      sync.RWMutex
	records []Record
}

func NewLog() *Log {
	return &Log{}
}

func (l *Log) Append(r Record) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	r.Offset = len(l.records)
	l.records = append(l.records, r)
	return r.Offset, nil
}

func (l *Log) Read(offset int) (Record, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if offset >= len(l.records) {
		return Record{}, ErrorOffsetNotFound
	}
	return l.records[offset], nil
}
