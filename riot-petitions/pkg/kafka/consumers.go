package kafka

import (
	"encoding/binary"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/zerolog/log"
)

type KafkaConsumer struct {
	Consumer   *kafka.Consumer
	CodecTable *CodecTable
}

type KafkaMessage struct {
	Key   []byte
	Value []byte
}

func NewKafkaConsumer(bootstrapServers string, codecTable *CodecTable) (*KafkaConsumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"group.id":          "consumerStatus",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	return &KafkaConsumer{
		Consumer:   consumer,
		CodecTable: codecTable,
	}, nil
}

func (kc *KafkaConsumer) ConsumeMessages(topic string, messageChan chan<- KafkaMessage) error {
	err := kc.Consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		return fmt.Errorf("error subscribing to topic: %w", err)
	}

	log.Info().Interface("topic", topic).Msg("Consumer created successfully!")
	go func() {
		for {
			msg, err := kc.Consumer.ReadMessage(-1)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				continue
			}

			messageChan <- KafkaMessage{Key: msg.Key, Value: msg.Value}
		}
	}()

	return nil
}

func (kc *KafkaConsumer) DecodeMessage(payload []byte) (interface{}, error) {
	return decodeMessage(payload, kc.CodecTable)
}

func (kc *KafkaConsumer) Close() {
	kc.Consumer.Close()
}

func decodeMessage(data []byte, codecTable *CodecTable) (interface{}, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("message too short")
	}

	// Extract schema ID (4 bytes after the magic byte)
	schemaID := binary.BigEndian.Uint32(data[1:5])

	// Get codec for this schema ID
	codec, err := codecTable.Codec(fmt.Sprintf("%d", schemaID))
	if err != nil {
		return nil, fmt.Errorf("failed to get codec: %w", err)
	}

	// Decode the actual Avro data (skipping the first 5 bytes)
	native, _, err := codec.NativeFromBinary(data[5:])
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %w", err)
	}

	// Check if native is a map or primitive
	if result, ok := native.(map[string]interface{}); ok {
		return result, nil
	}

	// Return the primitive type
	return native, nil
}

func decodeKey(data []byte, codecTable *CodecTable) (interface{}, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("key too short")
	}

	// Extract schema ID (4 bytes after the magic byte)
	schemaID := binary.BigEndian.Uint32(data[1:5])

	// Get codec for this schema ID
	codec, err := codecTable.Codec(fmt.Sprintf("%d", schemaID))
	if err != nil {
		return nil, fmt.Errorf("failed to get codec: %w", err)
	}

	// Decode the key data (skipping the first 5 bytes)
	native, _, err := codec.NativeFromBinary(data[5:])
	if err != nil {
		return nil, fmt.Errorf("failed to decode key: %w", err)
	}

	// Handle primitive types like string, int, etc.
	switch v := native.(type) {
	case string:
		return v, nil
	case int:
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported key type: %T", native)
	}
}

func (kc *KafkaConsumer) DecodeKey(payload []byte) (interface{}, error) {
	return decodeKey(payload, kc.CodecTable)
}

func DecodeValue(data []byte, codecTable *CodecTable) (map[string]interface{}, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("value too short")
	}

	// Extract schema ID (4 bytes after the magic byte)
	schemaID := binary.BigEndian.Uint32(data[1:5])

	// Get codec for this schema ID
	codec, err := codecTable.Codec(fmt.Sprintf("%d", schemaID))
	if err != nil {
		return nil, fmt.Errorf("failed to get codec: %w", err)
	}

	// Decode the actual Avro data (skipping the first 5 bytes)
	native, _, err := codec.NativeFromBinary(data[5:])
	if err != nil {
		return nil, fmt.Errorf("failed to decode value: %w", err)
	}

	result, ok := native.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("decoded value is not a map")
	}

	return result, nil
}

func (kc *KafkaConsumer) DecodeValue(payload []byte) (map[string]interface{}, error) {
	return DecodeValue(payload, kc.CodecTable)
}
