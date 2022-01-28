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

// Positive values are owed by Polka to the bank,
// Negative values are owed by the bank to Polka.
type Cache struct {
	ctx            context.Context
	list           *circularLinkedList
	bankIds        map[string]uint16
	banks          map[uint16]*int64            // maps bankids to bank balances (pointers)
	accounts       map[uint16]map[uint16]*int32 // maps bankids and accounts to account balances (pointers)
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
	bankChan chan<- *utils.BankBalance, // Channel to send bank balances
	accChan chan<- *utils.Balance, // Channel to send account balances
	bankRetChan <-chan *utils.BankBalance, // Channel to retrieve bank balances
	accRetChan <-chan *utils.Balance, // Channel to retrieve account balances
) (err error) {

	logger := log.New(os.Stderr, "[cache] ", log.LstdFlags|log.Lshortfile)
	// Prep bank dues singleton with number of banks
	bankNum := <-bankNumChan

	bankIds := make(map[string]uint16, bankNum)
	bankBalances := make(map[uint16]*int64, bankNum)
	accountBalances := make(map[uint16]map[uint16]*int32, bankNum)

	// Initialize circular linked list
	list := &circularLinkedList{}

	// Assign Cache singleton
	c = Cache{
		ctx:            ctx,
		list:           list,
		logger:         logger,
		bankIds:        bankIds,
		accChan:        accChan,
		banks:          bankBalances,
		bankChan:       bankChan,
		accounts:       accountBalances,
		quit:           make(chan bool),
		done:           make(chan bool),
		backupInterval: backupInterval,
	}

	// Atomically update bank balances retieved from database
	for balance := range bankRetChan {
		// Update circular linked list and bank id map
		c.list.add(balance.BankId)
		c.bankIds[balance.Name] = balance.BankId
		// Update value
		var bankPtr *int64 = c.banks[balance.BankId]
		atomic.StoreInt64(bankPtr, balance.Balance)
		// Make map for bank accounts
		c.accounts[balance.BankId] = make(map[uint16]*int32)
	}

	// Atomically update account balances retieved from database
	for balance := range accRetChan {
		// Update value
		var accPtr *int32 = c.accounts[balance.BankId][balance.Account]
		atomic.StoreInt32(accPtr, balance.Balance)
	}

	// Update periodically DB records at regular intervals
	go ManageDatabaseBackups()

	return
}

// UpdateDues changes the dues according to clearinghouse logic.
func UpdateDues(current *utils.SRBalance) (err error) {
	// Update counter and retrieve bank ids
	atomic.AddUint64(&counter, 1)
	senderId := c.bankIds[current.Sender.Name]
	receiverId := c.bankIds[current.Receiver.Name]

	// Update sender bank's balance
	senPtr := c.banks[senderId]
	atomic.AddInt64(senPtr, int64(current.Amount))

	// Update receiving bank's balance
	recPtr := c.banks[receiverId]
	atomic.AddInt64(recPtr, int64(current.Amount))

	// Update sending account's balance
	sAccPtr := c.accounts[senderId][current.Sender.Account]
	atomic.AddInt32(sAccPtr, current.Amount)

	// Update receiving account's balance
	rAccPtr := c.accounts[receiverId][current.Receiver.Account]
	atomic.AddInt32(rAccPtr, current.Amount)

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
			for i := 0; i < len(c.bankIds); i++ {
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

	for bankId, balance := range c.banks {
		c.bankChan <- &utils.BankBalance{
			BankId:  bankId,
			Balance: *balance,
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

	for account, balance := range bankAccounts {
		c.accChan <- &utils.Balance{
			BankId:  id,
			Account: account,
			Balance: *balance,
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
func PrintBalances(andAccounts bool) {

	log.Printf("Processed %d transactions.", counter)
	fmt.Printf("Bank balances:\n{\n")

	// Print bank balances
	for name, id := range c.bankIds {
		fmt.Printf("\t%s: %d\n", name, c.banks[id])
	}
	fmt.Println("}")

	// If specified, print account balances
	if andAccounts {
		fmt.Printf("}\nAccount balances:\n{\n")

		for name, id := range c.bankIds {
			fmt.Printf("\t%s: {\n", name)

			for account, amount := range c.accounts[id] {
				fmt.Printf("\t\t%d: %d\n", account, amount)
			}
			fmt.Println("\t}")
		}
		fmt.Println("}")
	}
}
