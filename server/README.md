# PottySense Web Server

## Overview

This is the main, central http web server which acts as the central hub for communications between the system and users. It includes both the backend and frontend together through the use of HTMX.

## Table of Contents
1. [Docker Setup]
2. [Installation and Setup](#installation-and-setup)
3. [External routes](#external-routes)

## Docker Setup


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
    - **Header** `X-PS-Header`
    - **Body:** none
    - **Expected output:**

        Status code: `200`

        ```json
        {
            "message": "Server is up and running!"
        }
        ```

### PottySense API

1. **Send Telegram message to TOs**
    - **Route:** `/ext/api`
    - **Method:** `POST`
    - **Header** `X-PS-Header`
    - **Body:** 
        ```json
        {
            "clientId": 0,
            "message": "",
            "messageType": ""
        }
        ```
        `messageType` accepts the following values: `alert`, `notification`, `complete`. Any other values will result in a regular message.

    - **Expected output:**
        ```json
        {
            "message": "All messages successfully sent."
        }
        {
            "message": "Some messages successfully sent."
        }
        ```
        The second response occurs when there is at least one failure when sending the telegram messages.

    - **Error responses:**
        - `500 Internal Server Error` if there are no TOs currently tracking this client in the system.
            ```json
            {
               "warning": "No TOs currently tracking this client."
            }
            ```