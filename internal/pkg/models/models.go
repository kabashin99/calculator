package models

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type Expression struct {
	ID     string  `json:"id"`
	Status string  `json:"status"`
	UserID string  `json:"user_id"`
	Result float64 `json:"result"`
}

type Task struct {
	ID            string   `json:"id"`
	ExpressionID  string   `json:"expression_id"`
	Operation     string   `json:"operation"`
	Operand1      float64  `json:"operand1"`
	Operand2      float64  `json:"operand2"`
	OperationTime int      `json:"operation_time"`
	Result        float64  `json:"result"`
	DependsOn     []string `json:"depends_on"`
	Status        string   `json:"status"`
}

/*
type Task struct {
	ID            string   `json:"id"`
	Arg1          float64  `json:"arg1"`
	Arg2          float64  `json:"arg2"`
	Operation     string   `json:"operation"`
	OperationTime int      `json:"operation_time"`
	Result        float64  `json:"result"`
	DependsOn     []string `json:"depends_on"`
}
*/
