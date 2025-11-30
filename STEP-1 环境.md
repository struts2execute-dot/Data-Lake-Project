# 数据湖项目（Kafka + Iceberg + MinIO + Trino 本地环境）

本项目基于 Docker 构建一个轻量级的数据湖（Lakehouse）环境，包含：

- **Kafka / Kafka Connect / Kafka UI**
- **MinIO（S3）**
- **Iceberg REST Catalog（元数据管理）**
- **PostgreSQL（Iceberg Catalog 后端）**
- **Trino（分布式 SQL 查询引擎）**

适用于本地开发测试、个人研究和企业 PoC（Proof of Concept）。

---

## 一、docker-compose 环境

将以下内容保存为 `docker-compose.yml`：

```yaml
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
  # 2) Iceberg REST Catalog（替代 Hive Metastore）
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
```

---

## 📌 二、插件准备

### 1. S3 Sink Connector（Confluent 官方）

下载地址：

```
https://www.confluent.io/hub/confluentinc/kafka-connect-s3
```

下载后存入：

```
./plugins/
```

并解压。

---

### 2. Iceberg Kafka Connector

下载：

```
https://github.com/databricks/iceberg-kafka-connect/releases
```

放入：

```
./plugins/
```

并解压。

---

## 📌 三、启动 / 停止 / 状态检查

- 启动所有服务：

  ```bash
  docker-compose up -d
  ```

- 停止所有服务：

  ```bash
  docker-compose down
  ```

- 查看容器运行状态：

  ```bash
  docker ps
  ```

- 查看 Kafka Connect 插件是否加载成功：

  ```bash
  curl http://localhost:8083/connector-plugins
  ```

---

## 📌 四、服务 Web 界面（默认端口）

| 服务 | 访问地址 |
|------|----------|
| Kafka UI | http://localhost:8080 |
| MinIO 控制台 | http://localhost:9090 |
| Trino Web UI | http://localhost:8085 |
| Iceberg REST Catalog | http://localhost:8181 |
| Kafka Connect 插件列表 | http://localhost:8083/connector-plugins |

---

## 📌 五、注意事项 / 建议

- **Iceberg Catalog 必须使用 PostgreSQL**，比 SQLite 稳定得多。
- **MinIO 仅用于本地开发**，生产环境应替换为 AWS S3 / 阿里云 OSS / GCS 等对象存储。
- 第一次启动若失败，可能是 PostgreSQL / MinIO / Kafka 等服务尚未完全 ready，可稍等片刻再重启相关容器。
- 由于服务较多，对机器资源有一定要求，建议运行环境至少有 **16GB RAM + 多核 CPU**。
- 该方案适合中小规模数据湖 / Lakehouse，用于测试 / 实验 / 验证架构。若用于生产，请酌情扩容、监控、权限/安全配置。

---

## ✅ 完整 README 模板 — 可以直接复制到 GitHub 仓库

