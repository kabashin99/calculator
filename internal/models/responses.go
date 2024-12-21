package models

type Request struct {
	Expression string `json:"expression"`
}

type SuccessResponse struct {
	Result string `json:"result"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
