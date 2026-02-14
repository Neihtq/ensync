package middleware

type EndpointProvider interface {
	GetEndpoint() string
}

type SubscriptionEndpointProvider struct{}

func (s SubscriptionEndpointProvider) GetEndpoint() string {
	return "http://localhost:8080/subscribe"
}

type MockEndpointProvider struct {
	FakeEndpoint string
}

func (m MockEndpointProvider) GetEndpoint() string {
	return m.FakeEndpoint
}
