# PottySense - Pi Server

## Overview

This is the http server which runs on the Pi, in which its main purpose is to listen out for signals from the web server to wake up the rest of the devices and to initiate the session.

## Table of Contents

1. [Installation and Setup](#installation-and-setup)
2. [API Usage](#api-usage)

## Installation and Setup

1. **Set up python virtual environment**

```bash
mkdir -p venv
python -m venv venv
```

Please type the following commands manually in the terminal to ensure that the virtual environment is being used.

```bash
source ./venv/bin/activate
pip install -r requirements.txt
```


2. **Create .env file**

Copy `.env.example` to a new file named `.env`.

```bash
cp .env.example .env
```

Update `.env` with the required environment variables.

3. **Start the server**

```bash
flask --app main run
```

Upon start up of the server, do look out for the first few messages which indicates errors in the `.env` config or for connection issues. Please rectify any connectivity issues before continuing to use this software.

## API Usage

Follow the address as advertised in the cli.

### Server health

1. **Check server health**
    - **Route:** `/`
    - **Method:** any
    - **Body:** none
    - **Expected output:**

        Status code: `200`

        ```json
        {
            "message": "Server is up and running!"
        }
        ```

### PottySense API

1. **Get the current client status**
    - **Route:** `/api`
    - **Method:** `GET`
    - **Body:** none
    - **Expected output:**

        Status code: `200`

        ```json
        {
            "clientId": 0,
            "timeElapsed": 0
        }
        ```
        `clientId` is the id of the client
        
        `timeElapsed` is the time duration in ***seconds*** since the session started

    - **Error responses:**
        - `400 Bad Request` if the session has yet to be started.
            ```json
            {
               "error": "No session found.",
            }
            ```

2. **Start the session**
    - **Route:** `/api`
    - **Method:** `POST`
    - **Body:** `application/json`
        ```json
        {
            "clientId": 0,
            "urination": 0,
            "defecation": 0
        }
        ```
        
        `clientId` is the id of the next client to enter the toilet
        
        `urination` is the time in ***seconds*** for the estimated time taken for this client to complete their urination

        `defecation` is the time in ***seconds*** for the estimated time taken for this client to complete their defecation

    - **Expected output:**

        Status code: `200`

        ```json
        {
            "message": "Timer 1 started.",
        }
        ```

    - **Error responses:**
        - `400 Bad Request` if the `Content-Type` has not been set to `application/json` or there are missing form values.
            ```json
            {
                "error": "Mismatched form type.",
            }

            {
                "error": "Missing clientId, urination or defecation in form data.",
            }
            ```      
        
3. **Terminate the session**
    - **Route:** `/api`
    - **Method:** `DELETE`
    - **Body:** none

    - **Expected output:**

        Status code: `200`

        ```json
        {
            "message": "Session terminated.",
        }
        ```