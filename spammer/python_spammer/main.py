import requests
import random
import json

from enums import *
import funcs 

data = funcs.make_request()

hello_response = requests.post(POLKA+HELLO)
transaction_response = requests.post(POLKA+TRANSACTION, data=data)

print(f"{hello_response.content}")
print(f"{transaction_response.content}")


