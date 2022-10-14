package main
//type Item interface {
//}

var _ Queue = (*ArrayQueue)(nil)

// Item the type of the queue
type ArrayQueue struct {
	items []string
}

type Queue interface {
	Enqueue(t string)
	Dequeue() *string
	IsEmpty() bool
	Size() int
	Peek() *string
}

// New creates a new ItemQueue
func NewArrayQueue() *ArrayQueue {
	//s.items =
	return &ArrayQueue{
		items: []string{},
	}
}

// Enqueue adds an Item to the end of the queue
func (s *ArrayQueue) Enqueue(t string) {
	s.items = append(s.items, t)
}

// dequeue
func (s *ArrayQueue) Dequeue() *string {
	if s.IsEmpty() {
		return nil
	}
	item := s.items[0] // 先进先出
	s.items = s.items[1:len(s.items)]

	return &item
}

func (s *ArrayQueue) IsEmpty() bool {
	return len(s.items) == 0
}

// Size returns the number of Items in the queue
func (s *ArrayQueue) Size() int {
	return len(s.items)
}

func (s *ArrayQueue) Peek() *string {
	if s.IsEmpty() {
		return nil
	}
	return &s.items[0]
}