package limiter

type Limiter interface {
	// Allow returns an error if the request should be rejected
	// should call returned function when request is consumed
	Allow() (func(), error)
}
