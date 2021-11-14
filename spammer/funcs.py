import random
import enums


def make_request(lo=1, hi=1000) -> dict:
    
    sum = random.randint(lo, hi)
    sender = random.choice(enums.BANKS)
    
    receiver = random.choice(enums.BANKS)
    while receiver == sender:
        receiver = random.choice(enums.BANKS)

    return {
        "sender": sender,
        "receiver": receiver,
        "sum": sum,
    }



