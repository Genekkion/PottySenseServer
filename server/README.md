# PottySense Web Server

## Overview

This is the main, central http web server which acts as the central hub for communications between the system and users. It includes both the backend and frontend together through the use of HTMX.

## Table of Contents

1. [Installation and Setup](#installation-and-setup)

## Installation and Setup

1. **Install `Go` (skip if installed)**

    To install `Go`, head over to the [official Go website](https://go.dev/dl/) and follow the instructions to install it.

    Run the following command to check if `Go` has been successfully installed.
    ```bash
    go version
    ```

2. **Create and edit the `.env` file**

    Copy the file `.env.example` into a new filed named `.env`. Open up the file using a text editor and add in the relevant information.

    ```bash
    cp .env.example .env
    ```

3. **Install dependencies & Build**

    Run the following commands to install dependencies and build the server.

    ```bash
    go install && go build
    ```

    An executable file named "PottySenseServer" (or something similar, depending on OS) should now be present in the same folder.

4. **Run the server**

    Run the executable and head over to the `LISTEN_ADDR` as specified in the `.env` file or as advertised in the terminal.

## External Routes

This server has some external routes which are ***not*** protected by CSRF so that the APIs are available to call.

However, they are protected through the use of the `X-PS-Header` http request header to minimise potential damage from attackers.


Follow the address as set in the `.env` file or as advertised in the terminal after running the application.

### Server health

1. **Check server health**
    - **Route:** `/ext`
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
    - **Route:** `/ext/api`
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