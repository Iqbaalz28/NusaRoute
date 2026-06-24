// Package orders resolves the owning customer (user id) of an order from its AWB by
// calling order-service. Notifications are produced from events that carry an AWB
// but not the recipient's user id, so this fills that gap so a notification can be
// addressed to the right user.
package orders

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Resolver struct {
	baseURL string
	client  *http.Client
}

func NewResolver(baseURL string) *Resolver {
	return &Resolver{baseURL: baseURL, client: &http.Client{Timeout: 3 * time.Second}}
}

type orderResp struct {
	Data struct {
		UserID string `json:"user_id"`
	} `json:"data"`
}

// UserIDForAWB returns the order owner's user id for the given AWB, or "" if it
// cannot be resolved (caller stores the notification anyway / logs).
func (r *Resolver) UserIDForAWB(ctx context.Context, awb string) (string, error) {
	if awb == "" {
		return "", fmt.Errorf("empty awb")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/api/v1/orders/awb/%s", r.baseURL, awb), nil)
	if err != nil {
		return "", err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("order lookup status %d", resp.StatusCode)
	}
	var out orderResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.Data.UserID, nil
}
