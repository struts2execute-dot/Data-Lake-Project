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