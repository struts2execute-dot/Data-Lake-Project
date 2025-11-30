# 消息总线

## 一 ODS数据流
ODS数据流的安全稳定，是整个架构的核心基础
- 数据总线：单一topic，单一connector，便于维护管理；
- offset+事物+精确一致性：数据不丢且不重；
- Parquet+ZSTD：列式存储+极限压缩；

![data-flow](https://github.com/struts2execute-dot/Data-Lake-Project/blob/main/img/data-flow.png)

## 二 采集器

### 2-1 http创建connector

- tasks.max：消费者数量
- topics：消息总线topic
- consumer.override.max.poll.records：每次拉取的数据条数
- errors.deadletterqueue.topic.name：异常数据topic
- iceberg.control.topic：事务与提交协调器topic，为了确保内部保协调指令的顺序，所以这里partitions的数量必须是1，不能去修改本topic的partitions数量。

```yaml
curl -X POST -H "Content-Type: application/json" \
  http://localhost:8083/connectors \
  -d '{
    "name": "events-iceberg-sink",
    "config": {
      "connector.class": "io.tabular.iceberg.connect.IcebergSinkConnector",
      "tasks.max": "1",

      "topics": "events",
      "iceberg.control.topic": "iceberg-control-events",

      "consumer.override.auto.offset.reset": "earliest",

      "iceberg.tables": "demo.dwd_biz_recharge,demo.dwd_biz_bet",
      "iceberg.tables.route-field": "event",
      "iceberg.table.demo.dwd_biz_recharge.route-regex": "recharge",
      "iceberg.table.demo.dwd_biz_bet.route-regex": "bet",

      "iceberg.tables.auto-create-enabled": "true",
      "iceberg.tables.evolve-schema-enabled": "true",

      "iceberg.catalog": "demo",
      "iceberg.catalog.type": "rest",
      "iceberg.catalog.uri": "http://iceberg-rest:8181",
      "iceberg.catalog.warehouse": "s3://warehouse/",
      "iceberg.catalog.client.region": "us-east-1",
      "iceberg.catalog.s3.endpoint": "http://minio:9000",
      "iceberg.catalog.s3.path-style-access": "true",
      "iceberg.catalog.s3.access-key-id": "admin",
      "iceberg.catalog.s3.secret-access-key": "admin12345",

      "key.converter": "org.apache.kafka.connect.storage.StringConverter",
      "key.converter.schemas.enable": "false",
      "value.converter": "org.apache.kafka.connect.json.JsonConverter",
      "value.converter.schemas.enable": "true",

      "errors.tolerance": "all",
      "errors.log.enable": "true",
      "errors.log.include.messages": "true",
      "errors.deadletterqueue.topic.name": "dlq-events",
      "errors.deadletterqueue.context.headers.enable": "true",
      "errors.deadletterqueue.topic.replication.factor": "1",
      "consumer.override.max.poll.records": "500",
      "behavior.on.null.values": "IGNORE",

      "iceberg.control.commit.interval-ms": "10000"
    }
  }'
```

### 2-1 维护connector

```
// 查看 连接器 状态
curl http://localhost:8083/connectors/events-iceberg-sink/status | jq .

// 删除 连接器
curl -X DELETE http://localhost:8083/connectors/events-iceberg-sink  
```

## 三 事件topic

- 需要关闭kafka自动创建topic的功能；
- topic需要我们手动创建；
- 模拟环境下，宿主机安装kafka客户端，使用命令进行创建

```
// kafka
cd /opt/homebrew/opt/kafka/bin
// 创建主线topic
kafka-topics --bootstrap-server localhost:19092 --create --topic events  --partitions 1  --replication-factor 1
// 创建事物提交topic
kafka-topics --bootstrap-server localhost:19092 --create --topic iceberg-control-events  --partitions 1  --replication-factor 1
```