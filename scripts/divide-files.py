"""divede-files"""

import subprocess
from pathlib import Path

if __name__ == "__main__":
    # Define the directory containing the .sh files
    directory = "./sample-data/players-file/"

    for file_path in Path(directory).iterdir():
        if file_path.suffix == ".txt":  # Check if the file has a .sh extension
            print(f"Executing: {file_path}")
            try:
                print(f"------------------ EXECUTING {file_path} --------------------")
                subprocess.run(["./scripts/run_docker.sh", "./scripts/Dockerfile", "./sample-data", "./scripts/produce_to.sh", "download-start", "/" + str(file_path)], check=True)
            except subprocess.CalledProcessError as e:
                print(f"------------------ ERROR {file_path} --------------------")
                print(e)

            print(f"------------------ END {file_path} --------------------")
