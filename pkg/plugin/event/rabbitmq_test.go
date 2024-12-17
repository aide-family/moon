package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"testing"
)

func TestNewEvent(t *testing.T) {
	connectRabbitMQ()
}

func connectRabbitMQ() {

	// 1. 建立连接
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// 2. 创建通道
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// 3. 声明队列（确保队列存在）
	queueName := "test_queue"
	q, err := ch.QueueDeclare(
		queueName, // 队列名
		true,      // 是否持久化
		false,     // 是否自动删除
		false,     // 是否独占
		false,     // 是否等待
		nil,       // 额外参数
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	log.Printf("Waiting for messages from queue: %s", q.Name)

	// 4. 注册消费者，消费消息
	msgs, err := ch.Consume(
		q.Name, // 队列名
		"",     // 消费者标识（空表示自动生成）
		false,  // 是否自动确认（建议设置为 false 手动确认）
		false,  // 是否独占
		false,  // 是否 no-local（不适用）
		false,  // 是否 no-wait
		nil,    // 额外参数
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	// 5. 处理消息
	forever := make(chan bool) // 保持程序运行
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			// 手动确认消息
			if err := d.Ack(false); err != nil {
				log.Printf("Failed to acknowledge message: %v", err)
			} else {
				log.Printf("Message acknowledged")
			}
		}
	}()

	log.Println(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever // 阻塞，等待消息
}
