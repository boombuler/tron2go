package main

import (
    "log"
    "sync"
)

type InputNode struct {
    Value Direction
}

// Queue is a basic FIFO queue based on a circular list that resizes as needed.
type InputQueue struct {
    nodes   []*InputNode
    head    int
    tail    int
    count   int
    mutex   *sync.Mutex
}

// Push adds a node to the queue.
func (q *InputQueue) Push(dir Direction) {
    if q.mutex != nil {
        q.mutex.Lock()
        defer q.mutex.Unlock()
    }
    n := &InputNode{Value: dir}
    if q.head == q.tail && q.count > 0 {
        nodes := make([]*InputNode, len(q.nodes)*2)
        copy(nodes, q.nodes[q.head:])
        copy(nodes[len(q.nodes)-q.head:], q.nodes[:q.head])
        q.head = 0
        q.tail = len(q.nodes)
        q.nodes = nodes
    }
    q.nodes[q.tail] = n
    q.tail = (q.tail + 1) % len(q.nodes)
    q.count++
    log.Printf("Pushed %v", dir)
}

// Pop removes and returns a node from the queue in first to last order.
func (q *InputQueue) Pop() Direction {
    if q.mutex != nil {
        q.mutex.Lock()
        defer q.mutex.Unlock()
    }
    if q.count == 0 {
        return NONE
    }
    node := q.nodes[q.head]
    q.head = (q.head + 1) % len(q.nodes)
    q.count--
    log.Printf("Poped %v", node.Value)
    return node.Value
}

func (q *InputQueue) Last() Direction {
    if q.mutex != nil {
        q.mutex.Lock()
        defer q.mutex.Unlock()
    }
    if q.count == 0 {
        return NONE
    }
    node := q.nodes[q.tail-1]
    return node.Value
}

func (q *InputQueue) Clear() {
    if q.mutex != nil {
        q.mutex.Lock()
        defer q.mutex.Unlock()
    }
    q.nodes = make([]*InputNode, len(q.nodes))
    q.head = 0
    q.tail = 0
    q.count = 0
}