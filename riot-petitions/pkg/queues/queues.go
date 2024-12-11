package queues

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Omashu-Data/api-data/riot-petitions/pkg/core"
	"github.com/Omashu-Data/api-data/riot-petitions/pkg/kafka"
	"github.com/Omashu-Data/api-data/riot-petitions/pkg/riot"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func ConsumeQueues(region core.Region, rdb *redis.Client, ctx context.Context, m *riot.RateLimitedClient, codecTable *kafka.CodecTable, p *kafka.KafkaProducer) {
	// the order of the list is the prioriy. This function can be updated to a RoundRobin style
	// or depending whch queue has the most
	queues := []string{
		fmt.Sprintf("%s-riotId", region),
		fmt.Sprintf("%s-matchIds", region),
		fmt.Sprintf("%s-match", region),
	}

	i := 0
	for {
		queue := queues[i%len(queues)] // cycle the queues!
		log.Debug().Str("queue name", queue).Msg("Checking")
		taskData, err := rdb.LPop(ctx, queue).Result()

		if err == redis.Nil { //check next queue
			log.Debug().Str("queue", queue).Msg("Empty queue. Going to the next one")
			i++
			if i%len(queues) == 0 {
				// Sleep when all queues have been checked and found empty
				sleepDuration := time.Duration(core.Config.POP_SLEEP_MILLISECONDS) * time.Millisecond
				log.Info().Dur("sleep_duration", sleepDuration).Msg("All queues empty. Sleeping for")
				time.Sleep(sleepDuration)
			}
			continue // Continue to the next iteration of the loop
		}

		if err != nil {
			log.Error().Err(err).Msg("Error occurred when popping from Redis")
			i++
			continue // log error and try next queue if an error happens
		}

		log.Info().Str("taskData", taskData).Msg("Data Obtained from Redis")
		log.Info().Str("queue", queue).Msg("Data found, processing")

		switch queue {
		case queues[0]:

			petitionData, err := kafka.DecodeValue([]byte(taskData), codecTable)
			if err != nil {
				log.Error().Err(err).Msg("error when decoding Redis obtained message")
				continue
			}

			puuid, ok := petitionData["puuid"].(string) // Check if "puuid" is a string
			if !ok || puuid == "" {
				log.Info().Msg("First time user")

				// Make petition to see if the riotId is valid
				puuid, err = riot.AccountsByPuuid(petitionData, m)
				if err != nil {
					log.Error().Err(err).Msg("Error obtaining puuid from riot.")
					continue
				}
			} else {
				// puuid exists, continue execution
				log.Info().Msgf("Updating user: %s", puuid)
			}

			// produce the result to kafka
			downloadStatus := createStatusObject(petitionData, puuid)
			err = p.ProduceMessage(core.Config.TOPIC_DOWNLOAD_STATUS, petitionData["omashuId"], downloadStatus)
			if err != nil {
				log.Error().Interface("topic", core.Config.TOPIC_DOWNLOAD_STATUS).Err(err).Msg("error updating the status")
				continue
			}

			// push all the month data per player
			err = pushMatchesIdPetitionsPerMonth(puuid, petitionData["omashuId"].(string), ctx, rdb, queues[1])
			if err != nil {
				log.Error().Err(err).Msg("Error when obtaining the matchesId per months.")
				continue
			}

		case queues[1]:
			log.Debug().Msg("Task is a matchId")

			var petition map[string]any
			err = json.Unmarshal([]byte(taskData), &petition)
			if err != nil {
				log.Error().Err(err).Msg("failed to decode JSON data")
				continue
			}

			puuid := petition["puuid"].(string)
			startTimestamp := int64(petition["startTimestamp"].(float64))
			endTimestamp := int64(petition["endTimestamp"].(float64))
			omashuId := petition["omashuId"].(string)

			log.Info().
				Str("puuid", puuid).
				Int64("start_timestamp", startTimestamp).
				Int64("end_timestamp", endTimestamp).
				Str("omashu_id", omashuId).
				Msg("Getting matches from start to end for puuid associated with omashuId")

			matchesId, err := riot.MatchesByPuuidMonthly(petition, m, region)
			if err != nil {
				log.Error().Err(err).Msg("error getting the matchesIds")
				continue
			}
			omashuID := petition["omashuId"].(string)
			puuid = petition["puuid"].(string)
			pushMatchIdToRetrieve(omashuID, puuid, queues[2], matchesId, ctx, rdb)

		case queues[2]:
			log.Debug().Msg("task is to download a match.")

			var petition map[string]any
			err = json.Unmarshal([]byte(taskData), &petition)
			if err != nil {
				log.Error().Err(err).Msg("failed to decode JSON data")
				continue
			}

			matchId, ok := petition["matchId"].(string)
			if !ok {
				log.Error().Msg("Failed to extract matchId from petition")
				continue
			}

			dataType, ok := petition["dataType"].(string)
			if !ok {
				log.Error().Msg("Failed to extract dataType from petition")
				continue
			}

			log.Info().
				Str("match_id", matchId).
				Str("data_type", dataType).
				Msg("Downloading match data")

			getDataFunc := func(q string) (map[string]any, error) {
				if q == string(core.POSTMATCH) {
					return riot.MatchDetailsByMatchId(petition, m, region)
				}
				return riot.MatchTimelineByMatchId(petition, m, region)
			}

			matchData, err := getDataFunc(petition["dataType"].(string))
			if err != nil {
				log.Error().Err(err).Msg("error getting match data")
				continue
			}

			// Convert matchData to JSON string
			matchDataJSON, err := json.Marshal(matchData)
			if err != nil {
				log.Error().Err(err).Msg("error marshalling match data to JSON")
				continue
			}

			// add new data and reuse the petition
			petition["obtainedAt"] = time.Now().Unix()
			petition["data"] = string(matchDataJSON)
			err = p.ProduceMessage(core.Config.TOPIC_MATCH_DATA, "", petition)
			if err != nil {
				log.Error().Interface("topic", core.Config.TOPIC_MATCH_DATA).Err(err).Msg("error publishing the results")
				continue
			}

			if core.Config.LOGS_TO_FILE {
				writePetitionToFile(petition)
			}

		default:
			log.Error().Msg("Non recognized queue to pull from...")
		}

		i = 0 // check the first queue again
	}
}
