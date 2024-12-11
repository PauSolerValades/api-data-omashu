import json

# Define file paths
input_file = "/Users/pausolervalades/Downloads/matches(1).txt"  # Change this to your actual input file


def utf16_decode(data_bytes):
    """Decode bytes to UTF-16 string."""
    return data_bytes.decode("utf-16")


def process_lines(input_file):
    """Process each line and separate the Avro types."""
    match_game_data = []
    match_to_store = []

    with open(input_file, "r") as file:
        for i, line in enumerate(file, start=1):
            try:
                record = json.loads(line.strip())
                if "data" in record:
                    print(f'{record["matchId"]} - GAME DATA FOUND')

                else:
                    # Decode POSTMATCH and BYTIME fields
                    print(f"{record['matchId']} - MATCH DATA FOUND")
                    record["POSTMATCH"] = (
                        record["POSTMATCH"].encode("latin1").decode("utf-16")
                    )

                    record["BYTIME"] = (
                        record["BYTIME"].encode("latin1").decode("utf-16")
                    )
                    match_to_store.append(record)
            except json.JSONDecodeError:
                print(f"Skipping invalid line {i}")

    return match_game_data, match_to_store


def main():
    # Process the input file
    _, match_to_store = process_lines(input_file)

    """# Save results to separate files
    with open("match_game_data.txt", "w") as game_data_file:
        for record in match_game_data:
            game_data_file.write(json.dumps(record) + "\n")
"""
    with open("match_to_store_decoded.txt", "w") as store_data_file:
        for record in match_to_store:
            store_data_file.write(json.dumps(record) + "\n")


if __name__ == "__main__":
    main()
