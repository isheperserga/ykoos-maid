package henrikapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"yk-dc-bot/internal/apperrors"
	"yk-dc-bot/internal/config"
	"yk-dc-bot/internal/logger"
	redisclient "yk-dc-bot/internal/redisclient"
)

const baseURL = "https://api.henrikdev.xyz/valorant"

type HenrikDevAPI struct {
	apiKey      string
	redisClient *redisclient.Client
	log         *logger.Logger
	httpClient  *http.Client
}

func NewHenrikDevAPI(cfg *config.Config, redisClient *redisclient.Client, log *logger.Logger) *HenrikDevAPI {
	return &HenrikDevAPI{
		apiKey:      cfg.HdevApiKey,
		redisClient: redisClient,
		log:         log,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (c *HenrikDevAPI) makeRequest(endpoint string) ([]byte, error) {
	req, err := http.NewRequest("GET", baseURL+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperrors.Wrap(err, "API_REQUEST_ERROR", "error making request")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, apperrors.Wrap(err, "API_STATUS_ERROR", fmt.Sprintf("API request failed with status code %d", resp.StatusCode))
	}

	return body, nil
}

type AccountData struct {
	Puuid        string   `json:"puuid"`
	Region       string   `json:"region"`
	AccountLevel int      `json:"account_level"`
	Name         string   `json:"name"`
	Tag          string   `json:"tag"`
	Card         string   `json:"card"`
	Title        string   `json:"title"`
	Platforms    []string `json:"platforms"`
	UpdatedAt    string   `json:"updated_at"`
}

type MMRData struct {
	Name        string `json:"name"`
	Tag         string `json:"tag"`
	CurrentData struct {
		CurrentTier         int    `json:"currenttier"`
		CurrentTierPatched  string `json:"currenttierpatched"`
		RankingInTier       int    `json:"ranking_in_tier"`
		MMRChangeToLastGame int    `json:"mmr_change_to_last_game"`
		Elo                 int    `json:"elo"`
		Images              struct {
			Small        string `json:"small"`
			Large        string `json:"large"`
			TriangleDown string `json:"triangle_down"`
			TriangleUp   string `json:"triangle_up"`
		} `json:"images"`
	} `json:"current_data"`
}

type Card struct {
	Small string `json:"small"`
	Large string `json:"large"`
	Wide  string `json:"wide"`
	ID    string `json:"id"`
}

type DetailedAccountData struct {
	Puuid         string `json:"puuid"`
	Region        string `json:"region"`
	AccountLevel  int    `json:"account_level"`
	Name          string `json:"name"`
	Tag           string `json:"tag"`
	Card          Card   `json:"card"`
	LastUpdate    string `json:"last_update"`
	LastUpdateRaw int64  `json:"last_update_raw"`
}

func (c *Card) UnmarshalJSON(data []byte) error {
	var cardString string
	if err := json.Unmarshal(data, &cardString); err == nil {
		c.Small = cardString
		c.Large = cardString
		c.Wide = cardString
		c.ID = cardString
		return nil
	}

	var cardObject struct {
		Small string `json:"small"`
		Large string `json:"large"`
		Wide  string `json:"wide"`
		ID    string `json:"id"`
	}
	if err := json.Unmarshal(data, &cardObject); err != nil {
		return err
	}
	c.Small = cardObject.Small
	c.Large = cardObject.Large
	c.Wide = cardObject.Wide
	c.ID = cardObject.ID
	return nil
}

func (c *HenrikDevAPI) GetAccountByNameTag(name, tag string) (*AccountData, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("account:%s:%s", name, tag)

	cachedData, err := c.redisClient.Get(ctx, cacheKey)
	if err == nil {
		var accountData AccountData
		if err := json.Unmarshal([]byte(cachedData), &accountData); err == nil {
			return &accountData, nil
		}
	}

	endpoint := fmt.Sprintf("/v2/account/%s/%s", name, tag)
	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var response struct {
		Status int         `json:"status"`
		Data   AccountData `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, apperrors.Wrap(err, "ACCOUNT_FETCH_ERROR", "Failed to fetch account data")
	}

	if response.Status != 200 {
		return nil, apperrors.Wrap(err, "ACCOUNT_FETCH_ERROR", fmt.Sprintf("Failed to fetch account data, status code: %d", response.Status))
	}

	cacheData, _ := json.Marshal(response.Data)
	c.redisClient.Set(ctx, cacheKey, string(cacheData), 12*time.Hour)

	return &response.Data, nil
}

func (c *HenrikDevAPI) GetMMRByPUUID(region, puuid string) (*MMRData, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("mmr:%s:%s", region, puuid)

	cachedData, err := c.redisClient.Get(ctx, cacheKey)
	if err == nil {
		var mmrData MMRData
		if err := json.Unmarshal([]byte(cachedData), &mmrData); err == nil {
			return &mmrData, nil
		}
	}

	endpoint := fmt.Sprintf("/v2/by-puuid/mmr/%s/%s", region, puuid)
	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var response struct {
		Status int     `json:"status"`
		Data   MMRData `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, apperrors.Wrap(err, "MMR_FETCH_ERROR", "Failed to fetch MMR data")
	}

	if response.Status != 200 {
		return nil, apperrors.Wrap(err, "MMR_FETCH_ERROR", "Failed to fetch MMR data")
	}

	cacheData, _ := json.Marshal(response.Data)
	if err := c.redisClient.Set(ctx, cacheKey, string(cacheData), 1*time.Minute); err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func (c *HenrikDevAPI) GetDetailedAccountByPUUID(puuid string) (*DetailedAccountData, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("detailed_account:%s", puuid)

	cachedData, err := c.redisClient.Get(ctx, cacheKey)
	if err == nil {
		var detailedAccountData DetailedAccountData
		if err := json.Unmarshal([]byte(cachedData), &detailedAccountData); err == nil {
			return &detailedAccountData, nil
		}
	}

	endpoint := fmt.Sprintf("/v2/by-puuid/account/%s", puuid)
	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var response struct {
		Status int                 `json:"status"`
		Data   DetailedAccountData `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, apperrors.Wrap(err, "DETAILED_ACCOUNT_FETCH_ERROR", "Failed to fetch detailed account data")
	}

	if response.Status != 200 {
		return nil, apperrors.Wrap(err, "DETAILED_ACCOUNT_FETCH_ERROR", "Failed to fetch detailed account data")
	}

	cacheData, _ := json.Marshal(response.Data)
	if err := c.redisClient.Set(ctx, cacheKey, string(cacheData), 4*time.Hour); err != nil {
		return nil, err
	}

	return &response.Data, nil
}
