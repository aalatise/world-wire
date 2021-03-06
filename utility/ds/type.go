package ds

type Void struct{}

type Iterator interface {
	HasMore() bool
	Next() (interface{}, error)
}

type Container interface {
	Add(interface{})
	Remove(interface{})
	Contains(interface{}) bool
	Size() int
}
