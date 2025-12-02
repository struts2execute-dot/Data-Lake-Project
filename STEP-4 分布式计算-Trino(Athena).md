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
通常的架构是一个Coordinator对应多个Worker