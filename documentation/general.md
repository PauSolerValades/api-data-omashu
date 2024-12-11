## Windows Compatibility
Several of the files of this repository need to be parsed for some kind of way, and if those (specifically _topics-and-schemas.txt_ and all of the bash scripts) were to be saved on a DOS based system, execution problems regarding the 'newline' characted in windows would be found.

To avoid it, in the _init-kafka_ Dockerfile the command ```sed -i 's/\r$//' file``` is ran in order to remove the control character ```\r``` before of the ```\n``` in the endline.

For all the files in the ```/avro-schemas```, when the contents get parsed at ```init-kafka/create_topics_with_schemas.sh```, the command ```tr -d '\r'``` is used to remove all the ```\r``` at the end of the lines.

Those are obviously more commans and computation spent in the parsing, but allows the compatibilty with native windows (more than the compatibility, it avoids accidents) and it has no effect on files written/saved in Unix based systems.

The init-kafka process has been updated to provide smooth cross-platform compatibility, especially when dealing with Windows environments where carriage return characters could previously cause execution errors in Docker containers. By incorporating `tr -d '\r'` and `sed -i 's/\r$//'` in both the script and Dockerfile, the repository is now fully functional in both Windows and Unix-like environments.

## Infrastructure Design

The infrascture design can be found in the _/design_ folders, containing the full system in the _api-data.drawio_, just the download process with the rate limiters in the _matches-download.drawio_ and the data processing in the _process-data.drawio_.


## Relationship between API-DATA and API-FRONT

The main design principle behind the two API design is to decongestionate potential calls to API-DATA making data already available duplicate in the API-FRONT. That specifically means that all the "results" data will be consumed by both API-DATA and API-FRONT and stored in both databases, DynamoDB for the api data and PostgreSQL for the FRONT. To acomplish that, DynamoDB and the FRONT will be subscribed to the same results topic as a consumers, making Kafka and the event driven arquitecture a perfect suit in that regard.

API-FRONT also iniciates the player match data obtention by writing into the _dowload-status_ topic.


