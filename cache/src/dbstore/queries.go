package dbstore

const (
	bankNumQ      = "SELECT COUNT(*) FROM banks;"
	bankRetrieveQ = "SELECT id, name, balance FROM banks;"
	accRetrieveQ  = `
		SELECT 	banks.name, 
				accounts.account, 
				accounts.balance
		FROM banks
		JOIN accounts ON 
		accounts.bank_id = banks.id;
	`
	updateBankBalanceQ = "UPDATE banks SET balance = $2 WHERE id = $1;"
	updateAccBalanceQ  = `
		INSERT INTO accounts (
			bank_id,
			account,
			balance
		) VALUES (
			$1,
			$2,
			$3
		) 
		ON CONFLICT (account, bank_id)
		DO UPDATE SET balance = $3
		WHERE accounts.bank_id = $1
		AND accounts.account = $2;
	`
)
