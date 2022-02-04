package main

import (
	"sync"
)

type PriorityQueue struct {
	items   []int64
	itemSet map[int64]struct{}
	lock    sync.RWMutex
}

func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{
		items:   []int64{},
		itemSet: make(map[int64]struct{}),
		lock:    sync.RWMutex{},
	}
}

func (q *PriorityQueue) Add(id int64) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if _, ok := q.itemSet[id]; ok {
		return
	}
	q.items = append(q.items, id)
	q.itemSet[id] = struct{}{}
}

func (q *PriorityQueue) Update(id int64) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if _, ok := q.itemSet[id]; !ok {
		return
	}

	pos := 0
	// find the element
	for i, item := range q.items {
		if item == id {
			pos = i
			break
		}
	}

	target := 0
	val := q.items[pos]
	q.items = append(q.items[:pos], q.items[pos+1:]...)
	newSlice := make([]int64, target+1)
	copy(newSlice, q.items[:target])
	newSlice[target] = val
	q.items = append(newSlice, q.items[target:]...)
}

func (q *PriorityQueue) Get() []int64 {
	return q.items
}
