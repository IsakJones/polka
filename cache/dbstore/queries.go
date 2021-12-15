package dbstore

const (
	getLatestTransaction = `
	SELECT
		sender.name,
		receiver.name,
		sending_account,
		receiving_account,
		dollar_amount,
		time
	FROM transactions
	JOIN banks sender ON sender.id=transactions.sending_bank_id
	JOIN banks receiver ON receiver.id=transactions.receiving_bank_id
	ORDER BY time;
	`
	updateDues = `
	UPDATE banks
	   SET balance = CASE name
	   				 WHEN $1 THEN balance - $3
					 WHEN $2 THEN balance + $3
					 END
	WHERE name = $1 OR name = $2;
	`
	addDues = `
	UPDATE banks SET balance = balance + $2
	WHERE name=$1;
	`
	subtractDues = `
	UPDATE banks SET balance = balance - $2
	WHERE name=$1;
	`
)
