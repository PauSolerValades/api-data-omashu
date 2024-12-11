import os

def write_lines_as_files(input_dir, output_dir):
    # Ensure the output directory exists
    if not os.path.exists(output_dir):
        os.makedirs(output_dir)

    # Iterate over all files in the input directory
    for filename in os.listdir(input_dir):
        if filename.endswith('.txt'):
            input_file_path = os.path.join(input_dir, filename)

            with open(input_file_path, 'r') as infile:
                for index, line in enumerate(infile, 1):  # Enumerate to give each line a unique number
                    # Clean up the line (optional, to avoid leading/trailing spaces)
                    clean_line = line.strip()

                    if clean_line:  # Only create a file if the line is not empty
                        output_file_name = f"{os.path.splitext(filename)[0]}_line_{index}.txt"
                        output_file_path = os.path.join(output_dir, output_file_name)

                        with open(output_file_path, 'w') as outfile:
                            outfile.write(clean_line)

                        print(f'Created file: {output_file_name} with content: "{clean_line}"')

if __name__ == "__main__":
    input_directory = './sample-data/players'  # Path to the directory containing the .txt files
    output_directory = './sample-data/players-file'  # Path to the directory where the new files will be written

    write_lines_as_files(input_directory, output_directory)
