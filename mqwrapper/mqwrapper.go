// Copyright 2018 Hurricanezwf. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mqwrapper

import (
	"errors"
	"fmt"
	"time"

	"github.com/Hurricanezwf/rabbitmq-go/mq"
)

type Config struct {
	// Ready 协调各组件之间的状态，当一切就绪之后，Ready将可读，此时该组件开始对外释放能力
	Ready <-chan struct{}

	MQUrl string

	// 生产者配置
	EnableProducer       bool
	ProducerExchange     string
	ProducerExchangeKind string
	ProducerQueue        string
	ProducerRouteKey     string

	// 消费者配置
	EnableConsumer       bool
	ConsumerExchange     string
	ConsumerExchangeKind string
	ConsumerQueue        string
	ConsumerRouteKey     string

	// 日志写入，如果为空将不记录日志
	Debug LogWriter
	Info  LogWriter
	Warn  LogWriter
	Error LogWriter
}

type MQWrapper struct {
	// Wrapper的ID
	id string

	// conf 配置
	conf *Config

	// MQ实例
	m *mq.MQ

	// MQ生产者
	producer *mq.Producer

	// MQ消费者
	consumer *mq.Consumer

	// MQ消息接收通道
	delivery chan mq.Delivery

	// MQ消息处理入口
	handlers map[int32]MsgHandler

	// MQ编解码器
	encoder MsgEncoder

	// 消费者控制开关
	blockConsumeCh chan struct{}
	stopConsumeCh  chan struct{}
}

type MsgHandler func(actionKey int32, msg []byte) error

func New() *MQWrapper {
	return &MQWrapper{
		handlers: make(map[int32]MsgHandler),
		encoder:  DefaultEncoder(),
	}
}

func (w *MQWrapper) Open(wrapperId string, conf *Config) error {
	var (
		err      error
		queue    *mq.MQ
		producer *mq.Producer
		consumer *mq.Consumer

		exbForProducer = []*mq.ExchangeBinds{
			{
				Exch: mq.DefaultExchange(conf.ProducerExchange, conf.ProducerExchangeKind),
				Bindings: []*mq.Binding{
					{
						RouteKey: conf.ProducerRouteKey,
						Queues: []*mq.Queue{
							mq.DefaultQueue(conf.ProducerQueue),
						},
					},
				},
			},
		}

		exbForConsumer = []*mq.ExchangeBinds{
			{
				Exch: mq.DefaultExchange(conf.ConsumerExchange, conf.ConsumerExchangeKind),
				Bindings: []*mq.Binding{
					{
						RouteKey: conf.ConsumerRouteKey,
						Queues: []*mq.Queue{
							mq.DefaultQueue(conf.ConsumerQueue),
						},
					},
				},
			},
		}
	)

	w.id = wrapperId
	w.conf = conf

	// 校验配置
	if err = w.ValidateConf(conf); err != nil {
		goto FINISH
	}

	// 连接MQ
	if queue, err = mq.New(conf.MQUrl).Open(); err != nil {
		goto FINISH
	}
	w.m = queue

	if conf.EnableProducer {
		// 新建Producer会话
		if producer, err = queue.Producer("ProduceTo"); err != nil {
			goto FINISH
		}
		if err = producer.SetExchangeBinds(exbForProducer).Open(); err != nil {
			goto FINISH
		}
		w.producer = producer
	}

	if conf.EnableConsumer {
		// 新建Consumer会话
		if consumer, err = queue.Consumer("ConsumeFrom"); err != nil {
			goto FINISH
		}

		w.delivery = make(chan mq.Delivery, 8)
		w.stopConsumeCh = make(chan struct{})

		if err = consumer.SetExchangeBinds(exbForConsumer).SetQos(1).SetMsgCallback(w.delivery).Open(); err != nil {
			goto FINISH
		}
		w.consumer = consumer

		// 循环消费
		go w.consumeFromLoop()
	}

FINISH:
	if err != nil {
		w.Close()
	}

	return err
}

func (w *MQWrapper) Close() error {
	if w.m != nil {
		w.m.Close()
	}
	if w.stopConsumeCh != nil {
		select {
		case <-w.stopConsumeCh:
			// do nothing
		default:
			close(w.stopConsumeCh)
		}
	}
	return nil
}

func (w *MQWrapper) RegistActionHandler(actionKey int32, f MsgHandler) error {
	if w.handlers == nil {
		w.handlers = make(map[int32]MsgHandler)
	}

	if f == nil {
		return fmt.Errorf("Msg handler for '%d' is nil", actionKey)
	}
	if _, exist := w.handlers[actionKey]; exist {
		return fmt.Errorf("Msg handler for '%d' had been existed", actionKey)
	}
	w.handlers[actionKey] = f
	return nil
}

func (w *MQWrapper) findHandler(actionKey int32) MsgHandler {
	return w.handlers[actionKey]
}

func (w *MQWrapper) Post(actionKey int32, msg []byte, retry int) error {
	if w.conf.EnableProducer == false {
		return errors.New("Producer is disabled")
	}

	<-w.conf.Ready

	b, err := w.encoder.Encode(actionKey, msg)
	if err != nil {
		return err
	}
	if w.conf.Debug != nil {
		w.conf.Debug.Println("%s post msg: %#v\nTotal: %dBytes", w.id, b, len(b))
	}

	mqMsg := mq.NewPublishMsg(b)
	mqMsg.ContentType = "application/octet-stream"

	for i := 0; i < retry+1; i++ {
		err = w.producer.Publish(w.conf.ProducerExchange, w.conf.ProducerRouteKey, mqMsg)
		if err == nil {
			break
		}
		if w.conf.Warn != nil {
			w.conf.Warn.Println("(%d) Try to post msg to proxy failed, %v", i, err)
		}
		time.Sleep(time.Second)
	}

	return err
}

func (w *MQWrapper) ValidateConf(conf *Config) error {
	if conf.Ready == nil {
		return errors.New("Ready flag in config is nil")
	}

	if len(conf.MQUrl) <= 0 {
		return errors.New("Missing 'MQUrl'")
	}

	if conf.EnableProducer {
		if len(conf.ProducerExchange) <= 0 {
			return errors.New("Missing 'ProducerExchange'")
		}
		if len(conf.ProducerQueue) <= 0 {
			return errors.New("Missing 'ProducerQueue'")
		}
		if len(conf.ProducerRouteKey) <= 0 {
			return errors.New("Missing 'ProdicerRoutekey'")
		}
		if len(conf.ProducerExchangeKind) <= 0 {
			return errors.New("Missing 'ProducerExchangeKind'")
		}
	}

	if conf.EnableConsumer {
		if len(conf.ConsumerExchange) <= 0 {
			return errors.New("Missin 'ConsumerExchange'")
		}
		if len(conf.ConsumerQueue) <= 0 {
			return errors.New("Missing 'ConsumerQueue'")
		}
		if len(conf.ConsumerRouteKey) <= 0 {
			return errors.New("Missing 'ConsumerRouteKey'")
		}
		if len(conf.ConsumerExchangeKind) <= 0 {
			return errors.New("Missing 'ConsumerExchangeKind'")
		}
	}
	return nil
}

func (w *MQWrapper) SetEncoder(e MsgEncoder) {
	if e != nil {
		w.encoder = e
	}
}

func (w *MQWrapper) consumeFromLoop() {
	defer close(w.delivery)

	<-w.conf.Ready

	for {
		select {
		case <-w.stopConsumeCh:
			return
		case d, ok := <-w.delivery:
			if !ok {
				return
			}
			go w.handleMsg(d)
		}
	}
}

func (w *MQWrapper) handleMsg(d mq.Delivery) {
	defer d.Ack(false)

	if w.conf.Debug != nil {
		w.conf.Debug.Println("%s receive msg: %#v\nTotal: %d bytes\n", w.id, d.Body, len(d.Body))
	}

	actionKey, msgBody, err := w.encoder.Decode(d.Body)
	if err != nil {
		if w.conf.Warn != nil {
			w.conf.Warn.Println(err.Error())
		}
		return
	}

	//glog.V(5).Infof("%s: Start handle '%s'", w.id, actionKey.String())

	if h := w.findHandler(actionKey); h == nil {
		if w.conf.Warn != nil {
			w.conf.Warn.Println("%s: No handler found for action '%d'", w.id, actionKey)
		}
	} else {
		h(actionKey, msgBody)
	}
}

type LogWriter interface {
	Println(format string, v ...interface{})
}
