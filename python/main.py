import os
import redis
from telegram import Update
import telegram
from telegram.ext import Application, CommandHandler, ContextTypes
import sqlite3
import functools


def set_env():
    file = open("./.env", "r")

    for line in file:
        entry = line.split("=")
        if len(entry) != 2:
            continue
        os.environ[entry[0]] = (entry[1])[:-1]


set_env()


redisDB = redis.Redis(host="localhost", port=int(os.environ["REDIS_PORT"]),
                      decode_responses=True, password=os.environ["REDIS_PASSWORD"])

db = sqlite3.connect(os.environ["DB_PATH"])
cursor = db.cursor()


def authorized(func):
    @functools.wraps(func)
    async def wrapper(update: Update, context: ContextTypes.DEFAULT_TYPE):
        if (update is None or update.message is None or
                update.message.from_user is None):
            await update.message.reply_text(
                "Error retrieving user data\! \nPlease try again later\.")
            return
        username = update.message.from_user.username

        res = cursor.execute(
            """SELECT id FROM Tofficers
            WHERE telegram = ?
            """, (username,)).fetchall()
        if len(res) == 0:
            await update.message.reply_text(
                text='Unauthorized user\.',
                parse_mode=telegram.constants.ParseMode.MARKDOWN_V2)
            return

        chatId = redisDB.get(username)
        if chatId is None:
            redisDB.set(name=username, value=update.message.chat_id)

        return await func(update, context)

    return wrapper


async def search_by_id(update: Update, context: ContextTypes.DEFAULT_TYPE, query: list[int]):
    for id in query:
        if id == 0:
            continue
        res = cursor.execute(
            """SELECT first_name, last_name
            FROM Clients
            WHERE id = ?
            """, (id, )).fetchall()
        if len(res) == 0:
            await update.message.reply_text(
                text='Search for "'+str(id)+'" returned no results\.',
                parse_mode=telegram.constants.ParseMode.MARKDOWN_V2)
            continue
        message = "*Search for id:* \[" + str(id) + "\]\n"
        for entry in res:
            message += "\[" + str(id) + "\] "
            message += entry[1] + " " + entry[2] + "\n"
        await update.message.reply_text(
            text=message, parse_mode=telegram.constants.ParseMode.MARKDOWN_V2)


async def search_by_name(update: Update, context: ContextTypes.DEFAULT_TYPE, query: list[str]):
    for name in query:
        query_name = name + "%"
        res = cursor.execute(
            """SELECT id, first_name, last_name
            FROM Clients
            WHERE first_name LIKE ? COLLATE NOCASE
            OR last_name LIKE ? COLLATE NOCASE
            """, (query_name, query_name)).fetchall()
        if len(res) == 0:
            await update.message.reply_text(
                text='Search for "'+name+'" returned no results\.',
                parse_mode=telegram.constants.ParseMode.MARKDOWN_V2)
            continue

        message = '*Search for name:* "' + name + '"\n'
        for entry in res:
            message += "\[" + str(entry[0]) + "\] "
            message += entry[1] + " " + entry[2] + "\n"
        await update.message.reply_text(
            text=message, parse_mode=telegram.constants.ParseMode.MARKDOWN_V2)


def parse_int(s: str) -> int:
    try:
        return int(s)
    except ValueError:
        return 0


@ authorized
async def search(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    args = (update.message.text.split(sep=" "))[1:]
    if args[0].lower() == "id":
        await search_by_id(update, context, [parse_int(s) for s in args[1:]])
    elif args[0].lower() == "name":
        await search_by_name(update, context, args[1:])
    else:
        await update.message.reply_text(
            text="Incorrect arguments supplied\! \nE\.g\. /search name John Jane \nE\.g\. /search id 1 \n",
            parse_mode=telegram.constants.ParseMode.MARKDOWN_V2)


def parse_time(seconds: int) -> str:
    return str(seconds // 60) + " min"


@ authorized
async def get_client(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    args = [parse_int(s) for s in (update.message.text.split(sep=" "))[1:]]
    for id in args:
        if id == 0:
            continue
        res = cursor.execute(
            """SELECT first_name, last_name, urination, defecation
            FROM Clients WHERE id = ?
            """, (id,)).fetchone()
        if res is None:
            await update.message.reply_text(
                text="No client found with id: \["+str(id)+"\]\n",
                parse_mode=telegram.constants.ParseMode.MARKDOWN_V2)
            continue
        message = "*"+res[0].capitalize() + " " + res[1].capitalize() + \
            "* \- " + "\[" + str(id) + "\]\n"
        message += "Average times:\n"
        message += "💧 \- "+parse_time(res[2])+"\n"
        message += "🚽 \- "+parse_time(res[3])

        await update.message.reply_text(
            text=message, parse_mode=telegram.constants.ParseMode.MARKDOWN_V2)


@ authorized
async def current(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    username = update.message.from_user.username
    res = cursor.execute(
        """SELECT Clients.id, Clients.first_name, Clients.last_name
        FROM Clients INNER JOIN Watch
        ON Clients.id = Watch.client_id
        INNER JOIN TOfficers
        ON TOfficers.id = Watch.to_id
        WHERE TOfficers.telegram = ?
        """, (username,)).fetchall()
    message = "*Currently tracking*\n"
    for entry in res:
        message += "[" + str(entry[0]) + "] "
        message += entry[1] + " " + entry[2] + "\n"

    await update.message.reply_text(
        text=message, parse_mode=telegram.constants.ParseMode.MARKDOWN_V2)


@ authorized
async def all_clients(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    res = cursor.execute(
        "SELECT id, first_name, last_name FROM Clients").fetchall()
    message = "*List of clients*\n"
    for entry in res:
        message += "[" + str(entry[0]) + "] "
        message += entry[1] + " " + entry[2] + "\n"
    await update.message.reply_text(
        text=message, parse_mode=telegram.constants.ParseMode.MARKDOWN_V2)


@ authorized
async def track_client(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    args = [parse_int(s) for s in (update.message.text.split(sep=" "))[1:]]
    username = update.message.from_user.username
    res = cursor.execute(
            """SELECT id FROM tofficers
            WHERE telegram = ?""",
            (username,)).fetchone()

    to_id = res[0]

    for id in args:
        if id == 0:
            continue

        res = cursor.execute(
            """INSERT OR IGNORE
            INTO Watch (to_id, client_id)
            VALUES (?, ?)""",
            (to_id, id))

        await update.message.reply_text(
            text="Tracking client \["+str(id)+"\]",
            parse_mode=telegram.constants.ParseMode.MARKDOWN_V2)


@ authorized
async def start(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if (update is not None and update.message is not None and
            update.message.from_user is not None):
        await update.message.reply_text(
            "Your account has been registered with PottySense! \nPlease return to the portal.")
    else:
        await update.message.reply_text(
            "Error registering your account with PottySense! \nPlease try again later.")


def main() -> None:
    """Start the bot."""
    TELEGRAM_BOT_TOKEN = os.environ["TELEGRAM_BOT_TOKEN"]
    application = Application.builder().token(TELEGRAM_BOT_TOKEN).build()
    application.add_handler(CommandHandler("start", start))
    application.add_handler(CommandHandler("clients", all_clients))
    application.add_handler(CommandHandler("current", current))
    application.add_handler(CommandHandler("search", search))
    application.add_handler(CommandHandler("id", get_client))
    application.add_handler(CommandHandler("track", track_client))
    application.run_polling(allowed_updates=Update.ALL_TYPES)


if __name__ == "__main__":
    main()
