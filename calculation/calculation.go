package calculation

import (
    "fmt"
    "strconv"
    "unicode"
)

// Типы токенов
const (
    TokenNumber = iota
    TokenPlus
    TokenMinus
    TokenMultiply
    TokenDivide
    TokenLParen
    TokenRParen
)

// Token представляет лексему (число, оператор или скобку).
type Token struct {
    Type  int
    Value string
}

// tokenize разбивает входную строку на токены.
func Tokenize(input string) []Token {
    var tokens []Token

    fmt.Printf("[Токенизация] Входная строка: %s\n", input) // Отладка: вывод входной строки

    for i := 0; i < len(input); {
        c := input[i]

        // Пропускаем пробелы
        if c == ' ' {
            i++
            continue
        }

        // Числа (включая десятичные)
        if unicode.IsDigit(rune(c)) || c == '.' {
            start := i
            for i < len(input) && (unicode.IsDigit(rune(input[i])) || input[i] == '.') {
                i++
            }
            token := Token{Type: TokenNumber, Value: input[start:i]}
            tokens = append(tokens, token)
            fmt.Printf("[Токенизация] Число: %s\n", token.Value) // Отладка: вывод числа
            continue
        }

        // Операторы и скобки
        switch c {
        case '+', '-', '*', '/', '(', ')':
            tokenType := map[byte]int{
                '+': TokenPlus, '-': TokenMinus, '*': TokenMultiply,
                '/': TokenDivide, '(': TokenLParen, ')': TokenRParen,
            }[c]
            tokens = append(tokens, Token{Type: tokenType, Value: string(c)})
            fmt.Printf("[Токенизация] Оператор/скобка: %c\n", c) // Отладка: вывод оператора или скобки
        default:
            fmt.Printf("[Ошибка] Неизвестный символ: '%c' (позиция %d)\n", c, i) // Отладка: вывод ошибки
        }
        i++
    }

    fmt.Printf("[Токенизация] Итоговый список токенов: %+v\n", tokens) // Отладка: вывод итогового списка токенов
    return tokens
}

// Node представляет узел AST.
type Node struct {
    Operator string  // "+", "-", "*", "/" или пустая строка для числа.
    Left     *Node   // левый операнд
    Right    *Node   // правый операнд
    Value    float64 // значение, если узел является числом
}

// Parser содержит токены и текущую позицию разбора.
type Parser struct {
    Tokens []Token
    pos    int
}

// current возвращает текущий токен.
func (p *Parser) Current() Token {
    if p.pos < len(p.Tokens) {
        return p.Tokens[p.pos]
    }
    return Token{Type: -1} // Маркер конца
}

// eat "съедает" токен заданного типа и переходит к следующему.
func (p *Parser) Eat(tokenType int) Token {
    token := p.Current()
    if token.Type == tokenType {
        p.pos++
        fmt.Printf("[Парсер] Считан токен: %+v\n", token) // Отладка: вывод считанного токена
        return token
    }
    panic(fmt.Sprintf("[Ошибка] Ожидался токен типа %d, но получен %v\n", tokenType, token))
}

// parseFactor обрабатывает число или выражение в скобках.
func (p *Parser) ParseFactor() *Node {
    token := p.Current()

    if token.Type == TokenNumber {
        p.Eat(TokenNumber)
        val, err := strconv.ParseFloat(token.Value, 64)
        if err != nil {
            panic(err)
        }
        fmt.Printf("[Парсер (число/скобка)] Узел числа: %v\n", val) // Отладка: вывод числа
        return &Node{Value: val}
    } else if token.Type == TokenLParen {
        p.Eat(TokenLParen)
        node := p.ParseExpression()
        p.Eat(TokenRParen)
        return node
    }

    panic("[Ошибка] Ожидалось число или '(', но получено что-то другое")
}

// parseTerm обрабатывает умножение и деление.
func (p *Parser) ParseTerm() *Node {
    node := p.ParseFactor()

    for {
        token := p.Current()
        if token.Type == TokenMultiply || token.Type == TokenDivide {
            p.Eat(token.Type)
            right := p.ParseFactor()
            fmt.Printf("[Парсер (умножение/деление)] Операция: %s с узлами (%v, %v)\n", token.Value, node, right) // Отладка: вывод операции
            if token.Type == TokenMultiply {
                node = &Node{Operator: "*", Left: node, Right: right, Value: node.Value * right.Value}
            } else {
                node = &Node{Operator: "/", Left: node, Right: right, Value: node.Value / right.Value}
            }
        } else {
            break
        }
    }

    return node
}

// parseExpression обрабатывает сложение и вычитание.
func (p *Parser) ParseExpression() *Node {
    node := p.ParseTerm()

    for {
        token := p.Current()
        if token.Type == TokenPlus || token.Type == TokenMinus {
            p.Eat(token.Type)
            right := p.ParseTerm()
            fmt.Printf("[Парсер (сложение/вычитание)] Операция: %s с узлами (%v, %v)\n", token.Value, node, right) // Отладка: вывод операции
            if token.Type == TokenPlus {
                node = &Node{Operator: "+", Left: node, Right: right, Value: node.Value + right.Value}
            } else {
                node = &Node{Operator: "-", Left: node, Right: right, Value: node.Value - right.Value}
            }
        } else {
            break
        }
    }

    return node
}

// printAST выводит AST в виде строки (для отладки).
func PrintAST(node *Node) string {
    if node == nil {
        return ""
    }
    if node.Operator == "" {
        return fmt.Sprintf("%v", node.Value)
    }
    return fmt.Sprintf("(%s %s %s)", PrintAST(node.Left), node.Operator, PrintAST(node.Right))
}