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

Copy `../.env.example` to a new file named `../.env`.

```bash
cp ../.env.example ../.env
```

Update `../.env` with the required environment variables.

3. **Start the server**

```bash
python main.py
```

Upon start up of the server, do look out for the first few messages which indicates errors in the `.env` config or for connection issues. Please rectify any connectivity issues before continuing to use this software.

## API Usage

Follow the address as advertised in the cli.

### Server health

1. **Check server health**
    - **Route:** `/`
    - **Method:** any
    - **Header** `X-PS-Header`
    - **Body:** none
    - **Expected output:**

        Status code: `200`

        ```json
        {
            "message": "Server is up and running!"
        }
        ```

### External API

1. **Get the current client status**
    - **Route:** `/api`
    - **Method:** `GET`
    - **Header** `X-PS-Header`
    - **Body:** none
    - **Expected output:**

        Status code: `200`

        ```json
        {
            "clientId": 0,
            "timeElapsed": 0,
            "phase": 0
        }
        ```
        `clientId` is the id of the client
        
        `timeElapsed` is the time duration in ***seconds*** since the session started

        `phase` is the phase of the session. `1` if session has started, `2` if client has entered toilet and is performing their toileting activities, `3` if client has completed toileting but has yet to leave

    - **Error responses:**
        - `400 Bad Request` if the session has yet to be started.
            ```json
            {
               "error": "No session found."
            }
            ```

2. **Start the session**
    - **Route:** `/api`
    - **Method:** `POST`
    - **Header** `X-PS-Header`
    - **Body:** `application/json`
        ```json
        {
            "clientId": 0,
            "businessType": "",
            "urination": 0,
            "defecation": 0
        }
        ```
        
        `clientId` is the id of the next client to enter the toilet
        
        `businessType` is the type of toileting business. Accepts either `urination` or `defecation`

        `urination` is the time in ***seconds*** for the estimated time taken for this client to complete their urination

        `defecation` is the time in ***seconds*** for the estimated time taken for this client to complete their defecation

    - **Expected output:**

        Status code: `200`

        ```json
        {
            "message": "Timer 1 started."
        }
        ```

    - **Error responses:**
        - `400 Bad Request` if the `Content-Type` has not been set to `application/json` or there are missing form values.
            ```json
            {
                "error": "Mismatched form type."
            }

            {
                "error": "Missing clientId, urination or defecation in form data."
            }
            ```      
        
3. **Terminate the session**
    - **Route:** `/api`
    - **Method:** `DELETE`
    - **Header** `X-PS-Header`
    - **Body:** none

    - **Expected output:**

        Status code: `200`

        ```json
        {
            "message": "Session terminated."
        }
        ```

### Internal API

The following routes are only to be called by the devices within the toilet.

1. **Client enter toilet**
    - **Route:** `/int`
    - **Method:** `GET`
    - **Header** `X-PS-Header`
    - **Body:** none
    - **Expected output:**

        Status code: `200`

        ```json
        {
            "message": "Timer 2 started."
        }
        ```

2. **Client finish toileting**
    - **Route:** `/int`
    - **Method:** `POST`
    - **Header** `X-PS-Header`
    - **Body:** `none`

    - **Expected output:**

        Status code: `200`

        ```json
        {
            "message": "Timer 3 started."
        }

3. **Client finish toileting**
    - **Route:** `/int`
    - **Method:** `PUT`
    - **Header** `X-PS-Header`
    - **Body:** `application/json`
        ```json
        {
            "businessType": ""
        }
        ```
        
        `businessType` is the type of toileting business. Accepts either `urination` or `defecation`


    - **Expected output:**

        Status code: `200`

        ```json
        {
            "message": "Business type updated."
        }

4. **Client left toilet**
    - **Route:** `/int`
    - **Method:** `DELETE`
    - **Header** `X-PS-Header`
    - **Body:** none

    - **Expected output:**

        Status code: `200`

        ```json
        {
            "message": "Session ended."
        }
        ```
