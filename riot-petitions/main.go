package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Omashu-Data/api-data/riot-petitions/internal"
	"github.com/Omashu-Data/api-data/riot-petitions/pkg/core"
	"github.com/Omashu-Data/api-data/riot-petitions/pkg/kafka"
	"github.com/Omashu-Data/api-data/riot-petitions/pkg/logic"
	"github.com/Omashu-Data/api-data/riot-petitions/pkg/queues"
	"github.com/Omashu-Data/api-data/riot-petitions/pkg/redis"
	"github.com/Omashu-Data/api-data/riot-petitions/pkg/riot"

	"github.com/rs/zerolog/log"
)

func main() {

	internal.SetupLogger(os.Stderr)

	// Starts the settings struct
	core.NewSettings()

	// Initialize Redis client
	ctx := context.Background()
	rdb := redis.InitializeRedis(core.Config.REDIS_HOST, core.Config.REDIS_PORT)
	defer rdb.Close()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	log.Info().Msg("Redis setup completed")

	// Inicialize codec table
	codecTable, err := kafka.NewCodecTable(kafka.CodecTableConfig{
		SchemaRegistry: core.Config.SchemaRegistryURL + "/schemas/ids/",
		MaxConnections: 1,
		Timeout:        10 * time.Second,
	})
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Failed to create CodecTable: %v", err))
		panic("Unable to create code table")
	}
	defer codecTable.Close()

	// inicialize client to make petitions to Riot
	riotClient, err := riot.NewRateLimitedClient("application", []core.Routes{core.ACCOUNT_V1, core.MATCH_V5})
	if err != nil {
		panic(fmt.Errorf("unable to create rate limited client: %v", err))
	}
	log.Info().Msg("Rate Limiter created Successfully")

	// iniciaize client to consume from kafka topics
	consumer, err := kafka.NewKafkaConsumer(core.Config.KafkaBootstrapServers, codecTable)
	if err != nil {
		panic(fmt.Errorf("unable to create kafka consumer: %v", err))
	}
	defer consumer.Close()
	log.Info().Msg("Kafka Consumer created successfully!")

	// inicialize producer to produce download results
	producer, err := kafka.NewKafkaProducer(core.Config.KafkaBootstrapServers, core.Config.SchemaRegistryURL)
	if err != nil {
		panic(fmt.Errorf("unable to create kafka producer: %v", err))
	}
	defer producer.Close()
	log.Info().Msg("Kafka Producer created successfully!")

	// Start consuming Redis queues
	log.Info().Msg("Starting Redis queue consumers")
	for i := 0; i <= core.Config.GOROUTINES_PER_REGION; i++ {
		go queues.ConsumeQueues(core.EUROPE, rdb, ctx, riotClient, codecTable, producer)
		go queues.ConsumeQueues(core.SEA, rdb, ctx, riotClient, codecTable, producer)
		go queues.ConsumeQueues(core.ASIA, rdb, ctx, riotClient, codecTable, producer)
		go queues.ConsumeQueues(core.AMERICAS, rdb, ctx, riotClient, codecTable, producer)
	}

	// Start Kafka consumer for 'download-status'
	log.Info().Msg("Starting 'download-status' Kafka consumer")
	logic.ConsumePuuids(consumer, rdb, ctx)

	// Block main thread
	select {}
}
