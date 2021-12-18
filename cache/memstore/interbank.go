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
	ctx            context.Context
	banks          *bankBalances
	accounts       *accountBalances
	quit           chan bool
	done           chan bool
	dbChan         chan<- *utils.BankBalance
	updateInterval time.Duration
}

type bankBalances struct {
	sync.RWMutex
	dues map[string]int64
}

type accountBalances struct {
	sync.RWMutex
	dues map[string]map[uint16]int
}

const (
	balanceInterval = time.Duration(1) * time.Second
)

var (
	memcache Cache // Declare cache singleton
	counter  uint64
)

func New(ctx context.Context, dbChan chan<- *utils.BankBalance, retreiveChan <-chan *utils.BankBalance) error {
	var err error

	// Prep bank dues singleton
	bankDues := &bankBalances{
		dues: make(map[string]int64),
	}

	// Prep bank dues singleton
	accountDues := &accountBalances{
		dues: make(map[string]map[uint16]int),
	}

	// Assign Cache singleton
	memcache = Cache{
		ctx:            ctx,
		banks:          bankDues,
		accounts:       accountDues,
		dbChan:         dbChan,
		quit:           make(chan bool),
		done:           make(chan bool),
		updateInterval: balanceInterval,
	}

	// Retrieve bank dues from database
	memcache.banks.Lock()
	for balance := range retreiveChan {
		memcache.banks.dues[balance.Name] = balance.Balance
		memcache.accounts.dues[balance.Name] = make(map[uint16]int)
	}
	memcache.banks.Unlock()

	// Update periodically
	go UpdateDatabaseBalances()

	return err
}

// UpdateDues changes the dues according to clearinghouse logic.
func UpdateDues(current *utils.SRBalance) error {
	// Update counter
	atomic.AddUint64(&counter, 1)

	// Update bank dues
	memcache.banks.Lock()

	memcache.banks.dues[current.Sender.Name] -= int64(current.Amount)
	memcache.banks.dues[current.Receiver.Name] += int64(current.Amount)

	memcache.banks.Unlock()

	// Update account dues
	memcache.accounts.Lock()

	memcache.accounts.dues[current.Sender.Name][current.Sender.Account] -= current.Amount
	memcache.accounts.dues[current.Receiver.Name][current.Receiver.Account] += current.Amount

	memcache.accounts.Unlock()

	return nil
}

func UpdateDatabaseBalances() {
	updateDB := func() {
		memcache.banks.RLock()

		for bank, balance := range memcache.banks.dues {
			memcache.dbChan <- &utils.BankBalance{
				Name:    bank,
				Balance: balance,
			}
		}
		memcache.banks.RUnlock()
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
	fmt.Printf("Bank dues:\n{\n")

	memcache.banks.RLock()
	for bank, due := range memcache.banks.dues {
		fmt.Printf("\t%s: %d\n", bank, due)
	}
	memcache.banks.RUnlock()

	fmt.Printf("}\nAccount dues:\n{\n")

	memcache.accounts.RLock()
	for bank, accounts := range memcache.accounts.dues {
		fmt.Printf("\t%s: {\n", bank)

		for account, amount := range accounts {
			fmt.Printf("\t\t%d: %d\n", account, amount)
		}
		fmt.Println("\t}")
	}
	fmt.Println("}")
	memcache.accounts.RUnlock()
}
