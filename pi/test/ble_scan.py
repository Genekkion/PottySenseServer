import asyncio
import time
from bleak import BleakScanner

mac_addr_ls = ["AC:23:3F:63:21:26","AC:23:3F:63:21:2A","B7:C3:57:2D:E9:48"]

async def main(addr_ls):
    devices = await BleakScanner.discover(return_adv=True)
    
    print("==========")
    print(time.strftime("%Y-%m-%d %H:%M:%S", time.localtime()))
    print("----------")
    #for addr, (device, adv_data) in devices.items():
    #  print(addr)
    #  print(device)
    #  print(adv_data)
    #  print("++++++++++")
    for addr in addr_ls:
        try:
            print(addr, "- RSSI =", devices[addr][1].rssi)
            print(devices[addr][1])
        except KeyError as ke:
            print("Beacon %s not found" % addr)

# print("Scanning for:")
# for mac_addr in mac_addr_ls:
#     print(mac_addr)
# print("...\n")
    
# while True:
#     asyncio.run(main(mac_addr_ls))

def run():
    while True:
        asyncio.run(main(mac_addr_ls))
