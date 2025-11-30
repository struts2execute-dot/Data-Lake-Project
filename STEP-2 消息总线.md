# 消息总线

## 一 ODS数据流
ODS数据流的安全稳定，是整个架构的核心基础
- 数据总线：单一topic，单一connector，便于维护管理；
- offset+事物+精确一致性：数据不丢且不重；
- Parquet+ZSTD：列式存储+极限压缩；

![data-flow](https://github.com/struts2execute-dot/Data-Lake-Project/blob/main/img/data-flow.png)
