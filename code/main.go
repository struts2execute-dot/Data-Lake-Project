package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/IBM/sarama"

	"go-test/kafks"
	"go-test/model"
	"go-test/model/payload"
)

func main() {
	producer := kafks.GetProducer()
	defer func() {
		if err := producer.Close(); err != nil {
			log.Printf("close producer error: %v", err)
		}
	}()

	// 投注
	events := GetBetEventData(1004, 10406)
	events = append(events, GetBetEventData(1003, 3001)...)
	SendEvent(events, "events", producer)

	// 充值
	events = make([]interface{}, 0)
	events = GetRechargeData(1004, 10406)
	events = append(events, GetRechargeData(1003, 3001)...)
	SendEvent(events, "events", producer)
}

func SendEvent(events []interface{}, topic string, producer sarama.SyncProducer) {
	for i, ev := range events {
		env := model.BuildEnvelope(ev)

		// ------ 打印完整 JSON（schema + payload）------
		prettyJSON, _ := json.MarshalIndent(env, "", "  ")
		log.Printf("\n================ EVENT %d JSON =================\n%s\n", i, prettyJSON)
		// ----------------------------------------------------------

		valueBytes, err := json.Marshal(env)
		if err != nil {
			log.Printf("marshal event %d failed: %v", i, err)
			continue
		}

		msg := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.ByteEncoder(valueBytes),
		}

		partition, offset, err := producer.SendMessage(msg)
		if err != nil {
			log.Printf("send event %d failed: %v", i, err)
			continue
		}

		log.Printf("sent event %d ok, partition=%d offset=%d", i, partition, offset)
	}
}

func GetRechargeData(platID int32, appID int32) []interface{} {
	strPtr := func(s string) *string { return &s }
	i32Ptr := func(v int32) *int32 { return &v }
	f64Ptr := func(v float64) *float64 { return &v }

	// 小工具函数

	logDate := "2025-11-25"
	hour := "18"

	// 这里你也可以用 time.Now() 算 ts，这里先保持和前面一致
	var (
		tsRecharge1 int64 = 1732527600000
		tsRecharge2 int64 = 1732527605000
		tsRecharge3 int64 = 1732527610000
	)
	_ = time.Now // 防止没用到 time 包被 IDE 提示

	events := []interface{}{
		// 1) 充值成功
		payload.Recharge{
			Payload: payload.Payload{
				Category:      "biz",
				Event:         "recharge",
				LogDate:       logDate,
				Hour:          hour,
				PlatID:        platID,
				AppID:         appID,
				Ts:            tsRecharge1,
				SchemaVersion: 1,
			},
			Status:   strPtr("success"),
			UserID:   i32Ptr(1001),
			OrderID:  strPtr("pay_20251125_0001"),
			Amount:   f64Ptr(100.5),
			Currency: strPtr("BRL"),
			Reason:   nil,
		},

		// 2) 充值失败
		payload.Recharge{
			Payload: payload.Payload{
				Category:      "biz",
				Event:         "recharge",
				LogDate:       logDate,
				Hour:          hour,
				PlatID:        platID,
				AppID:         appID,
				Ts:            tsRecharge2,
				SchemaVersion: 1,
			},
			Status:   strPtr("fail"),
			UserID:   i32Ptr(1002),
			OrderID:  strPtr("pay_20251125_0002"),
			Amount:   f64Ptr(50.0),
			Currency: strPtr("BRL"),
			Reason:   strPtr("PAYMENT_TIMEOUT"),
		},

		// 3) 充值成功（另一个 plat/app）
		payload.Recharge{
			Payload: payload.Payload{
				Category:      "biz",
				Event:         "recharge",
				LogDate:       logDate,
				Hour:          hour,
				PlatID:        platID,
				AppID:         appID,
				Ts:            tsRecharge3,
				SchemaVersion: 1,
			},
			Status:   strPtr("success"),
			UserID:   i32Ptr(1003),
			OrderID:  strPtr("pay_20251125_0003"),
			Amount:   f64Ptr(300.75),
			Currency: strPtr("BRL"),
			Reason:   nil,
		},
	}
	return events
}

func GetBetEventData(platID int32, appID int32) []interface{} {
	strPtr := func(s string) *string { return &s }
	i32Ptr := func(v int32) *int32 { return &v }
	f64Ptr := func(v float64) *float64 { return &v }

	// 小工具函数

	logDate := "2025-11-25"
	hour := "18"

	// 这里你也可以用 time.Now() 算 ts，这里先保持和前面一致
	var (
		tsBet1 int64 = 1732527602000
		tsBet2 int64 = 1732527607000
		tsBet3 int64 = 1732527612000
	)
	_ = time.Now // 防止没用到 time 包被 IDE 提示

	events := []interface{}{
		// 4) bet 1
		payload.Bet{
			Payload: payload.Payload{
				Category:      "biz",
				Event:         "bet",
				LogDate:       logDate,
				Hour:          hour,
				PlatID:        platID,
				AppID:         appID,
				Ts:            tsBet1,
				SchemaVersion: 1,
			},
			UserID:   i32Ptr(2001),
			OrderID:  strPtr("10001"),
			Amount:   f64Ptr(10),
			Currency: strPtr("BRL"),
			GameID:   i32Ptr(777),
		},

		// 5) bet 2
		payload.Bet{
			Payload: payload.Payload{
				Category:      "biz",
				Event:         "bet",
				LogDate:       logDate,
				Hour:          hour,
				PlatID:        platID,
				AppID:         appID,
				Ts:            tsBet2,
				SchemaVersion: 1,
			},
			UserID:   i32Ptr(2002),
			OrderID:  strPtr("10002"),
			Amount:   f64Ptr(10),
			Currency: strPtr("BRL"),
			GameID:   i32Ptr(777),
		},

		// 6) bet 3 (另一个 plat/app)
		payload.Bet{
			Payload: payload.Payload{
				Category:      "biz",
				Event:         "bet",
				LogDate:       logDate,
				Hour:          hour,
				PlatID:        platID,
				AppID:         appID,
				Ts:            tsBet3,
				SchemaVersion: 1,
			},
			UserID:   i32Ptr(2003),
			OrderID:  strPtr("10003"),
			Amount:   f64Ptr(20),
			Currency: strPtr("BRL"),
			GameID:   i32Ptr(888),
		},
	}
	return events
}
