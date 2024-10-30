package models

// Структура сообщения для получения токена
type TokenCheckMessage struct {
    Token    string `json:"token"`
    TempQueue string `json:"temp_queue"` // Имя временной очереди для отправки результата
}

// Структура ответа
type TokenCheckResponse struct {
    Valid  bool   `json:"valid"`
    UserID int    `json:"user_id"`
    Error  string `json:"error"`
}