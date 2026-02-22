package middleware

import "testing"

func TestEndpointProviderReturnCorrectEndpoint(t *testing.T) {
	// arrange
	endpointProvider := SubscriptionEndpointProvider{}

	// act
	endpoint := endpointProvider.GetEndpoint()

	// assert
	expected := "http://localhost:8080/subscribe"
	if endpoint != expected {
		t.Fatalf("Endpoint provider failed: expected %s but got %s", expected, endpoint)
	}
}
