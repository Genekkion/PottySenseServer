# import statements here
import asyncio
import time
from datetime import datetime, timedelta
from bleak import BleakScanner, BleakClient
import threading
import time
import subprocess
import serial
# import test.find_serial as find_serial
import httpx, multiprocessing as mp
import pygame

# hardware values
GESTURE_CHARACTERISTIC_UUID = "00002a56-0000-1000-8000-00805f9b34fb"
beacon_addr_ls = ["AC:23:3F:63:21:26", "AC:23:3F:63:21:2A"]

# define global variables for logic
ENTERED = False
LEFT = False
D15_NEAR = False
TIME_ENTERED = None
DETECTION = None
BUSINESS = ""
FLUSHED = False
WASHED_HANDS = False

stop_event = threading.Event()

# Initialize pygame mixer 
pygame.mixer.init()
flush_toilet = "./audio/flush_toilet.mp3"
pull_up = "./audio/pull_up.mp3"
wash_hands = "./audio/wash_hands.mp3"
well_done = "./audio/well_done.mp3"
 
# Function to play a single music file 
def play_music(music_file): 
    pygame.mixer.music.load(music_file) 
    pygame.mixer.music.play() 
    while pygame.mixer.music.get_busy():  # Wait for the music to finish playing 
        time.sleep(1) 

# define functions for each part of the flow here

# read serial
def find_serial():
    # Running the command and capturing the output
    result = subprocess.run('python3 -m serial.tools.list_ports', shell=True, capture_output=True, text=True)
    
    # Splitting the output into lines
    output_lines = result.stdout.splitlines()
    
    # Checking if there's at least one line of output to print
    if output_lines:
        first_line = output_lines[0]
        # print(f"First Serial Port: {first_line}")
        return first_line
    else:
        # print("No serial ports found.")
        return None

async def handle_ml():
    """ read serial and set global variable DETECTION to the value of the """
    global DETECTION, BUSINESS, FLUSHED, WASHED_HANDS
    global pull_up
    # port = find_serial.find_serial()
    port = find_serial()
    if port:
        port = port.strip()
        print(f"Using serial port: >{port}<")
    else:
        print("No serial port found")
        return
    print
    with serial.Serial(port=port, baudrate=115200, timeout=2) as s:
        print("Press Ctrl-C to stop")
        print("--------------------")

        # loop forever
        while True:
            # read a line from the serial port, as bytes
            line = s.readline()

            # Decode bytes to string
            line = line.decode('utf-8')
            line = line.rstrip()


            match line:
                case "":
                    print("just noise")
                    DETECTION = "noise"

                case "pee":
                    print("pee detected")
                    DETECTION = "pee"

                case "poo":
                    print("poo detected")
                    DETECTION = "poo"

                case "flushing":
                    print("flushing detected")
                    DETECTION = "flushing"

                case "washing":
                    print("washing detected")
                    DETECTION = "washing"
            
            if DETECTION == "pee" and BUSINESS == "" and not FLUSHED:
                BUSINESS = "pee"

                async with httpx.AsyncClient() as httpClient:
                    response = await httpClient.put(
                        "http://127.0.0.1:5000/int",
                        headers={
                            "X-PS-Header": "cs460slay"
                        },
                        json={
                            "businessType": "urination"
                        }
                    )
                    
                    print(response.read(), response.status_code)

            elif DETECTION == "poo" and BUSINESS != "poo" and not FLUSHED:
                BUSINESS = "poo"

                async with httpx.AsyncClient() as httpClient:
                    response = await httpClient.put(
                        "http://127.0.0.1:5000/int",
                        headers={
                            "X-PS-Header": "cs460slay"
                        },
                        json={
                            "businessType": "defecation"
                        }
                    )
                    
                    print(response.read(), response.status_code)
            elif DETECTION == "flushing" and not FLUSHED:
                FLUSHED = True

                play_music(pull_up)
                play_music(wash_hands)

                async with httpx.AsyncClient() as httpClient:
                    response = await httpClient.post(
                        "http://127.0.0.1:5000/int",
                        headers={
                            "X-PS-Header": "cs460slay"
                        }
                    )
                    
                    print(response.read(), response.status_code)
            elif DETECTION == "washing" and not WASHED_HANDS:
                WASHED_HANDS = True


async def ble_scan():
    global D15_NEAR
    print("Scanning for D15N beacons...")

    
    while not D15_NEAR:
        devices = await BleakScanner.discover(return_adv=True)
        # print("==========")
        # print(time.strftime("%Y-%m-%d %H:%M:%S", time.localtime()))
        # print("----------")

        for addr in beacon_addr_ls:
            try:
                print(addr, "- RSSI =", devices[addr][1].rssi)
                if devices[addr][1].rssi > -60:
                    print("D15N beacon near")

                    async with httpx.AsyncClient() as httpClient:
                        response = await httpClient.get(
                            "http://127.0.0.1:5000/int",
                            headers={
                                "X-PS-Header": "cs460slay"
                            }
                        )
                        
                        print(response.read(), response.status_code)
                    
                        if response.status_code == 200:
                            D15_NEAR = True

            except KeyError as ke:
                print("Beacon %s not found" % addr)
                pass



async def find_ges_ble():
    """ change entered when d15 and door openend, then sets TIME_ENTERED to 
    current time. ENTERED to true
    sets ENTERED to false when door closed after 15 seconds"""
    global LEFT, ENTERED, TIME_ENTERED
    global flush_toilet, wash_hands, well_done
    print("entered find gers ble")
    device_name = "GestureSensor"
    try:
        # Discover Bluetooth devices
        devices = await BleakScanner.discover()
        for device in devices:
            if device.name == device_name:
                address = device.address
                print(f"Found {device_name} at address {address}")
                # Connect to the device using its address
                
                # client = BleakClient(address)
                # try:
                #     await asyncio.wait_for(client.connect(), timeout=10)
                #     if client.is_connected():
                async with BleakClient(address) as client:
                    try:
                        print(f"Connected: {client.address} {client.is_connected}")

                        def notification_handler(sender, data):
                            global D15_NEAR, ENTERED, TIME_ENTERED, LEFT
                            global FLUSHED, WASHED_HANDS
                            msg = data.decode()
                            print("D15_NEAR:", D15_NEAR)
                            if D15_NEAR:
                                print(f"Gesture: {msg}")
                                if "DOOR_OPEN" in msg and not ENTERED:
                                    ENTERED = True
                                    TIME_ENTERED = datetime.now()
                                elif "DOOR_CLOSED" in msg and TIME_ENTERED is not None:
                                    time_difference = datetime.now() - TIME_ENTERED
                                    secs = time_difference.total_seconds()
                                    if ENTERED and secs > 15:
                                        print(
                                            f"User has left the bathroom after {secs}")
                                        ENTERED = False
                                        TIME_ENTERED = None
                                elif "DOOR_CLOSED" in msg:
                                    LEFT = True

                                    if FLUSHED and WASHED_HANDS:
                                        play_music(well_done)
                                    if not FLUSHED:
                                        play_music(flush_toilet)
                                    if not WASHED_HANDS:
                                        play_music(wash_hands)

                        while client.is_connected():
                            await client.start_notify(GESTURE_CHARACTERISTIC_UUID, notification_handler)
                            await asyncio.sleep(1)
                            await client.stop_notify(GESTURE_CHARACTERISTIC_UUID)
                        await client.disconnect()

                    except:
                        print("bopes")
    except:
        print("Error in find_ges_ble() function. Exiting...")

def run_async(async_func):
    asyncio.run(async_func())


def pee_poo_handwash_detection_and_door_sensing():
    global LEFT
    print("Please implement pee-poo-handwash detection and door sensing")
    # loop = asyncio.get_event_loop()
    # loop.run_until_complete(await find_ges_ble())
    pool = mp.Pool()
    pool.map(run_async, [find_ges_ble, handle_ml])
    pool.close()

    while not LEFT:
        pass
    
    pool.terminate()
    

# main loop here
# async def main_loop(*functions):
#     num_funcs = len(functions)
#     print(f"{num_funcs} {'function' if num_funcs == 1 else 'functions'} to run:")
#     for function in functions:
#         print(function)

#     if num_funcs == 0:
#         print("No functions to run. Exiting...")
#     else:
#         i = 0
#         while i < num_funcs:
#             await functions[i]()
#             i += 1

async def main():
    global D15_NEAR, ENTERED, LEFT, TIME_ENTERED, DETECTION
    global BUSINESS, FLUSHED, WASHED_HANDS
    try:
        await ble_scan()
        pee_poo_handwash_detection_and_door_sensing()

    except KeyboardInterrupt:
        print("main interrupted")
    finally:
        ENTERED = False
        LEFT = False
        D15_NEAR = False
        TIME_ENTERED = None
        DETECTION = None
        BUSINESS = ""
        FLUSHED = False
        WASHED_HANDS = False

        print("main ended, this process should close")

def run():
    asyncio.run(main())
# if __name__ == "__main__":
#     # run async server here
#     # --- HERE ---
#     while True:
#         # asyncio.run(main_loop(ble_scan, pee_poo_handwash_detection_and_door_sensing))

#         handle_ml()
#         asyncio.run(
#             main_loop(ble_scan, pee_poo_handwash_detection_and_door_sensing))

