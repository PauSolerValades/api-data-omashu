FROM python:3.11-slim

ENV PYTHONUNBUFFERED=1

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*

# Create a non-root user
RUN adduser --disabled-password --gecos '' appuser

WORKDIR /app
# TODO: AIXÒ ÉS SUPER CUTRE, MIRAR COM PENJAR LA LLIBRERIA AL PYPI/FER-HO AMB POETRY
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY ./modules/process_lol_matches-0.4.0-py3-none-any.whl .
RUN pip install process_lol_matches-0.4.0-py3-none-any.whl

# Copy the rest of the application code
COPY . .

RUN chmod +x run.sh
RUN chown -R appuser:appuser /app

USER appuser
CMD ["./run.sh"]
