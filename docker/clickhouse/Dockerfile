FROM clickhouse/clickhouse-server:latest

# Copy initialization scripts
COPY ./docker/clickhouse/init.sql /docker-entrypoint-initdb.d/
COPY ./db/clickhouse/schema.sql /docker-entrypoint-initdb.d/

# Set proper permissions
RUN chmod 644 /docker-entrypoint-initdb.d/*.sql

# Define volume locations (these will be named volumes when you run the container)
VOLUME ["/var/lib/clickhouse", "/var/log/clickhouse-server"]

# Expose ClickHouse ports
EXPOSE 8123 9000

# Set ulimits
CMD ["bash", "-c", "ulimit -n 262144 && /entrypoint.sh"]