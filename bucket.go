package htw

type (
	Node[T any] struct {
		Task Task[T]
		next *Node[T]
		prev *Node[T]
	}

	Bucket[T any] struct {
		head *Node[T]
		tail *Node[T]
	}
)

func NewBucket[T any]() *Bucket[T] {
	return &Bucket[T]{}
}

func (b *Bucket[T]) Add(task Task[T]) *Node[T] {
	node := &Node[T]{
		Task: task,
	}

	if b.tail == nil {
		b.head = node
		b.tail = node
	} else {
		b.tail.next = node
		node.prev = b.tail
		b.tail = node
	}

	return node
}

func (b *Bucket[T]) Remove(node *Node[T]) bool {
	if node == nil {
		return false
	}

	if node.prev == nil && node.next == nil && b.head != node {
		return false
	}

	if node.prev != nil {
		node.prev.next = node.next
	} else {
		b.head = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		b.tail = node.prev
	}

	node.next = nil
	node.prev = nil

	return true
}

func (b *Bucket[T]) Flush() (nodes []*Node[T]) {
	if b.head == nil {
		return
	}

	curr := b.head
	for curr != nil {
		nodes = append(nodes, curr)
		next := curr.next

		curr.next = nil
		curr.prev = nil

		curr = next
	}

	b.head = nil
	b.tail = nil

	return
}
