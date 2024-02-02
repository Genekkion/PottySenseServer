import os

from telegram import ForceReply, Update
from telegram.ext import Application, CommandHandler, ContextTypes, MessageHandler, filters

def set_env():
    file = open(".env", "r")

    for line in file:
        entry = line.split("=")
        if len(entry) != 2:
            continue
        os.environ[entry[0]] = (entry[1])[:-1]



async def command_template(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    pass

async def start(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if update is None or update.message is None:
        return
    await update.message.reply_text("Hi !")


async def help_command(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if update is None or update.message is None:
        return
    await update.message.reply_text("Help!")


async def echo(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if update is None or update.message is None or update.message.text is None:
        return
    await update.message.reply_text(update.message.text)


def main() -> None:
    """Start the bot."""
    TELEGRAM_BOT_TOKEN = os.environ["TELEGRAM_BOT_TOKEN"]
    print("token", TELEGRAM_BOT_TOKEN)
    # Create the Application and pass it your bot's token.
    application = Application.builder().token(TELEGRAM_BOT_TOKEN).build()

    # on different commands - answer in Telegram
    application.add_handler(CommandHandler("start", start))
    application.add_handler(CommandHandler("help", help_command))

    # on non command i.e message - echo the message on Telegram
    application.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, echo))

    # Run the bot until the user presses Ctrl-C
    application.run_polling(allowed_updates=Update.ALL_TYPES)


if __name__ == "__main__":
    set_env()
    main()
