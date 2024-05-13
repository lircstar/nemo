package collections

type Queue struct {
	list []interface{}
}

func (que *Queue) Enqueue(data interface{}) {

	que.list = append(que.list, data)
}

func (que *Queue) Count() int {
	return len(que.list)
}

func (que *Queue) Peek() interface{} {
	return que.list[0]
}

func (que *Queue) Dequeue() (ret interface{}) {

	if len(que.list) == 0 {
		return nil
	}

	ret = que.list[0]

	que.list = que.list[1:]

	return
}

// Loop to fetch all elements.
func (que *Queue) Range(f func(interface{})) {
	for i := range que.list {
		f(que.list[i])
	}
}

func NewQueue(size int) *Queue {

	return &Queue{
		list: make([]interface{}, 0, size),
	}
}
