package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sentioxyz/go-gecko/format"
	"github.com/sentioxyz/go-gecko/v3/types"
)

const defaultBaseURL = "https://api.coingecko.com/api/v3"

// Client struct
type Client struct {
	httpClient *http.Client

	baseURL       string
	authHeaderKey string
	apiKey        string
}

// NewClient create new client object
func NewClient(httpClient *http.Client) *Client {
	return NewClientWithBaseURL(httpClient, defaultBaseURL)
}

func NewClientWithBaseURL(httpClient *http.Client, baseURL string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{httpClient: httpClient, baseURL: baseURL}
}

func NewClientWithURLAndAuth(httpClient *http.Client, baseURL, authHeaderKey, apiKey string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{httpClient: httpClient, baseURL: baseURL, authHeaderKey: authHeaderKey, apiKey: apiKey}
}

// helper
// doReq HTTP client
func doReq(req *http.Request, client *http.Client) ([]byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if 200 != resp.StatusCode {
		return nil, fmt.Errorf("%s", body)
	}
	return body, nil
}

// MakeReq HTTP request helper
func (c *Client) MakeReq(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	if c.authHeaderKey != "" && c.apiKey != "" {
		req.Header.Add(c.authHeaderKey, c.apiKey)
	}
	resp, err := doReq(req, c.httpClient)
	if err != nil {
		return nil, err
	}
	return resp, err
}

// API

// Ping /ping endpoint
func (c *Client) Ping(ctx context.Context) (*types.Ping, error) {
	url := fmt.Sprintf("%s/ping", c.baseURL)
	resp, err := c.MakeReq(ctx, url)
	if err != nil {
		return nil, err
	}
	var data *types.Ping
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// SimpleSinglePrice /simple/price  Single ID and Currency (ids, vs_currency)
func (c *Client) SimpleSinglePrice(ctx context.Context, id string, vsCurrency string) (*types.SimpleSinglePrice, error) {
	return c.SimpleSinglePriceWithDate(ctx, id, vsCurrency, "")
}

// SimplePrice /simple/price Multiple ID and Currency (ids, vs_currencies)
func (c *Client) SimplePrice(ctx context.Context, ids []string, vsCurrencies []string) (*map[string]map[string]float32, error) {
	return c.SimplePriceWithDate(ctx, ids, vsCurrencies, []string{""})
}

// SimpleSinglePriceWithDate /simple/price  Single ID and Currency (ids, vs_currency)
func (c *Client) SimpleSinglePriceWithDate(ctx context.Context,
	id string, vsCurrency string, date string) (*types.SimpleSinglePrice, error) {
	idParam := []string{strings.ToLower(id)}
	vcParam := []string{strings.ToLower(vsCurrency)}
	dataParam := []string{strings.ToLower(date)}

	t, err := c.SimplePriceWithDate(ctx, idParam, vcParam, dataParam)
	if err != nil {
		return nil, err
	}
	curr := (*t)[id]
	if len(curr) == 0 {
		return nil, fmt.Errorf("id or vsCurrency not existed")
	}
	data := &types.SimpleSinglePrice{ID: id, Currency: vsCurrency, MarketPrice: curr[vsCurrency]}
	return data, nil
}

// SimplePriceWithDate /simple/price Multiple ID and Currency (ids, vs_currencies)
func (c *Client) SimplePriceWithDate(ctx context.Context,
	ids []string, vsCurrencies []string, dates []string) (*map[string]map[string]float32, error) {
	params := url.Values{}
	idsParam := strings.Join(ids[:], ",")
	vsCurrenciesParam := strings.Join(vsCurrencies[:], ",")

	params.Add("ids", idsParam)
	params.Add("vs_currencies", vsCurrenciesParam)
	params.Add("date", dates[0])

	url := fmt.Sprintf("%s/simple/price?%s", c.baseURL, params.Encode())
	resp, err := c.MakeReq(ctx, url)
	if err != nil {
		return nil, err
	}

	t := make(map[string]map[string]float32)
	err = json.Unmarshal(resp, &t)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// SimpleSupportedVSCurrencies /simple/supported_vs_currencies
func (c *Client) SimpleSupportedVSCurrencies(ctx context.Context) (*types.SimpleSupportedVSCurrencies, error) {
	url := fmt.Sprintf("%s/simple/supported_vs_currencies", c.baseURL)
	resp, err := c.MakeReq(ctx, url)
	if err != nil {
		return nil, err
	}
	var data *types.SimpleSupportedVSCurrencies
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CoinsList /coins/list
func (c *Client) CoinsList(ctx context.Context) (*types.CoinList, error) {
	url := fmt.Sprintf("%s/coins/list", c.baseURL)
	resp, err := c.MakeReq(ctx, url)
	if err != nil {
		return nil, err
	}

	var data *types.CoinList
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CoinsMarket /coins/market
func (c *Client) CoinsMarket(ctx context.Context, vsCurrency string, ids []string, order string, perPage int, page int, sparkline bool, priceChangePercentage []string) (*types.CoinsMarket, error) {
	if len(vsCurrency) == 0 {
		return nil, fmt.Errorf("vs_currency is required")
	}
	params := url.Values{}
	// vs_currency
	params.Add("vs_currency", vsCurrency)
	// order
	if len(order) == 0 {
		order = types.OrderTypeObject.MarketCapDesc
	}
	params.Add("order", order)
	// ids
	if len(ids) != 0 {
		idsParam := strings.Join(ids[:], ",")
		params.Add("ids", idsParam)
	}
	// per_page
	if perPage <= 0 || perPage > 250 {
		perPage = 100
	}
	params.Add("per_page", format.Int2String(perPage))
	params.Add("page", format.Int2String(page))
	// sparkline
	params.Add("sparkline", format.Bool2String(sparkline))
	// price_change_percentage
	if len(priceChangePercentage) != 0 {
		priceChangePercentageParam := strings.Join(priceChangePercentage[:], ",")
		params.Add("price_change_percentage", priceChangePercentageParam)
	}
	url := fmt.Sprintf("%s/coins/markets?%s", c.baseURL, params.Encode())
	resp, err := c.MakeReq(ctx, url)
	if err != nil {
		return nil, err
	}
	var data *types.CoinsMarket
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CoinsID /coins/{id}
func (c *Client) CoinsID(ctx context.Context, id string, localization bool, tickers bool, marketData bool, communityData bool, developerData bool, sparkline bool) (*types.CoinsID, error) {

	if len(id) == 0 {
		return nil, fmt.Errorf("id is required")
	}
	params := url.Values{}
	params.Add("localization", format.Bool2String(sparkline))
	params.Add("tickers", format.Bool2String(tickers))
	params.Add("market_data", format.Bool2String(marketData))
	params.Add("community_data", format.Bool2String(communityData))
	params.Add("developer_data", format.Bool2String(developerData))
	params.Add("sparkline", format.Bool2String(sparkline))
	url := fmt.Sprintf("%s/coins/%s?%s", c.baseURL, id, params.Encode())
	resp, err := c.MakeReq(ctx, url)
	if err != nil {
		return nil, err
	}

	var data *types.CoinsID
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CoinsIDTickers /coins/{id}/tickers
func (c *Client) CoinsIDTickers(ctx context.Context, id string, page int) (*types.CoinsIDTickers, error) {
	if len(id) == 0 {
		return nil, fmt.Errorf("id is required")
	}
	params := url.Values{}
	if page > 0 {
		params.Add("page", format.Int2String(page))
	}
	url := fmt.Sprintf("%s/coins/%s/tickers?%s", c.baseURL, id, params.Encode())
	resp, err := c.MakeReq(ctx, url)
	if err != nil {
		return nil, err
	}
	var data *types.CoinsIDTickers
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CoinsIDHistory /coins/{id}/history?date={date}&localization=false
func (c *Client) CoinsIDHistory(ctx context.Context, id string, date string, localization bool) (*types.CoinsIDHistory, error) {
	if len(id) == 0 || len(date) == 0 {
		return nil, fmt.Errorf("id and date is required")
	}
	params := url.Values{}
	params.Add("date", date)
	params.Add("localization", format.Bool2String(localization))

	url := fmt.Sprintf("%s/coins/%s/history?%s", c.baseURL, id, params.Encode())
	resp, err := c.MakeReq(ctx, url)
	if err != nil {
		return nil, err
	}
	var data *types.CoinsIDHistory
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CoinsIDMarketChart /coins/{id}/market_chart?vs_currency={usd, eur, jpy, etc.}&days={1,14,30,max}
func (c *Client) CoinsIDMarketChart(ctx context.Context, id string, vs_currency string, days string) (*types.CoinsIDMarketChart, error) {
	if len(id) == 0 || len(vs_currency) == 0 || len(days) == 0 {
		return nil, fmt.Errorf("id, vs_currency, and days is required")
	}

	params := url.Values{}
	params.Add("vs_currency", vs_currency)
	params.Add("days", days)

	url := fmt.Sprintf("%s/coins/%s/market_chart?%s", c.baseURL, id, params.Encode())
	resp, err := c.MakeReq(ctx, url)
	if err != nil {
		return nil, err
	}

	m := types.CoinsIDMarketChart{}
	err = json.Unmarshal(resp, &m)
	if err != nil {
		return &m, err
	}

	return &m, nil
}

// CoinsIDStatusUpdates

// CoinsIDContractAddress https://api.coingecko.com/api/v3/coins/{id}/contract/{contract_address}
// func CoinsIDContractAddress(id string, address string) (nil, error) {
// 	url := fmt.Sprintf("%s/coins/%s/contract/%s", c.baseURL, id, address)
// 	resp, err := request.MakeReq(url)
// 	if err != nil {
// 		return nil, err
// 	}
// }

// EventsCountries https://api.coingecko.com/api/v3/events/countries
func (c *Client) EventsCountries(ctx context.Context) ([]types.EventCountryItem, error) {
	url := fmt.Sprintf("%s/events/countries", c.baseURL)
	resp, err := c.MakeReq(ctx, url)
	if err != nil {
		return nil, err
	}
	var data *types.EventsCountries
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return nil, err
	}
	return data.Data, nil

}

// EventsTypes https://api.coingecko.com/api/v3/events/types
func (c *Client) EventsTypes(ctx context.Context) (*types.EventsTypes, error) {
	url := fmt.Sprintf("%s/events/types", c.baseURL)
	resp, err := c.MakeReq(ctx, url)
	if err != nil {
		return nil, err
	}
	var data *types.EventsTypes
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return nil, err
	}
	return data, nil

}

// ExchangeRates https://api.coingecko.com/api/v3/exchange_rates
func (c *Client) ExchangeRates(ctx context.Context) (*types.ExchangeRatesItem, error) {
	url := fmt.Sprintf("%s/exchange_rates", c.baseURL)
	resp, err := c.MakeReq(ctx, url)
	if err != nil {
		return nil, err
	}
	var data *types.ExchangeRatesResponse
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return nil, err
	}
	return &data.Rates, nil
}

// Global https://api.coingecko.com/api/v3/global
func (c *Client) Global(ctx context.Context) (*types.Global, error) {
	url := fmt.Sprintf("%s/global", c.baseURL)
	resp, err := c.MakeReq(ctx, url)
	if err != nil {
		return nil, err
	}
	var data *types.GlobalResponse
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return nil, err
	}
	return &data.Data, nil
}

// CoinsInfo /coins/{id}
func (c *Client) CoinInfo(ctx context.Context, id string) (*types.CoinInfo, error) {
	if len(id) == 0 {
		return nil, fmt.Errorf("id is required")
	}
	url := fmt.Sprintf("%s/coins/%s", c.baseURL, id)
	resp, err := c.MakeReq(ctx, url)
	if err != nil {
		return nil, err
	}
	var data *types.CoinInfo
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
