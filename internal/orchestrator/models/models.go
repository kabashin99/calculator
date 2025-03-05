package models

type Expression struct {
	ID     string `json:"id"`
	Status string `json:"status"` // pending, processing, done, error
	Result string `json:"result,omitempty"`
}

type Task struct {
	ID           string      `json:"id"`
	ExpressionID string      `json:"expression_id"`
	Arg1         interface{} `json:"arg1"`
	Arg2         interface{} `json:"arg2"`
	Operation    string      `json:"operation"`
	Status       string      `json:"status"`
	DependsOn    []string    `json:"depends_on"`
	Result       float64     `json:"result"`
}

type TaskResponse struct {
	ID            string      `json:"id"`
	Arg1          interface{} `json:"arg1"`
	Arg2          interface{} `json:"arg2"`
	Operation     string      `json:"operation"`
	OperationTime int         `json:"operation_time"`
}

type TaskResult struct {
	ID     string  `json:"id"`
	Result float64 `json:"result"`
}

type SuccessResponse struct {
	Message string `json:"message"` // Добавлено поле Message
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type Request struct {
	Expression string `json:"expression"`
}
