package memstore

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sekerez/polka/cache/src/utils"
)

const (
	backupInterval = time.Duration(1) * time.Second
)

var (
	c       cache // Declare cache singleton
	counter uint64
)

// Positive values are owed by Polka to the bank,
// Negative values are owed by the bank to Polka.
type cache struct {
	Ctx            context.Context
	List           *circularLinkedList
	Snap           *Snapshot
	Chans          *channels
	Logger         *log.Logger
	Balances       *balancesRW
	BackupInterval time.Duration
}

// channels contains channels relevant to the cache.
type channels struct {
	Quit     chan bool
	Done     chan bool
	AccChan  chan<- *utils.Balance
	BankChan chan<- *utils.BankBalance
}

type balancesRW struct {
	sync.RWMutex
	Banks map[string]*bank
}

// bank stores data relevant to each bank, including accounts.
// bank is only used in the main cache, not in any snapshot.
type bank struct {
	Id      uint16
	Balance *int64
	Accs    *accounts
}

// snapBank stores bank data relevant to a snapshot.
// snapBank does not include the bank's id or name.
type SnapBank struct {
	Balance  int64
	Accounts map[uint32]int32
}

// accounts maps account ids to pointers to balances.
type accounts struct {
	sync.RWMutex
	Mp map[uint32]*int32
}

// snapshot stores a synchronized snapshort of all balances.
// It stores integers and not pointers, since there's no need
// for concurrent access.
type Snapshot struct {
	ready     bool
	Banks     map[string]*SnapBank
	timestamp time.Time
}

// New initializes the cache struct.
func New(
	ctx context.Context,
	bankNumChan <-chan uint16,
	banksChan chan<- *utils.BankBalance, // Channel to send bank balances
	accountsChan chan<- *utils.Balance, // Channel to send account balances
	bankRetChan <-chan *utils.BankBalance, // Channel to retrieve bank balances
	accRetChan <-chan *utils.Balance, // Channel to retrieve account balances
) {

	logger := log.New(os.Stderr, "[cache] ", log.LstdFlags|log.Lshortfile)

	// Initialize circular linked list.
	list := newCircularLinkedList()

	// Initialize channels struct.
	chans := &channels{
		BankChan: banksChan,
		AccChan:  accountsChan,
		Quit:     make(chan bool),
		Done:     make(chan bool),
	}

	// Initialize balances struct
	bankNum := <-bankNumChan // TODO is this even necessary?
	balances := &balancesRW{
		Banks: make(map[string]*bank, bankNum),
	}

	// Initialize snapshot struct.
	snap := &Snapshot{
		ready: false,
		Banks: make(map[string]*SnapBank, bankNum),
	}

	// Assign Cache singleton
	c = cache{
		Ctx:            ctx,
		List:           list,
		Snap:           snap,
		Chans:          chans,
		Logger:         logger,
		Balances:       balances,
		BackupInterval: backupInterval,
	}

	c.Balances.Lock()
	// Atomically update bank balances retieved from database.
	for incomingBalance := range bankRetChan {
		// Declare variables for readability.
		bankId := incomingBalance.BankId
		bankName := incomingBalance.Name
		bankBalance := incomingBalance.Balance

		// If the bank is not in the cache map, allocate it.
		if _, exists := c.Balances.Banks[incomingBalance.Name]; !exists {
			c.Balances.Banks[bankName] = &bank{
				Id:      bankId,
				Balance: new(int64),
				Accs: &accounts{
					Mp: make(map[uint32]*int32),
				},
			}
		}
		cacheBank := c.Balances.Banks[bankName]
		// Update value
		atomic.StoreInt64(cacheBank.Balance, bankBalance)
		// Update circular linked list and bank id map
		c.List.add(bankName)
	}

	// Atomically update account balances retieved from database
	for incomingBalance := range accRetChan {
		// Declare variables for readability.
		bankName := incomingBalance.BankName
		accountNum := incomingBalance.Account
		accountBalance := incomingBalance.Balance

		// Update account in cache
		accs := c.Balances.Banks[bankName].Accs
		updateAccount(accs, accountNum, accountBalance)
	}

	c.Balances.Unlock()

	// Update periodically DB records at regular intervals
	go manageDatabaseBackups()
}

// Close shuts down the cache correctly,
// such that all current data is backed up.
func Close() (err error) {
	c.Chans.Quit <- true
	<-c.Chans.Done
	return
}

// UpdateBalances changes bank and account balances given an incoming payment.
func UpdateBalances(current *utils.SRBalance) error {

	// Update counter and retrieve bank ids
	atomic.AddUint64(&counter, 1)

	// Lock and unlock balances
	c.Balances.RLock()
	defer c.Balances.RUnlock()

	// Get bank structs
	senBank := c.Balances.Banks[current.Sender.Name]
	recBank := c.Balances.Banks[current.Receiver.Name]

	// Update sender bank's balance
	atomic.AddInt64(senBank.Balance, -int64(current.Amount)) // Amount is subtracted from sender...

	// Update receiving bank's balance
	atomic.AddInt64(recBank.Balance, int64(current.Amount)) // ... and added to receiver.

	// Update sender and receiver account balances
	updateAccount(senBank.Accs, current.Sender.Account, -current.Amount)
	updateAccount(recBank.Accs, current.Receiver.Account, current.Amount)

	return nil
}

// SettleSnapshot subtracts all balances by the balances stored in the snapshot.
// It is called when clearing payments.
func SettleSnapshot() error {
	// If there's no snapshot, return an error
	if !c.Snap.ready {
		return errors.New("no snapshot taken - must request a snapshot before settling payments.")
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
// It returns a pointer to the snapshot for performance reasons.
func GetSnapshot() (*Snapshot, error) {
	var sum int32
	var err error
	var snapbnk SnapBank // Stores the current snapbank

	c.Balances.Lock() // Lock to keep out other reader threads

	// Loop through banks, adding them one by one
	for name, bnk := range c.Balances.Banks {
		// Make the snapbank with the balance
		snapbnk = SnapBank{
			Balance: *bnk.Balance,
		}

		// Add all accounts
		bnk.Accs.RLock()
		for accNum, accBalance := range bnk.Accs.Mp {
			snapbnk.Accounts[accNum] = *accBalance
		}
		bnk.Accs.RUnlock()

		// Add snapbank to snapshot
		c.Snap.Banks[name] = &snapbnk
	}

	// Unlock, letting reader threads back in
	c.Balances.Unlock()

	// Check that each bank balance corresponds to the sum of its accounts
	for name, bnk := range c.Snap.Banks {
		sum = 0
		for _, accountBalance := range bnk.Accounts {
			sum += accountBalance
		}

		// Make the check, if so print error
		if bnk.Balance != int64(sum) {
			c.Logger.Printf("Error: account balances not synched with bank balance for %s", name)
			err = errors.New("incoherent snapshot")
		}
	}

	// Update time of snapshot and readiness
	c.Snap.timestamp = time.Now()
	c.Snap.ready = true

	return c.Snap, err
}

// PrintDues prints to the console how much Polka owes to
// each bank, listed in no specific order.
func PrintBalances(andAccounts bool) {

	log.Printf("Processed %d transactions.", counter)
	fmt.Printf("Bank balances:\n{\n")

	// Lock and unlock
	c.Balances.RLock()
	defer c.Balances.RUnlock()

	// Print bank balances
	for name, bnk := range c.Balances.Banks {
		// NB bnk.Balance is a pointer to an int
		fmt.Printf("\t%s: %d\n", name, *bnk.Balance)
	}
	fmt.Println("}")

	// If specified, print account balances
	if andAccounts {
		fmt.Printf("}\nAccount balances:\n{\n")

		for name, bnk := range c.Balances.Banks {
			fmt.Printf("\t%s: {\n", name)

			// Read lock the accounts
			bnk.Accs.RLock()

			for account, amount := range bnk.Accs.Mp {
				fmt.Printf("\t\t%d: %d\n", account, *amount)
			}
			fmt.Println("\t}")

			// Read unlock them
			bnk.Accs.RUnlock()
		}
		fmt.Println("}")
	}
}

// manageDatabaseBackups backs up cache data in the database.
// Every so often (e.g. one second) it backs up the bank dues.
// In between every other bank dues update, it updates only one
// bank's accounts balances for performance reasons.
func manageDatabaseBackups() {
	var (
		bankTicker    *time.Ticker
		accountTicker *time.Ticker
	)

	bankTicker = time.NewTicker(c.BackupInterval)
	time.Sleep(c.BackupInterval / 2)
	accountTicker = time.NewTicker(c.BackupInterval * 2)

	for {
		select {
		case <-c.Chans.Quit:
			backupDatabaseBankBalances()
			for i := uint16(0); i < c.List.Length; i++ {
				backupDatabaseAccountBalance()
			}
			c.Chans.Done <- true
			return
		case <-bankTicker.C:
			backupDatabaseBankBalances()
		case <-accountTicker.C:
			backupDatabaseAccountBalance()
		}
	}
}

// backupDatabaseBankBalances sends bank due data to
// the database connection through the bankChan channel.
func backupDatabaseBankBalances() {

	// Lock and unlock balances
	c.Balances.RLock()
	defer c.Balances.RUnlock()

	for _, bnk := range c.Balances.Banks {
		c.Chans.BankChan <- &utils.BankBalance{
			BankId:  bnk.Id,
			Balance: *bnk.Balance,
		}
	}
}

// backupDatabaseAccountBalances is a function with state, like a generator in Python.
// The goal is to update the account records of one bank at a time, in a specific order,
// so as to update each bank's accounts at regular intervals.
// The function sends account balance data to the database
// connection through the accChan channel.
func backupDatabaseAccountBalance() {

	// Lock and unlock balances
	c.Balances.RLock()
	defer c.Balances.RUnlock()

	// Get next bank
	c.List.next()
	name := c.List.getCurrent()
	bankAccounts := c.Balances.Banks[name].Accs

	// Read lock the bank accounts
	bankAccounts.RLock()
	defer bankAccounts.RUnlock()

	// Back up each account
	for account, balance := range bankAccounts.Mp {
		c.Chans.AccChan <- &utils.Balance{
			BankName: name,
			Account:  account,
			Balance:  *balance,
		}
	}
}

// addAccount adds a key to an accounts map in a
// thread-safe way.
func updateAccount(accs *accounts, accNum uint32, balance int32) {
	// Read lock and defer unlock
	accs.RLock()
	defer accs.RUnlock()

	// Check if the key exists, if not add it
	if _, exists := accs.Mp[accNum]; !exists {
		accs.RUnlock() // Release read lock to avoid deadlock
		accs.Lock()    // Lock to make key change
		accs.Mp[accNum] = new(int32)
		accs.Unlock()
		accs.RLock() // Read lock again to read val
	}
	accountPtr := accs.Mp[accNum]
	// Update value
	atomic.AddInt32(accountPtr, balance)
}
