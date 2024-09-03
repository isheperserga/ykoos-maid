package service

import (
	"fmt"
	"time"
	"yk-dc-bot/internal/apperrors"
	"yk-dc-bot/internal/config"
	"yk-dc-bot/internal/database"
	"yk-dc-bot/internal/henrikapi"
	"yk-dc-bot/internal/logger"
	"yk-dc-bot/internal/redisclient"
	"yk-dc-bot/internal/util"
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

func (s *Service) GetPlayerRankData(name, tag string, tracker *util.ProgressTracker) (*RankData, error) {
	time.Sleep(500 * time.Millisecond)

	tracker.SendUpdate(fmt.Sprintf("> right now, i'm fetching %s#%s's rank data", name, tag))
	accountData, err := s.HenrikAPI.GetAccountByNameTag(name, tag)
	if err != nil {
		tracker.SendError(err)
		return nil, apperrors.Wrap(err, "ACCOUNT_DATA_ERROR", "error fetching account data")
	}

	time.Sleep(700 * time.Millisecond)

	tracker.SendUpdate("> alright... just some more things...")
	mmrData, err := s.HenrikAPI.GetMMRByPUUID(accountData.Region, accountData.Puuid)
	if err != nil {
		tracker.SendError(err)
		return nil, apperrors.Wrap(err, "MMR_DATA_ERROR", "error fetching rank data")
	}

	time.Sleep(700 * time.Millisecond)

	tracker.SendUpdate("> oh, we can't forget about their card!")
	detailedAccountData, err := s.HenrikAPI.GetDetailedAccountByPUUID(accountData.Puuid)
	if err != nil {
		tracker.SendError(err)
		return nil, apperrors.Wrap(err, "DETAILED_ACCOUNT_DATA_ERROR", "error fetching detailed account data")
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

	tracker.SendDone()
	return rankData, nil
}
