package jd

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const baseURL = "https://api.jd.com/routerjson"

type Client struct {
	AppKey    string
	AppSecret string
	SiteID    string
	PID       string
	BaseURL   string
	HTTP      *http.Client
}

type APIResponse struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
}

func NewClient(appKey, appSecret, siteID, pid string) *Client {
	return &Client{
		AppKey:    appKey,
		AppSecret: appSecret,
		SiteID:    siteID,
		PID:       pid,
		BaseURL:   baseURL,
		HTTP: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) call(method string, params map[string]string) (*APIResponse, error) {
	reqParams := map[string]string{
		"method":      method,
		"app_key":     c.AppKey,
		"format":      "json",
		"v":           "1.0",
		"sign_method": "md5",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
	}
	for k, v := range params {
		reqParams[k] = v
	}
	reqParams["sign"] = c.sign(reqParams)

	form := buildForm(reqParams)

	var lastErr error
	for i := 0; i < 3; i++ {
		resp, err := c.doRequest(form)
		if err != nil {
			lastErr = err
			if i < 2 {
				time.Sleep(time.Duration(1<<i) * time.Second)
			}
			continue
		}
		if resp.Code == "0" {
			return resp, nil
		}
		if !isRetryable(resp.Code) {
			return resp, fmt.Errorf("jd api error: %s %s", resp.Code, resp.Message)
		}
		lastErr = fmt.Errorf("jd api error: %s %s", resp.Code, resp.Message)
		time.Sleep(time.Duration(1<<i) * time.Second)
	}
	return nil, fmt.Errorf("jd api retry exhausted: %w", lastErr)
}

func (c *Client) sign(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var sb strings.Builder
	sb.WriteString(c.AppSecret)
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString(params[k])
	}
	sb.WriteString(c.AppSecret)
	hash := md5.Sum([]byte(sb.String()))
	return strings.ToUpper(fmt.Sprintf("%x", hash))
}

func (c *Client) doRequest(form url.Values) (*APIResponse, error) {
	resp, err := c.HTTP.PostForm(c.BaseURL, form)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return &apiResp, nil
}

func buildForm(params map[string]string) url.Values {
	form := url.Values{}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		form.Add(k, params[k])
	}
	return form
}

func isRetryable(code string) bool {
	retryable := map[string]bool{
		"1001": true,
		"1002": true,
		"1003": true,
	}
	return retryable[code]
}

// CreateSubPID calls jd.union.open.user.pid.get
func (c *Client) CreateSubPID(unionID, pidName string) (string, error) {
	params := map[string]string{
		"360buy_param_json": fmt.Sprintf(
			`{"unionId":%s,"key":"%s","value":"%s","pidName":"%s"}`,
			c.PID, c.PID, c.SiteID, pidName),
	}
	resp, err := c.call("jd.union.open.user.pid.get", params)
	if err != nil {
		return "", err
	}
	return string(resp.Result), nil
}

// ConvertLink calls jd.union.open.goods.link.query
func (c *Client) ConvertLink(subPID, materialURL string) (string, error) {
	params := map[string]string{
		"360buy_param_json": fmt.Sprintf(
			`{"materialIds":"%s","subUnionId":"%s","positionId":%s}`,
			materialURL, subPID, c.SiteID),
	}
	resp, err := c.call("jd.union.open.goods.link.query", params)
	if err != nil {
		return "", err
	}
	return string(resp.Result), nil
}

// QueryOrders calls jd.union.open.order.query
func (c *Client) QueryOrders(startTime, endTime string, pageNo, pageSize int) (*APIResponse, error) {
	params := map[string]string{
		"360buy_param_json": fmt.Sprintf(
			`{"orderQueryType":"1","pageNo":%d,"pageSize":%d,"time":"%s","endTime":"%s"}`,
			pageNo, pageSize, startTime, endTime),
	}
	return c.call("jd.union.open.order.query", params)
}

// QueryBonusOrders calls jd.union.open.order.bonus.query
func (c *Client) QueryBonusOrders(startTime, endTime string, pageNo, pageSize int) (*APIResponse, error) {
	params := map[string]string{
		"360buy_param_json": fmt.Sprintf(
			`{"optType":1,"pageNo":%d,"pageSize":%d,"startTime":"%s","endTime":"%s"}`,
			pageNo, pageSize, startTime, endTime),
	}
	return c.call("jd.union.open.order.bonus.query", params)
}
