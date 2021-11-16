package transactions

import (
	"fmt"
	"sync"
)

// Dues registers how much Polka owes each bank.
// Positive values are owed by Polka to the bank,
// Negative values are owed by the bank to Polka.
// An RWMutex allows to simultaneously read and write.
type Register struct{
	sync.RWMutex
	Dues map[string]int64
}

// Update changes the dues according to clearinghouse logic.
func (reg * Register) update(current *Transaction) {

	// These operations make writing concurrently safe.
	reg.Lock()
	defer reg.Unlock()

	reg.Dues[current.Sender] -= current.Sum
	reg.Dues[current.Receiver] += current.Sum
}

// Init local register 
var register = Register{
	Dues: make(map[string]int64),
}

func UpdateDues(transactionChannel <-chan *Transaction) {
	
	for trans := range transactonChannel {
		go register.UpdateDues(trans)
	}
}

func PrintDues() (result map[string]int64) {
	register.RLock()
	defer register.RUnlock()

	fmt.Println("Dues:")
	for key, value := range register.Dues {
		fmt.Printf("%s: %d\n", key, value)
	}
	return
}

