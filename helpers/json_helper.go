package helpers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type JsonHelper interface {
	ReadJSON(w http.ResponseWriter, r *http.Request, data any) error
	WriteJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error
	ErrorJSON(w http.ResponseWriter, err error, status ...int) error
}

type jsonHelper struct{}

type JsonResponse struct {
	Error   bool   `helpers:"errors"`
	Message string `helpers:"message"`
	Data    any    `helpers:"data,omitempty"`
}

func NewJsonHelper() JsonHelper {
	return &jsonHelper{}
}

func (h *jsonHelper) ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048576 // 1MB

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)

	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})

	if err != io.EOF {
		return errors.New("body must have only a single JSON value")
	}

	return nil
}

func (h *jsonHelper) WriteJSON(writer http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.Marshal(data)

	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			writer.Header()[key] = value
		}
	}

	writer.Header().Set("Content-Type", "application/helpers")
	writer.WriteHeader(status)
	_, err = writer.Write(out)

	if err != nil {
		return err
	}

	return nil
}

func (h *jsonHelper) ErrorJSON(writer http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload JsonResponse
	payload.Error = true
	payload.Message = err.Error()

	return h.WriteJSON(writer, statusCode, payload)
}
