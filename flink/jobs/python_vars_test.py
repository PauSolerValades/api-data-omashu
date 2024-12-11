import sys
import site
import os

print("Python version:", sys.version)
print("Sys path:", sys.path)
print("Site packages:", site.getsitepackages())
print("PYTHONPATH:", os.environ.get("PYTHONPATH", "Not set"))

try:
    import process_lol_matches

    print("process_lol_matches is installed")
    print("process_lol_matches location:", process_lol_matches.__file__)
except ImportError as e:
    print("process_lol_matches is not installed or not found")
    print("Import error:", str(e))

# ... rest of your script

try:
    import pydantic_settings

    print("pydantic_settings is installed")
    print("pydantic_settings version:", pydantic_settings.__version__)
except ImportError:
    print("pydantic_settings is not installed")

try:
    import pydantic

    print("pydantic is installed")
    print("pydantic version:", pydantic.__version__)
except ImportError:
    print("pydantic is not installed")
