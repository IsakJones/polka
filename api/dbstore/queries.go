package dbstore

const (
	getTransByTrans = `
	SELECT 
		sending_bank_id AND
		receiving_bank_id AND
		sending_account AND
		receiving_account AND
		dollar_amount AND
		time 
	FROM transactions WHERE 
	sending_bank_id=(SELECT id FROM banks WHERE name=$1) AND
	receiving_bank_id=(SELECT id FROM banks WHERE name=$2) AND
	sending_account=$3 AND
	receiving_account=$4 AND
	dollar_amount=$5 AND
	time=$6;
	`
	// getLatestTrans = `
	// SELECT * FROM transactions
	// ORDER BY time
	// LIMIT 1;
	// `
	// getBankNames = `
	// SELECT sender.name, receiver.name
	// FROM banks sender
	// INNER JOIN transactions ON sender.id=$1
	// INNER JOIN banks receiver ON receiver.id=$2
	// LIMIT 1;
	// `

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

	insertTransSQL = `
	INSERT INTO transactions (
		sending_bank_id,
		receiving_bank_id,
		sending_account,
		receiving_account,
		dollar_amount,
		time
	) VALUES (
		(SELECT id FROM banks WHERE name=$1),
		(SELECT id FROM banks WHERE name=$2),
		$3,
		$4,
		$5,
		$6
	);
	`
)
