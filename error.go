package dockerclient

import "fmt"

// Error represents failures in the API. It represents a failure from the API.
type DockerClientError struct {
	Status  int
	Message string
}

func NewDockerClientError(code int, msg string) *DockerClientError {
	return &DockerClientError{Status: code, Message: msg}
}

func (e *DockerClientError) Error() string {
	return fmt.Sprintf("DockerrClient(%d): %s", e.Status, e.Message)
}
