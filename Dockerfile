FROM postgres:15

RUN apt-get update && apt-get install -y \
    postgresql-server-dev-15 \
    build-essential \
    && rm -rf /var/lib/apt/lists/* \
    && git clone https://github.com/pgvector/pgvector.git /pgvector \
    && cd /pgvector \
    && make \
    && make install
