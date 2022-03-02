package dbstore

const (
	getTransByTransQ = `
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
	getLatestPaymentQ = `
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
	insertPaymentQ = `
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
	deletePaymentQ = `
	DELETE FROM transactions WHERE 
	sending_bank_id=(SELECT id FROM banks WHERE name=$1) AND
	receiving_bank_id=(SELECT id FROM banks WHERE name=$2) AND
	sending_account=$3 AND
	receiving_account=$4 AND
	dollar_amount=$5 AND
	time=$6;
	`
)
