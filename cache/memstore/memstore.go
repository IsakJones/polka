package memstore

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sekerez/polka/cache/utils"
)

const (
	backupInterval = time.Duration(1) * time.Second
)

var (
	c       Cache // Declare cache singleton
	counter uint64
)

// Dues registers how much Polka owes each bank.
// Positive values are owed by Polka to the bank,
// Negative values are owed by the bank to Polka.
// An RWMutex allows to simultaneously read and write.
type Cache struct {
	ctx            context.Context
	list           *circularLinkedList
	bankId         map[string]uint16
	banks          *bankBalances
	accounts       map[uint16]*accountBalances
	logger         *log.Logger
	quit           chan bool
	done           chan bool
	accChan        chan<- *utils.Balance
	bankChan       chan<- *utils.BankBalance
	backupInterval time.Duration
}

// accountBalances records banks' balances.
type bankBalances struct {
	sync.RWMutex
	dues map[string]int64
}

// accountBalances records the balance of individual bank accounts.
type accountBalances struct {
	sync.RWMutex
	accBalances map[uint16]int // bank_id -> account -> balance
}

// New initializes the cache struct.
func New(
	ctx context.Context,
	bankNumChan <-chan uint16,
	bankChan chan<- *utils.BankBalance,
	accChan chan<- *utils.Balance,
	bankRetChan <-chan *utils.BankBalance,
	accRetChan <-chan *utils.Balance,
) (err error) {

	logger := log.New(os.Stderr, "[cache] ", log.LstdFlags|log.Lshortfile)
	// Prep bank dues singleton with number of banks
	bankNum := <-bankNumChan
	bankDues := &bankBalances{
		dues: make(map[string]int64, bankNum),
	}
	bankId := make(map[string]uint16, bankNum)

	// Prep bank dues singleton
	accountDues := make(map[uint16]*accountBalances, bankNum)

	// Make linked list
	list := &circularLinkedList{head: &node{}}

	// Assign Cache singleton
	c = Cache{
		ctx:            ctx,
		list:           list,
		logger:         logger,
		bankId:         bankId,
		accChan:        accChan,
		bankChan:       bankChan,
		banks:          bankDues,
		accounts:       accountDues,
		quit:           make(chan bool),
		done:           make(chan bool),
		backupInterval: backupInterval,
	}

	// Retreive bank balances from database.
	c.banks.Lock()
	for balance := range bankRetChan {
		c.list.add(balance.BankId)
		bankId[balance.Name] = balance.BankId
		c.banks.dues[balance.Name] = balance.Balance
		c.accounts[balance.BankId] = &accountBalances{
			accBalances: make(map[uint16]int),
		}
	}
	c.banks.Unlock()

	// Retrieve account balances from the database
	for balance := range accRetChan {
		bankAccounts := c.accounts[balance.BankId]
		bankAccounts.Lock()
		bankAccounts.accBalances[balance.Account] = balance.Balance
		bankAccounts.Unlock()
	}

	// Update periodically DB records at regular intervals
	go ManageDatabaseBackups()

	return
}

// UpdateDues changes the dues according to clearinghouse logic.
func UpdateDues(current *utils.SRBalance) (err error) {
	// Update counter
	atomic.AddUint64(&counter, 1)

	// Update bank dues
	c.banks.Lock()

	c.banks.dues[current.Sender.Name] -= int64(current.Amount)
	c.banks.dues[current.Receiver.Name] += int64(current.Amount)

	c.banks.Unlock()

	// Update account dues
	senderAccounts := c.accounts[c.bankId[current.Sender.Name]]

	senderAccounts.Lock()
	senderAccounts.accBalances[current.Sender.Account] -= current.Amount
	senderAccounts.Unlock()

	receiverAccounts := c.accounts[c.bankId[current.Receiver.Name]]
	receiverAccounts.Lock()
	receiverAccounts.accBalances[current.Receiver.Account] += current.Amount
	receiverAccounts.Unlock()

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

	bankTicker = time.NewTicker(c.backupInterval)
	time.Sleep(c.backupInterval / 2)
	accountTicker = time.NewTicker(c.backupInterval * 2)

	for {
		select {
		case <-c.quit:
			backupDatabaseBankBalances()
			for i := 0; i < len(c.bankId); i++ {
				backupDatabaseAccountBalance()
			}
			c.done <- true
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
	c.banks.RLock()
	defer c.banks.RUnlock()

	for bank, balance := range c.banks.dues {
		c.bankChan <- &utils.BankBalance{
			Name:    bank,
			Balance: balance,
		}
	}
}

// backupDatabaseAccountBalances is a function with state, like a generator in Python.
// The goal is to update the account records of one bank at a time, in a specific order,
// so as to update each bank's accounts at regular intervals.
// The function sends account balance data to the database
// connection through the accChan channel.
func backupDatabaseAccountBalance() {
	c.list.next()
	id := c.list.getCurrent()
	bankAccounts := c.accounts[id]
	bankAccounts.Lock()
	defer bankAccounts.Unlock()

	for account, balance := range bankAccounts.accBalances {
		c.accChan <- &utils.Balance{
			BankId:  id,
			Account: account,
			Balance: balance,
		}
	}
}

// Close shuts down the cache correctly,
// such that all current data is backed up.
func Close() (err error) {
	c.quit <- true
	<-c.done
	return
}

// PrintDues prints to the console how much Polka owes to
// each bank, listed in no specific order.
func PrintDues(andAccounts bool) {

	log.Printf("Processed %d transactions.", counter)
	fmt.Printf("Bank dues:\n{\n")

	c.banks.RLock()
	for bank, due := range c.banks.dues {
		fmt.Printf("\t%s: %d\n", bank, due)
	}
	c.banks.RUnlock()
	fmt.Println("}")

	if andAccounts {
		fmt.Printf("}\nAccount dues:\n{\n")

		for _, accBalance := range c.accounts {
			accBalance.RLock()
		}
		for bank, id := range c.bankId {
			fmt.Printf("\t%s: {\n", bank)

			for account, amount := range c.accounts[id].accBalances {
				fmt.Printf("\t\t%d: %d\n", account, amount)
			}
			fmt.Println("\t}")
		}
		fmt.Println("}")

		for _, accBalance := range c.accounts {
			accBalance.RUnlock()
		}
	}
}
