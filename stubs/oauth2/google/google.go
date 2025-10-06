package google

import "context"

type Token struct {
	AccessToken string
}

type TokenSource interface {
	Token() (*Token, error)
}

type dummyTokenSource struct{}

func (dummyTokenSource) Token() (*Token, error) {
	return &Token{AccessToken: "stub-token"}, nil
}

type Credentials struct {
	TokenSource TokenSource
}

func FindDefaultCredentials(ctx context.Context, scope ...string) (*Credentials, error) {
	return &Credentials{TokenSource: dummyTokenSource{}}, nil
}
