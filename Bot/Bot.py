from telegram import Update, ReplyKeyboardMarkup, InlineKeyboardButton, InlineKeyboardMarkup
from telegram.ext import CommandHandler, ApplicationBuilder, CallbackContext, MessageHandler, filters, CallbackQueryHandler
from dotenv import load_dotenv
import pandas as pd
import os
import sys
from thefuzz import process
import hashlib
import asyncio


def get_root_path(file_name):
    try:
        project_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
        path_to_file = os.path.join(project_root, file_name)
        return str(path_to_file)
    except (TypeError, FileNotFoundError) as e:
        print(f"Произошла ошибка: {e}")
        sys.exit(1)


load_dotenv(dotenv_path=get_root_path('data.env'))
token = os.getenv('Token')
EXCEL_FILE = get_root_path('books.xlsx')
df = pd.read_excel(EXCEL_FILE)


async def start(update: Update, context: CallbackContext) -> None:
    # Создаём меню кнопок для клавиатуры
    keyboard = [
        ["Информация", "Разделы", "Материалы всего раздела"],  # Две кнопки в одной строке
    ]
    reply_markup = ReplyKeyboardMarkup(keyboard, resize_keyboard=True)

    await update.message.reply_text('Привет! Я бот-библиотекарь. Чтобы узнать функционал бота '
                                    'используйте команду /info или воспользоваться функциональными кнопками.',
                                    reply_markup=reply_markup)


async def info(update: Update, context: CallbackContext) -> None:
    await update.message.reply_text('Это актуальный список функционала бота:'
                                    '\n/search [название книги или ключевые слова]')


def fuzzy_search(query, threshold=70):
    # Поиск по столбцу Title (Название книги)
    title_choices = df["Title"].dropna().unique()
    best_title = process.extractOne(query, title_choices)

    # Поиск по столбцу Subject (Раздел)
    subject_choices = df["Subject"].dropna().unique()
    best_subject = process.extractOne(query, subject_choices)

    # Определяем лучший результат среди Title и Subject
    best_match = None
    match_info = None  # Для хранения информации о соответствии

    if best_title and best_title[1] >= threshold:
        matches = df[df["Title"] == best_title[0]]
        if not matches.empty:
            best_match = matches.iloc[0]
            match_info = ("Title", best_title[1])  # Найдено по Title
    elif best_subject and best_subject[1] >= threshold:
        matches = df[df["Subject"] == best_subject[0]]
        if not matches.empty:
            best_match = matches.iloc[0]
            match_info = ("Subject", best_subject[1])  # Найдено по Subject

    # Если точных совпадений нет или результаты пусты
    if best_match is None or isinstance(best_match,
                                        pd.Series) and best_match.empty:  # Проверяем, что результат не пустой
        top_title = best_title if best_title else (None, 0)
        top_subject = best_subject if best_subject else (None, 0)

        # Выбираем наиболее близкий вариант, даже если он ниже порога
        if top_title[1] > top_subject[1]:
            match_info = ("Title", top_title[1])
            matches = df[df["Title"] == top_title[0]]
            if not matches.empty:
                best_match = matches.iloc[0]
        elif top_subject[1] > top_title[1]:
            match_info = ("Subject", top_subject[1])
            matches = df[df["Subject"] == top_subject[0]]
            if not matches.empty:
                best_match = matches.iloc[0]

    return best_match, match_info


# Обработчик команды /search
async def search(update: Update, context: CallbackContext):
    if not context.args:
        await update.message.reply_text("Используйте: /search [название книги или ключевые слова]")
        return

    query = " ".join(context.args)

    task = asyncio.create_task(async_fuzzy_search(query))  # Запускаем поиск в фоне
    result, match_info = await task  # Дожидаемся результата

    if result is not None:
        if match_info and match_info[1] >= 70:
            response = f"📚 *{result['Title']}*\n🔗 [Ссылка на книгу]({result['Link']})"
        else:
            response = (
                f"❗ Книга не найдена с точным соответствием.\n"
                f"Мы нашли наиболее подходящий результат:\n"
                f"📚 *{result['Title']}*\n🔗 [Ссылка на книгу]({result['Link']})\n"
                f"⚠️ Вероятность совпадения: {match_info[1]}%"
            )
    else:
        response = "❌ Книга или раздел не найдены. Попробуйте изменить запрос."

    await update.message.reply_text(response, parse_mode="Markdown")


async def async_fuzzy_search(query):
    loop = asyncio.get_running_loop()
    return await loop.run_in_executor(None, fuzzy_search, query)  # Запускаем в фоне


async def get_subjects(update: Update, context: CallbackContext):
    subjects = df["Subject"].unique()
    subjects = sorted(subjects)# Получаем уникальные предметы
    subject_list = "\n".join(subjects)  # Преобразуем список в строку

    max_length = 4000  # Безопасный лимит сообщения
    if len(subject_list) > max_length:
        # Разбиваем на части и отправляем по очереди
        chunk_size = 50  # Количество предметов в одном сообщении (можно настроить)
        for i in range(0, len(subjects), chunk_size):
            await update.message.reply_text("\n".join(subjects[i:i + chunk_size]))
    else:
        # Если сообщение короткое, отправляем одним текстом
        await update.message.reply_text(subject_list)



def title_search(query):
    subjects = df["Subject"].unique()

    if query in subjects:
        return df[df["Subject"] == query]["Title"].tolist()
    else:
        return None


async def sections(update: Update, context: CallbackContext):
    subjects = df["Subject"].unique()

    buttons = []
    for subject in subjects:
        callback_data = f"subject_{hashlib.md5(subject.encode()).hexdigest()[:10]}"
        buttons.append([InlineKeyboardButton(subject, callback_data=callback_data)])

    max_buttons = 50
    if "buttons_messages" not in context.user_data:
        context.user_data["buttons_messages"] = []

    tasks = []  # Список для хранения задач

    for i in range(0, len(buttons), max_buttons):
        reply_markup = InlineKeyboardMarkup(buttons[i:i + max_buttons])
        task = asyncio.create_task(update.message.reply_text(
            "Выберите раздел, чтобы увидеть все книги:", reply_markup=reply_markup
        ))
        tasks.append(task)

    responses = await asyncio.gather(*tasks)  # Запускаем все задачи одновременно

    for message in responses:
        context.user_data["buttons_messages"].append(message.message_id)




# Глобальный словарь для хранения соответствий хэшей и названий
subject_mapping = {hashlib.md5(subject.encode()).hexdigest()[:10]: subject for subject in df["Subject"].unique()}

async def section_selected(update: Update, context: CallbackContext):
    query = update.callback_query
    await query.answer()

    if "buttons_messages" in context.user_data:
        tasks = [asyncio.create_task(query.message.chat.delete_message(msg_id))
                 for msg_id in context.user_data["buttons_messages"]]
        await asyncio.gather(*tasks)  # Удаляем все кнопки одновременно

        context.user_data["buttons_messages"].clear()

    hash_key = query.data.split("_", 1)[1]
    subject = subject_mapping.get(hash_key, "Неизвестный раздел")

    books = title_search(subject)

    if books:
        books_text = "\n---------------------\n".join(books)
        response = f"Все книги в разделе \"{subject}\":\n{books_text}\n\n"
    else:
        response = "К сожалению, книги для этого раздела не найдены."

    max_length = 4000
    if len(response) > max_length:
        parts = [response[i:i + max_length] for i in range(0, len(response), max_length)]
        tasks = [asyncio.create_task(query.message.reply_text(part)) for part in parts]
        await asyncio.gather(*tasks)  # Отправляем все части параллельно
    else:
        await query.message.reply_text(response)  # Отправляем новым сообщением


async def button_handler(update: Update, context):
    text = update.message.text

    tasks = []

    if text == "Информация":
        tasks.append(asyncio.create_task(info(update, context)))
    elif text == "Разделы":
        tasks.append(asyncio.create_task(get_subjects(update, context)))
    elif text == "Материалы всего раздела":
        tasks.append(asyncio.create_task(sections(update, context)))
    else:
        tasks.append(asyncio.create_task(update.message.reply_text(
            "Извините, я не понимаю эту команду. Попробуйте ещё раз!"
        )))

    await asyncio.gather(*tasks)  # Запускаем все задачи параллельно


app = ApplicationBuilder().token(token).build()

# Добавляем обработчики команд
app.add_handler(CommandHandler('start', start))
app.add_handler(CommandHandler('info', info))
app.add_handler(CommandHandler("search", search))
app.add_handler(CommandHandler("subjects", get_subjects))
app.add_handler(CallbackQueryHandler(section_selected, pattern="subject_"))

app.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, button_handler))

app.run_polling()
