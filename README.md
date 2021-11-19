# Polka Payments

The goal is to simulate an online payments processor (think Stripe) through a concurrent Go REST API.

Currently, the server only accepts POST requests with three parameters: the paying bank (sender), the paid bank (receiver), and the sum of money transacted (sum).
All the server does is respond with the decoded information to the requests sender, and print how many requests it has processed. I'm working on implementing additional features.

There is also a spammer that sends requests to the server. Instead of sending them sequentially, the spammer first generates random transaction and encodes them,
then asks for permission to spam the server. Once permission is granted, it concurrently sends all the requests, putting a lot of pressure on the server!
