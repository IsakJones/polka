package memstore

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/sekerez/polka/utils"
)

// SettleSnapshot subtracts all balances by the balances stored in the snapshot.
// It is called when clearing payments.
func SettleSnapshot() error {
	// If there's no snapshot, return an error
	if c.Snap.Banks == nil {
		return errors.New("no snapshot taken - must request a snapshot before settling payments")
	}

	// Read because there's no synchronization required.
	c.Balances.RLock()
	defer c.Balances.RUnlock()

	for name, bnk := range c.Balances.Banks {
		snapBnk := c.Snap.Banks[name]

		// change balance and reset snapshot
		atomic.AddInt64(bnk.Balance, -snapBnk.Balance)
		snapBnk.Balance = 0

		// change all accounts
		bnk.Accs.RLock()
		for accNum, accPtr := range bnk.Accs.Mp {
			// change balance and reset snapshot
			atomic.AddInt32(accPtr, -snapBnk.Accounts[accNum])
			snapBnk.Accounts[accNum] = 0
		}
		bnk.Accs.RUnlock()
	}

	return nil
}

// GetSnapshot returns a snapshot of all balances in a given instant.
// It returns a pointer to the snapshot.
func GetSnapshot() (*utils.Snapshot, error) {
	var sum int32
	var err error
	var snapbnk *utils.SnapBank // Stores the current snapbank

	// Initialize banks
	snap := &utils.Snapshot{
		Banks: make(map[string]*utils.SnapBank),
	}

	c.Balances.Lock() // Lock to keep out other reader threads

	// Loop through banks, adding them one by one
	for name, bnk := range c.Balances.Banks {
		// Make the snapbank with the balance
		snapbnk = &utils.SnapBank{
			Balance:  *bnk.Balance,
			Accounts: make(map[uint32]int32),
		}

		// Add all accounts
		bnk.Accs.RLock()
		if len(bnk.Accs.Mp) == 0 {
			bnk.Accs.RUnlock()
			continue
		}
		for accNum, accBalance := range bnk.Accs.Mp {
			snapbnk.Accounts[accNum] = *accBalance
		}
		bnk.Accs.RUnlock()

		// Add snapbank to snapshot
		snap.Banks[name] = snapbnk
	}

	// Unlock, letting reader threads back in
	c.Balances.Unlock()

	// Set error and total sum variables
	err = errors.New("incoherent snapshot")
	totalSum := int64(0)

	// Check that each bank balance corresponds to the sum of its accounts
	for name, bnk := range snap.Banks {
		sum = 0
		for _, accountBalance := range bnk.Accounts {
			sum += accountBalance
		}

		// Make the check, if so print error
		if bnk.Balance != int64(sum) {
			c.Logger.Printf("Error: account balances not synched with bank balance for %s", name)
			c.Snap.Banks = nil
			return nil, err
		}
		totalSum += bnk.Balance
	}

	// check that all bank balances sum to zero
	if totalSum != 0 {
		c.Logger.Printf("Error: bank balances don't add to 0, but to ")
		c.Snap.Banks = nil
		return nil, err
	}

	// Update time of snapshot and readiness
	snap.Print()
	snap.Timestamp = time.Now()
	// Assign to cache
	c.Snap = snap

	c.Logger.Printf("finished sending back snapshot...")
	return snap, nil
}

// Cancels the last snapshot.
func CancelSnapshot() {
	c.Snap.Banks = nil
}
