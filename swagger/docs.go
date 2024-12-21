// Package docs provides API documentation.
//
// This is a sample API for a calculator.
//
//	Schemes:
//	http
//
//	Host: localhost:8080
//	BasePath: /api/v1
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
// swagger:meta
package swagger

// swagger:route POST /calculate calculate calculateRequest
//
// Calculates the given expression.
//
// responses:
//   200: calculateResponse
//   400: errorResponse
//   422: errorResponse
//   500: errorResponse

// swagger:parameters calculateRequest
type calculateRequestWrapper struct {
	// in: body
	Body struct {
		Expression string `json:"expression"`
	}
}

// swagger:response calculateResponse
type calculateResponseWrapper struct {
	// in: body
	Body struct {
		Result string `json:"result"`
	}
}

// swagger:response errorResponse
type errorResponseWrapper struct {
	// in: body
	Body struct {
		Error string `json:"error"`
	}
}
