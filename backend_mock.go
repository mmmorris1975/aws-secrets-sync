package main

import "fmt"

type mockBackend struct {
	kmsRequired bool
}

func newMockBackend() *mockBackend {
	return &mockBackend{kmsRequired: false}
}

// KmsRequired will always return false for the mock backend
func (b *mockBackend) KmsRequired() bool {
	return b.kmsRequired
}

// Store will always succeed, unless you pass a zero-length key or nil value (or zero-length string value)
func (b *mockBackend) Store(key string, value interface{}) error {
	if len(key) < 1 {
		return fmt.Errorf("invalid key")
	}

	if value == nil {
		return fmt.Errorf("invalid value")
	}

	switch t := value.(type) {
	case string:
		if len(t) < 1 {
			return fmt.Errorf("invalid value")
		}
	}

	return nil
}
