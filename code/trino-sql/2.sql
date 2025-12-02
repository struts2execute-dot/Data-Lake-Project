-- check connect
SHOW SCHEMAS FROM clickhouse;
SHOW TABLES FROM clickhouse.t_record_3001;

-- query ads data for ck
SELECT
    *
FROM
    clickhouse.t_record_3001.ads_biz_bet_hour
        LIMIT 10;

-- execute load data
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

