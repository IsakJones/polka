/*
Initially, the database is supposed to hold two tables:

 - banks: tracks the (10) largest US consumer banks, records their name and an id referenced in the next table;
 - transactions: tracks all processed transactions, records all information recorded in the Transaction struct.
*/

-- Creat banks table with built-in names
CREATE TABLE banks (
    id SERIAL,
    name VARCHAR(128),
    balance BIGINT,
    PRIMARY KEY (id),
    UNIQUE (name)
);

INSERT INTO banks (
    name,
    balance
)
VALUES
('JP Morgan Chase', 0),
('Bank of America', 0),
('Wells Fargo', 0),
('Citigroup', 0),
('U.S. Bancorp', 0),
('Truist Financial', 0),
('PNC Financial Services Group', 0),
('TD Group US', 0),
('Bank of New York Mellon', 0),
('Capital One Financial', 0);

-- Create accounts table with index
CREATE TABLE accounts (
    id SERIAL,
    account INT,
    balance INT,
    bank_id INT NOT NULL REFERENCES banks(id) ON DELETE CASCADE,
    UNIQUE (account, bank_id)
);

CREATE UNIQUE INDEX idu ON accounts(account, bank_id);

-- Create transactions table
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