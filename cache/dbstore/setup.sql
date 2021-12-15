/*
Initially, the database is supposed to hold two tables:

 - banks: tracks the (10) largest US consumer banks, records their name and an id referenced in the next table;
 - transactions: tracks all processed transactions, records all information recorded in the Transaction struct.
*/

CREATE TABLE banks (
    id SERIAL,
    name VARCHAR(128),
    balance BIGINT,
    PRIMARY KEY (id),
    UNIQUE (name)
);

INSERT INTO banks (
    name
)
VALUES
('JP Morgan Chase'),
('Bank of America'),
('Wells Fargo'),
('Citigroup'),
('U.S. Bancorp'),
('Truist Financial'),
('PNC Financial Services Group'),
('TD Group US'),
('Bank of New York Mellon'),
('Capital One Financial');


CREATE TABLE transactions (
    id SERIAL,
    sending_account INT,
    receiving_account INT,
    dollar_amount INT NOT NULL,
    time TIMESTAMP,
    PRIMARY KEY (id),
    sending_bank_id INT NOT NULL REFERENCES banks(id) ON DELETE CASCADE,
    receiving_bank_id INT NOT NULL REFERENCES banks(id) ON DELETE CASCADE,
    UNIQUE (
        sending_bank_id,
        receiving_bank_id,
        sending_account,
        receiving_account,
        dollar_amount,
        time
    )
);