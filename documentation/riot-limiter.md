# Riot Limiter

Riot limiter is a project in Go lang which aims to extract all the data of a given player of LoL using the RIOT's Developer API. It comunicates between Apache Kafka and Apache Flink to move the data arround, and uses Redis as a queue manager to store missing petitions to adapt to the rate limiter.

All the topics have an *avro* encoding (same name as the topic) which gets managed with the Confluentic's *schemaRegistry*, therefore NONE of the schemas must never be directly passed as a file.

## download-start and download-status
riot-limiter consumes from the *download-start* topics, where the webpage API has to produce a message to start the download process. This data must be in this format:

``` json
{
    "riotId": "MarkusMonkey#EUW", 
    "server": "EUW1", // Enum of all riot servers
    "puuid": null, 
    "omashuId": "0515a2f1-d138-4254-8bef-bcfbd48482ea"
}
```

where:
- riotId: parameter to find if it exists
- server: enum encoded which can take the following values per region:
  - EUROPE: EUW1, EUN1, RU, TR1, 
  - ASIA: KR, JP1
  - AMERICAS: NA1, BR1, LA1, LA2
  - SEA: OC1, PH2, SG2, TH2, TW2, VN2 
- puuid: if it's the first time a user data is downloaded, this parameter will came in as _null_. If it's not, it will have the value of the user and the system will skip the ACCOUNT-V1 petition to obtaint that data. TODO: implement the null.
- omashuId: unique identifier for all the players in the web, and also a unique identifier of the download.

After the reception of the message, the petition will be made to RIOT, and regardless of the status of the answer, a message is going to be produced to the *download-status* topic.

**download-status**
This topic serves as an status updater, where all the services that consume from it will be able to track the status of the download of a given player. To achive this behaviour, *riot-limiter* will produce to the *download-status* this messages:

``` json

"0515a2f1-d138-4254-8bef-bcfbd48482ea":
{
    "riotId": "MarkusMonkey#EUW", 
    "puuid": ,
    "status": "PROCESSING", //enum with diferent states
    "timestamp": 1719913045000, //when the petition response was processed
    "failureReason": null, //enum with several failure reasons
    "server": "EUW1"
}
```

This topic has two peculiarites:
- Key publication: the message is published with a key, which forces messages with the same key to go to the same predefined kafka partition.
- Clean-up policy: the clean up policy is aggressive, destroying whatever data **with the same key** that gets produced.

Chosing the unique identifier ```"omashuId"``` as key, allows the system to keep track of the status of a single user, as well to block user attemps to start another download while another is in progres.

The status enums are:
- PROCESSING: data has stated to download and it's sent to flink. Data will start to arrive in streaming.
- COMPLETED: all data to be downloaded has been downloaded. Processed data and statistics will eventually stop arriving.
- FAILED: something has gone wrong. When the status is failed, the failureReason enum can have one of the following values:
    - NOT_FOUND: Data was not found, ie this riotId does not exists.
    - RETRIES_EXCEEDED: The petition has exceeded all the retries.
    - NETWORK_ERROR: The petition has not been able to complete
    - UNKNOWN_ERROR: Something unexpected has happened.

## Middelware and Rate Limiter
All the petitions to RIOT pass through a middelware, which centralizes all the logic and petitions of the system. That middelware manages the _method_ and _application_ token exhaustion and it's associated waittimes with mutexes, which blocks which stops the goroutines that were tasked to receive all the data.

Backoff and retry strategis are implemented with native methods of the ```go-retryablehttp``` library, but those exclude the 429, due to being managed directly with the waitTimes provided by the RiotServer.

## Redis Queues

The tasks to execute by the middleware are stored in five Redis queues. Let's describe the data and workflow of each queue and then the algorithm:

### Region-riotId

This queue contains all the riotId for every player we want to verify. The petition is made to this following ACCOUNT-V1 endpoint: /riot/account/v1/accounts/by-riot-id/{gameName}/{tagLine}

Once the message from *download-start* is consumed, it's directly uploaded with it's encoded form to the Redis queue. Therefore, it's structure is the same in the AVRO schema we've already commented on. Then, when poped from the queue, the petition is done and puuid is obtained.

### Region-matchesId

When the puuid is obtained, we can obtain all the matchesIds of the player in the MATCH-V5 endpoint: /lol/match/v5/matches/by-puuid/{puuid}/ids

To obtain this data, it's not done arbirarily, but in a montly basis. The variable ```MIN_MONTH``` specifies which is the last month to download (format: YYYYMM) from now. For every month available, the following json structure gets encoded and pushed to the queue _region-matchesIds_:

``` json
{
    "puuid": ,
    "omashuId": "0515a2f1-d138-4254-8bef-bcfbd48482ea",
    "startTimestamp": 17725829283,
    "endTimestamp": 17725802164,
}
```

In that way, we obtain the player matches in a monthly basis automatically without the need for filtering.

### Region-matchData

The results of the latter petition is a list with all of the matchesIds. Those need to be downloaded individually twice: one for bytime data and another for postmatch, which are the following two endpoints of MATCH-V5: /lol/match/v5/matches/{matchId}/timeline and /lol/match/v5/matches/{matchId} respectively

So, to be able to achive both of the petitions to match, this towo json gets pushed to the _region-MatchData_ queue per matchId:

``` json
{
    "puuid": ,
    "omashuId": "0515a2f1-d138-4254-8bef-bcfbd48482ea",
    "matchId": "EUW_7719285618",
    "dataType": "BYTIME/POSTMATCH"
}
```

meaning that if the player has $n$ matchesIds, it ends up with $2n$ tasks in the Region-matchData queue.

When the tasks gets poped and the results received, those are produced to the _match-data_ topic with the following schema:

```

```



### Schema Encoding and Decoding
Here is the workflow of consuming a message from the Go file:
1. Inits a CodecTable (riot-petitions/kafka/avro_codec_table.go), which essentially a map which connects the name of the schema and it's schemaID. If the CodecTable does not have the ID in memory, asks for it to the schema-registry and gets stored into memory, and if it have just retrieves it.
2. When a message is Consumed, calls the decodeMessage function in *consumer_producer.go*, which extracts the schemaID (bytes 1 to 5) and looks it up in the CodecTable with the *CodecTable.Codec* method, which executes the behaviour descibed in point 1.
3. If 2 executes successfully, the rest of the message gets decoded with the *goavro* library as normally. If not, handles the error gracefully.


### Go Get problems:
- Set-up tu use SSH and HTTPS with personal tokens: https://gist.github.com/StevenACoffman/866b06ed943394fbacb60a45db5982f2

How to use SSH to go get:
1. Create an SSH key to have access to your repositories.
2. Execute the following commands
``` bash
git config --global url.git@github.com:.insteadOf https://github.com/
export GOPRIVATE=github.com/UserOrOrganitzation 
export GIT_SSH_COMMAND="ssh -i ~/.ssh/id_ed25519" #tell Go to look for your ssh key
```
