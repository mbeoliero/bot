package utils

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66"

type HTTPHandler struct {
	Method  string
	URL     string
	Header  map[string]string
	Limit   int64
	ReqBody any
}

func (h *HTTPHandler) Do() (*http.Response, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	if h.Method == "" {
		h.Method = http.MethodGet
	}

	bytes, err := json.Marshal(h.ReqBody)
	if err != nil {
		return nil, err
	}
	payload := strings.NewReader(string(bytes))

	req, err := http.NewRequest(h.Method, h.URL, payload)
	if err != nil {
		return nil, err
	}

	req.Header["User-Agent"] = []string{UserAgent}
	for k, v := range h.Header {
		req.Header.Set(k, v)
	}

	return client.Do(req)
}

func (h *HTTPHandler) Bytes() ([]byte, error) {
	resp, err := h.Do()
	if err != nil {
		return nil, err
	}

	body := resp.Body
	defer func() { _ = resp.Body.Close() }()

	return io.ReadAll(body)
}

func (h *HTTPHandler) Json() (gjson.Result, error) {
	resp, err := h.Do()
	if err != nil {
		return gjson.Result{}, err
	}

	body := resp.Body
	defer func() { _ = resp.Body.Close() }()

	var sb strings.Builder
	_, err = io.Copy(&sb, body)
	if err != nil {
		return gjson.Result{}, err
	}

	return gjson.Parse(sb.String()), nil
}
