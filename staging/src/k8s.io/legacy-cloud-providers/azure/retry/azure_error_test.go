// +build !providerless

/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package retry

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetError(t *testing.T) {
	now = func() time.Time {
		return time.Time{}
	}

	tests := []struct {
		code       int
		retryAfter int
		err        error
		expected   *Error
	}{
		{
			code:     http.StatusOK,
			expected: nil,
		},
		{
			code: http.StatusOK,
			err:  fmt.Errorf("some error"),
			expected: &Error{
				Retriable:      true,
				HTTPStatusCode: http.StatusOK,
				RawError:       fmt.Errorf("some error"),
			},
		},
		{
			code: http.StatusBadRequest,
			expected: &Error{
				Retriable:      true,
				HTTPStatusCode: http.StatusBadRequest,
				RawError:       fmt.Errorf("HTTP response: 400"),
			},
		},
		{
			code: http.StatusInternalServerError,
			expected: &Error{
				Retriable:      true,
				HTTPStatusCode: http.StatusInternalServerError,
				RawError:       fmt.Errorf("HTTP response: 500"),
			},
		},
		{
			code: http.StatusSeeOther,
			err:  fmt.Errorf("some error"),
			expected: &Error{
				Retriable:      true,
				HTTPStatusCode: http.StatusSeeOther,
				RawError:       fmt.Errorf("some error"),
			},
		},
		{
			code:       http.StatusTooManyRequests,
			retryAfter: 100,
			expected: &Error{
				Retriable:      true,
				HTTPStatusCode: http.StatusTooManyRequests,
				RetryAfter:     now().Add(100 * time.Second),
				RawError:       fmt.Errorf("HTTP response: 429"),
			},
		},
	}

	for _, test := range tests {
		resp := &http.Response{
			StatusCode: test.code,
			Header:     http.Header{},
		}
		if test.retryAfter != 0 {
			resp.Header.Add("Retry-After", fmt.Sprintf("%d", test.retryAfter))
		}
		rerr := GetError(resp, test.err)
		assert.Equal(t, test.expected, rerr)
	}
}

func TestGetStatusNotFoundAndForbiddenIgnoredError(t *testing.T) {
	now = func() time.Time {
		return time.Time{}
	}

	tests := []struct {
		code       int
		retryAfter int
		err        error
		expected   *Error
	}{
		{
			code:     http.StatusOK,
			expected: nil,
		},
		{
			code:     http.StatusNotFound,
			expected: nil,
		},
		{
			code:     http.StatusForbidden,
			expected: nil,
		},
		{
			code: http.StatusOK,
			err:  fmt.Errorf("some error"),
			expected: &Error{
				Retriable:      true,
				HTTPStatusCode: http.StatusOK,
				RawError:       fmt.Errorf("some error"),
			},
		},
		{
			code: http.StatusBadRequest,
			expected: &Error{
				Retriable:      true,
				HTTPStatusCode: http.StatusBadRequest,
				RawError:       fmt.Errorf("HTTP response: 400"),
			},
		},
		{
			code: http.StatusInternalServerError,
			expected: &Error{
				Retriable:      true,
				HTTPStatusCode: http.StatusInternalServerError,
				RawError:       fmt.Errorf("HTTP response: 500"),
			},
		},
		{
			code: http.StatusSeeOther,
			err:  fmt.Errorf("some error"),
			expected: &Error{
				Retriable:      true,
				HTTPStatusCode: http.StatusSeeOther,
				RawError:       fmt.Errorf("some error"),
			},
		},
		{
			code:       http.StatusTooManyRequests,
			retryAfter: 100,
			expected: &Error{
				Retriable:      true,
				HTTPStatusCode: http.StatusTooManyRequests,
				RetryAfter:     now().Add(100 * time.Second),
				RawError:       fmt.Errorf("HTTP response: 429"),
			},
		},
	}

	for _, test := range tests {
		resp := &http.Response{
			StatusCode: test.code,
			Header:     http.Header{},
		}
		if test.retryAfter != 0 {
			resp.Header.Add("Retry-After", fmt.Sprintf("%d", test.retryAfter))
		}
		rerr := GetStatusNotFoundAndForbiddenIgnoredError(resp, test.err)
		assert.Equal(t, test.expected, rerr)
	}
}

func TestShouldRetryHTTPRequest(t *testing.T) {
	tests := []struct {
		code     int
		err      error
		expected bool
	}{
		{
			code:     http.StatusBadRequest,
			expected: true,
		},
		{
			code:     http.StatusInternalServerError,
			expected: true,
		},
		{
			code:     http.StatusOK,
			err:      fmt.Errorf("some error"),
			expected: true,
		},
		{
			code:     http.StatusOK,
			expected: false,
		},
		{
			code:     399,
			expected: false,
		},
		{
			code:     202,
			expected: true,
		},
	}
	for _, test := range tests {
		resp := &http.Response{
			StatusCode: test.code,
		}
		res := shouldRetryHTTPRequest(resp, test.err)
		if res != test.expected {
			t.Errorf("expected: %v, saw: %v", test.expected, res)
		}
	}
}

func TestIsSuccessResponse(t *testing.T) {
	tests := []struct {
		code     int
		expected bool
	}{
		{
			code:     http.StatusNotFound,
			expected: false,
		},
		{
			code:     http.StatusInternalServerError,
			expected: false,
		},
		{
			code:     http.StatusOK,
			expected: true,
		},
		{
			code:     http.StatusAccepted,
			expected: false,
		},
	}

	for _, test := range tests {
		resp := http.Response{
			StatusCode: test.code,
		}
		res := isSuccessHTTPResponse(&resp)
		if res != test.expected {
			t.Errorf("expected: %v, saw: %v", test.expected, res)
		}
	}
}
