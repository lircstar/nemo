package collections

type Queue struct {
	list []any
}

func (que *Queue) Enqueue(data any) {

	que.list = append(que.list, data)
}

func (que *Queue) Count() int {
	return len(que.list)
}

func (que *Queue) Peek() any {
	return que.list[0]
}

func (que *Queue) Dequeue() (ret any) {

	if len(que.list) == 0 {
		return nil
	}

	ret = que.list[0]

	que.list = que.list[1:]

	return
}

// Range Loop to fetch all elements.
func (que *Queue) Range(f func(any)) {
	for i := range que.list {
		f(que.list[i])
	}
}

func NewQueue(size int) *Queue {

	return &Queue{
		list: make([]any, 0, size),
	}
}
