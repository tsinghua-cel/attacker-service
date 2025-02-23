package types

import "context"

type AISession interface {
	Ask(msg string) (string, error)
}

type AIEngine interface {
	NewSession(ctx context.Context, model string) AISession
}
