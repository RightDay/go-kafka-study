package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
)

func main() {
	topic := "comments"
	// kafka broker 주소 연결
	worker, err := connectConsumer([]string{"localhost:29092"})
	if err != nil {
		panic(err)
	}

	// 파티션 컨슈머 시작
	consumer, err := worker.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		panic(err)
	}

	fmt.Println("Consumer started ")

	//종료 신호 처리 채널 생성
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// 처리된 메세지 수 카운팅
	msgCount := 0

	// 종료 신호까지 메세지 처리
	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case err := <-consumer.Errors():
				fmt.Println(err) // 에러 발생시 출력
			case msg := <-consumer.Messages():
				msgCount++
				fmt.Printf("Received message Count %d: | Topic(%s) | Message(%s) \n", msgCount, string(msg.Topic), string(msg.Value))
			case <-sigchan:
				fmt.Println("Interrupt is detected")
				doneCh <- struct{}{}
			}
		}
	}()

	// 종료 신호 대기
	<-doneCh

	// 처리된 메세지 수 출력
	fmt.Println("Processed", msgCount, "messages")

	// Consumer 종료
	if err := worker.Close(); err != nil {
		panic(err)
	}

}

func connectConsumer(brokersUrl []string) (sarama.Consumer, error) {
	// Consumer 설정 생성
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// Consumer 연결 생성
	conn, err := sarama.NewConsumer(brokersUrl, config)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
