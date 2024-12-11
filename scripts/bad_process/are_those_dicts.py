import json

# Define file paths
input_file = ""  # Change this to your actual input file


def utf16_decode(data_bytes):
    """Decode bytes to UTF-16 string."""
    return data_bytes.decode("utf-16")


def process_lines(input_file):
    """Process each line and separate the Avro types."""

    with open(input_file, "r") as file:
        for i, line in enumerate(file, start=1):
            record = json.loads(line.strip())
            print(record["data"])

    return


def main():
    # Process the input file
    _ = process_lines(input_file)

    # Save results to separate files
    """with open("match_game_data.txt", "w") as game_data_file:
        for record in match_game_data:
            game_data_file.write(json.dumps(record) + "\n")
    with open("match_to_store_decoded.txt", "w") as store_data_file:
        for record in match_to_store:
            store_data_file.write(json.dumps(record) + "\n")
"""


if __name__ == "__main__":
    main()
