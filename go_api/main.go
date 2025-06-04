package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// Структуры данных
type Answer struct {
	ID         uuid.UUID `json:"id"`
	QuestionID uuid.UUID `json:"question_id"`
	Text       string    `json:"text"`
	Correct    bool      `json:"correct"`
}

type Question struct {
	ID     uuid.UUID `json:"id"`
	TestID uuid.UUID `json:"test_id"`
	Text   string    `json:"text"`
}

type Test struct {
	ID    uuid.UUID `json:"id"`
	Title string    `json:"title"`
}

type User struct {
	ID         uuid.UUID `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	GroupName  string    `json:"group_name"`
	IsAdmin    bool      `json:"is_admin"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Структуры для отображения данных
type TestWithQuestions struct {
	Test      Test
	Questions []QuestionWithAnswers
	Message   string // Добавлено для передачи сообщений
}

type QuestionWithAnswers struct {
	Question Question
	Answers  []Answer
}

// Data structure for the result page
type TestResultData struct {
	FirstName      string
	LastName       string
	GroupName      string
	CorrectAnswers int
	TotalQuestions int
}

// TestListData - структура для передачи данных в шаблон testlist.html
type TestListData struct {
	Tests       []Test // Здесь будут храниться тесты
	SearchQuery string // Здесь будет храниться строка поиска
}

var db *sql.DB

func init() {
	var err error
	connStr := "user=postgres password=root dbname=testplatform sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to PostgreSQL")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html", "templates/write.html")

	if err != nil {
		fmt.Fprintf(w, "Error loading template: %s", err)
		return
	}

	t.ExecuteTemplate(w, "index", nil)

}
func writeHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html", "templates/write.html")

	if err != nil {
		fmt.Fprintf(w, "Error loading template: %s", err)
		return
	}
	telegramID := r.URL.Query().Get("telegram_id")
	message := r.URL.Query().Get("message") // Получаем сообщение из параметров URL

	data := struct {
		Title      string
		TelegramID string
		Message    string // Добавляем поле Message
	}{
		Title:      "",
		TelegramID: telegramID,
		Message:    message, // Передаем сообщение в шаблон

	}

	t.ExecuteTemplate(w, "write", data)
}
func saveTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Получаем Telegram ID из формы
	telegramIDStr := r.Form.Get("admin_telegram_id")
	if telegramIDStr == "" {
		http.Redirect(w, r, "/write?message=ыВведите Telegram ID для проверки прав администратора.", http.StatusSeeOther)
		return
	}

	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/write?message=Неверный формат Telegram ID. Пожалуйста, введите числовое значение.", http.StatusSeeOther)
		return
	}

	// Проверяем, является ли пользователь администратором
	var isAdmin bool
	err = db.QueryRow("SELECT is_admin FROM users WHERE telegram_id = $1", telegramID).Scan(&isAdmin)
	if err == sql.ErrNoRows {
		http.Redirect(w, r, "/write?message=Пользователь с таким Telegram ID не найден.", http.StatusSeeOther)
		return
	} else if err != nil {
		log.Printf("Error checking user admin status: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !isAdmin {
		http.Redirect(w, r, "/write?message=У вас нет прав администратора для создания тестов.", http.StatusSeeOther)
		return
	}

	// Если пользователь является администратором, продолжаем сохранение теста
	title := r.Form.Get("test-title")

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Error starting transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	testID := uuid.New()

	_, err = tx.Exec("INSERT INTO tests (id, title) VALUES ($1, $2)", testID, title)
	if err != nil {
		http.Error(w, "Error inserting test: "+err.Error(), http.StatusInternalServerError)
		return
	}

	questionIndex := 1
	for {
		questionText := r.Form.Get(fmt.Sprintf("question-text-%d", questionIndex))
		if questionText == "" {
			break
		}

		questionID := uuid.New()
		_, err = tx.Exec("INSERT INTO questions (id, test_id, text) VALUES ($1, $2, $3)", questionID, testID, questionText)
		if err != nil {
			http.Error(w, "Error inserting question: "+err.Error(), http.StatusInternalServerError)
			return
		}

		answerIndex := 1
		for {
			answerText := r.Form.Get(fmt.Sprintf("answer-%d-%d", answerIndex, questionIndex))
			if answerText == "" {
				break
			}

			correctStr := r.Form.Get(fmt.Sprintf("correct-answer-%d", questionIndex))
			isCorrect := false
			if correctStr == fmt.Sprintf("%d", answerIndex) {
				isCorrect = true
			}

			answerID := uuid.New()
			_, err = tx.Exec("INSERT INTO answers (id, question_id, text, correct) VALUES ($1, $2, $3, $4)", answerID, questionID, answerText, isCorrect)
			if err != nil {
				http.Error(w, "Error inserting answer: "+err.Error(), http.StatusInternalServerError)
				return
			}
			answerIndex++
		}
		questionIndex++
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Error committing transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func getTests() ([]Test, error) {
	rows, err := db.Query("SELECT id, title FROM tests")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tests []Test
	for rows.Next() {
		var test Test
		if err := rows.Scan(&test.ID, &test.Title); err != nil {
			return nil, err
		}
		tests = append(tests, test)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tests, nil
}

func testListHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем поисковый запрос из URL-параметра 'search'
	searchQuery := r.URL.Query().Get("search")

	tests := []Test{}
	var rows *sql.Rows
	var err error

	if searchQuery != "" {
		// Если есть поисковый запрос, ищем тесты, содержащие его в названии
		rows, err = db.Query("SELECT id, title FROM tests WHERE title ILIKE $1 ORDER BY title ASC", "%"+searchQuery+"%")
	} else {
		// Если поискового запроса нет, получаем все тесты
		rows, err = db.Query("SELECT id, title FROM tests ORDER BY title ASC")
	}

	if err != nil {
		log.Printf("Error querying tests: %v", err)
		http.Error(w, "Unable to load tests", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var test Test
		if err := rows.Scan(&test.ID, &test.Title); err != nil {
			log.Printf("Error scanning test: %v", err)
			http.Error(w, "Unable to process tests", http.StatusInternalServerError)
			return
		}
		tests = append(tests, test)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating test rows: %v", err)
		http.Error(w, "Error loading tests data", http.StatusInternalServerError)
		return
	}

	// Подготавливаем данные для шаблона
	data := TestListData{
		Tests:       tests,
		SearchQuery: searchQuery, // Передаем поисковый запрос обратно в шаблон
	}

	t, err := template.ParseFiles("templates/testlist.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.ExecuteTemplate(w, "testlist", data)
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	testIDStr := r.URL.Path[len("/test/"):] // Извлекаем ID из URL

	testID, err := uuid.Parse(testIDStr) // Преобразуем строку в UUID
	if err != nil {
		http.Error(w, "Invalid Test ID", http.StatusBadRequest)
		return
	}

	t, err := template.ParseFiles("templates/test.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	testWithQuestions, err := getTestWithQuestions(testID) // Получаем данные теста
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Инициализируем Message пустым, если нет ошибок
	testWithQuestions.Message = ""
	t.ExecuteTemplate(w, "test", testWithQuestions) // Отображаем страницу теста
}

func getTestWithQuestions(testID uuid.UUID) (TestWithQuestions, error) {
	test := Test{ID: testID}
	err := db.QueryRow("SELECT title FROM tests WHERE id = $1", testID).Scan(&test.Title)
	if err != nil {
		return TestWithQuestions{}, err
	}
	log.Println("SQL (questions): SELECT id, text, test_id FROM questions WHERE test_id = ", testID)

	rows, err := db.Query("SELECT id, text, test_id FROM questions WHERE test_id = $1", testID)
	if err != nil {
		return TestWithQuestions{}, err
	}
	defer rows.Close()

	var questions []QuestionWithAnswers
	for rows.Next() {
		var question Question
		if err := rows.Scan(&question.ID, &question.Text, &question.TestID); err != nil {
			return TestWithQuestions{}, err
		}
		log.Println("SQL (answers): SELECT id, text, question_id, correct FROM answers WHERE question_id = ", question.ID)

		answerRows, err := db.Query("SELECT id, text, question_id, correct FROM answers WHERE question_id = $1", question.ID)
		if err != nil {
			return TestWithQuestions{}, err
		}
		defer answerRows.Close()

		var answers []Answer
		for answerRows.Next() {
			var answer Answer
			if err := answerRows.Scan(&answer.ID, &answer.Text, &answer.QuestionID, &answer.Correct); err != nil {
				return TestWithQuestions{}, err
			}
			answers = append(answers, answer)
		}
		if err := answerRows.Err(); err != nil {
			return TestWithQuestions{}, err
		}

		questions = append(questions, QuestionWithAnswers{Question: question, Answers: answers})
	}
	if err := rows.Err(); err != nil {
		return TestWithQuestions{}, err
	}
	log.Printf("TestWithQuestions: %+v", TestWithQuestions{Test: test, Questions: questions})

	return TestWithQuestions{Test: test, Questions: questions}, nil
}

func submitTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	testIDStr := r.URL.Path[len("/submit-test/"):]
	testID, err := uuid.Parse(testIDStr)
	if err != nil {
		http.Error(w, "Invalid Test ID", http.StatusBadRequest)
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	telegramIDStr := r.Form.Get("telegram_id")
	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Telegram ID", http.StatusBadRequest)
		return
	}
	firstName := r.Form.Get("first_name")
	lastName := r.Form.Get("last_name")
	groupName := r.Form.Get("group_name")

	// --- Начало логики проверки уже пройденного теста ---
	var existingTestResultCount int
	// Проверяем, существует ли уже запись о прохождении этого теста данным пользователем
	// Используем OR для проверки по полю: Telegram ID
	query := `
		SELECT COUNT(utr.id)
		FROM user_test_results utr
		JOIN users u ON utr.user_id = u.id
		WHERE utr.test_id = $1
		  AND u.telegram_id = $2
	`
	err = db.QueryRow(query, testID, telegramID).Scan(&existingTestResultCount)
	if err != nil && err != sql.ErrNoRows { // Обработка ошибок, кроме "нет строк"
		log.Printf("Error checking for existing test result: %v", err)
		http.Error(w, "Error checking test history", http.StatusInternalServerError)
		return
	}

	if existingTestResultCount > 0 {
		// Если тест уже пройден, отображаем сообщение об ошибке на странице теста
		testWithQuestions, err := getTestWithQuestions(testID)
		if err != nil {
			http.Error(w, "Error loading test details for error message", http.StatusInternalServerError)
			return
		}
		testWithQuestions.Message = "Этот тест уже был пройден."
		t, parseErr := template.ParseFiles("templates/test.html", "templates/header.html", "templates/footer.html")
		if parseErr != nil {
			http.Error(w, parseErr.Error(), http.StatusInternalServerError)
			return
		}
		t.ExecuteTemplate(w, "test", testWithQuestions)
		return // Важно остановить выполнение после отображения сообщения
	}
	// --- Конец логики проверки уже пройденного теста ---

	// Сохранение информации о пользователе в базу данных (остается без изменений)
	var userID uuid.UUID
	err = db.QueryRow("SELECT id FROM users WHERE telegram_id = $1", telegramID).Scan(&userID)

	if err == sql.ErrNoRows {
		newUser := User{
			ID:         uuid.New(),
			TelegramID: telegramID,
			FirstName:  firstName,
			LastName:   lastName,
			GroupName:  groupName,
			IsAdmin:    false,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		_, err = db.Exec("INSERT INTO users (id, telegram_id, first_name, last_name, group_name, is_admin, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
			newUser.ID, newUser.TelegramID, newUser.FirstName, newUser.LastName, newUser.GroupName, newUser.IsAdmin, newUser.CreatedAt, newUser.UpdatedAt)
		if err != nil {
			log.Printf("Error inserting new user: %v", err)
			http.Error(w, "Error saving user information", http.StatusInternalServerError)
			return
		}
		userID = newUser.ID
		log.Printf("New user registered: %s %s (%d), Group: %s", firstName, lastName, telegramID, groupName)
	} else if err != nil {
		log.Printf("Error checking for existing user: %v", err)
		http.Error(w, "Error processing user information", http.StatusInternalServerError)
		return
	} else {
		_, err = db.Exec("UPDATE users SET first_name = $1, last_name = $2, group_name = $3, updated_at = $4 WHERE id = $5",
			firstName, lastName, groupName, time.Now(), userID)
		if err != nil {
			log.Printf("Error updating existing user: %v", err)
			http.Error(w, "Error saving user information", http.StatusInternalServerError)
			return
		}
		log.Printf("Existing user updated: %s %s (%d), Group: %s", firstName, lastName, telegramID, groupName)
	}

	// Логика подсчета баллов
	correctAnswersCount := 0
	totalQuestions := 0

	testData, err := getTestWithQuestions(testID)
	if err != nil {
		log.Printf("Error getting test questions for scoring: %v", err)
		http.Error(w, "Error processing test", http.StatusInternalServerError)
		return
	}

	for _, qwa := range testData.Questions {
		totalQuestions++
		submittedAnswerIDStr := r.Form.Get(fmt.Sprintf("answer-%s", qwa.Question.ID.String()))
		if submittedAnswerIDStr == "" {
			continue
		}

		for _, answer := range qwa.Answers {
			if answer.Correct && answer.ID.String() == submittedAnswerIDStr {
				correctAnswersCount++
				break
			}
		}
	}

	log.Printf("Test %s submitted by user %s %s (Telegram ID: %d). Score: %d/%d correct answers.", testID, firstName, lastName, telegramID, correctAnswersCount, totalQuestions)

	_, err = db.Exec("INSERT INTO user_test_results (user_id, test_id, correct_answers, total_questions, submitted_at) VALUES ($1, $2, $3, $4, $5)",
		userID, testID, correctAnswersCount, totalQuestions, time.Now())
	if err != nil {
		log.Printf("Error saving test result: %v", err)
		// Здесь можно решить, что делать, если сохранение результата не удалось.
		// Возможно, стоит все равно перенаправить, но уведомить пользователя о проблеме.
		// Или отобразить сообщение об ошибке на текущей странице, как в случае с уже пройденным тестом.
	}
	// Render the result page
	resultData := TestResultData{
		FirstName:      firstName,
		LastName:       lastName,
		GroupName:      groupName,
		CorrectAnswers: correctAnswersCount,
		TotalQuestions: totalQuestions,
	}

	// Parse the templates for the result page
	t, parseErr := template.ParseFiles("templates/result.html", "templates/header.html", "templates/footer.html")
	if parseErr != nil {
		http.Error(w, parseErr.Error(), http.StatusInternalServerError)
		return
	}
	t.ExecuteTemplate(w, "result", resultData)
}

func main() {
	defer db.Close()
	fmt.Println("LISTEN PORT :8080")
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/write", writeHandler)
	http.HandleFunc("/save", saveTestHandler)
	http.HandleFunc("/test/", testHandler)
	http.HandleFunc("/testlist", testListHandler)
	http.HandleFunc("/submit-test/", submitTestHandler)

	http.ListenAndServe(":8080", nil)
}
