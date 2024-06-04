from telegram import Update
from telegram.ext import Updater, CommandHandler, CallbackContext

database = {}
Token = "7089762341:AAETKzgJfCeV5dmnZhJil5Q-MtgYkgNaBdI"
database_link = r'C:\Users\nikit\OneDrive\Рабочий стол\books.xlsx'


def start(update: Update, context: CallbackContext) -> None:
    update.message.reply_text('Привет! Я бот-библиотекарь. Чтобы узнать функционал бота используйте команду /info.')


def info(update: Update, context: CallbackContext) -> None:
    update.message.reply_text('Это актуальный список функционала бота:\n/search [название предмета] - выводится весь '
                              'доступный список литературы и полезных ссылок по данному предмету\n'
                              '/add | [название предмета] | [название книги] | [ссылка] - добавляет новую книгу в базу '
                              'по указанному предмету\n/find [название книги/автора/или иной ключевой информации] '
                              '- выводит ссылку на необходимую книгу (если она имеется)\n\nАктуальный список разделов:'
                              '\nВысшая математика\nФизическая химия\nАВМ\nПолезные ссылки')


def search(update: Update, context: CallbackContext) -> None:
    text = update.message.text.split(' ')[1:]
    subject = ' '.join(text)

    if subject in database:
        for book in database[subject]:
            update.message.reply_text(f'{book[0]}: {book[1]}')
    else:
        update.message.reply_text('Книги по данному предмету не найдено')


def find_book(update: Update, context: CallbackContext) -> None:
    text = update.message.text.split(' ')[1:]
    title = ' '.join(text)

    found = False
    for subject, books in database.items():
        for book in books:
            if title.lower() in book[0].lower():
                update.message.reply_text(f'Найдена книга "{book[0]}": {book[1]}.')
                found = True

    if not found:
        update.message.reply_text('Книга не найдена в базе данных.')


def add_book(update, context):
    chat_id = update.message.chat_id
    message = update.message.text
    book_info = message.split('|')[1:]

    for elem in book_info:
        elem = elem.strip()

    if len(book_info) == 3:
        key = book_info[0]
        if key not in database:
            database[key] = []

        database[key].append([book_info[1], book_info[2]])
        context.bot.send_message(chat_id=chat_id, text='Книга успешно добавлена!')
    else:
        context.bot.send_message(chat_id=chat_id, text='Ошибка! Пожалуйста проверьте формат ввода данных.')


def save_database_to_excel():
    import pandas as pd

    data = {'Subject': [], 'Title': [], 'Link': []}
    for subject, books in database.items():
        for book in books:
            data['Subject'].append(subject)
            data['Title'].append(book[0])
            data['Link'].append(book[1])

    df = pd.DataFrame(data)
    df.to_excel('books.xlsx',
                index=False)


def load_database_from_excel():
    import pandas as pd

    df = pd.read_excel('books.xlsx')
    for index, row in df.iterrows():
        key = row['Subject']
        if key not in database:
            database[key] = []

        database[key].append([row['Title'], row['Link']])


updater = Updater(Token, use_context=True)

updater.dispatcher.add_handler(CommandHandler('start', start))
updater.dispatcher.add_handler(CommandHandler('info', info))
updater.dispatcher.add_handler(CommandHandler('search', search))
updater.dispatcher.add_handler(CommandHandler('add', add_book))
updater.dispatcher.add_handler(CommandHandler('find', find_book))

load_database_from_excel()

updater.start_polling()
updater.idle()
