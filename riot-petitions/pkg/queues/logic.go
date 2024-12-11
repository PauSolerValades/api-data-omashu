package queues

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/rs/zerolog/log"

	"github.com/Omashu-Data/api-data/riot-petitions/pkg/core"
)

func createStatusObject(downloadStart map[string]any, puuid string) map[string]any {
	riotID := downloadStart["riotId"].(string)
	omashuID := downloadStart["omashuId"].(string)
	server := downloadStart["server"].(string)

	status := "PROCESSING"
	if puuid == "" {
		status = "FAILED"
	}

	timestamp := time.Now().Unix()
	return map[string]any{
		"riotId":    riotID,
		"server":    server,
		"omashuId":  omashuID,
		"puuid":     map[string]any{"string": puuid}, // format union type with null
		"status":    status,
		"timestamp": timestamp,
	}

}

type MonthTimestamps struct {
	StartTimestamp int64
	EndTimestamp   int64
}

func generateMonthlyTimestamps(minTime time.Time) []MonthTimestamps {
	var timestamps []MonthTimestamps

	// Start from the current time
	currentTime := time.Now().UTC()

	for currentTime.After(minTime) {
		// Calculate the start of the current month
		startOfMonth := time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, time.UTC)

		// Calculate the end of the current month
		endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

		// If the end of the month is after minTime, add the timestamps
		if endOfMonth.After(minTime) {
			timestamps = append(timestamps, MonthTimestamps{
				StartTimestamp: startOfMonth.Unix(),
				EndTimestamp:   endOfMonth.Unix(),
			})
		}

		// Move to the previous month
		currentTime = startOfMonth.AddDate(0, -1, 0)
	}

	// Reverse the slice so it's in chronological order
	for i := 0; i < len(timestamps)/2; i++ {
		j := len(timestamps) - 1 - i
		timestamps[i], timestamps[j] = timestamps[j], timestamps[i]
	}

	log.Info().
		Int("number_of_months", len(timestamps)).
		Time("first_month_start", time.Unix(timestamps[0].StartTimestamp, 0)).
		Time("last_month_end", time.Unix(timestamps[len(timestamps)-1].EndTimestamp, 0)).
		Msg("Generated monthly timestamps")

	return timestamps
}

func pushMatchesIdPetitionsPerMonth(puuid, omashuID string, ctx context.Context, rdb *redis.Client, queueName string) error {
	monthlyTimestamps := generateMonthlyTimestamps(core.Config.MIN_MONTH_TIME)

	for _, timestamps := range monthlyTimestamps {
		item := map[string]any{
			"puuid":          puuid,
			"omashuId":       omashuID,
			"startTimestamp": timestamps.StartTimestamp,
			"endTimestamp":   timestamps.EndTimestamp,
		}

		jsonData, err := json.Marshal(item)
		if err != nil {
			log.Error().Err(err).Msg("failed to encode petition to JSON")
			return err
		}

		err = rdb.LPush(ctx, queueName, jsonData).Err()
		if err != nil {
			log.Error().Err(err).Msg("failed to add item to queue")
			return err
		}
	}

	return nil
}

func pushMatchIdToRetrieve(omashuID, puuid, queue string, matchesId []string, ctx context.Context, rdb *redis.Client) error {
	for _, matchId := range matchesId {
		for _, dataType := range []core.DataType{core.POSTMATCH, core.BYTIME} {

			// TODO: CHECK WITH MATCHESID WE ALREADY HAVE ARE IN THE DATABASE

			item := map[string]any{
				"omashuId": omashuID,
				"puuid":    puuid,
				"matchId":  matchId,
				"dataType": dataType,
			}

			jsonData, err := json.Marshal(item)
			if err != nil {
				log.Error().Err(err).Msg("failed to encode petition to JSON")
				return err
			}

			err = rdb.LPush(ctx, queue, jsonData).Err()
			if err != nil {
				log.Error().Err(err).Str("queue", queue).Msg("failed to add item to")
				return err

			}
		}
	}

	return nil
}

func writePetitionToFile(petition map[string]any) error {
	// Convert petition to JSON
	jsonData, err := json.Marshal(petition)
	if err != nil {
		return fmt.Errorf("failed to marshal petition to JSON: %w", err)
	}

	// Open the file in append mode (create if not exists)
	file, err := os.OpenFile("/app/logs/petitions-responses.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Append a newline for readability between entries
	jsonData = append(jsonData, '\n')

	// Write JSON data to file
	_, err = file.Write(jsonData)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}
