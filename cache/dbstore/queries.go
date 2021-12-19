package dbstore

const (
	bankRetrieveQ = "SELECT name, balance FROM banks;"
	accRetrieveQ  = `
		SELECT
			bank.name,
			account,
			accounts.balance
		FROM accounts
		JOIN banks bank ON bank.id=accounts.bank_id;
	`
	updateBankBalanceQ = `
		UPDATE banks SET balance = $2
		WHERE name = $1;
	`
	updateAccBalanceQ = `
		INSERT INTO accounts (
			bank_id,
			account,
			balance
		) VALUES (
			(SELECT id FROM banks WHERE name=$1),
			$2,
			$3
		) 
		ON CONFLICT (account, bank_id)
		DO UPDATE SET balance = $3
		WHERE accounts.bank_id = (SELECT id FROM banks WHERE name=$1)
		AND accounts.account = $2;
	`
)
