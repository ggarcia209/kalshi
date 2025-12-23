package kalshi

import (
	"testing"

	"github.com/ggarcia209/kalshi/config"
	"github.com/spf13/viper"
)

func testClient(t *testing.T) *Client {
	c, err := NewClient(
		viper.GetString(config.KalshiDemoTradingHttpUrl),
		viper.GetString(config.KalshiApiKeyFile),
		viper.GetString(config.KalshiApiKey),
		viper.GetString(config.KalshiApiKeyId),
		viper.GetBool(config.KalshiApiKeyUseFile),
		viper.GetInt(config.KalshiRequestsPerSecond),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	return c
}
