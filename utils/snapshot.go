package utils

import (
	"fmt"
	"time"
)

// SnapBank stores bank data relevant to a snapshot.
// SnapBank does not include the bank's id or name.
type SnapBank struct {
	Balance  int64
	Accounts map[uint32]int32
}

// snapshot stores a synchronized snapshort of all balances.
// It stores integers and not pointers, since there's no need
// for concurrent access. Banks is nil unless a snapshot has
// been requested.
type Snapshot struct {
	Banks     map[string]*SnapBank
	Timestamp time.Time
}

func (snap *Snapshot) Print() {

	fmt.Println("Snapshot: ")
	fmt.Printf("Bank balances: \n{\n")

	// Print bank balances
	for name, bnk := range snap.Banks {
		// NB bnk.Balance is a pointer to an int
		fmt.Printf("\t%s: %d\n", name, bnk.Balance)
	}
	fmt.Println("}")

	// If specified, print account balances
	fmt.Printf("}\nAccount balances:\n{\n")

	for name, bnk := range snap.Banks {
		fmt.Printf("\t%s: {\n", name)

		for account, amount := range bnk.Accounts {
			fmt.Printf("\t\t%d: %d\n", account, amount)
		}
		fmt.Println("\t}")
	}
	fmt.Println("}")
}
