package logging

import (
	"context"
	"slices"
)

const (
	FieldLogID  = "log_id"
	FieldSource = "source"
	FieldJob    = "job"

	SourceAPI  = "api"
	SourceTick = "tick"
)

type contextFieldKey struct{}

func AddContextFields(ctx context.Context, args ...any) context.Context {
	existing, _ := ctx.Value(contextFieldKey{}).([]any)
	return context.WithValue(ctx, contextFieldKey{}, append(slices.Clone(existing), args...))
}

func ListContextFields(ctx context.Context) []any {
	fields, _ := ctx.Value(contextFieldKey{}).([]any)
	return fields
}
