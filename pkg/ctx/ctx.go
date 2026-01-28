package ctx

import "context"

type ctxUserID struct{}

func SetUserId(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxUserID{}, userID)

}

func UserId(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ctxUserID{}).(string)
	return id, ok
}
