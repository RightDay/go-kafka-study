package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Shopify/sarama"
	"github.com/gofiber/fiber/v2"
)

// Comment struct
type Comment struct {
	Text string `form:"text" json:"text"`
}

func main() {
	//fiber app 생성
	app := fiber.New()
	api := app.Group("/api/v1") // /api

	api.Post("/comments", createComment)

	app.Listen(":3000")

}

// kafka Producer 연결
func ConnectProducer(brokersUrl []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	conn, err := sarama.NewSyncProducer(brokersUrl, config)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func PushCommentToQueue(topic string, message []byte) error {
	// kafka broker 주소 연결
	brokersUrl := []string{"localhost:29092"}

	// kafka producer 연결
	producer, err := ConnectProducer(brokersUrl)
	if err != nil {
		return err
	}

	defer producer.Close()

	//메세지 생성
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	// 메세지 전송
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return err
	}

	fmt.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", topic, partition, offset)

	return nil
}

// 댓글 생성 핸들러 함수
func createComment(c *fiber.Ctx) error {
	// 댓글 정보를 담는 Comment 구조체 생성
	cmt := new(Comment)

	// 요청 Body Comment 구조체로 파싱
	if err := c.BodyParser(cmt); err != nil {
		log.Println(err)
		c.Status(400).JSON(&fiber.Map{
			"success": false,
			"message": err,
		})
		return err
	}

	// 댓글 정보 Json으로 변환
	cmtInBytes, err := json.Marshal(cmt)
	if err != nil {
		return err
	}

	PushCommentToQueue("comments", cmtInBytes)

	err = c.JSON(&fiber.Map{
		"success": true,
		"message": "Comment pushed successfully",
		"comment": cmt,
	})
	if err != nil {
		c.Status(500).JSON(&fiber.Map{
			"success": false,
			"message": "Error creating product",
		})
		return err
	}

	return err
}
