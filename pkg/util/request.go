package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

var (
	// ErrStatusOK send request without status ok response.
	ErrStatusOK = errors.New("failed to request without status ok")
)

// EncodeJSON transfer data from interface to bytes.
func EncodeJSON(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

// DecodeJSON transfer data from bytes to the map with string key and interface value.
func DecodeJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func DecodeJSONToMap(data []byte) (map[string]interface{}, error) {
	var m map[string]interface{}
	err := json.Unmarshal(data, &m)
	return m, err
}

// DecodeReader transfer io to the map with sting key and interface value.
func DecodeReader(r io.Reader, v interface{}) error {
	decode := json.NewDecoder(r)
	return decode.Decode(v)
}

// Request send the request to baseURL with method.
func Request(method, baseURL, path string, body interface{}, header map[string]string) ([]byte, int, error) {
	client := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(25 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*20)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}

	encodeBody, err := EncodeJSON(body)
	if err != nil {
		return nil, 0, err
	}

	url := fmt.Sprintf("%s%s", baseURL, path)
	r, err := http.NewRequest(method, url, bytes.NewBuffer(encodeBody))
	if err != nil {
		return nil, 0, err
	}

	r.Header.Set("Content-Type", "application/json")
	for k, v := range header {
		r.Header.Set(k, v)
	}

	w, err := client.Do(r)
	if err != nil {
		return nil, 0, err
	}
	defer w.Body.Close()

	data, err := ioutil.ReadAll(w.Body)
	if err != nil {
		return nil, 0, err
	}

	return data, w.StatusCode, nil
}

// Response return status to client.
func Response(data []byte, statusCode int, w http.ResponseWriter) {
	w.WriteHeader(statusCode)
	w.Write(data)
}

// ResponseError return error status to client.
func ResponseError(data string, httpCode int, stateCode int, w http.ResponseWriter) {
	w.WriteHeader(httpCode)
	d := fmt.Sprintf("{\"state\" : %d, \"error\" : \"%s\"}", stateCode, data)
	fmt.Fprintf(w, d)
}

// ParseForm ensures the request form is parsed even with invalid content types.
// If we don't do this, POST method without Content-type (even with empty body) will fail.
func ParseForm(r *http.Request) error {
	if r == nil {
		return nil
	}
	if err := r.ParseForm(); err != nil && !strings.HasPrefix(err.Error(), "mime:") {
		return err
	}
	return nil
}
