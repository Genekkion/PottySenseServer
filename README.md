# *PottySense*

## Introduction
PottySense is a system which features automated tracking of toileting activities.

The server portion consists of two parts, the web server which serves both the frontend and backend functionality, as well as a Telegram bot for ease of use.

## Run
The respective services are able to run individually in their respective folders, but both reference the `.env` file found at the root folder of the project. Additionally, an SQLite3 database and redis cache is required.

The project is also available via Docker and the instructions to use the Docker version is as follows:

1. **Ensure Docker compose is installed**
    
    Run the following commands to check if Docker compose is available on your system. Otherwise, please download it accordingly.
    ```bash
    docker compose version
    ```
    or
    ```bash
    docker-compose version
    ```

1. **Setup `sqlite` database**

    Run the following commands to get a `sqlite` database which has the schema applied to it.
    ```bash
    rm -f sqlite.db
	touch sqlite.db
	sqlite3 sqlite.db < schema.sql
    ```

1. **Create and populate `.env` file**
    Copy the sample `.env.example` file into the `.env` file and fill up the necessary data.
    ```bash
    cp .env.example .env
    ```

1. **Build the images**
    
    Edit the `docker-composse.yml` to expose the ports as desired. Otherwise, run the following command to begin building the images.
    ```bash
    docker compose up -d --build
    ```

1. **Run**

    By default, the ports of the redis and telegram bot containers are not exposed. The web client / server is available at port `3005`. Head over to `localhost:3005` to check it out.
