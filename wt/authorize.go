package wt

import "context"

// Authorize sets the JWT token of the WeTransfer client to issue
// authorized requests to the API
func Authorize(c *Client) error {
	req, err := c.NewRequest("POST", "authorize", nil)
	if err != nil {
		return err
	}

	var responseMessage struct {
		Success bool   `json:"success"`
		Token   string `json:"token,omitempty"`
	}

	_, err = c.Do(context.Background(), req, &responseMessage)
	if err != nil {
		return err
	}

	c.JWTAuthToken = responseMessage.Token
	return nil
}
