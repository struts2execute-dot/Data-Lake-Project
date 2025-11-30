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

## 四 消息体
因为做消息总线，需要配合插件的数据解析，所以必须采用Kafka Connect 官方标准格式：Schema + JSON，样例如下：

```json
{
  "schema": {
    "type": "struct",
    "name": "bet_record",
    "fields": [
      {
        "field": "category",
        "type": "string"
      },
      {
        "field": "event",
        "type": "string"
      },
      {
        "field": "log_date",
        "type": "string"
      },
      {
        "field": "hour",
        "type": "string"
      },
      {
        "field": "plat_id",
        "type": "int32"
      },
      {
        "field": "app_id",
        "type": "int32"
      },
      {
        "field": "ts",
        "type": "int64"
      },
      {
        "field": "schema_version",
        "type": "int32"
      },
      {
        "field": "user_id",
        "type": "int32",
        "optional": true
      },
      {
        "field": "order_id",
        "type": "string",
        "optional": true
      },
      {
        "field": "amount",
        "type": "double",
        "optional": true
      },
      {
        "field": "currency",
        "type": "string",
        "optional": true
      },
      {
        "field": "game_id",
        "type": "int32",
        "optional": true
      }
    ]
  },
  "payload": {
    "category": "biz",
    "event": "bet",
    "log_date": "2025-11-25",
    "hour": "18",
    "plat_id": 1,
    "app_id": 10406,
    "ts": 1732527602000,
    "schema_version": 1,
    "user_id": 2001,
    "order_id": "10001",
    "amount": 10,
    "currency": "BRL",
    "game_id": 777
  }
}
```

## 五 事件逻辑抽象

### 5-1 数据模型
对于开发来说，不同的业务需要抽象成一个统一的Schema + JSON格式的数据，确保逻辑简单，维护方便。

- ***事件模型***
```go
package payload

// 通用模型
type Payload struct {
    Category      string `json:"category"`
    Event         string `json:"event"`
    LogDate       string `json:"log_date"`
    Hour          string `json:"hour"`
    PlatID        int32  `json:"plat_id"`
    AppID         int32  `json:"app_id"`
    Ts            int64  `json:"ts"`
    SchemaVersion int32  `json:"schema_version"`
}

// 投注事件
type Bet struct {
    Payload
    UserID   *int32   `json:"user_id"`
    OrderID  *string  `json:"order_id"`
    Amount   *float64 `json:"amount"`
    Currency *string  `json:"currency"`
    GameID   *int32   `json:"game_id"`
}

// 充值事件
type Recharge struct {
    Payload
    Status   *string  `json:"status"`
    UserID   *int32   `json:"user_id"`
    OrderID  *string  `json:"order_id"`
    Amount   *float64 `json:"amount"`
    Currency *string  `json:"currency"`
    Reason   *string  `json:"reason"`
}
```

- ***schema模型***
```go
package schema

import (
    "reflect"
    "strings"
)

type Schema struct {
    Type   string  `json:"type"`
    Name   string  `json:"name"`
    Fields []Field `json:"fields"`
}

type Field struct {
    Field    string `json:"field"`
    Type     string `json:"type"`
    Optional bool   `json:"optional,omitempty"`
}

type Envelope struct {
    Schema  Schema      `json:"schema"`
    Payload interface{} `json:"payload"`
}

func BuildSchemaFromStruct(t reflect.Type, name string) Schema {
    fields := make([]Field, 0)
    var walk func(rt reflect.Type)
    walk = func(rt reflect.Type) {
       for i := 0; i < rt.NumField(); i++ {
          f := rt.Field(i)
          // 匿名嵌入（比如嵌入 Payload）
          if f.Anonymous {
             walk(f.Type)
             continue
          }
          // 非导出字段跳过
          if f.PkgPath != "" {
             continue
          }
          tag := f.Tag.Get("json")
          if tag == "-" {
             continue
          }
          jsonName := strings.Split(tag, ",")[0]
          if jsonName == "" {
             jsonName = f.Name
          }
          // optional 规则：指针类型 或 tag 里含 omitempty
          isOptional := false
          ft := f.Type
          if ft.Kind() == reflect.Ptr {
             isOptional = true
             ft = ft.Elem()
          }
          if strings.Contains(tag, "omitempty") {
             isOptional = true
          }
          var typeStr string
          switch ft.Kind() {
          case reflect.String:
             typeStr = "string"
          case reflect.Int32:
             typeStr = "int32"
          case reflect.Int, reflect.Int64:
             typeStr = "int64"
          case reflect.Float32:
             typeStr = "float"
          case reflect.Float64:
             typeStr = "double"
          case reflect.Bool:
             typeStr = "boolean"
          default:
             // 复杂类型后面需要的话再拓展
             continue
          }
          fields = append(fields, Field{
             Field:    jsonName,
             Type:     typeStr,
             Optional: isOptional,
          })
       }
    }
    walk(t)
    return Schema{
       Type:   "struct",
       Name:   name,
       Fields: fields,
    }
}
```

- ***事件注册***
```go
package model

import (
    "go-test/model/payload"
    "go-test/model/schema"
    "reflect"
)

// SchemaRegistry 业务类型 -> 对应的 Schema
var SchemaRegistry = map[string]schema.Schema{}

func RegisterSchema(eventType string, sample interface{}) {
    t := reflect.TypeOf(sample)
    if t.Kind() == reflect.Ptr {
       t = t.Elem()
    }
    s := schema.BuildSchemaFromStruct(t, eventType+"_record")
    SchemaRegistry[eventType] = s
}

func init() {
    // event 类型直接用 Event 字段的值
    RegisterSchema("recharge", payload.Recharge{})
    RegisterSchema("bet", payload.Bet{})
}
```

- ***消息体编译***
```go
package model

import (
    "go-test/model/schema"
    "reflect"
)

func BuildEnvelope(data interface{}) *schema.Envelope {
    v := reflect.ValueOf(data)
    if v.Kind() == reflect.Ptr {
       v = v.Elem()
    }
    // 所有业务 struct 都嵌入了 Payload，有 Event 字段
    field := v.FieldByName("Event")
    if !field.IsValid() || field.Kind() != reflect.String {
       panic("BuildEnvelope: data has no string field `Event`")
    }
    eventType := field.String()

    s, ok := SchemaRegistry[eventType]
    if !ok {
       panic("BuildEnvelope: schema not registered for event = " + eventType)
    }

    return &schema.Envelope{
       Schema:  s,
       Payload: data,
    }
}
```
至此，如果需要新增事件，只需要做：
- 在payload中新增事件的模型；
- 对新事件进行注册
  RegisterSchema("bet", payload.Bet{})

### 5-2 模型测试

```go
package kafks

import (
  "github.com/IBM/sarama"
  "log"
  "os"
)

func GetProducer() sarama.SyncProducer {
  // === 基本配置 ===
  brokers := []string{"localhost:19092"} // 如果不是这个地址，改这里

  logger := log.New(os.Stdout, "[producer] ", log.LstdFlags|log.Lmicroseconds)

  // === sarama Producer 配置 ===
  cfg := sarama.NewConfig()
  cfg.Producer.RequiredAcks = sarama.WaitForAll
  cfg.Producer.Retry.Max = 3
  cfg.Producer.Return.Successes = true
  cfg.Producer.Return.Errors = true
  cfg.Version = sarama.V2_5_0_0 // 一般都兼容，实在不行再调版本号

  producer, err := sarama.NewSyncProducer(brokers, cfg)
  if err != nil {
    logger.Fatalf("create producer failed: %v", err)
  }
  return producer
}
```

- 运行main函数。
- 查看kafka-ui。
- 查看s3存储。
  - data目录：ods列式压缩数据
  - metadata：表元数据

![docker](https://github.com/struts2execute-dot/Data-Lake-Project/blob/main/img/s3.png)

## 六总结
通过消息总线，我们可以将事件数据稳定的存储到S3存储中，为下一步的分布式计算做好准备。