package proxy

import "context"

type requestTagsKey struct{}

// withRequestTags stores the given tags slice in the context.
func withRequestTags(ctx context.Context, tags []string) context.Context {
	return context.WithValue(ctx, requestTagsKey{}, tags)
}

// RequestTagsFromContext retrieves the tags stored by the tagging middleware.
// It returns nil if no tags are present.
func RequestTagsFromContext(ctx context.Context) []string {
	v, _ := ctx.Value(requestTagsKey{}).([]string)
	return v
}
