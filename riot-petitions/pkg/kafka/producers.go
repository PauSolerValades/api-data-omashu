package kafka

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/linkedin/goavro"

	"github.com/rs/zerolog/log"
)

type KafkaProducer struct {
	Producer             *kafka.Producer
	SchemaRegistryURL    string
	Codecs               map[string]*goavro.Codec // TODO: probably we should define (and streamiline) the CodecTable
	SchemaRegistryClient schemaregistry.Client
}

func NewKafkaProducer(bootstrapServers, schemaRegistryURL string) (*KafkaProducer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	config := schemaregistry.NewConfig(schemaRegistryURL)
	client, err := schemaregistry.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema registry client: %w", err)
	}

	return &KafkaProducer{
		Producer:             producer,
		SchemaRegistryURL:    schemaRegistryURL,
		Codecs:               make(map[string]*goavro.Codec),
		SchemaRegistryClient: client,
	}, nil
}

func (kp *KafkaProducer) getSchema(topic string) (string, error) {
	url := fmt.Sprintf("%s/subjects/%s-value/versions/latest", kp.SchemaRegistryURL, topic)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch schema for topic %s: %w", topic, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read schema response for topic %s: %w", topic, err)
	}

	var schemaResp struct {
		Schema string `json:"schema"`
	}
	if err := json.Unmarshal(body, &schemaResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal schema response for topic %s: %w", topic, err)
	}

	if schemaResp.Schema == "" {
		return "", fmt.Errorf("empty schema received for topic %s", topic)
	}

	return schemaResp.Schema, nil
}

func (kp *KafkaProducer) getCodec(topic string) (*goavro.Codec, error) {
	if codec, ok := kp.Codecs[topic]; ok {
		return codec, nil
	}

	schema, err := kp.getSchema(topic)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema for topic %s: %w", topic, err)
	}

	codec, err := goavro.NewCodec(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to create codec for topic %s: %w", topic, err)
	}

	kp.Codecs[topic] = codec
	return codec, nil
}

func (kp *KafkaProducer) ProduceMessage(topic string, key any, value map[string]any) error {

	// fetches if has been used before or creates it
	codec, err := kp.getCodec(topic)
	if err != nil {
		return fmt.Errorf("failed to get codec for topic %s: %w", topic, err)
	}

	avroData, err := codec.BinaryFromNative(nil, value)
	if err != nil {
		return fmt.Errorf("failed to serialize message for topic %s: %w", topic, err)
	}

	// Get the schema ID from the Schema Registry
	schemaID, err := kp.getSchemaID(topic)
	if err != nil {
		return fmt.Errorf("failed to get schema ID: %w", err)
	}

	// write the message in bytes
	var msg bytes.Buffer
	msg.WriteByte(0x00)                                    // Magic byte
	binary.Write(&msg, binary.BigEndian, uint32(schemaID)) // Schema ID as 4-byte integer
	msg.Write(avroData)                                    // Avro serialized data

	// Create
	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          msg.Bytes(),
	}

	if key != nil {
		switch key := key.(type) {
		case string:
			kafkaMsg.Key = []byte(key)
		case int:
			kafkaMsg.Key = []byte(strconv.Itoa(key))
		default:
			panic("not implemented for complex types yet")
		}
	}

	err = kp.Producer.Produce(kafkaMsg, nil)
	log.Info().
		Str("topic", topic).
		Msg("Produced message")
	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	return nil
}

func (kp *KafkaProducer) Close() {
	kp.Producer.Close()
}

func (kp *KafkaProducer) getSchemaID(topic string) (int, error) {
	// The subject in Schema Registry usually follows the pattern <topic>-value or <topic>-key
	subject := fmt.Sprintf("%s-value", topic)

	// Fetch the latest schema version metadata from the Schema Registry
	schemaMetadata, err := kp.SchemaRegistryClient.GetLatestSchemaMetadata(subject)
	if err != nil {
		return 0, fmt.Errorf("failed to get schema metadata for subject %s: %w", subject, err)
	}

	// Return the schema ID from the metadata
	return schemaMetadata.ID, nil
}
