package scaler

import (
	"context"
)

type MockScaler struct {
	name string
}

func NewMockScaler() *MockScaler {
	return &MockScaler{
		name: "Mock Scaler",
	}
}

func (m *MockScaler) Name() string {
	return m.name
}

func (m *MockScaler) OnTick(ctx context.Context, metric int) error {
	return nil
}
