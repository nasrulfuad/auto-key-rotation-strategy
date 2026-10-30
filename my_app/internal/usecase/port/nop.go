package port

import "context"

type NopLogger struct{}

func (NopLogger) Debug(context.Context, string, ...any) {}
func (NopLogger) Info(context.Context, string, ...any)  {}
func (NopLogger) Warn(context.Context, string, ...any)  {}
func (NopLogger) Error(context.Context, string, ...any) {}
func (n NopLogger) With(...any) Logger                  { return n }
