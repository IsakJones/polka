package main

import (
	"fmt"
	"sync"
)

func main() {
	m := mm{num: 5}
	fmt.Printf("%d", m.f())
}

type mm struct {
	sync.RWMutex
	num int
}

func (m *mm) f() int {
	m.Lock()
	res := m.num
	defer m.Unlock()
	return res
}

// func main() {
// 	var i uint16
// 	exp := &circularLinkedList{head: &node{}}

// 	for i = 1; i <= 10; i++ {
// 		exp.add(i)
// 	}
// 	for i = 1; i <= 10; i++ {
// 		exp.next()
// 	}

// }
