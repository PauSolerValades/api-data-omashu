/**
This file contains all the functions to fetch schemas from the registry-schema, which implement avro format.
The code has been extracted from the following github issue https://github.com/linkedin/goavro/issues/97
which is also listed in the documentation. With some changes have been done to implement JSON reading
from the shcema registry directly.

The CodecTable is a hashmap between the id and the codec implemeting the goswarm fo have it memoized for
extremely fast access. If the schema is not found in the table, it gets fetched from the schema-registry
and lodaded with the name as the key.
*/

package kafka

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/karrick/goswarm"
	"github.com/linkedin/goavro"
)

const (
	defaultMaxConns = 5
	defaultTimeout  = time.Duration(0)
)

// ErrDownloadSchema is returned when the library cannot download the specified schema ID.
type ErrDownloadSchema struct {
	SchemaID string
	Err      error
}

func (e ErrDownloadSchema) Error() string {
	return fmt.Sprintf("cannot download schema %q: %s", e.SchemaID, e.Err)
}

// ErrSchemaRegistryNotSet is returned when the client attempts to create a CodecTable instance without specifying the schema registry.
var ErrSchemaRegistryNotSet = errors.New("cannot create CodecTable without non-empty SchemaRegistry string")

// ErrCompilingSchema is returned when the schema fails to compile.
type ErrCompilingSchema struct {
	SchemaID string
	Err      error
}

func (e ErrCompilingSchema) Error() string {
	return fmt.Sprintf("cannot compile schema %q: %s", e.SchemaID, e.Err)
}

// CodecTableConfig structure allows upstream to specify parameters required when creating a CodecTable structure.
type CodecTableConfig struct {
	SchemaRegistry string
	MaxConnections int
	Timeout        time.Duration // nanoseconds
}

// CodecTable structure provides a memoized mapping of Avro schema IDs to a compiled goavro.Codec for that schema.
type CodecTable struct {
	querier *goswarm.Simple // map from schemaID to codec
}

// NewCodecTable returns a new CodecTable instance.
func NewCodecTable(config CodecTableConfig) (*CodecTable, error) {
	var err error
	if config.SchemaRegistry == "" {
		return nil, ErrSchemaRegistryNotSet
	}
	if !strings.HasSuffix(config.SchemaRegistry, "/ids/") {
		return nil, errors.New("SchemaRegistry format is invalid. Must end with '/ids/'")
	}
	if config.MaxConnections < 0 {
		return nil, fmt.Errorf("cannot create CodecTable when MaxConnections is less than 0: %d", config.MaxConnections)
	}
	if config.MaxConnections == 0 {
		config.MaxConnections = defaultMaxConns
	}
	if config.Timeout < 0 {
		return nil, fmt.Errorf("cannot create CodecTable when Timeout is less than 0: %v", config.Timeout)
	}
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}
	s, err := goswarm.NewSimple(&goswarm.Config{Lookup: makeLookupCodec(config)})
	if err != nil {
		return nil, err
	}
	return &CodecTable{querier: s}, nil
}

// Close releases resources used by the CodecTable, and returns any error associated with releasing those resources.
func (ct *CodecTable) Close() error {
	return ct.querier.Close()
}

// Codec returns the pointer to a goavro.Codec instance from a schema ID.
func (ct *CodecTable) Codec(schemaID string) (*goavro.Codec, error) {
	codec, err := ct.querier.Query(schemaID)
	if err != nil {
		fmt.Printf("Error fetching codec for schema ID %s: %v\n", schemaID, err)
		return nil, err
	}
	fmt.Printf("Successfully fetched codec for schema ID %s\n", schemaID)
	return codec.(*goavro.Codec), nil
}

func makeLookupCodec(config CodecTableConfig) func(string) (interface{}, error) {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: int(config.MaxConnections),
		},
		Timeout: time.Duration(config.Timeout),
	}
	return func(schemaID string) (interface{}, error) {
		resp, err := client.Get(config.SchemaRegistry + url.QueryEscape(schemaID))
		if err != nil {
			return nil, ErrDownloadSchema{schemaID, err}
		}
		// got a response from this server, so commit to reading all of it
		defer func(iorc io.ReadCloser) {
			// so we can reuse connections via Keep-Alive
			io.Copy(io.Discard, iorc)
			iorc.Close()
		}(resp.Body)
		if resp.StatusCode != http.StatusOK {
			var message string
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				message = fmt.Sprintf(": cannot read response body: %s", err)
			} else if len(body) > 0 {
				message = ": " + string(body)
			}
			return nil, ErrDownloadSchema{schemaID, fmt.Errorf("schema registry status code: %d%s", resp.StatusCode, message)}
		}
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, ErrDownloadSchema{schemaID, err}
		}
		schemaStr, ok := result["schema"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to find 'schema' key in response for schema ID %s", schemaID)
		}
		fmt.Printf("Fetched schema for ID %s: %s\n", schemaID, schemaStr)
		return goavro.NewCodec(schemaStr)
	}
}
