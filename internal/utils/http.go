package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type JsonHttpClient struct {
	http                *http.Client
	requestTransformers []JsonHttpClientRequestTransformerFn
}

type JsonHttpClientOption = func(*JsonHttpClient)
type JsonHttpClientHttpClientConfigurationFn = func(*http.Client)
type JsonHttpClientRequestTransformerFn = func(*http.Request) error

func NewJsonHttpClient(opts ...JsonHttpClientOption) *JsonHttpClient {
	http := &http.Client{Transport: http.DefaultTransport}
	client := &JsonHttpClient{
		http: http,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func WithHttpConfiguration(fn JsonHttpClientHttpClientConfigurationFn) JsonHttpClientOption {
	return func(c *JsonHttpClient) {
		fn(c.http)
	}
}

func WithRequestTransformer(fn JsonHttpClientRequestTransformerFn) JsonHttpClientOption {
	return func(c *JsonHttpClient) {
		c.requestTransformers = append(c.requestTransformers, fn)
	}
}

func (c *JsonHttpClient) Get(ctx context.Context, path string, input interface{}, output interface{}) error {
	req, err := c.newRequest(ctx, http.MethodGet, path, input)
	if err != nil {
		return err
	}
	return c.do(req, output)
}

func (c *JsonHttpClient) Post(ctx context.Context, path string, input interface{}, output interface{}) error {
	req, err := c.newRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return err
	}
	return c.do(req, output)
}

func (c *JsonHttpClient) Put(ctx context.Context, path string, input interface{}, output interface{}) error {
	req, err := c.newRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return err
	}
	return c.do(req, output)
}

func (c *JsonHttpClient) Patch(ctx context.Context, path string, input interface{}, output interface{}) error {
	req, err := c.newRequest(ctx, http.MethodPatch, path, input)
	if err != nil {
		return err
	}
	return c.do(req, output)
}

func (c *JsonHttpClient) Delete(ctx context.Context, path string, input interface{}, output interface{}) error {
	req, err := c.newRequest(ctx, http.MethodDelete, path, input)
	if err != nil {
		return err
	}
	return c.do(req, output)
}

func (c *JsonHttpClient) newRequest(ctx context.Context, method, path string, input interface{}) (*http.Request, error) {
	bs := []byte{}
	if input != nil {
		bs2, err := json.Marshal(input)
		if err != nil {
			return nil, NewJsonHttpClientError("failed to marshal request body: %v", err)
		}
		bs = bs2
	}
	reqBody := bytes.NewReader(bs)
	req, err := http.NewRequest(method, path, reqBody)
	for _, reqTransformer := range c.requestTransformers {
		if err := reqTransformer(req); err != nil {
			return nil, NewJsonHttpClientError("failed to apply HTTP request transformer: %v", err)
		}
	}
	if err != nil {
		return nil, NewJsonHttpClientError("failed to create request: %v", err)
	}
	req = req.WithContext(ctx)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *JsonHttpClient) do(req *http.Request, output interface{}) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		responseBody := string(bs)
		return NewJsonHttpClientRequestError(resp.StatusCode, responseBody, "request failed with status code %d: %s", resp.StatusCode, responseBody)
	}
	err = json.Unmarshal(bs, output)
	if err != nil {
		responseBody := string(bs)
		return NewJsonHttpClientRequestError(resp.StatusCode, responseBody, "request could not be unmarshalled: %v", err)
	}
	return nil
}

var _ error = (*JsonHttpClientError)(nil)

type JsonHttpClientError struct {
	Message      string
	StatusCode   int
	ResponseBody string
}

func (e JsonHttpClientError) Error() string {
	return e.Message
}

func NewJsonHttpClientError(message string, args ...interface{}) JsonHttpClientError {
	return JsonHttpClientError{
		Message:      fmt.Sprintf(message, args...),
		StatusCode:   0,
		ResponseBody: "",
	}
}

func NewJsonHttpClientRequestError(statusCode int, responseBody string, message string, args ...interface{}) JsonHttpClientError {
	return JsonHttpClientError{
		Message:      fmt.Sprintf(message, args...),
		StatusCode:   statusCode,
		ResponseBody: responseBody,
	}
}

func IsJsonHttpClientError(err error) bool {
	_, ok := err.(JsonHttpClientError)
	return ok
}

func IsJsonHttpClientErrorWithResponseText(err error, containsText string) bool {
	e, ok := err.(JsonHttpClientError)
	return ok && strings.Contains(e.ResponseBody, containsText)
}
