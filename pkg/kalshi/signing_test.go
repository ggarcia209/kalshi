//go:build integration

package kalshi

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ggarcia209/kalshi/config"
	"github.com/spf13/viper"
)

const (
	TestTicker = "KXBTC15M"
)

// go test -v -tags=integration --run TestIntegrationAuthenticated
func TestIntegrationAuthenticated(t *testing.T) {
	var tests = []struct {
		name    string
		req     SettlementsRequest
		wantErr error
	}{
		{
			name: "success",
			req:  SettlementsRequest{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli, err := NewClient(
				viper.GetString(config.KalshiLiveTradingHttpUrl),
				viper.GetString(config.KalshiApiKeyId),
				viper.GetString(config.KalshiApiKeyFile),
				viper.GetString(config.KalshiApiKey),
				viper.GetViper().GetBool(config.KalshiApiKeyUseFile),
			)
			if err != nil {
				t.Errorf("NewKalshiClient: %v", err)
				return
			}

			settlements, err := cli.GetSettlements(context.Background(), tt.req)
			if err != nil {
				t.Errorf("cli.GetSettlements: %v", err)
				return
			}

			js, err := json.MarshalIndent(settlements, "", "\t")
			if err != nil {
				t.Errorf("test %s: json.MarshalIndent: %v", tt.name, err)
				return
			}
			_ = js

			t.Logf("test %s settlements:\n%s", tt.name, js)
		})
	}
}

// go test -v -tags=integration --run TestIntegrationUnauthenticated
func TestIntegrationUnauthenticated(t *testing.T) {
	var tests = []struct {
		name    string
		req     MarketsRequest
		wantErr error
	}{
		{
			name: "success",
			req: MarketsRequest{
				SeriesTicker: TestTicker,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli, err := NewClient(
				viper.GetString(config.KalshiLiveTradingHttpUrl),
				viper.GetString(config.KalshiApiKeyId),
				viper.GetString(config.KalshiApiKeyFile),
				viper.GetString(config.KalshiApiKey),
				viper.GetViper().GetBool(config.KalshiApiKeyUseFile),
			)
			if err != nil {
				t.Errorf("NewKalshiClient: %v", err)
				return
			}

			markets, err := cli.Markets(context.Background(), tt.req)
			if err != nil {
				t.Errorf("cli.Markets: %v", err)
				return
			}

			js, err := json.MarshalIndent(markets, "", "\t")
			if err != nil {
				t.Errorf("test %s: json.MarshalIndent: %v", tt.name, err)
				return
			}
			_ = js

			t.Logf("test %s markets:\n%s", tt.name, js)
		})
	}
}
