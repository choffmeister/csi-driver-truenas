package truenas

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	"github.com/choffmeister/csi-driver-truenas/internal/utils"
)

type TruenasHttpClient struct {
	http          *utils.JsonHttpClient
	BaseURL       string
	ApiKey        string
	TLSSkipVerify bool
}

func NewTruenasHttpClient(baseUrl string, apiKey string, tlsSkipVerify bool) *TruenasHttpClient {
	baseUrlOpt := utils.WithRequestTransformer(func(r *http.Request) error {
		fullUrl, err := url.Parse(fmt.Sprintf("%s/api/v2.0%s", baseUrl, r.URL.String()))
		if err != nil {
			return err
		}
		r.URL = fullUrl
		return nil
	})
	apiKeyOpt := utils.WithRequestTransformer(func(r *http.Request) error {
		r.Header.Set("Authorization", "Bearer "+apiKey)
		return nil
	})
	tlsSkipVerifyOpt := utils.WithHttpConfiguration(func(c *http.Client) {
		c.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: tlsSkipVerify},
		}
	})
	client := utils.NewJsonHttpClient(tlsSkipVerifyOpt, baseUrlOpt, apiKeyOpt)

	return &TruenasHttpClient{
		http:          client,
		BaseURL:       baseUrl,
		ApiKey:        apiKey,
		TLSSkipVerify: tlsSkipVerify,
	}
}

type PoolDataset struct {
	Id       string        `json:"id"`
	Type     string        `json:"type"`
	Name     string        `json:"name"`
	Pool     string        `json:"pool"`
	Children []PoolDataset `json:"children"`
}

// https://www.truenas.com/docs/api/rest.html#api-PoolDataset-poolDatasetGet
func (c *TruenasHttpClient) PoolDatasetGet(ctx context.Context, limit int, offset int) (*[]PoolDataset, error) {
	res := []PoolDataset{}
	if err := c.http.Get(ctx, fmt.Sprintf("/pool/dataset?limit=%d&offest=%d", limit, offset), nil, &res); err != nil {
		return nil, fmt.Errorf("unable to call PoolDatasetGet: %w", err)
	}
	return &res, nil
}

// https://www.truenas.com/docs/api/rest.html#api-PoolDataset-poolDatasetIdIdGet
func (c *TruenasHttpClient) PoolDatasetIdIdGet(ctx context.Context, id string) (*PoolDataset, error) {
	res := PoolDataset{}
	if err := c.http.Get(ctx, "/pool/dataset/id/"+url.QueryEscape(id), nil, &res); err != nil {
		return nil, fmt.Errorf("unable to call PoolDatasetGet: %w", err)
	}
	return &res, nil
}

// https://www.truenas.com/docs/api/rest.html#api-PoolDataset-poolDatasetPost
func (c *TruenasHttpClient) PoolDatasetPost(ctx context.Context, name string, volsize int64) (*PoolDataset, error) {
	req := struct {
		Type    string `json:"type"`
		Name    string `json:"name"`
		Volsize int64  `json:"volsize"`
	}{
		Type:    "VOLUME",
		Name:    name,
		Volsize: volsize,
	}
	res := PoolDataset{}
	if err := c.http.Post(ctx, "/pool/dataset", &req, &res); err != nil {
		return nil, fmt.Errorf("unable to call PoolDatasetPost: %w", err)
	}
	return &res, nil
}

// https://www.truenas.com/docs/api/rest.html#api-PoolDataset-poolDatasetPut
func (c *TruenasHttpClient) PoolDatasetPutVolsize(ctx context.Context, id string, volsize int64) (*PoolDataset, error) {
	req := struct {
		Volsize int64 `json:"volsize"`
	}{
		Volsize: volsize,
	}
	res := PoolDataset{}
	if err := c.http.Put(ctx, "/pool/dataset/id/"+url.QueryEscape(id), &req, &res); err != nil {
		return nil, fmt.Errorf("unable to call PoolDatasetPut: %w", err)
	}
	return &res, nil
}

// https://www.truenas.com/docs/api/rest.html#api-PoolDataset-poolDatasetPut
func (c *TruenasHttpClient) PoolDatasetPutComments(ctx context.Context, id string, comments string) (*PoolDataset, error) {
	req := struct {
		Comments string `json:"comments"`
	}{
		Comments: comments,
	}
	res := PoolDataset{}
	if err := c.http.Put(ctx, "/pool/dataset/id/"+url.QueryEscape(id), &req, &res); err != nil {
		return nil, fmt.Errorf("unable to call PoolDatasetPut: %w", err)
	}
	return &res, nil
}

// https://www.truenas.com/docs/api/rest.html#api-PoolDataset-poolDatasetIdIdDelete
func (c *TruenasHttpClient) PoolDatasetIdIdDelete(ctx context.Context, id string, recursive bool, force bool) error {
	opts := struct {
		Recursive bool `json:"recursive"`
		Force     bool `json:"force"`
	}{
		Recursive: recursive,
		Force:     force,
	}
	var res interface{}
	if err := c.http.Delete(ctx, "/pool/dataset/id/"+url.QueryEscape(id), &opts, &res); err != nil {
		return fmt.Errorf("unable to call PoolDatasetIdIdDelete: %w", err)
	}
	return nil
}

type ISCSIExtent struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Disk        string `json:"disk"`
	InsecureTPC bool   `json:"insecure_tpc"`
}

// https://www.truenas.com/docs/api/rest.html#api-IscsiExtent-iscsiExtentGet
func (c *TruenasHttpClient) ISCSIExtentGet(ctx context.Context, limit int) (*[]ISCSIExtent, error) {
	res := []ISCSIExtent{}
	if err := c.http.Get(ctx, fmt.Sprintf("/iscsi/extent?limit=%d", limit), nil, &res); err != nil {
		return nil, fmt.Errorf("unable to call ISCSIExtentGet: %w", err)
	}
	return &res, nil
}

// https://www.truenas.com/docs/api/rest.html#api-IscsiExtent-iscsiExtentPost
func (c *TruenasHttpClient) ISCSIExtentPost(ctx context.Context, name string, disk string) (*ISCSIExtent, error) {
	req := struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Disk        string `json:"disk"`
		InsecureTPC bool   `json:"insecure_tpc"`
	}{
		Name:        name,
		Type:        "DISK",
		Disk:        disk,
		InsecureTPC: false,
	}
	res := ISCSIExtent{}
	if err := c.http.Post(ctx, "/iscsi/extent", &req, &res); err != nil {
		return nil, fmt.Errorf("unable to call ISCSIExtentPost: %w", err)
	}
	return &res, nil
}

type ISCSITargetGroup struct {
	PortalId    int `json:"portal"`
	InitiatorId int `json:"initiator"`
}
type ISCSITarget struct {
	Id     int                `json:"id"`
	Name   string             `json:"name"`
	Groups []ISCSITargetGroup `json:"groups"`
}

// https://www.truenas.com/docs/api/rest.html#api-IscsiTarget-iscsiTargetGet
func (c *TruenasHttpClient) ISCSITargetGet(ctx context.Context, limit int) (*[]ISCSITarget, error) {
	res := []ISCSITarget{}
	if err := c.http.Get(ctx, fmt.Sprintf("/iscsi/target?limit=%d", limit), nil, &res); err != nil {
		return nil, fmt.Errorf("unable to call ISCSITargetGet: %w", err)
	}
	return &res, nil
}

// https://www.truenas.com/docs/api/rest.html#api-IscsiTarget-iscsiTargetPost
func (c *TruenasHttpClient) ISCSITargetPost(ctx context.Context, name string, portalId int, initiatorId int) (*ISCSITarget, error) {
	req := struct {
		Name   string             `json:"name"`
		Groups []ISCSITargetGroup `json:"groups"`
	}{
		Name: name,
		Groups: []ISCSITargetGroup{
			{
				PortalId:    portalId,
				InitiatorId: initiatorId,
			},
		},
	}
	res := ISCSITarget{}
	if err := c.http.Post(ctx, "/iscsi/target", &req, &res); err != nil {
		return nil, fmt.Errorf("unable to call ISCSITargetPost: %w", err)
	}
	return &res, nil
}

type ISCSITargetExtend struct {
	Id     int `json:"id"`
	Target int `json:"target"`
	Extent int `json:"extent"`
	LUNId  int `json:"lunid"`
}

// https://www.truenas.com/docs/api/rest.html#api-IscsiTargetExtent-iscsiTargetExtentGet
func (c *TruenasHttpClient) ISCSITargetExtendGet(ctx context.Context, limit int) (*[]ISCSITargetExtend, error) {
	res := []ISCSITargetExtend{}
	if err := c.http.Get(ctx, fmt.Sprintf("/iscsi/targetextent?limit=%d", limit), nil, &res); err != nil {
		return nil, fmt.Errorf("unable to call ISCSITargetExtendGet: %w", err)
	}
	return &res, nil
}

// https://www.truenas.com/docs/api/rest.html#api-IscsiTargetextent-iscsiTargetextentPost
func (c *TruenasHttpClient) ISCSITargetExtendPost(ctx context.Context, targetId int, extentId int) (*ISCSITargetExtend, error) {
	req := struct {
		Target int `json:"target"`
		Extent int `json:"extent"`
	}{
		Target: targetId,
		Extent: extentId,
	}
	res := ISCSITargetExtend{}
	if err := c.http.Post(ctx, "/iscsi/targetextent", &req, &res); err != nil {
		return nil, fmt.Errorf("unable to call ISCSITargetExtendPost: %w", err)
	}
	return &res, nil
}
