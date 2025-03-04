package main

import (
    "fmt"
    "net/http"
    "github.com/gulovv/web_calculator/handler"
)

func main() {
    fmt.Println("Запуск сервера Оркестратора...")

    // Добавление всех эндпоинтов
    http.HandleFunc("/api/v1/calculate", handler.AddTask)         // Для добавления новой задачи
    http.HandleFunc("/api/v1/task", handler.GetTask)              // Для получения задачи агентом
    http.HandleFunc("/api/v1/task/result", handler.UpdateTaskResult) // Для обновления результата задачи
    http.HandleFunc("/api/v1/tasks/delete", handler.DeleteAllTasks) // Удаление всех задач
    http.HandleFunc("/api/v1/expressions/", handler.GetExpressionByID)
    http.HandleFunc("/api/v1/expressions", handler.GetAllExpressions)

    // Запуск сервера
    fmt.Println("Сервер Оркестратора запущен на http://localhost:8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        fmt.Println("Ошибка сервера:", err)
    }
}