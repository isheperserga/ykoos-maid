package trngg

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"yk-dc-bot/internal/apperrors"
	"yk-dc-bot/internal/config"
	"yk-dc-bot/internal/logger"
	redisclient "yk-dc-bot/internal/redisclient"

	"github.com/gospider007/ja3"
	"github.com/gospider007/requests"
	"github.com/tidwall/gjson"
)

const baseURL = "https://api.tracker.gg/api/v2/valorant/standard/profile/riot"

type TrackerAPI struct {
	redisClient *redisclient.Client
	log         *logger.Logger
	httpClient  *requests.Client
	proxies     []string // TODO: implement usage for when we're getting rate limited
}

func NewTrackerAPI(cfg *config.Config, redisClient *redisclient.Client, log *logger.Logger) *TrackerAPI {
	client, _ := requests.NewClient(context.TODO())

	return &TrackerAPI{
		redisClient: redisClient,
		log:         log,
		httpClient:  client,
	}
}

type PlayerData struct {
	User           string `json:"user,omitempty"`
	AvatarUrl      string `json:"avatarUrl,omitempty"`
	Wins           string `json:"wins,omitempty"`
	Losses         string `json:"losses,omitempty"`
	WinPct         string `json:"winPct,omitempty"`
	HsPct          string `json:"hsPct,omitempty"`
	KdRatio        string `json:"kdRatio,omitempty"`
	DamagePerRound string `json:"damagePerRound,omitempty"`
	TimePlayed     string `json:"timePlayed,omitempty"`
	Rank           string `json:"rank,omitempty"`
	RankIconUrl    string `json:"rankIconUrl,omitempty"`
	IsPrivate      bool   `json:"isPrivate,omitempty"`
}

func (t *TrackerAPI) GetPlayerTrackerData(username, tagline string) (*PlayerData, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("tracker:%s:%s", username, tagline)

	cachedData, err := t.redisClient.Get(ctx, cacheKey)
	if err == nil {
		var playerData PlayerData
		if err := json.Unmarshal([]byte(cachedData), &playerData); err == nil {
			return &playerData, nil
		}
	}

	playerData, err := t.fetchPlayerData(username, tagline)
	if err != nil {
		return nil, err
	}

	cacheData, _ := json.Marshal(playerData)
	if err := t.redisClient.Set(ctx, cacheKey, string(cacheData), 30*time.Minute); err != nil {
		t.log.Error("Failed to cache player data", "error", err)
	}

	return playerData, nil
}

func (t *TrackerAPI) fetchPlayerData(username, tagline string) (*PlayerData, error) {
	url := fmt.Sprintf("%s/%s%%23%s", baseURL, username, tagline)

	headers := map[string]string{
		"Accept":          "application/json",
		"Accept-Encoding": "gzip, deflate, br",
		"Accept-Language": "en-US,en;q=0.9",
		"Connection":      "keep-alive",
		"Host":            "api.tracker.gg",
		"Origin":          "https://tracker.gg",
		"Referer":         fmt.Sprintf("https://tracker.gg/valorant/profile/riot/%s%%23%s/overview", username, tagline),
		"Sec-Fetch-Dest":  "empty",
		"Sec-Fetch-Mode":  "cors",
		"Sec-Fetch-Site":  "cross-site",
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36 OPR/109.0.0.0",
	}

	ja3str := "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,18-13-65281-65037-35-23-27-5-43-45-16-17513-51-10-0-11-21,29-23-24,0"
	ja3Spec, _ := ja3.CreateSpecWithStr(ja3str)

	for attempts := 0; attempts < 5; attempts++ {
		resp, err := t.httpClient.Get(context.TODO(), url, requests.RequestOption{
			Headers: headers,
			Ja3:     true,
			Ja3Spec: ja3Spec,
		})

		if err != nil {
			t.log.Error("Failed to fetch player data", "error", err)
			continue
		}

		if strings.Contains(resp.Text(), "scrape our website") || strings.Contains(resp.Text(), "You are being rate lim") {
			t.log.Warn("Rate limited, retrying", "attempt", attempts+1)
			time.Sleep(time.Second)
			continue
		}

		return t.parsePlayerData(resp.Text())
	}

	return nil, apperrors.New("TRACKER_FETCH_ERROR", "Failed to fetch player data after multiple attempts")
}

func (t *TrackerAPI) parsePlayerData(jsonBody string) (*PlayerData, error) {
	data := gjson.Get(jsonBody, "data")

	if strings.Contains(jsonBody, "CollectorResultStatus::Private") {
		return &PlayerData{IsPrivate: true}, nil
	}

	segments := data.Get("segments").Array()
	if len(segments) == 0 {
		return nil, apperrors.New("TRACKER_PARSE_ERROR", "No segments found in player data")
	}

	stats := segments[0].Get("stats")

	playerData := &PlayerData{
		User:           data.Get("platformInfo.platformUserIdentifier").String(),
		AvatarUrl:      data.Get("platformInfo.avatarUrl").String(),
		Wins:           stats.Get("matchesWon.displayValue").String(),
		Losses:         stats.Get("matchesLost.displayValue").String(),
		WinPct:         stats.Get("matchesWinPct.displayValue").String(),
		HsPct:          stats.Get("headshotsPercentage.displayValue").String(),
		KdRatio:        stats.Get("kDRatio.displayValue").String(),
		DamagePerRound: stats.Get("damagePerRound.displayValue").String(),
		TimePlayed:     stats.Get("timePlayed.displayValue").String(),
		Rank:           stats.Get("rank.metadata.tierName").String(),
		RankIconUrl:    stats.Get("rank.metadata.iconUrl").String(),
		IsPrivate:      false,
	}

	return playerData, nil
}
