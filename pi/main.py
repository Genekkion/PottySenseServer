import asyncio
import time
from functools import wraps
from flask import Flask, request, jsonify, current_app
import os
from dotenv import load_dotenv
import requests as rq

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


# Initial setup of flask app
def create_app():
    SECRET_HEADER = os.getenv("SECRET_HEADER")
    if SECRET_HEADER is None:
        print("Required env variable SECRET_HEADER not set. Exiting.")
        exit()
    SERVER_ADDR = os.getenv("SERVER_ADDR")
    if SERVER_ADDR is None:
        print("Required env variable SERVER_ADDR not set. Exiting.")
        exit()
    start_session_threshold = os.getenv("START_SESSION_THRESHOLD")
    if start_session_threshold is None:
        print(
            "Optional env variable START_SESSION_THRESHOLD not set. Defaulting to 300s"
        )
    else:
        global START_SESSION_THRESHOLD
        START_SESSION_THRESHOLD = start_session_threshold

    SERVER_ADDR += "/ext"
    try:
        if rq.get(SERVER_ADDR).status_code == HTTP_STATUS_OK:
            print("Test connection with server successful.")
        else:
            print(
                "WARNING: Error during test connection with server. Please check for connection issues!"
            )
    except rq.RequestException as err:
        print(
            "WARNING: Error during test connection with server. Please check for connection issues!"
        )
        print("Error: ", err)

    app = Flask(__name__)

    # Add global variables to app
    with app.app_context():
        # Constants
        current_app.config["SECRET_HEADER"] = SECRET_HEADER
        current_app.config["SERVER_ADDR"] = SERVER_ADDR + "/api"

        # Default values for numerical values will be -1
        current_app.config["timestamp_1"] = -1
        current_app.config["timestamp_2"] = -1
        current_app.config["current_client_id"] = -1
        current_app.config["time_urination"] = -1
        current_app.config["time_defecation"] = -1
        current_app.config["async_tasks"] = []
    return app


app = create_app()


def METHOD_NOT_ALLOWED_REPLY():
    return (
        jsonify(
            {
                "error": "Method not allowed.",
            }
        ),
        HTTP_STATUS_METHOD_NOT_ALLOWED,
    )


async def send_tele_message(message: str, message_type: str):
    url = current_app.config["SERVER_ADDR"]
    try:
        current_client_id = current_app.config["current_client_id"]
        data = {
            "message": "Client "
            + str(current_client_id)
            + " has yet to enter the toilet."
        }
        # send post request to server
        # for a messgae

        response = rq.post(url, json=data)

    except rq.RequestException:
        pass


async def first_timer_warning():
    print("first_timer_warning - start sleeping")
    await asyncio.sleep(START_SESSION_THRESHOLD)
    # TODO:
    url = ""
    try:
        current_client_id = current_app.config["current_client_id"]
        data = {
            "message": "Client "
            + str(current_client_id)
            + " has yet to enter the toilet."
        }
        # send post request to server
        # for a messgae

        response = rq.post(url, json=data)

    except rq.RequestException:
        pass

    pass


# Wraps all the routes
# All routes need to include the
# "X-PS-Header" for the request
def auth_wrapper(func):
    @wraps(func)
    def wrapper_func(*args, **kwargs):
        secret_header = request.headers.get("X-PS-Header")
        if (
            secret_header is None
            or secret_header != current_app.config["SECRET_HEADER"]
        ):
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
def index_handler():
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
        current_client_id = current_app.config.get("current_client_id")
        timestamp_1 = current_app.config.get("timestamp_1")

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

            current_app.config["time_urination"] = int(json_urination)
            current_app.config["time_defecation"] = int(json_defecation)
            current_app.config["current_client_id"] = int(json_urination)
            current_app.config["timestamp_1"] = time.time()

        except ValueError:
            return (
                jsonify(
                    {
                        "error": "clientId, urination and defecation should be integers.",
                    }
                ),
                HTTP_STATUS_BAD_REQUEST,
            )

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
        current_app.config["timestamp_1"] = -1
        current_app.config["timestamp_2"] = -1
        current_app.config["current_client_id"] = -1
        current_app.config["time_urination"] = -1
        current_app.config["time_defecation"] = -1
        current_app.config["async_tasks"] = []

        for task in current_app.config["async_tasks"]:
            task.cacnel()
        current_app.config["async_tasks"].clear()
        return (
            jsonify(
                {
                    "message": "Session terminated.",
                }
            ),
            HTTP_STATUS_OK,
        )

    # Catch all return statement
    # Only "PUT" should reach here
    return METHOD_NOT_ALLOWED_REPLY()
