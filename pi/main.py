import asyncio
import time
from functools import wraps
from flask import Flask, request, jsonify, current_app
import os
from dotenv import load_dotenv
import requests as rq

load_dotenv()

HTTP_STATUS_OK = 200
HTTP_STATUS_BAD_REQUEST = 400
HTTP_STATUS_UNAUTHORIZED = 401
HTTP_STATUS_METHOD_NOT_FOUND = 404
HTTP_STATUS_METHOD_NOT_ALLOWED = 405

# Threshold for when to alert the server
# since the server's 1st ping
START_SESSION_THRESHOLD = 5 * 60


# Initial setup of flask app
def create_app():
    app = Flask(__name__)

    # Add global variables to app
    with app.app_context():
        # Default values for numerical values will be -1
        current_app.config["timestamp_1"] = -1
        current_app.config["timestamp_2"] = -1
        current_app.config["current_client_id"] = -1
        current_app.config["async_tasks"] = []
    return app


app = create_app()


def METHOD_NOT_FOUND_REPLY():
    return (
        jsonify(
            {
                "error": "Method not allowed.",
            }
        ),
        HTTP_STATUS_METHOD_NOT_FOUND,
    )

async def send_tele_message():
    # TODO:
    url = ""
    try:
        current_client_id = current_app.config["current_cliend_id"]
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


async def first_timer_warning():
    print("first_timer_warning - start sleeping")
    await asyncio.sleep(START_SESSION_THRESHOLD)

    


# Wraps all the routes
# All routes need to include the
# "X-PS-Header" for the request
def auth_wrapper(func):
    @wraps(func)
    def wrapper_func(*args, **kwargs):
        secret_header = request.headers.get("X-PS-Header")
        if secret_header is None or secret_header != os.getenv("SECRET_HEADER"):
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


@app.route("/", methods=["GET", "POST"])
@auth_wrapper
def index_handler():
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
                    "timeElapsed": int(time.time() - timestamp_1),
                }
            ),
            HTTP_STATUS_OK,
        )
    elif request.method == "POST":
        if request.json is None:
            return (
                jsonify(
                    {
                        "error": "Mismatched form type.",
                    }
                ),
                HTTP_STATUS_BAD_REQUEST,
            )

        try:
            current_app.config["current_client_id"] = int(request.json.get("clientId"))
            current_app.config["timestamp_1"] = time.time()

        except ValueError:
            return (
                jsonify(
                    {
                        "error": "clientId should be an integer.",
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
    # the current session
    elif request.method == "DELETE":
        pass

    return METHOD_NOT_FOUND_REPLY()


@app.route("/client", methods=["GET", "POST"])
@auth_wrapper
def client_handler():
    if request.method == "":

        pass
    elif request.method == "":
        pass
    return METHOD_NOT_FOUND_REPLY()
