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
type Cache struct {
	sync.RWMutex
	ctx            context.Context
	dues           map[string]int64
	quit           chan bool
	done           chan bool
	dbChan         chan<- *utils.BankBalance
	updateInterval time.Duration
}

const (
	balanceInterval = time.Duration(1) * time.Second
)

var (
	memcache Cache
	counter  uint64
)

func New(ctx context.Context, dbChan chan<- *utils.BankBalance, retreiveChan <-chan *utils.BankBalance) error {
	var err error
	// Init local register
	memcache = Cache{
		ctx:            ctx,
		dues:           make(map[string]int64),
		dbChan:         dbChan,
		quit:           make(chan bool),
		done:           make(chan bool),
		updateInterval: balanceInterval,
	}

	// Retrieve dues from database
	memcache.Lock()
	for balance := range retreiveChan {
		memcache.dues[balance.Name] = balance.Balance
	}
	memcache.Unlock()

	// Update periodically
	go UpdateDatabaseBalances()

	return err
}

// UpdateDues changes the dues according to clearinghouse logic.
func UpdateDues(current *utils.SRBalance) error {
	// Update counter
	atomic.AddUint64(&counter, 1)

	// These operations make writing concurrently safe.
	memcache.Lock()
	defer memcache.Unlock()

	memcache.dues[current.Sender] -= int64(current.Amount)
	memcache.dues[current.Receiver] += int64(current.Amount)

	return nil
}

func UpdateDatabaseBalances() {
	updateDB := func() {
		memcache.RLock()

		for bank, balance := range memcache.dues {
			memcache.dbChan <- &utils.BankBalance{
				Name:    bank,
				Balance: balance,
			}
		}
		memcache.RUnlock()
	}

	ticker := time.NewTicker(memcache.updateInterval)
	for {
		select {
		case <-memcache.quit:
			updateDB()
			memcache.done <- true
			return
		case <-ticker.C:
			updateDB()
		}
	}
}

func Close() (err error) {
	memcache.quit <- true
	<-memcache.done
	return
}

// PrintDues prints to the console how much Polka owes to
// each bank, listed in no specific order.
func PrintDues() {

	log.Printf("Processed %d transactions.", counter)
	fmt.Printf("Dues:\n{\n")

	memcache.RLock()
	for bank, due := range memcache.dues {
		fmt.Printf("%s: %d\n", bank, due)
	}
	defer memcache.RUnlock()

	fmt.Printf("}\n")
}
