package conn

import (
	"fmt"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
)

var kafkaEndpoints = []string{"localhost:9092"}

func Test_push(t *testing.T) {
	topic := "test"

	groupId := "consumer-" + uuid.New().String()
	fmt.Println(groupId)

	//consumer, err := NewKafkaConsumer(kafkaEndpoints, groupId)
	//if err != nil {
	//	fmt.Println("kafka消费者初始化失败")
	//	return
	//}
	//defer consumer.Close()

	//if err = consumer.SubscribeTopics([]string{topic}, func(consumer *kafka.Consumer, event kafka.Event) error {
	//	fmt.Println("kafka消费者开始消费", event)
	//	return nil
	//}); err != nil {
	//	fmt.Println("kafka消费者订阅失败")
	//	return
	//}

	//go func() {
	//	event := consumer.Poll(1000)
	//	for event != nil {
	//		switch e := event.(type) {
	//		case *kafka.Message:
	//			fmt.Printf("Message on %s:\n%s\n", e.TopicPartition, string(e.Value))
	//		case kafka.PartitionEOF:
	//			fmt.Printf("Reached %v\n", e)
	//		case kafka.Error:
	//			fmt.Printf("%% Error: %v\n", e)
	//		default:
	//			fmt.Printf("Ignored %v\n", e)
	//		}
	//		event = consumer.Poll(1000)
	//	}
	//}()

	producer, err := NewKafkaProducer(kafkaEndpoints)
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
			}, nil)

			if err != nil {
				fmt.Println("kafka生产失败", err)
				break
			}
			fmt.Println("kafka生产成功", count)
			time.Sleep(1 * time.Second)
			count++
			go producer.Flush(15 * 1000)
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
