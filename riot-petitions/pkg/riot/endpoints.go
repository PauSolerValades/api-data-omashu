/**


 */

package riot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Omashu-Data/api-data/riot-petitions/pkg/core"
	"github.com/rs/zerolog/log"
)

// creates the header containing the RIOT_API_KEY to make petitions
// to RIOT APIs
func createRequest(url, method string) (*http.Request, error) {
	// Create a new request
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	// Add the API key to the header
	req.Header.Set("X-Riot-Token", core.Config.API_KEY)

	// Log specific fields of the request
	log.Info().
		Str("method", req.Method).
		Str("url", req.URL.String()).
		Str("api_key", req.Header.Get("X-Riot-Token")).
		Msg("Request created")

	return req, nil
}

// makes the petition to account
func AccountsByPuuid(downloadStart map[string]interface{}, m *RateLimitedClient) (string, error) {

	riotId := downloadStart["riotId"].(string)
	server := downloadStart["server"].(string)

	tagLineAndGameName := strings.Split(riotId, "#")
	if len(tagLineAndGameName) != 2 {
		log.Error().Interface("riotId", riotId).Msg("Invalid riotID provided.")
		return "", fmt.Errorf("invalid riotID provided %s", riotId)
	}

	url := fmt.Sprintf(string(core.BASE_RIOT_URL), core.ServerToRegion[core.Server(server)]) + string(core.ACCOUNT_V1) + fmt.Sprintf(string(core.ROUTE_BY_RIOTID), tagLineAndGameName[0], tagLineAndGameName[1])
	log.Info().Interface("url", url).Msg("Url to make petition")
	request, err := createRequest(url, "GET")
	if err != nil {
		log.Error().Interface("err", err).Msg("Error writing the Request petition")
	}
	// Log specific fields of the request
	log.Debug().
		Str("method", request.Method).
		Str("url", request.URL.String()).
		Str("api_key", request.Header.Get("X-Riot-Token")).
		Msg("Request created")

	response, err := m.Do(request, string(core.ACCOUNT_V1))
	if err != nil {
		log.Printf("Error making the request: %v", err)
		return "", err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading the response body: %v", err)
		return "", err
	}

	var respData map[string]interface{}
	err = json.Unmarshal(body, &respData)
	if err != nil {
		log.Error().Err(err).Msg("Error unmarshaling the response body")
		return "", err
	}

	log.Info().Interface("puuid", respData["puuid"]).Msg("Response from Riot")

	return respData["puuid"].(string), nil
}

func MatchesByPuuidMonthly(data map[string]any, m *RateLimitedClient, region core.Region) ([]string, error) {
	puuid := data["puuid"].(string)
	startTimestamp := int64(data["startTimestamp"].(float64))
	endTimestamp := int64(data["endTimestamp"].(float64))

	var allMatchIds []string
	start := 0
	count := 100 // Maximum allowed by the API

	for {
		params := fmt.Sprintf("?startTime=%d&endTime=%d&queue=%s&type=%s&start=%d&count=%d", startTimestamp, endTimestamp, core.QUEUE_TYPE, core.MATCH_TYPE, start, count)
		url := fmt.Sprintf(string(core.BASE_RIOT_URL), string(region)) + string(core.MATCH_V5) + fmt.Sprintf("by-puuid/%s/ids", puuid) + params

		request, err := createRequest(url, "GET")
		if err != nil {
			log.Error().Err(err).Msg("Error creating request")
			return nil, err
		}

		response, err := m.Do(request, string(core.MATCH_V5))
		if err != nil {
			log.Error().Err(err).Msg("Error making the request")
			return nil, err
		}
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			log.Error().Err(err).Msg("Error reading the response body")
			return nil, err
		}

		var matchIds []string
		err = json.Unmarshal(body, &matchIds)
		if err != nil {
			log.Error().Err(err).Msg("Error unmarshaling the response body")
			return nil, err
		}

		allMatchIds = append(allMatchIds, matchIds...)

		// If we received fewer matches than requested, we've reached the end
		if len(matchIds) < count {
			break
		}

		// Move to the next page
		start += count
	}

	return allMatchIds, nil
}

func MatchDetailsByMatchId(data map[string]any, m *RateLimitedClient, region core.Region) (map[string]any, error) {
	matchId := data["matchId"].(string)

	url := fmt.Sprintf(string(core.BASE_RIOT_URL), string(region)) + string(core.MATCH_V5) + matchId

	request, err := createRequest(url, "GET")
	if err != nil {
		log.Error().Err(err).Msg("Error creating request")
		return nil, err
	}

	response, err := m.Do(request, string(core.MATCH_V5))
	if err != nil {
		log.Error().Err(err).Msg("Error making the request")
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading the response body")
		return nil, err
	}

	var matchDetails map[string]interface{}
	err = json.Unmarshal(body, &matchDetails)
	if err != nil {
		log.Error().Err(err).Msg("Error unmarshaling the response body")
		return nil, err
	}

	log.Info().Msg("Response from Riot")

	return matchDetails, nil
}

func MatchTimelineByMatchId(data map[string]interface{}, m *RateLimitedClient, region core.Region) (map[string]interface{}, error) {
	matchId := data["matchId"].(string)

	url := fmt.Sprintf(string(core.BASE_RIOT_URL), string(region)) + string(core.MATCH_V5) + matchId + "/timeline"

	request, err := createRequest(url, "GET")
	if err != nil {
		log.Error().Err(err).Msg("Error creating request")
		return nil, err
	}

	response, err := m.Do(request, string(core.MATCH_V5))
	if err != nil {
		log.Error().Err(err).Msg("Error making the request")
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading the response body")
		return nil, err
	}

	var matchTimeline map[string]interface{}
	err = json.Unmarshal(body, &matchTimeline)
	if err != nil {
		log.Error().Err(err).Msg("Error unmarshaling the response body")
		return nil, err
	}

	log.Info().Msg("Response from Riot")

	return matchTimeline, nil
}
