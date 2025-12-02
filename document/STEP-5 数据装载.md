# 数据装载
目的：把trino中的DWS冷数据装载到clickhouse中，作为ck的业务数据提供给客户端查询

## 第一 准备工作
### 2-1 在clickhouse中建立ads表
```sql
CREATE TABLE t_record_3001.ads_biz_bet_hour (
    log_date     Date,
    hour         UInt8,
    plat_id      Int32,
    app_id       Int32,
    currency     String,
    game_id      Int32,
    amount       Float64,
    version      UInt64 DEFAULT toUnixTimestamp64Milli(now64())
)
ENGINE = ReplacingMergeTree(version)
PARTITION BY log_date
ORDER BY (log_date, hour, plat_id, app_id, currency, game_id);
```

### 2-2 trino环境配置
* 找到/trino/catalog/（见前面的环境搭建），新建clickhouse.properties，并编辑连接内容。
```text
connector.name=clickhouse
connection-url=jdbc:clickhouse://{ck-host}:8123
connection-user=default
connection-password=test123456
```
* 重启 Trino 容器
```text
docker restart trino-coordinator trino-worker-1 trino-worker-2
```
* 验证连接，进入trino sql窗口。确定在trino连接上，能够查询到ck中刚新建的表
```sql
-- check connect
SHOW SCHEMAS FROM clickhouse;
SHOW TABLES FROM clickhouse.t_record_3001;

-- query ads data for ck
SELECT 
    * 
FROM 
    clickhouse.t_record_3001.ads_biz_bet_hour 
LIMIT 10;
```

## 二 装载数据
已知
- ck中已经建立好了表ads_biz_bet_hour，用于接收trino中dws_biz_bet_hour的数据。
- trino和ck的网络信息连接成功。
- 至此，可以准备进行 跨catalog 数据load，打开trino sql窗口
```sql
INSERT INTO clickhouse.t_record_3001.ads_biz_bet_hour (
    log_date,
    hour,
    plat_id,
    app_id,
    currency,
    game_id,
    amount,
    version
)
SELECT
    log_date,
    CAST(hour AS smallint),  -- Iceberg: integer → CK: UInt8 → Trino 映射为 smallint
    plat_id,
    app_id,
    CAST(currency AS varbinary), -- Iceberg: varchar → CK String → Trino 映射为 varbinary
    game_id,
    amount,
    CAST(to_unixtime(current_timestamp) * 1000 AS decimal(20,0)) AS version
FROM iceberg.demo.dws_biz_bet_hour
WHERE log_date = DATE '2025-11-25'
  AND hour = 18;
```
从ck中进行验证
![project](https://github.com/struts2execute-dot/Data-Lake-Project/blob/main/img/ck_check.png)

## 三 ck多模组配置
在 ./trino/catalog/ 下建多个配置文件，一个配置代表一个plat组。多个 catalog 的“名字”是由 文件名 决定的，根据文件名进行数据库访问即可）。
* clickhouse_1003.properties
```text
connector.name=clickhouse
connection-url=jdbc:clickhouse://ck-1003:8123
connection-user=default
connection-password=xxx
```
* clickhouse_1004.properties
```text
connector.name=clickhouse
connection-url=jdbc:clickhouse://ck-1004:8123
connection-user=default
connection-password=xxx
```

至此：整个数湖的数据调度逻辑全部完成。