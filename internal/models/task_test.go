package models

import (
	"testing"
)

func TestWebhookPayload_Validate(t *testing.T) {
	//table driven tests
	//An array of anonymous structs defining our test cases
	tests := []struct {
		name        string         // The name of the sub-test
		payload     WebhookPayload // The input data to test
		expectedErr string         // The error message we expect (or "" if it should pass)
	}{
		{
			name: "Valid Payload",
			payload: WebhookPayload{
				URL:   "https://api.client.com/hooks",
				Event: "payment.success",
				Data:  map[string]interface{}{"order_id": 123},
			},
			expectedErr: "", // Expecting NO error
		},
		{
			name: "Missing URL",
			payload: WebhookPayload{
				Event: "payment.success",
				Data:  map[string]interface{}{"order_id": 123},
			},
			expectedErr: "webhook URL is required",
		},
		{
			name: "Missing Event",
			payload: WebhookPayload{
				URL:  "https://api.client.com/hooks",
				Data: map[string]interface{}{"order_id": 123},
			},
			expectedErr: "event type is required",
		},
		{
			name: "Missing Data",
			payload: WebhookPayload{
				URL:   "https://api.client.com/hooks",
				Event: "payment.success",
			},
			expectedErr: "data payload is required",
		},
	}

	//Iterate through the table and run each test
	for _, tc := range tests {
		// t.Run creates a "sub-test" with the name we defined
		t.Run(tc.name, func(t *testing.T) {

			err := tc.payload.Validate()

			//expected an error, but didn't get one
			if tc.expectedErr != "" && err == nil {
				t.Fatalf("expected error '%s', but got nil", tc.expectedErr)
			}

			// got an error, but it was the wrong one
			if tc.expectedErr != "" && err.Error() != tc.expectedErr {
				t.Errorf("expected error '%s', but got '%v'", tc.expectedErr, err)
			}

			//expected NO error, but we got one
			if tc.expectedErr == "" && err != nil {
				t.Fatalf("expected no error, but got: %v", err)
			}
		})
	}
}
