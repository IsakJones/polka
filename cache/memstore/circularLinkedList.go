package memstore

/*
The circular linked list is a data structure with state.
It is used to back up single banks' accounts in the database to improve performance.
*/

type node struct {
	val  uint16
	next *node
}

type circularLinkedList struct {
	current *node
}

func (cll *circularLinkedList) add(id uint16) {
	if cll.current == nil { // This works because no bank's id can be 0, being serial!
		cll.current = &node{val: id}
		cll.current.next = cll.current
		return
	}

	nxt := cll.current.next
	cll.current.next = &node{
		val:  id,
		next: nxt,
	}
}

func (cll *circularLinkedList) getCurrent() uint16 {
	return cll.current.val
}

func (cll *circularLinkedList) next() {
	cll.current = cll.current.next
}
