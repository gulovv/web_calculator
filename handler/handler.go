package handler

import (
	"sync"
	"regexp"
	"net/http"
	"fmt"
	"encoding/json"
	"strconv"

)
type Task struct {
    ID         int     `json:"id"`
    Expression string  `json:"expression"`
    Result     float64 `json:"result,omitempty"`
    Status     string  `json:"status"`
}

var (
    TaskQueue       []Task
    CompletedTasks  = make(map[int]Task) // Хранилище завершённых задач
    TaskMutex       sync.Mutex
    TaskIDCounter   int
)

// Проверка деления на ноль (отлавливает случаи: "/0", "/ 0", "/0.0", "/ 0.000")
var DivisionByZeroRegex = regexp.MustCompile(`/\s*0(?:\.0+)?\b`)

// Проверка на недопустимые числа с ведущими нулями, например "08" или "001"
// Допускается "0" и "0.xxx", но не "0" за которыми сразу идут цифры без точки.
var InvalidNumberRegex = regexp.MustCompile(`\b0\d+\b`)

// Проверка на повторяющиеся операторы, включая пробелы между ними
var InvalidOperatorsRegex = regexp.MustCompile(`[+\-*/]\s*([+\-*/])`)

// Проверка на числа с запятыми
var InvalidCommaInNumberRegex = regexp.MustCompile(`\d+,\d+`)
//_______________________________________________________________________________________________________________________________

// 1) Эндпоинт для добавления новой задачи
func AddTask(w http.ResponseWriter, r *http.Request) {
    var newTask Task

    // Декодирование JSON-запроса
    err := json.NewDecoder(r.Body).Decode(&newTask)
    if err != nil {
        fmt.Println("Ошибка декодирования данных задачи:", err)
        http.Error(w, "Некорректные данные задачи", http.StatusUnprocessableEntity)
        return
    }
    
   
    // Проверка на повторяющиеся операторы
    if InvalidOperatorsRegex.MatchString(newTask.Expression) {
        fmt.Println("Ошибка: выражение содержит повторяющиеся операторы:", newTask.Expression)
        http.Error(w, "Выражение не должно содержать повторяющиеся операторы", http.StatusUnprocessableEntity) // 422
        return
    }
    // Проверка на запятую 
    if InvalidCommaInNumberRegex.MatchString(newTask.Expression) {
        fmt.Println("Ошибка: выражение содержит недопустимую запятую в числе:", newTask.Expression)
        http.Error(w, "Запятая в числе недопустима", http.StatusUnprocessableEntity)
        return
    }
    
    // Проверка деления на ноль
    if DivisionByZeroRegex.MatchString(newTask.Expression) {
        fmt.Println("Ошибка: деление на ноль в выражении:", newTask.Expression)
        http.Error(w, "Деление на ноль невозможно", http.StatusUnprocessableEntity)
        return
    }
    
    // Проверка на недопустимые числа (например, 08)
    if InvalidNumberRegex.MatchString(newTask.Expression) {
        fmt.Println("Ошибка: выражение содержит числа с ведущими нулями:", newTask.Expression)
        http.Error(w, "Числа с ведущими нулями недопустимы", http.StatusUnprocessableEntity)
        return
    }
    
    // Проверка, что выражение не пустое
    if newTask.Expression == "" {
        fmt.Println("Ошибка: выражение отсутствует")
        http.Error(w, "Выражение не должно быть пустым", http.StatusUnprocessableEntity)
        return
    }


    // Добавление задачи в очередь
    TaskMutex.Lock()
    defer TaskMutex.Unlock()

    TaskIDCounter++
    newTask.ID = TaskIDCounter
    newTask.Status = "pending"
    TaskQueue = append(TaskQueue, newTask)

    fmt.Printf("Задача добавлена: ID=%d, Выражение=%s, Статус=%s\n", newTask.ID, newTask.Expression, newTask.Status)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    err = json.NewEncoder(w).Encode(map[string]int{"id": newTask.ID})
    if err != nil {
        fmt.Println("Ошибка при отправке ответа:", err)
        http.Error(w, "Ошибка при обработке запроса", http.StatusInternalServerError) // 500
    }
}
//_______________________________________________________________________________________________________________________________

// 2) Эндпоинт для получения выражения по ID
func GetExpressionByID(w http.ResponseWriter, r *http.Request) {
    defer func() {
        if r := recover(); r != nil {
            http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError) // 500
        }
    }()

    idStr := r.URL.Path[len("/api/v1/expressions/"):] // Парсим ID из URL
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Некорректный идентификатор", http.StatusBadRequest) // 400
        return
    }

    TaskMutex.Lock()
    defer TaskMutex.Unlock()

    // Проверяем, есть ли задача в истории
    task, exists := CompletedTasks[id]
    if !exists {
        http.Error(w, "Выражение не найдено", http.StatusNotFound) // 404
        return
    }

    response := map[string]Task{"expression": task}

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
//_______________________________________________________________________________________________________________________________

// 3) Эндпоинт для получения списка всех выражений
func GetAllExpressions(w http.ResponseWriter, r *http.Request) {
    // Защищаем доступ к данным с помощью мьютекса
    TaskMutex.Lock()
    defer TaskMutex.Unlock()

    // Если нет завершённых задач, возвращаем пустой список
    if len(CompletedTasks) == 0 {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string][]Task{"expressions": {}})
        return
    }

    // Формируем список задач для ответа
    var expressions []Task
    for _, task := range CompletedTasks {
        expressions = append(expressions, task)
    }

    // Отправляем ответ
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string][]Task{"expressions": expressions})
}
//_______________________________________________________________________________________________________________________________

// 4) Эндпоинт для получения задачи
func GetTask(w http.ResponseWriter, r *http.Request) {
    TaskMutex.Lock()
    defer TaskMutex.Unlock()

    if len(TaskQueue) == 0 {
        http.NotFound(w, r)
        return
    }

    task := TaskQueue[0]

    // Если задача уже завершена, не меняем её статус
    if task.Status == "completed" {
        fmt.Println("Задача уже завершена, пропускаем её")
    } else {
        // Меняем статус на "in-progress", если задача не завершена
        task.Status = "in-progress" // Статус на английском
        TaskQueue[0] = task // Обновляем очередь
    }

    fmt.Printf("Задача получена: ID=%d, Выражение=%s, Статус=%s\n", task.ID, task.Expression, task.Status)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]Task{"task": task})
}
//_______________________________________________________________________________________________________________________________

// 5) Эндпоинт для обновления результата задачи
func UpdateTaskResult(w http.ResponseWriter, r *http.Request) {
    defer func() {
        if r := recover(); r != nil {
            http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
        }
    }()

    var updatedTask Task
    err := json.NewDecoder(r.Body).Decode(&updatedTask)
    if err != nil {
        http.Error(w, "Некорректные данные задачи", http.StatusBadRequest)
        return
    }

    TaskMutex.Lock()
    defer TaskMutex.Unlock()

    for i, task := range TaskQueue {
        if task.ID == updatedTask.ID {
            if task.Status == "completed" {
                http.Error(w, "Задача уже завершена", http.StatusBadRequest)
                return
            }

            // Обновляем задачу и переносим в историю
            TaskQueue[i].Result = updatedTask.Result
            TaskQueue[i].Status = "completed"
            CompletedTasks[TaskQueue[i].ID] = TaskQueue[i] // Сохраняем в историю

            fmt.Printf("Задача обновлена и сохранена в истории: ID=%d, Результат=%f\n", TaskQueue[i].ID, TaskQueue[i].Result)

            // Удаляем из очереди
            TaskQueue = append(TaskQueue[:i], TaskQueue[i+1:]...)

            // Отправляем обновлённую задачу в ответ
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusOK)

            json.NewEncoder(w).Encode(CompletedTasks[updatedTask.ID])
            return
        }
    }

    http.NotFound(w, r)
}
//_______________________________________________________________________________________________________________________________
// 6) Эндпоинт для удаления всех задач
func DeleteAllTasks(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Получен запрос на удаление всех задач")

    TaskMutex.Lock()
    defer TaskMutex.Unlock()

    // Очистка очереди задач
    TaskQueue = []Task{}

    // Очистка завершённых задач
    CompletedTasks = make(map[int]Task)
    TaskIDCounter = 0

    // Проверяем, что данные очищены
    fmt.Printf("Очистили TaskQueue: %v, CompletedTasks: %v\n", TaskQueue, CompletedTasks)

    fmt.Println("Все задачи удалены")

    // Отправка подтверждения об удалении
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Все задачи были удалены"))
}
//_______________________________________________________________________________________________________________________________
