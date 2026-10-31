package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/gateway"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
)

const maximumResponseBytes = 1 << 20

type PartnerEntitlementClient struct {
	baseURL string
	client  *http.Client
}

var _ gateway.PartnerEntitlementGateway = (*PartnerEntitlementClient)(nil)

func NewPartnerEntitlementClient(baseURL string, opts ...Option) *PartnerEntitlementClient {
	client := newOptions(opts).httpClient
	return &PartnerEntitlementClient{baseURL: strings.TrimRight(baseURL, "/"), client: client}
}

type upsertEntitlementRequest struct {
	Products string        `json:"products"`
	Offer    *offerRequest `json:"offer,omitempty"`
}

type offerRequest struct {
	FreeTrial string `json:"free_trial"`
}

type entitlementResponse struct {
	Provider      string        `json:"provider"`
	SubjectID     string        `json:"subject_id"`
	EntitlementID string        `json:"entitlement_id"`
	Products      string        `json:"products"`
	Offer         offerResponse `json:"offer"`
	Status        string        `json:"status"`
	TerminatedAt  *time.Time    `json:"terminated_at"`
	LastUpdatedAt time.Time     `json:"last_updated_at"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type offerResponse struct {
	FreeTrial string `json:"free_trial"`
}

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *PartnerEntitlementClient) Upsert(ctx context.Context, in gateway.PartnerEntitlementUpsert) (domain.PartnerEntitlement, error) {
	request := upsertEntitlementRequest{Products: in.Change.Products}
	if in.Change.FreeTrial != "" {
		request.Offer = &offerRequest{FreeTrial: string(in.Change.FreeTrial)}
	}
	var response entitlementResponse
	if err := c.do(ctx, http.MethodPut, entitlementPath(in.Change.SubjectID, in.Change.EntitlementID), in.Token, request, &response); err != nil {
		return domain.PartnerEntitlement{}, err
	}
	return toPartnerEntitlement(response), nil
}

func (c *PartnerEntitlementClient) Terminate(ctx context.Context, in gateway.PartnerEntitlementTerminate) (domain.PartnerEntitlement, error) {
	var response entitlementResponse
	if err := c.do(ctx, http.MethodPost, entitlementPath(in.SubjectID, in.EntitlementID)+"/terminate", in.Token, nil, &response); err != nil {
		return domain.PartnerEntitlement{}, err
	}
	return toPartnerEntitlement(response), nil
}

func (c *PartnerEntitlementClient) do(ctx context.Context, method, path, token string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return domain.Internal(err)
		}
		reader = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return domain.Internal(err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Accept", "application/json")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := c.client.Do(request)
	if err != nil {
		return domain.PartnerUnavailable(err)
	}
	defer func() { _ = response.Body.Close() }()
	payload, err := io.ReadAll(io.LimitReader(response.Body, maximumResponseBytes))
	if err != nil {
		return domain.PartnerUnavailable(err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var envelope errorResponse
		_ = json.Unmarshal(payload, &envelope)
		return domain.PartnerRequestRejected(response.StatusCode, envelope.Error.Code, envelope.Error.Message)
	}
	if err := json.Unmarshal(payload, out); err != nil {
		return domain.PartnerUnavailable(fmt.Errorf("failed to decode partner_app response: %w", err))
	}
	return nil
}

func entitlementPath(subjectID, entitlementID string) string {
	return "/subjects/" + url.PathEscape(subjectID) + "/entitlements/" + url.PathEscape(entitlementID)
}

func toPartnerEntitlement(response entitlementResponse) domain.PartnerEntitlement {
	return domain.PartnerEntitlement{
		Provider:      response.Provider,
		SubjectID:     response.SubjectID,
		EntitlementID: response.EntitlementID,
		Products:      response.Products,
		FreeTrial:     domain.FreeTrialStatus(response.Offer.FreeTrial),
		Status:        domain.EntitlementStatus(response.Status),
		TerminatedAt:  response.TerminatedAt,
		LastUpdatedAt: response.LastUpdatedAt,
		CreatedAt:     response.CreatedAt,
		UpdatedAt:     response.UpdatedAt,
	}
}
