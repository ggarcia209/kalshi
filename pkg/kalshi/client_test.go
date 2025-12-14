package kalshi

import (
	"testing"
	"time"

	"github.com/ggarcia209/kalshi/config"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

var rateLimit = rate.NewLimiter(rate.Every(time.Second), 10-1)

func testClient(t *testing.T) *Client {
	c, err := NewClient(
		viper.GetString(config.KalshiDemoTradingHttpUrl),
		viper.GetString(config.KalshiApiKeyFile),
		viper.GetString(config.KalshiApiKey),
		viper.GetString(config.KalshiApiKeyId),
		viper.GetBool(config.KalshiApiKeyUseFile),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	c.WriteRatelimit = rateLimit

	return c
}
