import os
import redis
import hiredis
from telegram import ForceReply, Update
from telegram.ext import Application, CommandHandler, ContextTypes, MessageHandler, filters

def set_env():
    file = open("./.env", "r")

    for line in file:
        entry = line.split("=")
        if len(entry) != 2:
            continue
        os.environ[entry[0]] = (entry[1])[:-1]

set_env()


redisDB = redis.Redis(host ="localhost", port=6379, decode_responses=True,
                    password=os.environ["REDIS_PASSWORD"])

async def command_template(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    pass

async def start(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if update is None or update.message is None:
        return
    user = update.message.from_user
    if user is None:
        return
    else:
        redisDB.set(user.username.__str__(), update.message.chat_id)

    await update.message.reply_text("Your account has been registered with PottySense! \nPlease return to the portal.")


def main() -> None:
    """Start the bot."""
    TELEGRAM_BOT_TOKEN = os.environ["TELEGRAM_BOT_TOKEN"]
    application = Application.builder().token(TELEGRAM_BOT_TOKEN).build()
    application.add_handler(CommandHandler("start", start))
    application.run_polling(allowed_updates=Update.ALL_TYPES)


if __name__ == "__main__":
    main()
