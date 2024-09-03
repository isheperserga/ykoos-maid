package service

import (
	"fmt"
	"yk-dc-bot/internal/config"
	"yk-dc-bot/internal/database"
	"yk-dc-bot/internal/henrikapi"
	"yk-dc-bot/internal/logger"
	"yk-dc-bot/internal/redisclient"
)

type Service struct {
	DB          *database.Database
	Log         *logger.Logger
	RedisClient *redisclient.Client
	HenrikAPI   *henrikapi.HenrikDevAPI
}

func NewService(db *database.Database, log *logger.Logger, redisClient *redisclient.Client, cfg *config.Config, henrikAPI *henrikapi.HenrikDevAPI) *Service {
	return &Service{
		DB:          db,
		Log:         log,
		RedisClient: redisClient,
		HenrikAPI:   henrikAPI,
	}
}

type RankData struct {
	AccountName string
	AccountTag  string
	Rank        string
	RR          int
	LastGameRR  int
	CardURL     string
}

func (s *Service) GetPlayerRankData(name, tag string) (*RankData, error) {
	accountData, err := s.HenrikAPI.GetAccountByNameTag(name, tag)
	if err != nil {
		s.Log.Error("Error getting account data", "error", err)
		return nil, fmt.Errorf("error fetching account data: %w", err)
	}

	mmrData, err := s.HenrikAPI.GetMMRByPUUID(accountData.Region, accountData.Puuid)
	if err != nil {
		s.Log.Error("Error getting MMR data", "error", err)
		return nil, fmt.Errorf("error fetching rank data: %w", err)
	}

	detailedAccountData, err := s.HenrikAPI.GetDetailedAccountByPUUID(accountData.Puuid)
	if err != nil {
		s.Log.Error("Error getting detailed account data", "error", err)
	}

	rankData := &RankData{
		AccountName: accountData.Name,
		AccountTag:  accountData.Tag,
		Rank:        mmrData.CurrentData.CurrentTierPatched,
		RR:          mmrData.CurrentData.RankingInTier,
		LastGameRR:  mmrData.CurrentData.MMRChangeToLastGame,
	}

	if detailedAccountData != nil {
		rankData.CardURL = fmt.Sprintf("https://media.valorant-api.com/playercards/%s/smallart.png", detailedAccountData.Card.Small)
	} else {
		rankData.CardURL = mmrData.CurrentData.Images.Large
	}

	return rankData, nil
}
