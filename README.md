# Polka Payments

The goal is to simulate an online payments processor (think Stripe) through a concurrent Go REST API.

List of added features:

 - Spammer
  - Accepts a maximum number of workers
  - Add max worker pool and number of transactions as parameter command-line flags
  - spams given number of randomly generated transactions concurrently to API
  - warns if request response is an error
  - send hello request to check if API service works
  - client with custom settings

 - API service
  - /transaction DELETE deletes a given transaction from DB transactions table and sends according request to cache
  - /transaction POST sends given transaction information to cache and updates DB transactions table
  - /transaction GET to get the latest transaction sent (slow, requires sort by time)
  - /hello GET endpoint to check if spammer can connect to API
  - client with custom settings

 - Cache
  - /balance POST updates memory bank and account ledgers 
  - every second the cache updates the DB banks table with its own memory ledger
  - loads bank balances from DB on startup
  - listens for updates from multiple API services

 - DB
  - transactions records all transactions send through /transaction POST to API
   - records name and account number of sender and receiver, dollar amount sent, and time
  - banks records bank names and bank balances with Polka (<0 then owes to Polka, >0 then Polka owes)


List of features to add:

 - Spammer
  - Test for DELETE requests

 - Load balancer
  - Add load balancer
  - Add authentication

 - API
  - Add /transaction PATCH endpoint for different parameters (name or account of sender or receiver)
  - Detect timeout parameter for context

 - Cache
  - Send account balance updates to DB
  - Implement clearing mechanism

 - DB
  - Add account balances (how? relationships? Ask Andrew)

 - General
  - Currency info and currency conversion according to official API
  - Fee calculation + interest rate for banks before clearing (?)
  - Polka liquidity tracker (?)
