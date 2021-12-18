package memstore

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sekerez/polka/cache/utils"
)

// Dues registers how much Polka owes each bank.
// Positive values are owed by Polka to the bank,
// Negative values are owed by the bank to Polka.
// An RWMutex allows to simultaneously read and write.
type Register struct {
	sync.RWMutex
	ctx            context.Context
	quit           <-chan bool
	dues           map[string]int64
	dbChan         chan<- utils.BankBalance
	updateInterval time.Duration
}

var (
	register Register
	counter  uint64
)

func New(ctx context.Context, interval time.Duration, quit <-chan bool, dbChan chan<- utils.BankBalance) error {
	var err error
	// Init local register
	register = Register{
		ctx:            ctx,
		quit:           quit,
		dues:           make(map[string]int64),
		dbChan:         dbChan,
		updateInterval: interval,
	}
	// Update periodically
	go UpdateDatabaseBalances()

	return err
}

// UpdateDues changes the dues according to clearinghouse logic.
func UpdateDues(current *utils.SRBalance) error {
	// Update counter
	atomic.AddUint64(&counter, 1)

	// These operations make writing concurrently safe.
	register.Lock()
	defer register.Unlock()

	register.dues[current.Sender] -= int64(current.Amount)
	register.dues[current.Receiver] += int64(current.Amount)

	return nil
}

func UpdateDatabaseBalances() {
	ticker := time.NewTicker(register.updateInterval)

	for {
		select {
		case <-register.quit:
			return
		case <-ticker.C:
			register.Lock()

			for bank, balance := range register.dues {
				register.dbChan <- utils.BankBalance{
					Name:    bank,
					Balance: balance,
				}
				register.dues[bank] = 0
			}
			register.Unlock()

		}
	}
}

// PrintDues prints to the console how much Polka owes to
// each bank, listed in no specific order.
func PrintDues() {

	log.Printf("Processed %d transactions.", counter)
	fmt.Printf("Dues:\n{\n")

	register.RLock()
	for bank, due := range register.dues {
		fmt.Printf("%s: %d\n", bank, due)
	}
	defer register.RUnlock()

	fmt.Printf("}\n")
}
