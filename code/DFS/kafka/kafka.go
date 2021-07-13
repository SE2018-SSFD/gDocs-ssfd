package kafka

import (
	"DFS/util"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

type consumerGroupHandler struct {
	name string
}
var kafkaHosts = []string{
	"kafka1:9092",
	"kafka2:9093",
	"kafka3:9094",
}


func (consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	filename := path.Join(string(h.name), "log.dat")
	fd, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		logrus.Warnf("OpenFile error :%s\n", err)
		return err
	}
	defer fd.Close()
	for msg := range claim.Messages() {
		//fmt.Printf("%s Message topic:%q partition:%d offset:%d  value:%s\n",h.name, msg.Topic, msg.Partition, msg.Offset, string(msg.Value))
		var ml util.MasterLog
		json.Unmarshal(msg.Value, &ml)
		logrus.Debugf("AppendLog : %d", ml.OpType)
		enc := json.NewEncoder(fd)
		logrus.Infoln(ml)
		err = enc.Encode(ml)
		if err != nil {
			logrus.Warnf("AppendLogError :%s", err)
			return err
		}
		// 手动确认消息
		sess.MarkMessage(msg, "")
	}
	return nil
}

func handleErrors(group *sarama.ConsumerGroup) {
	for err := range (*group).Errors() {
		fmt.Println("ERROR", err)
	}
}

func Consume(group *sarama.ConsumerGroup, name string) error {
	fmt.Println(name + "start")
	ctx := context.Background()
	for {
		topics := []string{util.MasterTopicName}
		handler := consumerGroupHandler{name: name}
		err := (*group).Consume(ctx, topics, handler)
		if err != nil {
			logrus.Debug(err)
			return err
		}
	}
}

func MakeConsumerGroup(MastAddr string) (*sarama.ConsumerGroup, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = false
	config.Version = sarama.V0_11_0_2
	brokers := kafkaHosts
	cg, err := sarama.NewConsumerGroup(brokers, "testMaster"+MastAddr, config)
	if err != nil {
		logrus.Fatal(err)
		return nil, err
	}
	return &cg, err
}

func MakeProducer(MastAddr string) (*sarama.AsyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Version = sarama.V0_11_0_2

	producer, err := sarama.NewAsyncProducer(kafkaHosts, config)
	if err != nil {
		fmt.Printf("create producer error :%s\n", err.Error())
		return nil, err
	}
	return &producer, nil
}
