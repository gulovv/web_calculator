package test

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "strings"
    "github.com/gulovv/web_calculator/handler"
)

func TestAddTask(t *testing.T) {
    tests := []struct {
        name           string
        body           string
        expectedStatus int
    }{
        {
            name:           "Successful task addition",
            body:           `{"expression": "5 + 3"}`,
            expectedStatus: http.StatusCreated,
        },
        {
            name:           "Invalid expression",
            body:           `{"expression": "5 & 3"}`,
            expectedStatus: http.StatusUnprocessableEntity,
        },
        {
            name:           "Empty expression",
            body:           `{"expression": ""}`,
            expectedStatus: http.StatusUnprocessableEntity,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Logf("Запуск теста: %s", tt.name)
            req := httptest.NewRequest("POST", "/api/v1/calculate", strings.NewReader(tt.body))
            w := httptest.NewRecorder()

            t.Logf("Отправляем запрос с телом: %s", tt.body)
            handler.AddTask(w, req)

            t.Logf("Ответ получен с кодом: %d", w.Code)
            if status := w.Code; status != tt.expectedStatus {
                t.Errorf("Ожидался статус %d, но получили %d", tt.expectedStatus, status)
            }
        })
    }
}

func TestGetExpressionByID(t *testing.T) {
    // Prepare the handler with a task in the CompletedTasks map
    handler.CompletedTasks = map[int]handler.Task{
       1: {ID: 1, Expression: "3 + 2", Status: "completed", Result: 5},
    }

    tests := []struct {
        name           string
        id             string
        expectedStatus int
    }{
        {
            name:           "Task found",
            id:             "1",
            expectedStatus: http.StatusOK,
        },
        {
            name:           "Task not found",
            id:             "999",
            expectedStatus: http.StatusNotFound,
        },
        {
            name:           "Invalid ID format",
            id:             "abc",
            expectedStatus: http.StatusBadRequest,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Logf("Запуск теста: %s", tt.name)
            req := httptest.NewRequest("GET", "/api/v1/expressions/"+tt.id, nil)
            w := httptest.NewRecorder()

            t.Logf("Отправляем запрос для ID: %s", tt.id)
            handler.GetExpressionByID(w, req)

            t.Logf("Ответ получен с кодом: %d", w.Code)
            if status := w.Code; status != tt.expectedStatus {
                t.Errorf("Ожидался статус %d, но получили %d", tt.expectedStatus, status)
            }
        })
    }
}

func TestGetAllExpressions(t *testing.T) {
    handler.TaskQueue = []handler.Task{
        {ID: 1, Expression: "3 + 2", Status: "completed", Result: 5},
    }

    req := httptest.NewRequest("GET", "/api/v1/expressions", nil)
    w := httptest.NewRecorder()

    t.Log("Отправляем запрос на получение всех выражений")
    handler.GetAllExpressions(w, req)

    t.Logf("Ответ получен с кодом: %d", w.Code)
    if status := w.Code; status != http.StatusOK {
        t.Errorf("Ожидался статус %d, но получили %d", http.StatusOK, status)
    }

    var response map[string][]handler.Task
    if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
        t.Fatalf("Ожидался корректный JSON, но возникла ошибка: %v", err)
    }
    t.Logf("Полученный ответ: %+v", response)
}

func TestGetTask(t *testing.T) {
    handler.TaskQueue = []handler.Task{
        {ID: 1, Expression: "2 + 2", Status: "pending"},
    }

    req := httptest.NewRequest("GET", "/api/v1/task", nil)
    w := httptest.NewRecorder()

    t.Log("Отправляем запрос для получения задачи")
    handler.GetTask(w, req)

    t.Logf("Ответ получен с кодом: %d", w.Code)
    if status := w.Code; status != http.StatusOK {
        t.Errorf("Ожидался статус %d, но получили %d", http.StatusOK, status)
    }
}

func TestUpdateTaskResult(t *testing.T) {
    handler.CompletedTasks = map[int]handler.Task{
       1: {ID: 1, Expression: "2 + 2", Status: "pending"},
    }

    tests := []struct {
        name           string
        body           string
        expectedStatus int
    }{
        {
            name:           "Successful result update",
            body:           `{"id": 1, "result": 4}`,
            expectedStatus: http.StatusOK,
        },
        {
            name:           "Task already completed",
            body:           `{"id": 1, "result": 10}`,
            expectedStatus: http.StatusBadRequest,
        },
        {
            name:           "Invalid task ID",
            body:           `{"id": 999, "result": 5}`,
            expectedStatus: http.StatusNotFound,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Logf("Запуск теста: %s", tt.name)
            req := httptest.NewRequest("POST", "/api/v1/task/result", strings.NewReader(tt.body))
            w := httptest.NewRecorder()

            t.Logf("Отправляем запрос с телом: %s", tt.body)
            handler.UpdateTaskResult(w, req)

            t.Logf("Ответ получен с кодом: %d", w.Code)
            if status := w.Code; status != tt.expectedStatus {
                t.Errorf("Ожидался статус %d, но получили %d", tt.expectedStatus, status)
            }
        })
    }
}

func TestDeleteAllTasks(t *testing.T) {
    handler.TaskQueue = []handler.Task{
        {ID: 1, Expression: "3 + 3", Status: "pending"},
    }

    req := httptest.NewRequest("DELETE", "/api/v1/tasks/delete", nil)
    w := httptest.NewRecorder()

    t.Log("Отправляем запрос на удаление всех задач")
    handler.DeleteAllTasks(w, req)

    t.Logf("Ответ получен с кодом: %d", w.Code)
    if status := w.Code; status != http.StatusOK {
        t.Errorf("Ожидался статус %d, но получили %d", http.StatusOK, status)
    }

    t.Logf("Проверяем количество задач в очереди: %d", len(handler.TaskQueue))
    if len(handler.TaskQueue) != 0 {
        t.Errorf("Ожидалась пустая очередь задач, но осталось %d задач", len(handler.TaskQueue))
    }
}