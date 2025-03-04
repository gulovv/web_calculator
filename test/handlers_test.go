package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gulovv/web_calculator/cmd" // Путь к основным обработчикам
)

// Тестирование добавления задачи
func TestAddTask(t *testing.T) {
	task := map[string]string{
		"expression": "2 * 9 + 8",
	}

	taskJSON, _ := json.Marshal(task)

	req, err := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(taskJSON))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(AddTask) // Используем обработчик из orchestrator

	handler.ServeHTTP(rr, req)

	// Проверка кода ответа
	if rr.Code != http.StatusCreated {
	t.Errorf("Expected status %v, got %v", http.StatusCreated, rr.Code)
	}

	// Проверка JSON-ответа
	var response map[string]int
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
	t.Errorf("Expected a valid JSON response, got error: %v", err)
	}

	if _, exists := response["id"]; !exists {
	t.Errorf("Expected field 'id' in response")
	}
}

// Тестирование получения задачи
func TestGetTask(t *testing.T) {
 task := orchestrator.Task{
  ID:         1,
  Expression: "2 * 9 + 8",
  Status:     "pending",
 }

 orchestrator.TaskQueue = append(orchestrator.TaskQueue, task)

 req, err := http.NewRequest("GET", "/api/v1/task", nil)
 if err != nil {
  t.Fatal(err)
 }

 rr := httptest.NewRecorder()
 handler := http.HandlerFunc(orchestrator.GetTask)

 handler.ServeHTTP(rr, req)

 // Проверка кода ответа
 if rr.Code != http.StatusOK {
  t.Errorf("Expected status %v, got %v", http.StatusOK, rr.Code)
 }

 var response map[string]orchestrator.Task
 err = json.NewDecoder(rr.Body).Decode(&response)
 if err != nil {
  t.Errorf("Expected a valid JSON response, got error: %v", err)
 }

 if response["task"].ID != task.ID {
  t.Errorf("Expected task ID %v, got %v", task.ID, response["task"].ID)
 }
}

// Тестирование обновления результата задачи
func TestUpdateTaskResult(t *testing.T) {
 taskResult := map[string]string{
  "result": "26", // Результат вычисления
 }

 taskResultJSON, _ := json.Marshal(taskResult)

 req, err := http.NewRequest("POST", "/api/v1/task/result", bytes.NewBuffer(taskResultJSON))
 if err != nil {
  t.Fatal(err)
 }

 rr := httptest.NewRecorder()
 handler := http.HandlerFunc(orchestrator.UpdateTaskResult)

 handler.ServeHTTP(rr, req)

 // Проверка кода ответа
 if rr.Code != http.StatusOK {
  t.Errorf("Expected status %v, got %v", http.StatusOK, rr.Code)
 }

 var response map[string]string
 err = json.NewDecoder(rr.Body).Decode(&response)
 if err != nil {
  t.Errorf("Expected a valid JSON response, got error: %v", err)
 }

 if response["status"] != "updated" {
  t.Errorf("Expected task status 'updated', got %v", response["status"])
 }
}

// Тестирование удаления всех задач
func TestDeleteAllTasks(t *testing.T) {
 req, err := http.NewRequest("POST", "/api/v1/tasks/delete", nil)
 if err != nil {
  t.Fatal(err)
 }

 rr := httptest.NewRecorder()
 handler := http.HandlerFunc(orchestrator.DeleteAllTasks)

 handler.ServeHTTP(rr, req)

 // Проверка кода ответа
 if rr.Code != http.StatusOK {
  t.Errorf("Expected status %v, got %v", http.StatusOK, rr.Code)
 }

 // Проверка, что очередь задач пуста
 if len(orchestrator.TaskQueue) != 0 {
  t.Errorf("Expected no tasks, but found %v", len(orchestrator.TaskQueue))
 }
}

// Тестирование получения выражения по ID
func TestGetExpressionByID(t *testing.T) {
 req, err := http.NewRequest("GET", "/api/v1/expressions/1", nil)
 if err != nil {
  t.Fatal(err)
 }

 rr := httptest.NewRecorder()
 handler := http.HandlerFunc(orchestrator.GetExpressionByID)

 handler.ServeHTTP(rr, req)

 // Проверка кода ответа
 if rr.Code != http.StatusOK {
  t.Errorf("Expected status %v, got %v", http.StatusOK, rr.Code)
 }

 var response map[string]string
 err = json.NewDecoder(rr.Body).Decode(&response)
 if err != nil {
  t.Errorf("Expected a valid JSON response, got error: %v", err)
 }

 // Проверка, что ответ содержит expression
 if _, exists := response["expression"]; !exists {
  t.Errorf("Expected field 'expression' in response")
 }
}

// Тестирование получения всех выражений
func TestGetAllExpressions(t *testing.T) {
 req, err := http.NewRequest("GET", "/api/v1/expressions", nil)
 if err != nil {
  t.Fatal(err)
 }

 rr := httptest.NewRecorder()
 handler := http.HandlerFunc(orchestrator.GetAllExpressions)

 handler.ServeHTTP(rr, req)

 // Проверка кода ответа
 if rr.Code != http.StatusOK {
  t.Errorf("Expected status %v, got %v", http.StatusOK, rr.Code)
 }

 var response []map[string]string
 err = json.NewDecoder(rr.Body).Decode(&response)
 if err != nil {
  t.Errorf("Expected a valid JSON response, got error: %v", err)
 }

 // Проверка, что есть хотя бы одно выражение в ответе
 if len(response) == 0 {
  t.Errorf("Expected non-empty response for expressions")
 }
}