package models

type Expression struct {
	ID     string
	Status string // "pending", "processing", "done", "error"
	Result float64
}

type Task struct {
	ID           string
	ExpressionID string
	Arg1         float64
	Arg2         float64
	Operation    string
	Status       string
}

type TaskResponse struct {
	ID            string  `json:"id"`
	ExpressionID  string  `json:"expression_id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

type TaskResult struct {
	ID     string  `json:"id"`
	Result float64 `json:"result"`
}
