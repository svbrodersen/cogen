package main

import (
	"cogen/evaluator"
	"cogen/generator"
	"cogen/lexer"
	"cogen/object"
	"cogen/parser"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GenerateRequest struct {
	Program string `json:"program"`
	Delta   []int  `json:"delta"`
}

type EvaluateRequest struct {
	Program string   `json:"program"`
	Args    []string `json:"args"`
}

type Response struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to read body: %v", err))
		return
	}

	var req GenerateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		sendError(w, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	if req.Program == "" {
		sendError(w, "Program is required")
		return
	}

	l := lexer.New(req.Program)
	p := parser.New(l)
	c := generator.New(p)
	generated, err := c.Gen(req.Delta)
	if err != nil {
		sendError(w, fmt.Sprintf("Generation error: %v", err))
		return
	}

	sendResult(w, generated.String())
}

func evaluateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to read body: %v", err))
		return
	}

	var req EvaluateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		sendError(w, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	if req.Program == "" {
		sendError(w, "Program is required")
		return
	}

	l := lexer.New(req.Program)
	p := parser.New(l)
	env := object.NewEnvironment()

	parsedProgram := p.ParseProgram()
	if len(p.Errors()) != 0 {
		sendError(w, p.GetErrorMessage())
		return
	}

	if len(parsedProgram.Variables) > 0 {
		expectedArgs := len(parsedProgram.Variables)
		if len(req.Args) < expectedArgs {
			sendError(w, fmt.Sprintf("Program expects %d arguments, got %d", expectedArgs, len(req.Args)))
			return
		}

		for i, input := range parsedProgram.Variables {
			val := parseArgument(req.Args[i])
			env.Set(input.Ident.Value, val)
		}
	}

	e := evaluator.New(parsedProgram)
	evaluated := e.Eval(parsedProgram, env)
	if evaluated != nil {
			sendResult(w, evaluated.String())
	} else {
		sendResult(w, "nil")
	}
}

func parseArgument(arg string) object.Object {
	l := lexer.New(arg)
	p := parser.New(l)
	parseResult := p.ParseExpression(parser.LOWEST)
	env := object.NewEnvironment()
	e := evaluator.Evaluator{Program: nil}
	return e.Eval(parseResult, env)
}

func sendError(w http.ResponseWriter, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	resp := Response{Error: errMsg}
	json.NewEncoder(w).Encode(resp)
}

func sendResult(w http.ResponseWriter, result string) {
	w.Header().Set("Content-Type", "application/json")
	resp := Response{Result: result}
	json.NewEncoder(w).Encode(resp)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func staticFiles() http.Handler {
	return http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/generate", generateHandler)
	mux.HandleFunc("/api/evaluate", evaluateHandler)

	mux.Handle("/", http.FileServer(http.Dir("static")))
	mux.Handle("/static/", staticFiles())

	handler := corsMiddleware(mux)

	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
