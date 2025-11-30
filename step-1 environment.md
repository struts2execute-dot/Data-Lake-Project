📌 1. Components Overview

This project demonstrates a minimal local setup of a modern Lakehouse Architecture using Docker:

Component	Description
Zookeeper	Coordination service for Kafka.
Kafka	Distributed message queue, supports batch consumption, offset management, and plugin ecosystem.
Kafka-UI	Web dashboard for monitoring Kafka topics, consumers and offsets.
Kafka-Connect	Responsible for consuming Kafka messages and sinking them to S3/MinIO (Parquet / Iceberg).
MinIO	Distributed object storage (S3-compatible). Production-ready alternative: AWS S3.
Iceberg-REST	REST catalog service for Apache Iceberg (alternative to Hive Metastore).
Postgres	Iceberg catalog backend (replaces SQLite, more stable).
Trino	Distributed SQL engine for analytics and ad-hoc querying.
📌 2. Docker Compose Setup

The following configuration has been tested and verified locally.
For production environments, replace MinIO with AWS S3 and scale components horizontally.

Project directory:

/workspace-bigdata/offline/kafka-s3-demo

docker-compose.yml
version: '3.8'

services:

# -------------------------------
# 0) Zookeeper & Kafka & Connect
# -------------------------------
zookeeper:
image: confluentinc/cp-zookeeper:7.5.0
container_name: zookeeper
environment:
ZOOKEEPER_CLIENT_PORT: 2181
ZOOKEEPER_TICK_TIME: 2000

kafka:
image: confluentinc/cp-kafka:7.5.0
container_name: kafka
depends_on:
- zookeeper
ports:
- "9092:9092"
- "19092:19092"
environment:
KAFKA_BROKER_ID: 1
KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
KAFKA_LISTENERS: PLAINTEXT_INTERNAL://0.0.0.0:9092,PLAINTEXT_EXTERNAL://0.0.0.0:19092
KAFKA_ADVERTISED_LISTENERS: PLAINTEXT_INTERNAL://kafka:9092,PLAINTEXT_EXTERNAL://localhost:19092
KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: PLAINTEXT_INTERNAL:PLAINTEXT,PLAINTEXT_EXTERNAL:PLAINTEXT
KAFKA_INTER_BROKER_LISTENER_NAME: PLAINTEXT_INTERNAL
KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
KAFKA_AUTO_CREATE_TOPICS_ENABLE: "false"
KAFKA_LOG_RETENTION_HOURS: 168
KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR: 1
KAFKA_TRANSACTION_STATE_LOG_MIN_ISR: 1

kafka-ui:
image: provectuslabs/kafka-ui:latest
container_name: kafka-ui
ports:
- "8080:8080"
depends_on:
- kafka
environment:
KAFKA_CLUSTERS_0_NAME: "local"
KAFKA_CLUSTERS_0_BOOTSTRAPSERVERS: "kafka:9092"

connect:
image: confluentinc/cp-kafka-connect:7.5.0
container_name: connect
depends_on:
- kafka
ports:
- "8083:8083"
environment:
CONNECT_BOOTSTRAP_SERVERS: "kafka:9092"
CONNECT_REST_ADVERTISED_HOST_NAME: "connect"
CONNECT_REST_PORT: "8083"
CONNECT_GROUP_ID: "connect-cluster"
CONNECT_CONFIG_STORAGE_TOPIC: "connect-configs"
CONNECT_OFFSET_STORAGE_TOPIC: "connect-offsets"
CONNECT_STATUS_STORAGE_TOPIC: "connect-status"
CONNECT_CONFIG_STORAGE_REPLICATION_FACTOR: "1"
CONNECT_OFFSET_STORAGE_REPLICATION_FACTOR: "1"
CONNECT_STATUS_STORAGE_REPLICATION_FACTOR: "1"
CONNECT_KEY_CONVERTER: "org.apache.kafka.connect.storage.StringConverter"
CONNECT_KEY_CONVERTER_SCHEMAS_ENABLE: "false"
CONNECT_VALUE_CONVERTER: "org.apache.kafka.connect.json.JsonConverter"
CONNECT_VALUE_CONVERTER_SCHEMAS_ENABLE: "true"
CONNECT_PLUGIN_PATH: "/usr/share/java,/usr/share/confluent-hub-components,/connect-plugins"
volumes:
- ./plugins:/connect-plugins
command:
- bash
- -c
- |
echo "Installing S3 sink connector from Confluent Hub..."
confluent-hub install --no-prompt confluentinc/kafka-connect-s3:latest
echo "Starting Kafka Connect..."
/etc/confluent/docker/run

# -------------------------------
# 1) MinIO (S3 Object Storage)
# -------------------------------
minio:
image: minio/minio:latest
container_name: minio
environment:
MINIO_ROOT_USER: admin
MINIO_ROOT_PASSWORD: admin12345
command: server /data --console-address ":9090"
ports:
- "9000:9000"
- "9090:9090"
volumes:
- ./minio-data:/data

# -----------------------------------------------
# 2) Iceberg REST Catalog (replace Hive Metastore)
# -----------------------------------------------
iceberg-postgres:
image: postgres:15-alpine
container_name: iceberg-postgres
environment:
POSTGRES_DB: iceberg
POSTGRES_USER: iceberg
POSTGRES_PASSWORD: iceberg123
volumes:
- ./iceberg-pg-data:/var/lib/postgresql/data

iceberg-rest:
image: tabulario/iceberg-rest:latest
container_name: iceberg-rest
depends_on:
- minio
- iceberg-postgres
environment:
CATALOG_URI: jdbc:postgresql://iceberg-postgres:5432/iceberg?user=iceberg&password=iceberg123
CATALOG_WAREHOUSE: s3://warehouse/
CATALOG_IO__IMPL: org.apache.iceberg.aws.s3.S3FileIO
CATALOG_S3_ENDPOINT: http://minio:9000
CATALOG_S3_PATH__STYLE__ACCESS: "true"
AWS_ACCESS_KEY_ID: admin
AWS_SECRET_ACCESS_KEY: admin12345
AWS_REGION: us-east-1
ports:
- "8181:8181"

# -------------------------------
# 3) Trino Coordinator
# -------------------------------
trino-coordinator:
image: trinodb/trino:latest
container_name: trino-coordinator
depends_on:
- iceberg-rest
- minio
ports:
- "8085:8085"
volumes:
- ./trino/coordinator:/etc/trino
- ./trino/catalog:/etc/trino/catalog

# -------------------------------
# 4) Trino Worker Nodes
# -------------------------------
trino-worker-1:
image: trinodb/trino:latest
container_name: trino-worker-1
depends_on:
- trino-coordinator
volumes:
- ./trino/worker:/etc/trino
- ./trino/catalog:/etc/trino/catalog

trino-worker-2:
image: trinodb/trino:latest
container_name: trino-worker-2
depends_on:
- trino-coordinator
volumes:
- ./trino/worker:/etc/trino
- ./trino/catalog:/etc/trino/catalog

networks:
default:
name: iceberg_net
driver: bridge


💡 If any service fails to start, it's usually because its dependency wasn't fully initialized.
Simply restart the container after a few seconds.

📌 3. Kafka Connect Plugin Preparation
3.1 S3 Sink Connector (Confluent)

Download from:
https://www.confluent.io/hub/confluentinc/kafka-connect-s3

Ensure version matches your Kafka Connect (e.g., 7.5.0)

Extract it into:

./plugins/

3.2 Iceberg Kafka Connector

Download from:
https://github.com/databricks/iceberg-kafka-connect/releases

Place and extract into the same:

./plugins/

📌 4. Maintenance & Operations
# Start all services
docker-compose up -d

# Check running containers
docker ps

# List loaded Kafka Connect plugins
curl http://localhost:8083/connector-plugins

# Stop all containers
docker-compose down

📌 5. Service Health Checklist

Kafka UI: http://localhost:8080

MinIO Console: http://localhost:9090

Trino Web UI: http://localhost:8085

Iceberg REST Catalog: http://localhost:8181

Kafka Connect Plugins: GET http://localhost:8083/connector-plugins