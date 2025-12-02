--view table ddl
SHOW CREATE TABLE iceberg.demo.dwd_biz_bet;

--view describe
DESCRIBE iceberg.demo.dwd_biz_bet;

--view partition info
SELECT * FROM iceberg.demo."dwd_biz_bet$partitions"   LIMIT 20;

--query data size
select count(1) from iceberg.demo.dwd_biz_bet;
select count(1) from iceberg.demo.dwd_biz_recharge;

select * from iceberg.demo.dwd_biz_bet where log_date='2025-11-25';

select
    log_date,
    hour,
    plat_id,
    app_id,
    currency,
    game_id,
    sum(amount) as amount
from iceberg.demo.dwd_biz_bet
where
    log_date='2025-11-25'
  and hour='18'
group by
    log_date,
    hour,
    plat_id,
    app_id,
    currency,
    game_id;


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

select * from iceberg.demo.dws_biz_bet_hour;
