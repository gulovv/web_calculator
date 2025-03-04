package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "regexp"
    "strconv"
)

type Task struct {
    ID         int     `json:"id"`
    Expression string  `json:"expression"`
    Result     float64 `json:"result,omitempty"`
    Status     string  `json:"status"`
}

var (
    taskQueue       []Task
    completedTasks  = make(map[int]Task) // Хранилище завершённых задач
    taskMutex       sync.Mutex
    taskIDCounter   int
)

var validExpressionRegex = regexp.MustCompile(`^\s*[\d\+\-\*/$begin:math:text$$end:math:text$\s]+\s*$`)



//_______________________________________________________________________________________________________________________________

// 1) Эндпоинт для добавления новой задачи
func addTask(w http.ResponseWriter, r *http.Request) {
    var newTask Task

    // Декодирование JSON-запроса
    err := json.NewDecoder(r.Body).Decode(&newTask)
    if err != nil {
        fmt.Println("Ошибка декодирования данных задачи:", err)
        http.Error(w, "Некорректные данные задачи", http.StatusUnprocessableEntity) // 422
        return
    }
    // Проверка корректности выражения
    if !validExpressionRegex.MatchString(newTask.Expression) {
        fmt.Println("Ошибка: выражение содержит недопустимые символы:", newTask.Expression)
        http.Error(w, "Выражение содержит недопустимые символы", http.StatusUnprocessableEntity) // 422
        return
    }

    // Проверка, что выражение не пустое
    if newTask.Expression == "" {
        fmt.Println("Ошибка: выражение отсутствует")
        http.Error(w, "Выражение не должно быть пустым", http.StatusUnprocessableEntity) // 422
        return
    }

    // Добавление задачи в очередь
    taskMutex.Lock()
    defer taskMutex.Unlock()

    taskIDCounter++
    newTask.ID = taskIDCounter
    newTask.Status = "pending"
    taskQueue = append(taskQueue, newTask)

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
func getExpressionByID(w http.ResponseWriter, r *http.Request) {
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

    taskMutex.Lock()
    defer taskMutex.Unlock()

    // Проверяем, есть ли задача в истории
    task, exists := completedTasks[id]
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
func getAllExpressions(w http.ResponseWriter, r *http.Request) {
    // Защищаем доступ к данным с помощью мьютекса
    taskMutex.Lock()
    defer taskMutex.Unlock()

    // Если нет завершённых задач, возвращаем пустой список
    if len(completedTasks) == 0 {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string][]Task{"expressions": {}})
        return
    }

    // Формируем список задач для ответа
    var expressions []Task
    for _, task := range completedTasks {
        expressions = append(expressions, task)
    }

    // Отправляем ответ
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string][]Task{"expressions": expressions})
}

//_______________________________________________________________________________________________________________________________


// 4) Эндпоинт для получения задачи
func getTask(w http.ResponseWriter, r *http.Request) {
    //fmt.Println("Получен запрос на получение задачи")

    taskMutex.Lock()
    defer taskMutex.Unlock()

    if len(taskQueue) == 0 {
        //fmt.Println("Нет доступных задач")
        http.NotFound(w, r)
        return
    }

    task := taskQueue[0]

    // Если задача уже завершена, не меняем её статус
    if task.Status == "completed" {
        fmt.Println("Задача уже завершена, пропускаем её")
    } else {
        // Меняем статус на "in-progress", если задача не завершена
        task.Status = "in-progress" // Статус на английском
        taskQueue[0] = task // Обновляем очередь
    }

    fmt.Printf("Задача получена: ID=%d, Выражение=%s, Статус=%s\n", task.ID, task.Expression, task.Status)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]Task{"task": task})
}


//_______________________________________________________________________________________________________________________________

// 5) Эндпоинт для обновления результата задачи
func updateTaskResult(w http.ResponseWriter, r *http.Request) {
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

    taskMutex.Lock()
    defer taskMutex.Unlock()

    for i, task := range taskQueue {
        if task.ID == updatedTask.ID {
            if task.Status == "completed" {
                http.Error(w, "Задача уже завершена", http.StatusBadRequest)
                return
            }

            // Обновляем задачу и переносим в историю
            taskQueue[i].Result = updatedTask.Result
            taskQueue[i].Status = "completed"
            completedTasks[taskQueue[i].ID] = taskQueue[i] // Сохраняем в историю

            fmt.Printf("Задача обновлена и сохранена в истории: ID=%d, Результат=%f\n", taskQueue[i].ID, taskQueue[i].Result)

            // Удаляем из очереди
            taskQueue = append(taskQueue[:i], taskQueue[i+1:]...)

            // Отправляем обновлённую задачу в ответ
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusOK)

            json.NewEncoder(w).Encode(completedTasks[updatedTask.ID])
            return
        }
    }

    http.NotFound(w, r)
}

//_______________________________________________________________________________________________________________________________


// 6) Эндпоинт для удаления всех задач
func deleteAllTasks(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Получен запрос на удаление всех задач")

    taskMutex.Lock()
    defer taskMutex.Unlock()

    // Очистка очереди задач
    taskQueue = []Task{}

    fmt.Println("Все задачи удалены")

    // Отправка подтверждения об удалении
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Все задачи были удалены"))
}

//_______________________________________________________________________________________________________________________________


func main() {
    fmt.Println("Запуск сервера Оркестратора...")

    // Добавление всех эндпоинтов
    http.HandleFunc("/api/v1/calculate", addTask)         // Для добавления новой задачи
    http.HandleFunc("/api/v1/task", getTask)              // Для получения задачи агентом
    http.HandleFunc("/api/v1/task/result", updateTaskResult) // Для обновления результата задачи
    http.HandleFunc("/api/v1/tasks/delete", deleteAllTasks) // Удаление всех задач
    http.HandleFunc("/api/v1/expressions/", getExpressionByID)
    http.HandleFunc("/api/v1/expressions", getAllExpressions)


    // Запуск сервера
    fmt.Println("Сервер Оркестратора запущен на http://localhost:8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        fmt.Println("Ошибка сервера:", err)
    }
}