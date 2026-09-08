package clients

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type CPGWClient struct {
	httpClient *http.Client
	logger     *slog.Logger
	baseURL    string
	account    AccountInfo
}

func NewCPGWClient(logger *slog.Logger, httpClient *http.Client, baseURL string) *CPGWClient {
	client := &CPGWClient{logger: logger, httpClient: httpClient, baseURL: baseURL}
	if err := client.Build_AccountInfo(); err != nil {
		client.logger.Warn("failed to build account info: %w", slog.String("error", err.Error()))
	}
	return client
}

func (cpgw *CPGWClient) Start_tickle(ctx context.Context) {
	go func() {
		cpgw.logger.InfoContext(ctx, "starting periodic tickle")
		for {
			select {
			case <-ctx.Done():
				cpgw.logger.WarnContext(ctx, "tickle context exit", slog.String("error", ctx.Err().Error()), slog.String("cause", context.Cause(ctx).Error()))
				return
			case <-time.After(time.Hour):
				err := cpgw.Do_tickle()
				if err != nil {
					cpgw.logger.WarnContext(ctx, "failed to call tickle api", slog.String("err", err.Error()))
				}
			}
		}
	}()
}

// Do_tickle is the Brokerage Keep-Alive Ping. It corresponds to
// https://www.interactivebrokers.com/docs/web-api/api-reference/trading/trading-session/get-session-token
func (cpgw *CPGWClient) Do_tickle() error {
	apisuffix := "/tickle"
	url := cpgw.baseURL + apisuffix
	resp, err := cpgw.httpClient.Post(url, "application/json", strings.NewReader("{}"))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	print_helper(resp, url)
	return nil
}

// Do_iserver_auth_ssodh_init initializes a brokerage session. It corresponds to
// https://www.interactivebrokers.com/docs/web-api/api-reference/trading/trading-session/initialize-session
func (cpgw *CPGWClient) Do_iserver_auth_ssodh_init() error {
	apisuffix := "/iserver/auth/ssodh/init"
	url := cpgw.baseURL + apisuffix
	resp, err := cpgw.httpClient.Post(url, "application/json", bytes.NewBufferString("{}"))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	print_helper(resp, url)
	return nil
}

// Do_iserver_account_orders corresponds to
// https://www.interactivebrokers.com/docs/web-api/api-reference/trading/trading-orders/get-open-orders
func (cpgw *CPGWClient) Do_iserver_account_orders() error {
	apisuffix := "/iserver/account/orders"
	url := cpgw.baseURL + apisuffix
	// First run doesn't return any results
	_, _ = cpgw.httpClient.Get(url)
	// Actual run expected to return something
	resp, err := cpgw.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed on %s: %w", apisuffix, err)
	}
	defer resp.Body.Close()
	print_helper(resp, url)
	return nil
}

// Do_portfolio2_positions corresponds to
// https://www.interactivebrokers.com/docs/web-api/v1/endpoints/portfolio/positions-new
func (cpgw *CPGWClient) Do_portfolio2_positions() error {
	apisuffix := "/portfolio2/" + cpgw.account.AccountId + "/positions"
	url := cpgw.baseURL + apisuffix
	resp, err := cpgw.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed on %s: %w", apisuffix, err)
	}
	defer resp.Body.Close()
	print_helper(resp, url)
	return nil
}

// Do_portfolio_positions_by_conid corresponds to
// https://www.interactivebrokers.com/docs/web-api/v1/endpoints/portfolio/position-contract-info
func (cpgw *CPGWClient) Do_portfolio_positions_by_conid(conid int) error {
	apisuffix := "/portfolio/positions/" + fmt.Sprintf("%d", conid)
	url := cpgw.baseURL + apisuffix
	resp, err := cpgw.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to Get %s: %w", apisuffix, err)
	}
	defer resp.Body.Close()

	positions := make(map[string][]PositionByConId)
	if err := json.UnmarshalRead(resp.Body, &positions); err != nil {
		return fmt.Errorf("failed to unmarshal positions for conid JSON: %w", err)
	}
	fmt.Printf("positions for conid by account: %+v\n", positions)

	return nil
}

// Do_portfolio_accounts corresponds to
// https://www.interactivebrokers.com/docs/web-api/v1/endpoints/portfolio/portfolio-accounts
func (cpgw *CPGWClient) Do_portfolio_accounts() error {
	apisuffix := "/portfolio/accounts"
	url := cpgw.baseURL + apisuffix
	resp, err := cpgw.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to Get %s: %w", apisuffix, err)
	}
	defer resp.Body.Close()
	print_helper(resp, url)
	return nil
}

func (cpgw *CPGWClient) Build_AccountInfo() error {
	apisuffix := "/portfolio/accounts"
	url := cpgw.baseURL + apisuffix
	resp, err := cpgw.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed on %s: %w", apisuffix, err)
	}
	defer resp.Body.Close()

	accounts := make([]AccountInfo, 0, 1)
	err = json.UnmarshalRead(resp.Body, &accounts)
	if err != nil {
		return fmt.Errorf("failed to unmarshal portfolio accounts JSON: %w", err)
	}
	fmt.Printf("portfolio accounts: %+v\n", accounts)
	cpgw.account = accounts[0]
	return nil
}

func (cpgw *CPGWClient) Account() AccountInfo {
	return cpgw.account
}

type AccountInfo struct {
	Id           string `json:"id"`
	AccountId    string `json:"accountId"`
	TradingType  string `json:"tradingType"`
	AccountTitle string `json:"accountTitle"`
	Desc         string `json:"desc"`
}

type PositionByConId struct {
	ConId        int     `json:"conid"`
	Currency     string  `json:"currency"`
	ContractDesc string  `json:"contractDesc"`
	Position     float64 `json:"position"`
	AssetClass   string  `json:"assetClass"`
}

func print_helper(resp *http.Response, endpoint string) {
	fmt.Println()
	fmt.Println("endpoint:", endpoint, "->", time.Now().Local().Format(time.RFC3339), "->", resp.Status)
	if endpoint == "/v1/api/tickle" {
		return
	}
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("body:", string(body))
}
