package concourse

import (
	"fmt"
	"io"
	"net/http"
)

const maxResponseSize = 10 * 1024 * 1024

func readAll(response *http.Response) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	return body, nil
}
