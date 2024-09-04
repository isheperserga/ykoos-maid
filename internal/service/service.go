package service

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"yk-dc-bot/internal/apperrors"
	"yk-dc-bot/internal/config"
	"yk-dc-bot/internal/database"
	"yk-dc-bot/internal/henrikapi"
	"yk-dc-bot/internal/logger"
	"yk-dc-bot/internal/redisclient"
	"yk-dc-bot/internal/trngg"
	"yk-dc-bot/internal/util"
)

type Service struct {
	DB          *database.Database
	Log         *logger.Logger
	RedisClient *redisclient.Client
	HenrikAPI   *henrikapi.HenrikDevAPI
	TrackerAPI  *trngg.TrackerAPI
}

func NewService(db *database.Database, log *logger.Logger, redisClient *redisclient.Client, cfg *config.Config, henrikAPI *henrikapi.HenrikDevAPI) *Service {
	return &Service{
		DB:          db,
		Log:         log,
		RedisClient: redisClient,
		HenrikAPI:   henrikAPI,
		TrackerAPI:  trngg.NewTrackerAPI(cfg, redisClient, log),
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
		appErr := apperrors.Wrap(err, "ACCOUNT_DATA_ERROR", "error fetching account data", "There was an error. Please try again later.")
		if errors.As(err, &appErr) && strings.Contains(appErr.Message, "not found") {
			appErr = apperrors.New("ACCOUNT_DATA_ERROR", "Couldn't find the account via API", "Account with this Riot ID not found")
		}
		tracker.SendError(appErr)
		return nil, appErr
	}

	time.Sleep(700 * time.Millisecond)

	tracker.SendUpdate("> alright... just some more things...")
	mmrData, err := s.HenrikAPI.GetMMRByPUUID(accountData.Region, accountData.Puuid)
	if err != nil {
		tracker.SendError(apperrors.Wrap(err, "MMR_DATA_ERROR", "error fetching rank data", "There was an error. Please try again later."))
		return nil, apperrors.Wrap(err, "MMR_DATA_ERROR", "error fetching rank data")
	}

	time.Sleep(700 * time.Millisecond)

	tracker.SendUpdate("> oh, we can't forget about their card!")
	detailedAccountData, err := s.HenrikAPI.GetDetailedAccountByPUUID(accountData.Puuid)
	if err != nil {
		tracker.SendError(apperrors.Wrap(err, "DETAILED_ACCOUNT_DATA_ERROR", "error fetching detailed account data", "There was an error. Please try again later."))
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

func (s *Service) GetPlayerTrackerData(name, tag string, tracker *util.ProgressTracker) (*trngg.PlayerData, error) {
	time.Sleep(500 * time.Millisecond)

	tracker.SendUpdate(fmt.Sprintf("> right now, i'm fetching %s#%s's tracker data", name, tag))
	playerData, err := s.TrackerAPI.GetPlayerTrackerData(name, tag)
	if err != nil {
		tracker.SendError(apperrors.Wrap(err, "TRACKER_DATA_ERROR", "error fetching tracker data", "There was an error. Please try again later."))
		return nil, apperrors.Wrap(err, "TRACKER_DATA_ERROR", "error fetching tracker data")
	}

	tracker.SendDone()
	return playerData, nil
}
