import asyncio
import time
from bleak import BleakScanner, BleakClient

peripheral_addr = "B7:C3:57:2D:E9:48"

async def main(addr):
	async with BleakClient(addr) as client:
		print("Connected to %s" % addr)
		try:
			while True:
				print(await client.read_gatt_char("19b10001-e8f2-537e-4f6c-d104768a1214"))
				time.sleep(1)
		except:
			client.disconnect()

asyncio.run(main(peripheral_addr))
