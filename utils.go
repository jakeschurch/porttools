package porttools

// selection sort for positions.
func selectionSort(A []*Position) []*Position {
	for i := 0; i < len(A)-1; i++ {
		min := i
		for j := i + 1; j < len(A); j++ {
			if A[j].Ticker < A[min].Ticker {
				min = j
			}
		}
		key := A[i]
		A[i] = A[min]
		A[min] = key
	}
	return A
}

func filter(positions []Position, key string) []Position {
	filtered := make([]Position, 0)

	for _, position := range positions {
		if position.Ticker == key {
			filtered = append(filtered, position)
		}
	}
	return filtered
}

func findKey(A []string, toFind string) bool {
	for _, key := range A {
		if key == toFind {
			return true
		}
	}
	return false
}

// Queue is an implementation of a FIFO container type.
type Queue struct {
	len    int
	values []*Node
}

// Enqueue stores value in the queue.
func (queue *Queue) Enqueue(node *Node) {
	queue.len++
	if queue.len-1 == 0 {
		queue.values[0] = node
	} else {
		queue.values[queue.len] = node
	}
}

// Dequeue removes and returns value from the queue.
func (queue *Queue) Dequeue() (node *Node) {
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
