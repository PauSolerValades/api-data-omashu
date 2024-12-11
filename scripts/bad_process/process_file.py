def is_bytes_section(line):
    """
    Check if the line contains 'POSTMATCH' or 'BYTIME', indicating it might have byte-encoded content.
    """
    # If the line contains 'POSTMATCH' or 'BYTIME', we treat it as bytes
    return b'POSTMATCH' in line or b'BYTIME' in line

def process_file(input_file, bytes_output, non_bytes_output):
    """
    Process the file by reading it in binary mode (without decoding),
    then check each line to determine if it's byte-encoded or regular text.
    Write byte-like lines (containing 'POSTMATCH' or 'BYTIME') to one file,
    and the rest to another file.
    """
    with open(input_file, 'rb') as source_file, \
         open(bytes_output, 'w', encoding='utf-8') as bytes_file, \
         open(non_bytes_output, 'w', encoding='utf-8') as non_bytes_file:
        
        for line in source_file:
            # Check if the line contains byte-like data
            if is_bytes_section(line):
                # Write the byte-encoded lines, decoding with UTF-16
                bytes_file.write(line.decode('utf-16', errors='replace') + '\n')
            else:
                # Write the regular data (decoded as UTF-8)
                non_bytes_file.write(line.decode('utf-8', errors='replace') + '\n')

# Example usage
input_file = './output/matches.txt'
bytes_output = './output/bytime_postmatch.txt'  # Will contain 'POSTMATCH' or 'BYTIME' data
non_bytes_output = './output/regular_data.txt'  # Will contain regular data

process_file(input_file, bytes_output, non_bytes_output)
