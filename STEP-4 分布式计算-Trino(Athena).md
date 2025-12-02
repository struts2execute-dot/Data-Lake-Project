# 分布式计算
自建Trino对标AWS的Athena，核心原理应该还是从mapreduce演化而来的，逻辑差不多。

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