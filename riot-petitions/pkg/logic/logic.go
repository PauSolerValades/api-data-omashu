package logic

import (
	"context"
	"fmt"

	"github.com/Omashu-Data/api-data/riot-petitions/pkg/core"
	"github.com/Omashu-Data/api-data/riot-petitions/pkg/kafka"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func ConsumePuuids(c *kafka.KafkaConsumer, rdb *redis.Client, ctx context.Context) {
	puuidExistsChan := make(chan kafka.KafkaMessage)
	go func() {
		err := c.ConsumeMessages(core.Config.TOPIC_DOWNLOAD_START, puuidExistsChan)
		if err != nil {
			log.Error().Interface("topic", core.Config.TOPIC_DOWNLOAD_START).Msg("error creating consumer")
		}
	}()

	log.Info().Str("topic", core.Config.TOPIC_DOWNLOAD_START).Msg("Kafka consumption started!")
	for kfkMsg := range puuidExistsChan {
		log.Info().Msg("Message consumed from kafka!")

		// Decode value, which has a schemaID and is expected to be a complex structure (map)
		decodedMsgValue, err := c.DecodeValue(kfkMsg.Value)
		if err != nil {
			log.Error().Err(err).Msg("Failed to decode value from schemaID")
			continue
		}

		// convert the server into the proper region for making the data petitions
		region := core.ServerToRegion[core.Server(decodedMsgValue["server"].(string))]

		regionQueueName := fmt.Sprintf("%s-riotId", region)
		log.Debug().Str("Pushed to queue", regionQueueName)
		// we push the original encoded message to avoid another encoding
		err = rdb.LPush(ctx, regionQueueName, string(kfkMsg.Value)).Err()
		if err != nil {
			log.Error().Err(err).Msg("failed to add item to queue")
		}
	}
}
