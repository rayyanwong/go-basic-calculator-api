package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"
)

type NumbersStruct struct {
	Number1 int `json:"number1"`
	Number2 int `json:"number2"`
}
type ResponseStruct struct {
	Result int `json:"result"`
}
type DivideResponseStruct struct {
	Quotient  int `json:"quotient"`
	Remainder int `json:"remainder"`
}

func dumpRequest(w http.ResponseWriter, req *http.Request) {
	dump, err := httputil.DumpRequest(req, true) // The 'true' argument dumps the body
	if err != nil {
		slog.Error("Error while dumping request", "error", err)
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	slog.Info("Received request", "dump", dump)
}

func writeResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func decodeJSONBody[T any](w http.ResponseWriter, req *http.Request) (*T, bool) {
	if req.Method == http.MethodPost {
		defer req.Body.Close()
		var payload T
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			slog.Error("Error while decoding JSON", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil, false
		}
		return &payload, true
	}
	slog.Warn("Method not allowed", "method", req.Method)
	http.Error(w, "Method not allowed!", http.StatusMethodNotAllowed)
	return nil, false
}

func helloHandler(w http.ResponseWriter, req *http.Request) {
	dumpRequest(w, req)
	if req.Method == http.MethodPost {
		fmt.Fprint(w, "Received POST request!")
	} else {
		fmt.Fprintf(w, "Request received and logged!")
	}
}

func addHandler(w http.ResponseWriter, req *http.Request) {
	dumpRequest(w, req)
	if val, ok := decodeJSONBody[NumbersStruct](w, req); ok {
		totalSum := val.Number1 + val.Number2
		writeResponse(w, ResponseStruct{Result: totalSum})
	}
}

func subtractHandler(w http.ResponseWriter, req *http.Request) {
	dumpRequest(w, req)
	if val, ok := decodeJSONBody[NumbersStruct](w, req); ok {
		res := val.Number1 - val.Number2
		writeResponse(w, ResponseStruct{Result: res})
	}
}

func multiplyHandler(w http.ResponseWriter, req *http.Request) {
	dumpRequest(w, req)
	if val, ok := decodeJSONBody[NumbersStruct](w, req); ok {
		result := val.Number1 * val.Number2
		writeResponse(w, ResponseStruct{Result: result})
	}
}

func divideHandler(w http.ResponseWriter, req *http.Request) {
	dumpRequest(w, req)
	if val, ok := decodeJSONBody[NumbersStruct](w, req); ok {
		if val.Number2 == 0 {
			slog.Error("Number 2 cannot be 0!", "reason", "cannot divide by 0")
			http.Error(w, "Number 2 value cannot be 0!", http.StatusBadRequest)
			return
		}
		quotient := val.Number1 / val.Number2
		remainder := val.Number1 % val.Number2

		writeResponse(w, DivideResponseStruct{Quotient: quotient, Remainder: remainder})
	}
}

func sumHandler(w http.ResponseWriter, req *http.Request) {
	dumpRequest(w, req)
	if nums, ok := decodeJSONBody[[]int](w, req); ok {
		var res int
		for _, val := range *nums {
			res += val
		}
		writeResponse(w, ResponseStruct{Result: res})
	}
}

func main() {
	// Handlers
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/subtract", subtractHandler)
	http.HandleFunc("/multiply", multiplyHandler)
	http.HandleFunc("/divide", divideHandler)
	http.HandleFunc("/sum", sumHandler)

	slog.Info("Starting server...", "port", 8080)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Server error: ", err)
	}
}
