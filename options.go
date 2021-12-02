package opaclient

type Option func(client *Client) error

// OptEnableGoogleIAP enables Google Cloud IAP client.
// If the option is enabled, HTTP client is renewed request by request.
func OptEnableGoogleIAP() Option {
	return func(client *Client) error {
		client.enableGoogleIAP = true
		return nil
	}
}
