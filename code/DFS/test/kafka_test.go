package test

import (
	"DFS/kafka"
	"DFS/util"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
	"strconv"
	"sync"
	"testing"
	"time"
)
//func InitKafkaTest() (c *client.Client,m []*master.Master,csList []*chunkserver.ChunkServer){
//	logrus.SetLevel(logrus.DebugLevel)
//	//delete old ckp
//	util.DeleteFile("../log/checkpoint.dat")
//	//delete old log
//	util.DeleteFile("../log/log.dat")
//	// Init client
//	c = client.InitClient(util.CLIENTADDR, util.MASTER1ADDR)
//	go func(){c.Serve()}()
//
//	// Register some virtual chunkServers
//	_, err := os.Stat("cs")
//	if err == nil {
//		err := os.RemoveAll("cs")
//		if err != nil {
//			logrus.Fatalf("mkdir %v error\n", "cs")
//		}
//	}
//	err = os.Mkdir("cs", 0755)
//	if err != nil {
//		logrus.Fatalf("mkdir %v error\n", "cs")
//	}
//	csList = make([]*chunkserver.ChunkServer, 5)
//	for index,port := range []int{3000,3001,3002,3003,3004}{
//		addr := util.Address("127.0.0.1:" + strconv.Itoa(port))
//		csList[index] = chunkserver.InitChunkServer(string(addr), "cs/cs"+strconv.Itoa(port),  util.MASTER1ADDR)
//		//util.AssertNil(t,err)
//	}
//
//	m = make([]*master.Master,3)
//	m[0],_ = master.InitMaster(util.MASTER1ADDR, "../log1")
//	m[1],_ = master.InitMaster(util.MASTER2ADDR, "../log2")
//	m[2],_ = master.InitMaster(util.MASTER3ADDR, "../log3")
//	for i:= 0;i < 3;i++ {
//		go func(idx int){m[idx].Serve()}(i)
//	}
//	time.Sleep(500*time.Millisecond)
//	return
//}
//
//func TestLog(t *testing.T) {
//	logrus.SetLevel(logrus.DebugLevel)
//	// Init master and client
//	client,m,cs := InitKafkaTest()
//}


func TestKafka(t *testing.T) {
	var group []*sarama.ConsumerGroup = make([]*sarama.ConsumerGroup,3)
	var masterAddrs []string
	masterAddrs = append(masterAddrs,util.MASTER1ADDR)
	masterAddrs = append(masterAddrs,util.MASTER2ADDR)
	masterAddrs = append(masterAddrs,util.MASTER3ADDR)
	var err error
	for i:= 0;i < 3;i++ {
		go func(idx int) {
			group[idx],err = kafka.MakeConsumerGroup(masterAddrs[idx])
			if err != nil {
				logrus.Debug(err)
				t.Fail()
			}
			err = kafka.Consume(group[idx],"testMaster"+masterAddrs[idx])
		}(i)
	}

	time.Sleep(1* time.Second)
	producer,err:= kafka.MakeProducer(masterAddrs[0])
	if err != nil {
		logrus.Debug(err)
		t.Fail()
	}
	msg := &sarama.ProducerMessage{
		Topic: "my_topic1",
		Key:   sarama.StringEncoder("go_test"),
	}
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		for i:= 0;i < 3;i++ {
			value := "this is message" + strconv.Itoa(i)
			msg.Value = sarama.ByteEncoder(value)
			fmt.Printf("input [%s]\n", value)
			(*producer).Input() <- msg

			select {
			case suc := <-(*producer).Successes():
				fmt.Printf("offset: %d,  timestamp: %s\n", suc.Offset, suc.Timestamp.String())
			case fail := <-(*producer).Errors():
				fmt.Printf("err: %s\n", fail.Err.Error())
			}
			wg.Done()
		}
	}()
	wg.Wait()
	time.Sleep(1* time.Second)
	(*group[1]).Close()

	wg.Add(3)
	go func() {
		for i:= 3;i < 6;i++ {
			value := "this is message" + strconv.Itoa(i)
			msg.Value = sarama.ByteEncoder(value)
			fmt.Printf("input [%s]\n", value)
			(*producer).Input() <- msg

			select {
			case suc := <-(*producer).Successes():
				fmt.Printf("offset: %d,  timestamp: %s\n", suc.Offset, suc.Timestamp.String())
			case fail := <-(*producer).Errors():
				fmt.Printf("err: %s\n", fail.Err.Error())
			}
			wg.Done()
		}
	}()
	wg.Wait()
	group[1],err = kafka.MakeConsumerGroup(masterAddrs[1])
	if err != nil {
		logrus.Debug(err)
		t.Fail()
	}
	go kafka.Consume(group[1],"testMaster"+masterAddrs[1])
	for i:= 6;i < 9;i++ {
		value := "this is message" + strconv.Itoa(i)
		msg.Value = sarama.ByteEncoder(value)
		fmt.Printf("input [%s]\n", value)
		(*producer).Input() <- msg

		select {
		case suc := <-(*producer).Successes():
			fmt.Printf("offset: %d,  timestamp: %s\n", suc.Offset, suc.Timestamp.String())
		case fail := <-(*producer).Errors():
			fmt.Printf("err: %s\n", fail.Err.Error())
		}
	}

	time.Sleep(1 * time.Second)
	for i := 0;i < 3;i++ {
		(*group[i]).Close()
	}
	(*producer).Close()
}
