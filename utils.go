package porttools

// Queue is an implementation of a FIFO container type.
type Queue struct {
	len    int
	values []*Node
}

// Push stores value in the queue.
func (queue *Queue) Push(node *Node) {
	queue.len++
	if queue.len-1 == 0 {
		queue.values[0] = node
		return
	}
	queue.values[queue.len] = node
	return
}

// Pop removes and returns value from the queue.
func (queue *Queue) Pop() (node *Node) {
	queue.len--
	if queue.len+1 == 0 {
		return nil
	}
	node = queue.values[0]
	queue.values = queue.values[0:]
	return node
}

// NewQueue instantiates a new Queue.
func NewQueue() *Queue {
	q := Queue{
		len: 0,
	}
	return &q
}

// Node represents data stored in a container.
type Node struct {
	data interface{}
}

// NewNode instantiates a new Node.
func NewNode(data interface{}) *Node {
	return &Node{data: data}
}

// Kwarg struct allows for add'l args/attrs to a class or func.
// NOTE: is this really needed?
type Kwarg struct {
	name  string
	value interface{}
}
