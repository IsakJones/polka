package memstore

import (
	"fmt"
	"sync"

	"github.com/IsakJones/polka/api/utils"
)

// Dues registers how much Polka owes each bank.
// Positive values are owed by Polka to the bank,
// Negative values are owed by the bank to Polka.
// An RWMutex allows to simultaneously read and write.
type Register struct {
	sync.RWMutex
	Dues map[string]int64
}

// Init local register
var register = Register{
	Dues: make(map[string]int64),
}

// UpdateDues changes the dues according to clearinghouse logic.
func UpdateDues(current utils.Transaction) {

	// These operations make writing concurrently safe.
	register.Lock()
	defer register.Unlock()

	register.Dues[current.GetSenBank()] -= int64(current.GetAmount())
	register.Dues[current.GetRecBank()] += int64(current.GetAmount())
}

// PrintDues prints to the console how much Polka owes to
// each bank, listed in no specific order.
func PrintDues() {

	fmt.Printf("Dues:\n{\n")

	register.RLock()
	for key, value := range register.Dues {
		fmt.Printf("%s: %d\n", key, value)
	}
	defer register.RUnlock()

	fmt.Printf("}\n")
}
