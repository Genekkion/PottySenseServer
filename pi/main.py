import asyncio
import time
from functools import wraps
from quart import Quart, request, jsonify
import os
from dotenv import load_dotenv
import httpx

load_dotenv()

# Constants
HTTP_STATUS_OK = 200
HTTP_STATUS_BAD_REQUEST = 400
HTTP_STATUS_UNAUTHORIZED = 401
HTTP_STATUS_NOT_FOUND = 404
HTTP_STATUS_METHOD_NOT_ALLOWED = 405

MESSAGE_TYPE_ALERT = "alert"
MESSAGE_TYPE_MESSAGE = "message"
MESSAGE_TYPE_COMPLETE = "complete"

# Threshold for when to alert the server
# since the server's 1st ping
START_SESSION_THRESHOLD = 300  # 300s, 5 minutes


async def test_server_connection(url: str, secret_header: str):
    async with httpx.AsyncClient() as client:
        response = await client.get(
            url,
            headers={
                "X-PS-Header": secret_header,
            },
        )
        if response.status_code == HTTP_STATUS_OK:
            print("Test connection with server successful.")
        else:
            print(
                "WARNING: Error during test connection with server. Please check for connection issues!"
            )


# Initial setup of flask app
def create_app():
    SECRET_HEADER = os.getenv("SECRET_HEADER")
    if SECRET_HEADER is None or SECRET_HEADER == "":
        print("Required env variable SECRET_HEADER not set. Exiting.")
        exit()
    SERVER_ADDR = os.getenv("SERVER_ADDR")
    if SERVER_ADDR is None or SERVER_ADDR == "":
        print("Required env variable SERVER_ADDR not set. Exiting.")
        exit()
    start_session_threshold = os.getenv("START_SESSION_THRESHOLD")
    if start_session_threshold is None or start_session_threshold == "":
        print(
            "Optional env variable START_SESSION_THRESHOLD not set. Defaulting to 300s."
        )
    else:
        global START_SESSION_THRESHOLD
        try:
            START_SESSION_THRESHOLD = int(start_session_threshold)
        except ValueError:
            print(
                "Invalid value for env variable START_SESSION_THRESHOLD, defaulting to 300s."
            )

    # test_server_connection(SERVER_ADDR + "/ext", SECRET_HEADER)

    app = Quart(__name__)

    # Constants
    app.config["SECRET_HEADER"] = SECRET_HEADER
    app.config["SERVER_ADDR"] = SERVER_ADDR
    app.config["HEADER_CONFIG"] = {
        "Content-Type": "application/json",
        "X-PS-Header": SECRET_HEADER,
    }

    # Default values for numerical values will be -1
    app.config["timestamp_1"] = -1
    app.config["timestamp_2"] = -1
    app.config["current_client_id"] = -1
    app.config["time_urination"] = -1
    app.config["time_defecation"] = -1

    return app


app = create_app()


async def send_tele_message(message: str, message_type: str):
    current_client_id = app.config.get("current_client_id")
    data = {
        "clientId": current_client_id,
        "message": message,
        "messageType": message_type,
    }

    async with httpx.AsyncClient() as client:
        response = await client.post(
            app.config.get("SERVER_ADDR") + "/ext/api",
            json=data,
            headers=app.config.get("HEADER_CONFIG"),
        )
        print(response)  # print response status code


async def first_timer_warning():
    print("first_timer_warning - start sleeping")
    await asyncio.sleep(START_SESSION_THRESHOLD)
    send_tele_message(
        "Client has yet to have entered the toilet",
        "alert",
    )


# Wraps all the routes
# All routes need to include the
# "X-PS-Header" for the request
def auth_wrapper(func):
    @wraps(func)
    def wrapper_func(*args, **kwargs):
        secret_header = request.headers.get("X-PS-Header")
        if secret_header is None or secret_header != app.config.get("SECRET_HEADER"):
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


@app.route("/")
async def index_handler():
    await test_server_connection(
        app.config.get("SERVER_ADDR") + "/ext",
        app.config.get("SECRET_HEADER"),
    )

    return (
        jsonify(
            {
                "message": "Server is up and running!",
            }
        ),
        HTTP_STATUS_OK,
    )


@app.route("/api", methods=["GET", "POST", "PUT", "DELETE"])
@auth_wrapper
def api_handler():
    if request.method == "GET":
        current_client_id = app.config.get("current_client_id")
        timestamp_1 = app.config.get("timestamp_1")

        if current_client_id == -1 or timestamp_1 == -1:
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
                    "clientId": int(current_client_id),
                    "timeElapsed": int(time.time() - timestamp_1),
                }
            ),
            HTTP_STATUS_OK,
        )

    # For server to obtain the client id
    elif request.method == "POST":
        if not request.is_json or request.json is None:
            return (
                jsonify(
                    {
                        "error": "Mismatched form type.",
                    }
                ),
                HTTP_STATUS_BAD_REQUEST,
            )

        try:
            json_client_id = request.json.get("clientId")
            json_urination = request.json.get("urination")
            json_defecation = request.json.get("defecation")
            if (
                json_client_id is None
                or json_urination is None
                or json_defecation is None
            ):
                return (
                    jsonify(
                        {
                            "error": "Missing clientId, urination or defecation in form data.",
                        }
                    ),
                    HTTP_STATUS_BAD_REQUEST,
                )

            app.config["time_defecation"] = int(json_defecation)
            app.config["current_client_id"] = int(json_urination)
            app.config["timestamp_1"] = time.time()
            app.config["time_urination"] = int(json_urination)

        except ValueError:
            return (
                jsonify(
                    {
                        "error": "clientId, urination and defecation should be integers.",
                    }
                ),
                HTTP_STATUS_BAD_REQUEST,
            )

        app.config["async_tasks"].append(asyncio.create_task(first_timer_warning()))

        return (
            jsonify(
                {
                    "message": "Timer 1 started.",
                }
            ),
            HTTP_STATUS_OK,
        )

    # A call to this route will terminate
    # the current session and reset all
    # parameters
    elif request.method == "DELETE":
        app.config["timestamp_1"] = -1
        app.config["timestamp_2"] = -1
        app.config["current_client_id"] = -1
        app.config["time_urination"] = -1
        app.config["time_defecation"] = -1

        return (
            jsonify(
                {
                    "message": "Session terminated.",
                }
            ),
            HTTP_STATUS_OK,
        )

    # Catch all return statement
    return (
        jsonify(
            {
                "error": "Method not allowed.",
            }
        ),
        HTTP_STATUS_METHOD_NOT_ALLOWED,
    )


app.run()
