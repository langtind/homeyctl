package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/langtind/homey-cli/internal/config"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{
		baseURL: cfg.BaseURL(),
		token:   cfg.Token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Devices

func (c *Client) GetDevices() (json.RawMessage, error) {
	return c.doRequest("GET", "/api/manager/devices/device/", nil)
}

func (c *Client) GetDevice(id string) (json.RawMessage, error) {
	return c.doRequest("GET", "/api/manager/devices/device/"+id, nil)
}

func (c *Client) SetCapability(deviceID, capability string, value interface{}) error {
	body := map[string]interface{}{"value": value}
	_, err := c.doRequest("PUT", fmt.Sprintf("/api/manager/devices/device/%s/capability/%s", deviceID, capability), body)
	return err
}

// Flows

func (c *Client) GetFlows() (json.RawMessage, error) {
	return c.doRequest("GET", "/api/manager/flow/flow/", nil)
}

func (c *Client) GetAdvancedFlows() (json.RawMessage, error) {
	return c.doRequest("GET", "/api/manager/flow/advancedflow/", nil)
}

func (c *Client) TriggerFlow(id string) error {
	_, err := c.doRequest("POST", fmt.Sprintf("/api/manager/flow/flow/%s/trigger", id), nil)
	return err
}

func (c *Client) TriggerAdvancedFlow(id string) error {
	_, err := c.doRequest("POST", fmt.Sprintf("/api/manager/flow/advancedflow/%s/trigger", id), nil)
	return err
}

func (c *Client) CreateFlow(flow map[string]interface{}) (json.RawMessage, error) {
	return c.doRequest("POST", "/api/manager/flow/flow/", flow)
}

func (c *Client) CreateAdvancedFlow(flow map[string]interface{}) (json.RawMessage, error) {
	return c.doRequest("POST", "/api/manager/flow/advancedflow/", flow)
}

func (c *Client) DeleteFlow(id string) error {
	_, err := c.doRequest("DELETE", "/api/manager/flow/flow/"+id, nil)
	return err
}

func (c *Client) DeleteAdvancedFlow(id string) error {
	_, err := c.doRequest("DELETE", "/api/manager/flow/advancedflow/"+id, nil)
	return err
}

// Flow cards

func (c *Client) GetFlowTriggers() (json.RawMessage, error) {
	return c.doRequest("GET", "/api/manager/flow/flowcardtrigger/", nil)
}

func (c *Client) GetFlowConditions() (json.RawMessage, error) {
	return c.doRequest("GET", "/api/manager/flow/flowcardcondition/", nil)
}

func (c *Client) GetFlowActions() (json.RawMessage, error) {
	return c.doRequest("GET", "/api/manager/flow/flowcardaction/", nil)
}

// Zones

func (c *Client) GetZones() (json.RawMessage, error) {
	return c.doRequest("GET", "/api/manager/zones/zone/", nil)
}

// Apps

func (c *Client) GetApps() (json.RawMessage, error) {
	return c.doRequest("GET", "/api/manager/apps/app/", nil)
}

func (c *Client) GetApp(id string) (json.RawMessage, error) {
	return c.doRequest("GET", "/api/manager/apps/app/"+id, nil)
}

func (c *Client) RestartApp(id string) error {
	_, err := c.doRequest("POST", fmt.Sprintf("/api/manager/apps/app/%s/restart", id), nil)
	return err
}

// Notifications

func (c *Client) SendNotification(text string) error {
	// Use flow card action to create notification
	body := map[string]interface{}{
		"args": map[string]string{"text": text},
	}
	_, err := c.doRequest("POST", "/api/manager/flow/flowcardaction/homey:manager:notifications/homey:manager:notifications:create_notification/run", body)
	return err
}

func (c *Client) GetNotifications() (json.RawMessage, error) {
	return c.doRequest("GET", "/api/manager/notifications/notification/", nil)
}

// RunFlowCardAction runs any flow card action
func (c *Client) RunFlowCardAction(uri, id string, args map[string]interface{}) (json.RawMessage, error) {
	body := map[string]interface{}{
		"args": args,
	}
	return c.doRequest("POST", fmt.Sprintf("/api/manager/flow/flowcardaction/%s/%s/run", uri, id), body)
}

// Logic variables

func (c *Client) GetVariables() (json.RawMessage, error) {
	return c.doRequest("GET", "/api/manager/logic/variable/", nil)
}

func (c *Client) GetVariable(id string) (json.RawMessage, error) {
	return c.doRequest("GET", "/api/manager/logic/variable/"+id, nil)
}

func (c *Client) SetVariable(id string, value interface{}) error {
	body := map[string]interface{}{"value": value}
	_, err := c.doRequest("PUT", "/api/manager/logic/variable/"+id, body)
	return err
}

func (c *Client) CreateVariable(name string, varType string, value interface{}) (json.RawMessage, error) {
	body := map[string]interface{}{
		"name":  name,
		"type":  varType,
		"value": value,
	}
	return c.doRequest("POST", "/api/manager/logic/variable/", body)
}

func (c *Client) DeleteVariable(id string) error {
	_, err := c.doRequest("DELETE", "/api/manager/logic/variable/"+id, nil)
	return err
}

// System

func (c *Client) GetSystem() (json.RawMessage, error) {
	return c.doRequest("GET", "/api/manager/system/", nil)
}

func (c *Client) Reboot() error {
	_, err := c.doRequest("POST", "/api/manager/system/reboot/", nil)
	return err
}

// Users

func (c *Client) GetUsers() (json.RawMessage, error) {
	return c.doRequest("GET", "/api/manager/users/user/", nil)
}

// Insights (logs/history)

func (c *Client) GetInsights() (json.RawMessage, error) {
	return c.doRequest("GET", "/api/manager/insights/log/", nil)
}
