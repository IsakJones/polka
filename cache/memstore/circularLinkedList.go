package memstore

/*
The circular linked list is a data structure with state.
It backs up single banks' accounts in the database to improve performance.
*/

type node struct {
	val  uint16
	next *node
}

type circularLinkedList struct {
	head    *node
	current *node
}

func (cll *circularLinkedList) add(id uint16) {
	switch {
	case cll.head.val == 0: // This works because no bank's id can be 0, being serial!
		cll.head.val = id
	case cll.current == nil:
		cll.head.next = &node{
			val:  id,
			next: cll.head,
		}
		cll.current = cll.head.next
	default:
		cll.current.next = &node{
			val:  id,
			next: cll.head,
		}
		cll.current = cll.current.next
	}
}

func (cll *circularLinkedList) getCurrent() uint16 {
	return cll.current.val
}

func (cll *circularLinkedList) next() {
	cll.current = cll.current.next
	cll.head = cll.head.next
}
