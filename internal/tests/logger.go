package tests

import (
	"context"
	"log/slog"
	"testing"
)

type TestHandler struct {
	t testing.TB
}

func NewTestHandler(t testing.TB) *TestHandler {
	return &TestHandler{t: t}
}

func (h *TestHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *TestHandler) Handle(ctx context.Context, r slog.Record) error {
	h.t.Logf("[%s] %s", r.Level, r.Message)
	return nil
}

func (h *TestHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *TestHandler) WithGroup(name string) slog.Handler {
	return h
}
