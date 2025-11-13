package server

import (
	"context"
	"testing"
)

// run and validate the server through tests

func TestCreateUser(t *testing.T) {
	s := New()
	id, err := s.CreateUser(context.Background(), "test")
	_, _ = id, err
}
