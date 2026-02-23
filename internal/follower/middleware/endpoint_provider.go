package middleware

type EndpointProvider interface {
	GetEndpoint() string
}

type FollowersEndpointProvider struct{}

func (s FollowersEndpointProvider) GetEndpoint() string {
	return "http://localhost:8080/followers"
}

type MockEndpointProvider struct {
	FakeEndpoint string
}

func (m MockEndpointProvider) GetEndpoint() string {
	return m.FakeEndpoint
}
