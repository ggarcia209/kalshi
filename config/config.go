package config

import "github.com/spf13/viper"

// config for running test locally only
const (
	KalshiLiveTradingHttpUrl = "KALSHI_LIVE_TRADING_HTTP_URL"
	KalshiDemoTradingHttpUrl = "KALSHI_DEMO_TRADING_HTTP_URL"
	KalshiApiKeyId           = "KALSHI_API_KEY_ID"
	KalshiApiKey             = "KALSHI_API_KEY"
	KalshiApiKeyFile         = "KALSHI_API_KEY_FILE"
	KalshiApiKeyUseFile      = "KALSHI_API_KEY_USE_FILE"
	KalshiRequestsPerSecond  = "KALSHI_REQUESTS_PER_SECOND"
)

func init() {
	viper.SetDefault(KalshiLiveTradingHttpUrl, "https://api.elections.kalshi.com/trade-api/v2")
	viper.SetDefault(KalshiDemoTradingHttpUrl, "https://demo-api.kalshi.co/trade-api/v2")
	viper.SetDefault(KalshiApiKeyUseFile, false)
	viper.SetDefault(KalshiRequestsPerSecond, 20)
	viper.AutomaticEnv()
}
