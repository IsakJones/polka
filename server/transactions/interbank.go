package transactions

import (
	"fmt"
	"sync"

	//"github.com/IsakJones/polka/server/utils"
)

// Dues registers how much Polka owes each bank.
// Positive values are owed by Polka to the bank,
// Negative values are owed by the bank to Polka.
// An RWMutex allows to simultaneously read and write.
type Register struct{
	sync.RWMutex
	Dues map[string]int64
}

// Init local register 
var register = Register{
	Dues: make(map[string]int64),
}

// Update changes the dues according to clearinghouse logic.
func UpdateDues(current *Transaction) {

	// These operations make writing concurrently safe.
	register.Lock()
	defer register.Unlock()

	register.Dues[current.Sender] -= int64(current.Sum)
	register.Dues[current.Receiver] += int64(current.Sum)
}


// func UpdateDues(transactionChannel <-chan *Transaction) {
	
// 	for trans := range transactonChannel {
// 		go register.UpdateDues(trans)
// 	}
// }

func PrintDues() (result map[string]int64) {
	register.RLock()
	defer register.RUnlock()

	fmt.Printf("Dues:\n{\n")
	for key, value := range register.Dues {
		fmt.Printf("%s: %d\n", key, value)
	}
	fmt.Println("}\n")
	// Test successfull: values always sum to 0
	// fmt.Printf("Values sum to: %d\n\n", utils.SumMap64(register.Dues))
	return
}

