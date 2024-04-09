import asyncio
from bleak import BleakClient, BleakScanner

# device_name = "B7:C3:57:2D:E9:48"
# device_name = "90:BF:5A:5E:05:AA"

GESTURE_CHARACTERISTIC_UUID = "00002a56-0000-1000-8000-00805f9b34fb"


async def find_ges_ble():
    device_name = "GestureSensor"
    try:
        # Discover Bluetooth devices
        devices = await BleakScanner.discover()
        for device in devices:
            if device.name == device_name:
                address = device.address
                print(f"Found {device_name} at address {address}")
                # Connect to the device using its address
                async with BleakClient(address) as client:
                    print(f"Connected: {client.is_connected}")

                    def notification_handler(sender, data):
                        print(f"Gesture: {data.decode()}")

                    while True:
                        await client.start_notify(GESTURE_CHARACTERISTIC_UUID, notification_handler)
                        await asyncio.sleep(1)
                        await client.stop_notify(GESTURE_CHARACTERISTIC_UUID)
    except:
        print("Error in find_ges_ble() function. Exiting...")
              

if __name__ == "__main__":
    loop = asyncio.get_event_loop()
    loop.run_until_complete(find_ges_ble())
