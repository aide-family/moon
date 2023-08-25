package conn

import (
	"fmt"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-kratos/kratos/v2/log"
)

var kafkaEndpoints = []string{"localhost:9092"}

func Test_push(t *testing.T) {
	topic := "test"

	consumer, err := NewKafkaConsumer(kafkaEndpoints, []string{topic}, log.DefaultLogger)
	if err != nil {
		fmt.Println("kafka消费者初始化失败")
		return
	}
	defer consumer.Close()

	consumer.Consume(func(et *kafka.Message) bool {
		fmt.Printf("message at topic:%v time:%d %s = %s\n", et.TopicPartition, et.Timestamp.Unix(), string(et.Key), string(et.Value))
		return true
	})

	producer, err := NewKafkaProducer(kafkaEndpoints, log.DefaultLogger)
	if err != nil {
		fmt.Println("kafka生产者初始化失败")
		return
	}
	defer producer.Close()
	go func() {
		count := 0
		for {
			err = producer.Produce(&kafka.Message{
				Key:   []byte(fmt.Sprintf("Key-%d", count)),
				Value: []byte(fmt.Sprintf(`{"count":"%d"}`, count)),
				TopicPartition: kafka.TopicPartition{
					Topic:     &topic,
					Partition: 0, // kafka.PartitionAny
				},
			})
			if err != nil {
				fmt.Println("kafka生产失败", err)
			}
			time.Sleep(3 * time.Second)
			count++
		}
	}()

	time.Sleep(20 * time.Second)
}

func TestName(t *testing.T) {
	timeStr := "2023-08-24T14:07:36.237Z"
	// 2023-08-24T14:07:36.237Z
	tm2, _ := time.Parse("2006-01-02T15:04:05Z07:00", timeStr)
	fmt.Println(tm2.Unix())

	// 0001-01-01T00:00:00Z
	tm3, _ := time.Parse("2006-01-02T15:04:05Z07:00", "0001-01-01T00:00:00Z")
	fmt.Println(tm3.Unix())

	loc, _ := time.LoadLocation("Local")
	startsAt, _ := time.ParseInLocation(time.DateTime, timeStr, loc)
	fmt.Println(startsAt.Unix())
}
