package handler

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// Add these two interfaces for testability (mocking)
type converterIface interface {
	TextToSQL(ctx context.Context, question string) (string, string, error)
}

type executorIface interface {
	Run(ctx context.Context, sql string) ([]map[string]any, error)
}

// QueryHandler holds the dependencies needed to handle query requests
type QueryHandler struct {
	conv converterIface
	exec executorIface
}

func NewQueryHandler(conv converterIface, exec executorIface) *QueryHandler {
	return &QueryHandler{conv: conv, exec: exec}
}

// queryRequest is the expected JSON body for POST /query
type queryRequest struct {
	Question string `json:"question"`
}

// queryResponse is the JSON body returned on success
type queryResponse struct {
	SQL         string           `json:"sql"`
	Explanation string           `json:"explanation"`
	Result      []map[string]any `json:"result"`
}

// errorResponse is the JSON body returned on failure
type errorResponse struct {
	Error string `json:"error"`
	Stage string `json:"stage"` // "parse", "generate", or "execute"
}

// Query handles POST /query
func (h *QueryHandler) Query(c fiber.Ctx) error {
	// Step 1: Parse the request body
	var req queryRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse{
			Error: "invalid request body — expected JSON with a 'question' field",
			Stage: "parse",
		})
	}

	// Step 2: Validate input
	if strings.TrimSpace(req.Question) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse{
			Error: "question is required and cannot be empty",
			Stage: "parse",
		})
	}

	// Step 3: Convert question to SQL
	sql, explanation, err := h.conv.TextToSQL(c.Context(), req.Question)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(errorResponse{
			Error: err.Error(),
			Stage: "generate",
		})
	}

	// Step 4: Execute the SQL
	result, err := h.exec.Run(c.Context(), sql)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse{
			Error: err.Error(),
			Stage: "execute",
		})
	}

	// Step 5: Return success response
	return c.JSON(queryResponse{
		SQL:         sql,
		Explanation: explanation,
		Result:      result,
	})
}
