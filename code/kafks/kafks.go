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
