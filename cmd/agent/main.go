package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
    "github.com/gulovv/web_calculator/calculation"
)

type Response struct {
    Task Task `json:"task"`
}

type Task struct {
    ID         int     `json:"id"`
    Expression string  `json:"expression"`
    Result     float64 `json:"result,omitempty"`
    Status     string  `json:"status"`
}

//_______________________________________________________________________________________________________________________________

// Функция, которая получает задачу и отправляет результат
func agent() {

    defer func() {
        if r := recover(); r != nil {
            fmt.Printf("Ошибка при вычислении выражения: %v\n", r)
        }
    }()
    for {

        
        // Получаем задачу
        resp, err := http.Get("http://orchestrator:8080/api/v1/task") // Получаем задачу с оркестратора
        if err != nil {
            fmt.Println("Ошибка получения задачи:", err)
            time.Sleep(2 * time.Second) // Пауза перед повторной попыткой
            continue
        }

        // Чтение ответа
        var response Response
        body, err := io.ReadAll(resp.Body) // Вместо ioutil.ReadAll
        if err != nil {
            fmt.Println("Ошибка чтения ответа:", err)
            continue
        }

        // Для диагностики — выводим тело ответа
        //fmt.Println("Тело ответа:", string(body))

        // Парсим ответ в структуру Response
        err = json.Unmarshal(body, &response)
        if err != nil {
            //fmt.Println("Ошибка парсинга задачи:", err)
            continue
        }

        // Получаем задачу из структуры Response
        task := response.Task
        

        // Проверяем, если задача уже завершена
        if task.Status == "completed" {
            fmt.Println("Задача уже завершена, пропускаем её.")
            time.Sleep(1 * time.Second) // Даем время на обновление очереди
            continue
        }

        // Выводим математическое выражение
        fmt.Println("Полученное выражение:", task.Expression)
        tokens := calculation.Tokenize(task.Expression)

        

        // Парсинг: построение AST
        parser := calculation.Parser{Tokens: tokens}
        ast := parser.ParseExpression() // исправленный вызов функции

        fmt.Println("Результат вычисления:", ast.Value)

        // Обновляем результат задачи
        task.Result = ast.Value
        task.Status = "completed" // Обновляем статус задачи на "completed"
        taskData, _ := json.Marshal(task)


        // Отправляем результат на оркестратор
        resp, err = http.Post("http://orchestrator:8080/api/v1/task/result", "application/json", bytes.NewBuffer(taskData))
        if err != nil {
            fmt.Println("Ошибка отправки результата:", err)
            continue
        }
        resp.Body.Close()

        

        fmt.Printf("Задача: %s, Результат: %f\n", task.Expression, task.Result)

        // Пауза перед следующим запросом
        time.Sleep(1 * time.Second)
    }
}

//_______________________________________________________________________________________________________________________________


func main(){
	agent()
}