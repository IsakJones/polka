package memstore

import (
	"context"
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
	Snap           *snapshot
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
	Accs    accounts
}

// snapBank stores bank data relevant to a snapshot.
// snapBank does not include the bank's id or name.
type snapBank struct {
	Balance int64
	Accs    map[uint32]int32
}

// accounts maps account ids to pointers to balances.
type accounts map[uint32]*int32

// snapshot stores a synchronized snapshort of all balances.
// It stores integers and not pointers, since there's no need
// for concurrent access.
type snapshot struct {
	ready     bool
	banks     map[string]*snapBank
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
	snap := &snapshot{
		ready: false,
		banks: make(map[string]*snapBank, bankNum),
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
				Accs:    make(accounts),
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

		// If the bank is not in the cache map, allocate it.
		if _, exists := c.Balances.Banks[bankName].Accs[accountNum]; !exists {
			c.Balances.Banks[bankName].Accs[accountNum] = new(int32)
		}
		accountPtr := c.Balances.Banks[bankName].Accs[accountNum]
		// Update value
		atomic.StoreInt32(accountPtr, accountBalance)
	}

	c.Balances.Unlock()

	// Update periodically DB records at regular intervals
	go ManageDatabaseBackups()
}

// UpdateBalances changes bank and account balances given an incoming payment.
func UpdateBalances(current *utils.SRBalance) (err error) {

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

	// Check if sending account exists
	if _, exists := senBank.Accs[current.Sender.Account]; exists {
		senBank.Accs[current.Sender.Account] = new(int32)
	}
	senAcc := senBank.Accs[current.Sender.Account]
	atomic.AddInt32(senAcc, -current.Amount)

	// Check if receiving account exists
	if _, exists := recBank.Accs[current.Receiver.Account]; exists {
		recBank.Accs[current.Receiver.Account] = new(int32)
	}
	recAcc := recBank.Accs[current.Receiver.Account]
	atomic.AddInt32(recAcc, current.Amount)

	return
}

// ManageDatabaseBackups backs up cache data in the database.
// Every so often (e.g. one second) it backs up the bank dues.
// In between every other bank dues update, it updates only one
// bank's accounts balances for performance reasons.
func ManageDatabaseBackups() {
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

	// Back up each account
	for account, balance := range bankAccounts {
		c.Chans.AccChan <- &utils.Balance{
			BankName: name,
			Account:  account,
			Balance:  *balance,
		}
	}
}

// Close shuts down the cache correctly,
// such that all current data is backed up.
func Close() (err error) {
	c.Chans.Quit <- true
	<-c.Chans.Done
	return
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
		// NB c.banks[id] is a pointer to an int
		fmt.Printf("\t%s: %d\n", name, *bnk.Balance)
	}
	fmt.Println("}")

	// If specified, print account balances
	if andAccounts {
		fmt.Printf("}\nAccount balances:\n{\n")

		for name, bnk := range c.Balances.Banks {
			fmt.Printf("\t%s: {\n", name)

			for account, amount := range bnk.Accs {
				fmt.Printf("\t\t%d: %d\n", account, *amount)
			}
			fmt.Println("\t}")
		}
		fmt.Println("}")
	}
}
