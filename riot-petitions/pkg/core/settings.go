package core

import (
	"fmt"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

type Settings struct {
	KAFKA_BROKER           string `envconfig:"KAFKA_BROKER"`
	KAFKA_PORT             int    `envconfig:"KAFKA_PORT"`
	SCHEMA_REGISTRY_HOST   string `envconfig:"SCHEMA_REGISTRY_HOST"`
	SCHEMA_REGISTRY_PORT   int    `envconfig:"SCHEMA_REGISTRY_PORT"`
	REDIS_PORT             int    `envconfig:"REDIS_PORT"`
	REDIS_HOST             string `envconfig:"REDIS_HOST"`
	API_KEY                string `envconfig:"RIOT_API_KEY"`
	TOPIC_DOWNLOAD_STATUS  string `envconfig:"TOPIC_DOWNLOAD_STATUS"`
	TOPIC_DOWNLOAD_START   string `envconfig:"TOPIC_DOWNLOAD_START"`
	TOPIC_MATCH_DATA       string `envconfig:"TOPIC_MATCH_DATA"`
	POP_SLEEP_MILLISECONDS int    `envconfig:"POP_SLEEP_MILLISECONDS"`
	GOROUTINES_PER_REGION  int    `envconfig:"GOROUTINES_PER_REGION"`
	MIN_MONTH              string `envconfig:"MIN_MONTH"`
	LOGS_TO_FILE           bool   `envconfig:"LOGS_TO_FILE"`
	MIN_MONTH_TIME         time.Time
	KafkaBootstrapServers  string
	SchemaRegistryURL      string
}

var Config Settings

type Region string
type Server string
type serverToRegion map[Server]Region
type Routes string
type Params string
type DataType string

const (
	EUROPE   Region = "europe"
	AMERICAS Region = "americas"
	ASIA     Region = "asia"
	SEA      Region = "sea"
	ESPORTS  Region = "esports"

	EUW Server = "EUW1"
	EUN Server = "EUN1"
	RU  Server = "RU"
	TR  Server = "TR1"
	NA  Server = "NA1"
	BR  Server = "BR1"
	LA1 Server = "LA1"
	LA2 Server = "LA2"
	KR  Server = "KR"
	JP  Server = "JP1"
	OC  Server = "OC1"
	PH  Server = "PH2"
	SG  Server = "SG2"
	TH  Server = "TH2"
	TW  Server = "TW2"
	VN  Server = "VN2"

	BASE_RIOT_URL   Routes = "https://%s.api.riotgames.com/"
	ACCOUNT_V1      Routes = "riot/account/v1/"
	MATCH_V5        Routes = "lol/match/v5/matches/"
	ROUTE_BY_RIOTID Routes = "accounts/by-riot-id/%s/%s"

	MATCH_TYPE Params = "ranked"
	QUEUE_TYPE Params = "420"

	POSTMATCH DataType = "POSTMATCH"
	BYTIME    DataType = "BYTIME"
)

var ServerToRegion = serverToRegion{
	EUW: EUROPE,
	EUN: EUROPE,
	RU:  EUROPE,
	TR:  EUROPE,
	NA:  AMERICAS,
	BR:  AMERICAS,
	LA1: AMERICAS,
	LA2: AMERICAS,
	KR:  ASIA,
	JP:  ASIA,
	OC:  SEA,
	PH:  SEA,
	SG:  SEA,
	TH:  SEA,
	TW:  SEA,
	VN:  SEA,
}

func timestampFromString(dateStr string) (time.Time, error) {
	// Parse the string as "YYYYMM"
	date, err := time.Parse("200601", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format: %v", err)
	}

	return date, nil
}

func NewSettings() {
	if err := envconfig.Process("myapp", &Config); err != nil {
		log.Fatal().Interface("err", err).Msg("Error processing environment variables")
		os.Exit(1)
	}
	Config.MIN_MONTH_TIME, _ = timestampFromString(Config.MIN_MONTH)
	Config.KafkaBootstrapServers = fmt.Sprintf("%s:%d", Config.KAFKA_BROKER, Config.KAFKA_PORT)
	Config.SchemaRegistryURL = fmt.Sprintf("http://%s:%d", Config.SCHEMA_REGISTRY_HOST, Config.SCHEMA_REGISTRY_PORT)
}
