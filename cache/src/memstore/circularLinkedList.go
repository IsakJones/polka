package memstore

/*
The circular linked list is a data structure with state.
It backs up single banks' accounts in the database to improve performance.
*/

type node struct {
	bankName string
	next     *node
}

type circularLinkedList struct {
	current *node
	Length  uint16
}

func newCircularLinkedList() *circularLinkedList {
	return &circularLinkedList{
		current: nil,
		Length:  0,
	}
}

func (cll *circularLinkedList) add(name string) {
	if cll.current == nil { // This works because no bank's id can be 0, being serial!
		cll.current = &node{bankName: name}
		cll.current.next = cll.current
		return
	}

	nxt := cll.current.next
	cll.current.next = &node{
		bankName: name,
		next:     nxt,
	}

	cll.Length++
}

func (cll *circularLinkedList) getCurrent() string {
	return cll.current.bankName
}

func (cll *circularLinkedList) next() {
	cll.current = cll.current.next
}
