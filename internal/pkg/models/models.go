package models

type Expression struct {
	ID     string  `json:"id"`
	Status string  `json:"status"` // "pending", "processing", "done", "error"
	Result float64 `json:"result"`
}

type Task struct {
	ID            string   `json:"id"`
	Arg1          float64  `json:"arg1"`
	Arg2          float64  `json:"arg2"`
	Operation     string   `json:"operation"`
	OperationTime int      `json:"operation_time"`
	Result        float64  `json:"result"`
	DependsOn     []string `json:"depends_on"` // Зависимости от других задач
}
