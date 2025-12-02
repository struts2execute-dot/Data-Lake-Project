# 分布式计算
本章核心是使用自建Trino对已经完成存储的文件进行分析，核心原理应该还是从mapreduce演化而来的，逻辑差不多。
* 首先需要通过dbeaver或者dbVisualizer(推荐，dbeaver连接会有一些莫名其妙的问题)连接到trino。
- 地址：localhost:8085
- 用户名：任意，但必填
- 密码：无
* 连接以后即可进行DWS数据聚合。

## 一 介绍
![project](https://github.com/struts2execute-dot/Data-Lake-Project/blob/main/img/cal.png)

### 1-1-1 Coordinator + Worker架构
```text
+-------------------+
|   Coordinator     |
|  (负责调度工作)    |
+-------------------+
         |
         v
+--------+---------+--------+
| Worker | Worker | Worker |
+--------+---------+--------+
```
- Coordinator：解析 SQL、规划执行计划、调度 Worker
- Worker：执行实际计算（scan、filter、join、aggregate）
- 通常的架构是一个Coordinator对应多个Worker

### 1-1-2 核心原理
1. 执行sql
```text
SELECT event, COUNT(*)
FROM table
WHERE dt='2025-01-01'
GROUP BY event;
```
2. 运行逻辑
```text
Coordinator：查 dt=2025-01-01
↓
Coordinator：生成执行计划
↓
Workers：并行扫描 S3 / MinIO 上的数据
↓
Workers：做 filter / project
↓
Workers：做部分聚合（Partial Aggregation）
↓
Coordinator：收集所有 Worker 部分聚合结果，合并成最终结果
↓
返回 SQL 结果
```

## DWS聚合逻辑
- 按小时调度，使用merge语法作upsert操作，以便支持数据重跑
```sql
--view table ddl
SHOW CREATE TABLE iceberg.demo.dwd_biz_bet;

--view describe
DESCRIBE iceberg.demo.dwd_biz_bet;

--create dws table
CREATE TABLE iceberg.demo.dws_biz_bet_hour (
   log_date   date,
   hour       integer,
   plat_id    integer,
   app_id     integer,
   currency   varchar,
   game_id    integer,
   amount     double
)
WITH (
   format = 'PARQUET',
   format_version = 2,
   compression_codec = 'ZSTD',
   location = 's3://warehouse/demo/dws_biz_bet_hour',
   partitioning = ARRAY['log_date', 'hour']
);

--dispatch
MERGE INTO iceberg.demo.dws_biz_bet_hour t
USING (
    SELECT 
         CAST(log_date AS date)    AS log_date,
         CAST(hour AS integer)     AS hour,
         plat_id,
         app_id,
         currency,
         game_id,
         SUM(amount)               AS amount
    FROM iceberg.demo.dwd_biz_bet 
    WHERE 
         log_date = '2025-11-25' 
         AND hour = '18'
    GROUP BY 
         log_date,
         hour,
         plat_id,
         app_id,
         currency,
         game_id
) s
ON  t.log_date  = s.log_date
AND t.hour      = s.hour
AND t.plat_id   = s.plat_id
AND t.app_id    = s.app_id
AND t.currency  = s.currency
AND t.game_id   = s.game_id
WHEN MATCHED THEN
    UPDATE SET 
        amount = s.amount    
WHEN NOT MATCHED THEN
    INSERT (log_date, hour, plat_id, app_id, currency, game_id, amount)
    VALUES (s.log_date, s.hour, s.plat_id, s.app_id, s.currency, s.game_id, s.amount);
    
--check dispatch result
select * from iceberg.demo.dws_biz_bet_hour;
```
![project](https://github.com/struts2execute-dot/Data-Lake-Project/blob/main/img/dws.png)
