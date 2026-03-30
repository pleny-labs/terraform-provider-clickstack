package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL       string
	APIKey        string
	SessionCookie string
	APIBasePath   string
	HTTPClient    *http.Client
}

func NewClient(baseURL, apiKey, sessionCookie, apiBasePath string) *Client {
	if apiBasePath == "" {
		apiBasePath = "/api"
	}
	return &Client{
		BaseURL:       baseURL,
		APIKey:        apiKey,
		SessionCookie: sessionCookie,
		APIBasePath:   apiBasePath,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// apiPath constructs the full API path by prepending the configured base path.
func (c *Client) apiPath(resource string) string {
	return c.APIBasePath + resource
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if c.SessionCookie != "" {
		req.Header.Set("Cookie", "connect.sid="+c.SessionCookie)
	} else if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// dataWrapper is a generic wrapper for v2 API responses that contain a "data" field.
type dataWrapper[T any] struct {
	Data T `json:"data"`
}

// unmarshalResponse tries to unmarshal a v2 response. The v2 API may wrap responses
// in a {"data": ...} envelope. We try the wrapper first; if the "data" field is
// zero-valued we fall back to direct unmarshal for compatibility.
func unmarshalResponse[T any](body []byte) (T, error) {
	// Try direct unmarshal first (works for non-wrapped responses and v1 endpoints)
	var direct T
	if err := json.Unmarshal(body, &direct); err == nil {
		// Check if it might be wrapped
		var wrapper dataWrapper[T]
		if json.Unmarshal(body, &wrapper) == nil {
			// Heuristic: if we got a valid wrapper, check if a raw map has a "data" key
			var raw map[string]json.RawMessage
			if json.Unmarshal(body, &raw) == nil {
				if _, hasData := raw["data"]; hasData {
					return wrapper.Data, nil
				}
			}
		}
		return direct, nil
	}
	return direct, fmt.Errorf("unmarshaling response: %s", string(body))
}

// --- Dashboards ---

type Dashboard struct {
	ID                 string             `json:"_id,omitempty"`
	Name               string             `json:"name"`
	Tiles              []Tile             `json:"tiles"`
	Tags               []string           `json:"tags,omitempty"`
	Filters            []Filter           `json:"filters,omitempty"`
	SavedQuery         *string            `json:"savedQuery,omitempty"`
	SavedQueryLanguage *string            `json:"savedQueryLanguage,omitempty"`
	SavedFilterValues  []SavedFilterValue `json:"savedFilterValues,omitempty"`
}

type Tile struct {
	ID     string     `json:"id,omitempty"`
	X      float64    `json:"x"`
	Y      float64    `json:"y"`
	W      float64    `json:"w"`
	H      float64    `json:"h"`
	Config TileConfig `json:"config"`
}

type TileConfig struct {
	Name           string       `json:"name,omitempty"`
	DisplayType    string       `json:"displayType"`
	Source         string       `json:"source,omitempty"`
	Select         []SelectItem `json:"select,omitempty"`
	GroupBy        string       `json:"groupBy,omitempty"`
	Where          string       `json:"where,omitempty"`
	WhereLanguage  string       `json:"whereLanguage,omitempty"`
	Granularity    string       `json:"granularity,omitempty"`
	Fields         []string     `json:"fields,omitempty"`
	Content        *string      `json:"content,omitempty"`
	SortOrder      *string      `json:"sortOrder,omitempty"`
	NumberFormat   *NumberFormat `json:"numberFormat,omitempty"`
}

type SelectItem struct {
	AggFn                string   `json:"aggFn"`
	ValueExpression      string   `json:"valueExpression,omitempty"`
	AggCondition         string   `json:"aggCondition,omitempty"`
	AggConditionLanguage string   `json:"aggConditionLanguage,omitempty"`
	Alias                string   `json:"alias,omitempty"`
	Level                *float64 `json:"level,omitempty"`
}

type NumberFormat struct {
	Output            string `json:"output,omitempty"`
	Mantissa          *int   `json:"mantissa,omitempty"`
	ThousandSeparated *bool  `json:"thousandSeparated,omitempty"`
}

type Filter struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	Expression string `json:"expression"`
	SourceID   string `json:"sourceId"`
}

type SavedFilterValue struct {
	Type      string `json:"type,omitempty"`
	Condition string `json:"condition"`
}

func (c *Client) ListDashboards(ctx context.Context) ([]Dashboard, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, c.apiPath("/dashboards"), nil)
	if err != nil {
		return nil, err
	}
	return unmarshalResponse[[]Dashboard](resp)
}

func (c *Client) GetDashboard(ctx context.Context, id string) (*Dashboard, error) {
	dashboards, err := c.ListDashboards(ctx)
	if err != nil {
		return nil, err
	}
	for _, d := range dashboards {
		if d.ID == id {
			return &d, nil
		}
	}
	return nil, fmt.Errorf("dashboard %s not found", id)
}

func (c *Client) CreateDashboard(ctx context.Context, d Dashboard) (*Dashboard, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, c.apiPath("/dashboards"), d)
	if err != nil {
		return nil, err
	}
	result, err := unmarshalResponse[Dashboard](resp)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateDashboard(ctx context.Context, id string, d Dashboard) (*Dashboard, error) {
	resp, err := c.doRequest(ctx, http.MethodPut, c.apiPath("/dashboards/")+id, d)
	if err != nil {
		return nil, err
	}
	result, err := unmarshalResponse[Dashboard](resp)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteDashboard(ctx context.Context, id string) error {
	_, err := c.doRequest(ctx, http.MethodDelete, c.apiPath("/dashboards/")+id, nil)
	return err
}

// --- Alerts ---

type Alert struct {
	ID                    string        `json:"_id,omitempty"`
	Threshold             float64       `json:"threshold"`
	Interval              string        `json:"interval"`
	ThresholdType         string        `json:"thresholdType"`
	Source                string        `json:"source"`
	Channel               AlertChannel  `json:"channel"`
	DashboardID           *string       `json:"dashboardId,omitempty"`
	TileID                *string       `json:"tileId,omitempty"`
	SavedSearchID         *string       `json:"savedSearchId,omitempty"`
	GroupBy               *string       `json:"groupBy,omitempty"`
	ScheduleOffsetMinutes *int          `json:"scheduleOffsetMinutes,omitempty"`
	ScheduleStartAt       *string       `json:"scheduleStartAt,omitempty"`
	Name                  *string       `json:"name,omitempty"`
	Message               *string       `json:"message,omitempty"`
	State                 string        `json:"state,omitempty"`
	TeamID                string        `json:"teamId,omitempty"`
	Silenced              *AlertSilence `json:"silenced,omitempty"`
	CreatedAt             string        `json:"createdAt,omitempty"`
	UpdatedAt             string        `json:"updatedAt,omitempty"`
}

type AlertChannel struct {
	Type      string `json:"type"`
	WebhookID string `json:"webhookId"`
}

type AlertSilence struct {
	By    string `json:"by"`
	At    string `json:"at"`
	Until string `json:"until"`
}

func (c *Client) ListAlerts(ctx context.Context) ([]Alert, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, c.apiPath("/alerts"), nil)
	if err != nil {
		return nil, err
	}
	return unmarshalResponse[[]Alert](resp)
}

func (c *Client) GetAlert(ctx context.Context, id string) (*Alert, error) {
	alerts, err := c.ListAlerts(ctx)
	if err != nil {
		return nil, err
	}
	for _, a := range alerts {
		if a.ID == id {
			return &a, nil
		}
	}
	return nil, fmt.Errorf("alert %s not found", id)
}

func (c *Client) CreateAlert(ctx context.Context, a Alert) (*Alert, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, c.apiPath("/alerts"), a)
	if err != nil {
		return nil, err
	}
	result, err := unmarshalResponse[Alert](resp)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateAlert(ctx context.Context, id string, a Alert) (*Alert, error) {
	resp, err := c.doRequest(ctx, http.MethodPut, c.apiPath("/alerts/")+id, a)
	if err != nil {
		return nil, err
	}
	result, err := unmarshalResponse[Alert](resp)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteAlert(ctx context.Context, id string) error {
	_, err := c.doRequest(ctx, http.MethodDelete, c.apiPath("/alerts/")+id, nil)
	return err
}

// --- Sources ---

type Source struct {
	ID                                    string              `json:"_id"`
	Name                                  string              `json:"name"`
	Kind                                  string              `json:"kind"`
	Connection                            string              `json:"connection"`
	From                                  *SourceFrom         `json:"from,omitempty"`
	QuerySettings                         []SourceQuerySetting `json:"querySettings,omitempty"`
	DefaultTableSelectExpression          string              `json:"defaultTableSelectExpression,omitempty"`
	TimestampValueExpression              string              `json:"timestampValueExpression,omitempty"`
	ServiceNameExpression                 string              `json:"serviceNameExpression,omitempty"`
	SeverityTextExpression                string              `json:"severityTextExpression,omitempty"`
	BodyExpression                        string              `json:"bodyExpression,omitempty"`
	EventAttributesExpression             string              `json:"eventAttributesExpression,omitempty"`
	ResourceAttributesExpression          string              `json:"resourceAttributesExpression,omitempty"`
	DisplayedTimestampValueExpression     string              `json:"displayedTimestampValueExpression,omitempty"`
	MetricSourceID                        string              `json:"metricSourceId,omitempty"`
	TraceSourceID                         string              `json:"traceSourceId,omitempty"`
	TraceIDExpression                     string              `json:"traceIdExpression,omitempty"`
	SpanIDExpression                      string              `json:"spanIdExpression,omitempty"`
	ImplicitColumnExpression              string              `json:"implicitColumnExpression,omitempty"`
}

type SourceFrom struct {
	DatabaseName string `json:"databaseName"`
	TableName    string `json:"tableName"`
}

type SourceQuerySetting struct {
	Setting string `json:"setting"`
	Value   string `json:"value"`
}

func (c *Client) ListSources(ctx context.Context) ([]Source, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, c.apiPath("/sources"), nil)
	if err != nil {
		return nil, err
	}
	return unmarshalResponse[[]Source](resp)
}

// --- Webhooks ---

type Webhook struct {
	ID          string            `json:"_id,omitempty"`
	Name        string            `json:"name"`
	Service     string            `json:"service"`
	URL         string            `json:"url"`
	Description string            `json:"description,omitempty"`
	QueryParams map[string]string `json:"queryParams,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        string            `json:"body,omitempty"`
	CreatedAt   string            `json:"createdAt,omitempty"`
	UpdatedAt   string            `json:"updatedAt,omitempty"`
}

func (c *Client) ListWebhooks(ctx context.Context) ([]Webhook, error) {
	// The internal API requires a service filter; query all known service types
	resp, err := c.doRequest(ctx, http.MethodGet, c.apiPath("/webhooks")+"?service[]=slack&service[]=generic&service[]=incidentio", nil)
	if err != nil {
		return nil, err
	}
	return unmarshalResponse[[]Webhook](resp)
}

func (c *Client) GetWebhook(ctx context.Context, id string) (*Webhook, error) {
	// List all and find by ID since individual GET may not be available
	webhooks, err := c.ListWebhooks(ctx)
	if err != nil {
		return nil, err
	}
	for _, w := range webhooks {
		if w.ID == id {
			return &w, nil
		}
	}
	return nil, fmt.Errorf("webhook %s not found", id)
}

func (c *Client) CreateWebhook(ctx context.Context, w Webhook) (*Webhook, error) {
	// Webhook mutations use the v1 /api/webhooks path, not the v2 base path
	resp, err := c.doRequest(ctx, http.MethodPost, "/api/webhooks", w)
	if err != nil {
		return nil, err
	}
	var webhook Webhook
	if err := json.Unmarshal(resp, &webhook); err != nil {
		return nil, fmt.Errorf("unmarshaling webhook: %w", err)
	}
	return &webhook, nil
}

func (c *Client) UpdateWebhook(ctx context.Context, id string, w Webhook) (*Webhook, error) {
	// Webhook mutations use the v1 /api/webhooks path, not the v2 base path
	resp, err := c.doRequest(ctx, http.MethodPut, "/api/webhooks/"+id, w)
	if err != nil {
		return nil, err
	}
	var webhook Webhook
	if err := json.Unmarshal(resp, &webhook); err != nil {
		return nil, fmt.Errorf("unmarshaling webhook: %w", err)
	}
	return &webhook, nil
}

func (c *Client) DeleteWebhook(ctx context.Context, id string) error {
	// Webhook mutations use the v1 /api/webhooks path, not the v2 base path
	_, err := c.doRequest(ctx, http.MethodDelete, "/api/webhooks/"+id, nil)
	return err
}
