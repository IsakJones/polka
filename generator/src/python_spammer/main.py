import requests
import time

from enums import *
import funcs 

from multiprocessing import Pool, cpu_count


    
def send_request(data):
    # hello_response = requests.post(POLKA+HELLO)
    transaction_response = requests.post(POLKA+TRANSACTION, json=data)

    # print(f"{hello_response.content}")
    print(f"{transaction_response.content}")

if __name__ == "__main__":

    data = [funcs.make_request() for i in range(2000)]


    pool = Pool(max(cpu_count()//2, 1))

    # start = time.time()
    pool.map(func=send_request, iterable=data)
    # end = time.time()
    # print(end - start)







