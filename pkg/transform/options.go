package transform

type Option func(*transformOptions)

type transformOptions struct {
	userID bool
}

// Also pass user id when transforming object.
func WithUserID() Option {
	return func(o *transformOptions) {
		o.userID = true
	}
}
