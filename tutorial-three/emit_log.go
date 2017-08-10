// 本文档为 RabbitMQ `发布/订阅` 模式的生产者示例代码。
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	// Go RabbitMQ 客户端包
	"github.com/streadway/amqp"
)

// 工具函数，打印错误信息并使程序 Panic 异常退出。
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func main() {
	// 使用 guest 用户及密码连接 RabbitMQ，使用默认 vhost `/`。连接 RabbitMQ 消息头为 amqp。
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	// 使用 defer 函数关闭连接
	defer conn.Close()

	// 创建一个信道（channel），用于传输消息。
	// Note:
	// 1. 必须首先连接到 RabbitMQ，才能消费或发布消息，所以我们必须在应用程序和 RabbitMQ 代理
	// 代理服务器之间创建一条 TCP 连接，建立连接过程如前所示。
	// 2. 一旦 TCP 连接打开（你通过了验证），应用程序就可以创建一条 AMQP。信道是建立在“真是的”
	// TCP 连接内的虚拟连接。AMQP 命令都是通过信道发送出去的。每条信道都会被指派一个唯一 ID（
	// AMQP 库会帮你记住 ID）。不论是发送消息、订阅队列或是接收消息，这些动作都是通过信道完成的。
	// 3. 引入信道的原因：操作系统建立和销毁 TCP 会话的代价是高昂的。假设应用程序从队列消费消息，
	// 并根据服务需求合理调度线程。若只进行 TCP 连接，那么每个线程都需要自行连接到 Rabbit。也就
	// 说高峰期有每秒成百上千的连接。这不仅造成 TCP 连接的巨大浪费，而且操作系统每秒只能建立有限
	// 数量的连接。RabbitMQ 的做法是，线程启动后，会在线程的连接上创建一条信道，也就获得了连接
	// 到 Rabbit 上的私密通信路径，而不会给操作系统的 TCP 栈造成额外负担。
	// 4. 可以将一条 TCP 连接想象成电缆，而 AMQP 信道就像是电缆中一条条独立的光纤束。
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	// 使用 defer 关闭信道
	defer ch.Close()

	// 在信道上声明一个名叫`logs`的交换机，类型为`fanout`。若同名的交换机已经存在，则不做任何
	// 处理。
	err = ch.ExchangeDeclare(
		"logs",   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	// 从命令行参数中获取要发送的消息内容
	body := bodyFrom(os.Args)

	// 发布消息到交换机`logs`上，使用空的路由键，消息为纯文本消息
	err = ch.Publish(
		"logs", // exchange
		"",     // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain", // 创建的消息是纯文本的
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")

	// 发送成功后在终端显示发送的内容
	log.Printf(" [x] Sent %s", body)
}

// 工具函数，从标准输入中获取消息内容，若传入的只有一个参数，设置默认值为`hello`
func bodyFrom(args []string) string {
	var s string
	if (len(args) < 2) || os.Args[1] == "" {
		s = "hello"
	} else {
		s = strings.Join(args[1:], " ")
	}
	return s
}