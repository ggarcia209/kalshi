package kalshi

import "context"

//go:generate mockgen -destination ./client_mock.go -package=kalshi . KalshiClientLogic
type KalshiClientLogic interface {
	// Auth
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	Logout(ctx context.Context) error

	// exchange
	ExchangeStatus(ctx context.Context) (*ExchangeStatusResponse, error)
	ExchangeSchedule(ctx context.Context) (*ExchangeScheduleResponse, error)

	// market
	Events(ctx context.Context, req EventsRequest) (*EventsResponse, error)
	Event(ctx context.Context, event string) (*EventResponse, error)
	Market(ctx context.Context, ticker string) (*Market, error)
	Markets(ctx context.Context, req MarketsRequest) (*MarketsResponse, error)
	MarketOrderBook(ctx context.Context, ticker string) (*OrderBook, error)
	MarketHistory(ctx context.Context, ticker string, req MarketHistoryRequest) (*MarketHistoryResponse, error)
	Series(ctx context.Context, seriesTicker string) (*Series, error)
	Trades(ctx context.Context, req TradesRequest) (*TradesResponse, error)

	// orders
	CreateOrder(ctx context.Context, req CreateOrderRequest) (*Order, error)
	CancelOrder(ctx context.Context, orderID string) (*Order, error)
	DecreaseOrder(ctx context.Context, orderID string, req DecreaseOrderRequest) (*Order, error)
	GetOrder(ctx context.Context, orderID string) (*Order, error)
	GetOrders(ctx context.Context, req OrdersRequest) (*OrdersResponse, error)
	GetBalance(ctx context.Context) (Cents, error)
	GetFills(ctx context.Context, req FillsRequest) (*FillsResponse, error)
	GetPositions(ctx context.Context, req PositionsRequest) (*PositionsResponse, error)
	GetSettlements(ctx context.Context, req SettlementsRequest) (*SettlementsResponse, error)
}
