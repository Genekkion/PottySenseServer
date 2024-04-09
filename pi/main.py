import asyncio
import time
from functools import wraps
from quart import Quart, request, jsonify
import os
from dotenv import load_dotenv
import httpx
import multiprocessing as mp, datetime
import toilet_functions as toilet

load_dotenv(dotenv_path="../.env")

# Constants
HTTP_STATUS_OK: int = 200
HTTP_STATUS_BAD_REQUEST: int = 400
HTTP_STATUS_UNAUTHORIZED: int = 401
HTTP_STATUS_NOT_FOUND: int = 404
HTTP_STATUS_METHOD_NOT_ALLOWED: int = 405

MESSAGE_TYPE_ALERT: str = "alert"
MESSAGE_TYPE_MESSAGE: str = "message"
MESSAGE_TYPE_COMPLETE: str = "complete"

# Config key naming convention
TIMESTAMP_1: str = "timestamp_1"  # Session start
TIMESTAMP_2: str = "timestamp_2"  # Client enters
TIMESTAMP_3: str = "timestamp_3"  # Client finishes business
CLIENT_ID: str = "client_id"
BUSINESS_TYPE: str = "business_type"
TIME_URINATION: str = "time_urination"
TIME_DEFECATION: str = "time_defecation"
TIMER: str = "timer_1"
TIMER_2: str = "timer_2"
SECRET_HEADER: str = "secret_header"
SERVER_ADDR: str = "server_addr"
HEADER_CONFIG: str = "header_config"
TIMER_1_THRESHOLD: str = "timer_1_threshold"
TIMER_3_THRESHOLD: str = "timer_3_threshold"
PHASE: str = "phase"

# Special constants
HEADER_NAME: str = "X-PS-Header"

running_processes = []

def clear_processes():
    global running_processes
    if len(running_processes) > 0:
        for p in running_processes:
            p.terminate()
        running_processes = []

"""
In total there will be 3 timers started by this server
and will send notification to the TO where necessary.

Timer 1:
WHEN: started upon this server receives notice of a new
client entering
TRIGGER: default of 300 seconds (5 minutes) or
    otherwise changed from env

Timer 2:
WHEN: if client takes more time than usual for urination / defecation
TRIGGER: value from web server


Timer 3:
WHEN: if client takes a long time AFTER they have finished their business
TRIGGER: default of 300 seconds (5 minutes) or
    otherwise changed from env
"""


def reset_config(config: dict) -> None:
    # Default values for numerical values will be -1
    config[TIMESTAMP_1] = -1
    config[TIMESTAMP_2] = -1
    config[TIMESTAMP_3] = -1
    config[CLIENT_ID] = -1
    config[TIME_URINATION] = -1
    config[TIME_DEFECATION] = -1
    config[PHASE] = -1

    config[TIMER] = None
    config[TIMER_2] = None


# Initial setup of flask app
def create_app() -> Quart:
    secret_header = os.getenv("SECRET_HEADER")

    if secret_header is None or secret_header == "":
        print("Required env variable SECRET_HEADER not set. Exiting.")
        exit()

    server_addr = os.getenv("SERVER_ADDR")

    if server_addr is None or server_addr == "":
        print("Required env variable SERVER_ADDR not set. Exiting.")
        exit()

    timer_1_threshold = os.getenv("PI_TIMER_START_THRESHOLD")
    if timer_1_threshold is None or timer_1_threshold == "":
        print(
            "Optional env variable PI_TIMER_START_THRESHOLD not set. Defaulting to 300s."
        )
        timer_1_threshold = 300
    else:
        try:
            timer_1_threshold = int(timer_1_threshold)
        except ValueError:
            print(
                "Invalid value for env variable PI_TIMER_START_THRESHOLD, defaulting to 300s."
            )
            timer_1_threshold = 300

    timer_3_threshold = os.getenv("PI_TIMER_END_THRESHOLD")
    if timer_3_threshold is None or timer_3_threshold == "":
        print(
            "Optional env variable PI_TIMER_END_THRESHOLD not set. Defaulting to 300s."
        )
        timer_3_threshold = 300
    else:
        try:
            timer_3_threshold = int(timer_3_threshold)
        except ValueError:
            print(
                "Invalid value for env variable PI_TIMER_END_THRESHOLD, defaulting to 300s."
            )
            timer_3_threshold = 300

    app: Quart = Quart(__name__)

    # Constants
    app.config[SECRET_HEADER] = secret_header
    app.config[SERVER_ADDR] = "https://" + server_addr
    app.config[HEADER_CONFIG] = {
        "Content-Type": "application/json",
        "X-PS-Header": secret_header,
    }
    app.config[TIMER_1_THRESHOLD] = timer_1_threshold
    app.config[TIMER_3_THRESHOLD] = timer_3_threshold
    reset_config(app.config)
    return app


app: Quart = create_app()


# Returns the appropriate http status code
async def send_tele_message(
    message: str,
    message_type: str,
    # ,silent_message: bool = True
) -> int:
    current_client_id = app.config.get("client_id")
    data = {
        "clientId": current_client_id,
        "message": message,
        "messageType": message_type,
        # "silentMessage": silent_message,
    }

    async with httpx.AsyncClient() as client:
        response = await client.post(
            app.config[SERVER_ADDR] + "/ext/api",
            json=data,
            headers=app.config[HEADER_CONFIG],
        )
        # print response status code
        print(response.read(), response.status_code)
        return response.status_code


# Returns True if timer and message sends successfully,
# or False if interrupted or errors occured when sending message
async def start_timer_1(duration: int):
    print("Session timer started for", duration, "seconds")
    try:
        await asyncio.sleep(duration)
        print("timer 1 time up")
        clear_processes()
        return (
            await send_tele_message(
                "Client "
                + str(app.config[CLIENT_ID])
                + " has yet to have entered the toilet.",
                "alert",
                # silent_message=False,
            )
            == HTTP_STATUS_OK
        )
    except asyncio.CancelledError:
        print("timer 1 cancelled")
        return False


# Returns True if timer and message sends successfully,
# or False if interrupted or errors occured when sending message
async def start_timer_2(duration: int, business_type: str):
    print(business_type + " timer started")
    try:
        await asyncio.sleep(duration)
    except asyncio.CancelledError:
        print("timer 2 cancelled")
        return False
    print("timer 2 time up")
    clear_processes()
    return (
        await send_tele_message(
            "Client "
            + str(app.config[CLIENT_ID])
            + " is taking too long for "
            + business_type
            + ".",
            "alert",
            # silent_message=False,
        )
        == HTTP_STATUS_OK
    )


# Returns True if timer and message sends successfully,
# or False if interrupted or errors occured when sending message
async def start_timer_3(duration: int):
    print("Toileting finished, end timer started")
    try:
        await asyncio.sleep(duration)
    except asyncio.CancelledError:
        print("timer 3 cancelled")
        return False
    print("timer 3 time up")
    clear_processes()
    return (
        await send_tele_message(
            "Client "
            + str(app.config[CLIENT_ID])
            + " has finished their business and is taking an unusually long time to leave.",
            "alert",
            # silent_message=False,
        )
        == HTTP_STATUS_OK
    )


# Wraps all the routes
# All routes need to include the
# HEADER_NAME for the request
def auth_wrapper(func):
    @wraps(func)
    def wrapper_func(*args, **kwargs):
        secret_header = request.headers.get(HEADER_NAME)

        if secret_header is None or secret_header != app.config[SECRET_HEADER]:
            return (
                jsonify(
                    {
                        "error": "Unauthorized.",
                    }
                ),
                HTTP_STATUS_UNAUTHORIZED,
            )
        return func(*args, **kwargs)

    return wrapper_func


# Returns the http status code from the web server
async def test_server_connection(url: str, secret_header: str) -> int:
    async with httpx.AsyncClient() as client:
        response = await client.get(
            url,
            headers={
                HEADER_NAME: secret_header,
            },
        )
        message: str = ""
        if response.status_code == HTTP_STATUS_OK:
            message = "Test connection with server successful."
        else:
            message = "WARNING: Error during test connection with server. Please check for connection issues!"

        print(message)
        return response.status_code


@app.route("/")
async def index_handler():
    web_server_status_code: int = await test_server_connection(
        app.config[SERVER_ADDR] + "/ext",
        app.config[SECRET_HEADER],
    )

    return (
        jsonify(
            {
                "message": "Server is up and running!",
                "webServer": web_server_status_code,
            }
        ),
        HTTP_STATUS_OK,
    )


@auth_wrapper
@app.get("/api")
async def api_handler_get():
    # No session found since these parameters would
    # have been set if there was one
    if app.config[CLIENT_ID] == -1 or app.config[TIMESTAMP_1] == -1:
        return (
            jsonify(
                {
                    "error": "No session found.",
                }
            ),
            HTTP_STATUS_BAD_REQUEST,
        )

    return (
        jsonify(
            {
                "clientId": int(app.config[CLIENT_ID]),
                "timeElapsed": int(time.time() - app.config[TIMESTAMP_1]),
                # timeElapsed will be rounded down to int from float
                "phase": int(app.config[PHASE]),
            }
        ),
        HTTP_STATUS_OK,
    )


@auth_wrapper
@app.post("/api")
async def api_handler_post():
    if not request.is_json or await request.get_json() is None:
        return (
            jsonify(
                {
                    "error": "Mismatched form type.",
                }
            ),
            HTTP_STATUS_BAD_REQUEST,
        )

    try:
        data = await request.get_json()
        json_client_id = data.get("clientId")
        json_business_type = data.get("businessType")
        json_urination = data.get("urination")
        json_defecation = data.get("defecation")
        if json_client_id is None or json_urination is None or json_defecation is None:
            return (
                jsonify(
                    {
                        "error": "Missing clientId, urination or defecation in form data.",
                    }
                ),
                HTTP_STATUS_BAD_REQUEST,
            )

        app.config[CLIENT_ID] = int(json_client_id)
        app.config[BUSINESS_TYPE] = json_business_type
        app.config[TIME_URINATION] = int(json_urination)
        app.config[TIME_DEFECATION] = int(json_defecation)
        app.config[TIMESTAMP_1] = time.time()
        app.config[PHASE] = 1

    except ValueError:
        return (
            jsonify(
                {
                    "error": "clientId, urination and defecation should be integers.",
                }
            ),
            HTTP_STATUS_BAD_REQUEST,
        )

    # STOP CURRENT TIMER1 if have
    # START NEW TIMER
    if app.config[TIMER] is not None and not app.config[TIMER].done():
        app.config[TIMER].cancel()

    clear_processes()
    print(running_processes)
    p1 = mp.Process(target=toilet.run) 
    running_processes.append(p1)
    p1.start()

    app.config[TIMER] = asyncio.create_task(
        start_timer_1(int(app.config[TIMER_1_THRESHOLD]))
    )

    return (
        jsonify(
            {
                "message": "Timer 1 started.",
            }
        ),
        HTTP_STATUS_OK,
    )


@auth_wrapper
@app.delete("/api")
async def api_handler_delete():
    # A call to this route will terminate
    # the current session and reset all
    # parameters
    if app.config[TIMER] is not None and not app.config[TIMER].done():
        app.config[TIMER].cancel()
    clear_processes()
    reset_config(app.config)

    return (
        jsonify(
            {
                "message": "Session terminated.",
            }
        ),
        HTTP_STATUS_OK,
    )


# For use for internal communication within the toilet
@auth_wrapper
@app.get("/int")
async def int_handler_get():
    # For beacon sensing process to call to start
    # the timer for within toilet

    # timer_1 is reset since there is client has entered
    if not app.config[TIMER].done():
        app.config[TIMER].cancel()

    app.config[TIMER] = asyncio.create_task(
        start_timer_2(
            app.config[
                (
                    TIME_URINATION
                    if app.config[BUSINESS_TYPE] == "urination"
                    else TIME_DEFECATION
                )
            ],
            app.config[BUSINESS_TYPE],
        )
    )
    app.config[TIMESTAMP_2] = time.time()
    app.config[PHASE] = 2
    return (
        jsonify(
            {
                "message": "Timer 2 started.",
            }
        ),
        HTTP_STATUS_OK,
    )


@auth_wrapper
@app.post("/int")
async def int_handler_post():
    # For the controllers to inform this server
    # that the client has completed their business

    # timer_1 is reset since there is client has entered
    if app.config[TIMER] is not None and not app.config[TIMER].done():
        app.config[TIMER].cancel()

    app.config[TIMER] = asyncio.create_task(
        start_timer_3(
            app.config[TIMER_3_THRESHOLD],
        )
    )
    app.config[TIMESTAMP_3] = time.time()
    app.config[PHASE] = 3
    return (
        jsonify(
            {
                "message": "Timer 3 started.",
            }
        ),
        HTTP_STATUS_OK,
    )


@auth_wrapper
@app.put("/int")
async def int_handler_put():
    # For the controllers to update this server
    # of the type of business

    if not request.is_json or await request.get_json() is None:
        return (
            jsonify(
                {
                    "error": "Mismatched form type.",
                }
            ),
            HTTP_STATUS_BAD_REQUEST,
        )

    try:
        data = await request.get_json()
        json_business_type = data.get("businessType")
        if json_business_type is None:
            return (
                jsonify(
                    {
                        "error": "Missing businessType in form data.",
                    }
                ),
                HTTP_STATUS_BAD_REQUEST,
            )

        app.config[BUSINESS_TYPE] = json_business_type

    except ValueError:
        return (
            jsonify(
                {
                    "error": "clientId, urination and defecation should be integers.",
                }
            ),
            HTTP_STATUS_BAD_REQUEST,
        )

    # timer_1 is reset since there is client has entered
    if app.config[TIMER] is not None and not app.config[TIMER].done():
        app.config[TIMER].cancel()

    time_elapsed: float = time.time() - app.config[TIMESTAMP_2]

    app.config[TIMER] = asyncio.create_task(
        start_timer_2(
            app.config[
                (
                    TIME_URINATION
                    if app.config[BUSINESS_TYPE] == "urination"
                    else TIME_DEFECATION
                )
            ]
            - time_elapsed,
            app.config[BUSINESS_TYPE],
        )
    )

    return (
        jsonify(
            {
                "message": "Business type updated.",
            }
        ),
        HTTP_STATUS_OK,
    )


@auth_wrapper
@app.delete("/int")
async def int_handler():
    # For the controllers to inform this server
    # that the client has left the toilet

    # timer_1 is reset since there is client has entered
    if app.config[TIMER] is not None and not app.config[TIMER].done():
        app.config[TIMER].cancel()

    # TODO: Update server with the new durations

    status_code: int = await send_tele_message(
        "Client has completed their toileting and has left the toilet.",
        MESSAGE_TYPE_COMPLETE,
    )
    print(status_code)

    reset_config(app.config)
    return (
        jsonify(
            {
                "message": "Session ended.",
            }
        ),
        HTTP_STATUS_OK,
    )


if __name__ == "__main__":
    app.run(host=os.getenv("PI_HOST"), port=os.getenv("PI_PORT"))
