package models

// Структура для запроса
type Request struct {
	Expression string `json:"expression"`
}

// Структура для успешного ответа
type SuccessResponse struct {
	Result string `json:"result"`
}

// Структура для ошибки
type ErrorResponse struct {
	Error string `json:"error"`
}
