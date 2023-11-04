package gdelay

import (
	"log"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

type Order struct {
}

func (o *Order) GetInfo(id string) string {
	log.Printf("order.GetInfo id %s", id)
	return "order info"
}

func (o *Order) GetList() string {
	log.Println("order.GetList")
	return "order GetList"
}

var globalDelay *Delay

func TestConcurrentDelay(t *testing.T) {
	rand.Seed(time.Now().UnixMicro())
	delay := NewDelay(3)
	delay.Start()

	for i := 0; i < 50; i++ {
		r := rand.Intn(10) + 1
		delay.DelayAdd(&DelayParam{
			Duration: time.Now().Add(time.Second * time.Duration(r)).Unix(),
			fun: func() {
				log.Println("hello world1", i)
			},
		})
	}

	for {

	}
}

func TestDelay(t *testing.T) {
	globalDelay = NewDelay(20)
	globalDelay.Start()

	globalDelay.DelayAdd(&DelayParam{
		Duration: time.Now().Add(time.Second * 3).Unix(),
		fun: func() {
			log.Println("hello world1")
		},
	})

	globalDelay.DelayAdd(&DelayParam{
		Duration: time.Now().Add(time.Second * 6).Unix(),
		fun: func() {
			log.Println("hello world panic")
			panic("inner panic")
		},
	})

	globalDelay.DelayAdd(&DelayParam{
		Duration: time.Now().Add(time.Second * 10).Unix(),
		fun: func(name string) {
			println("name: ", name)
			log.Println("hello world1")
		},
		funcParam: []reflect.Value{reflect.ValueOf("Tom")},
	})

	globalDelay.DelayAdd(&DelayParam{
		Duration: time.Now().Add(time.Second * 3).Unix(),
		fun: func() {
			time.Sleep(time.Second * 3)
			log.Println("hello world 2")
		},
	})
	globalDelay.DelayAdd(&DelayParam{
		Duration: time.Now().Add(time.Second * 3).Unix(),
		fun: func() {
			log.Println("hello world3")
		},
	})
	globalDelay.DelayAdd(&DelayParam{
		Duration:   time.Now().Add(time.Second * 5).Unix(),
		obj:        &Order{},
		methodName: "GetList",
	})

	globalDelay.DelayAdd(&DelayParam{
		Duration:    time.Now().Add(time.Second * 5).Unix(),
		obj:         &Order{},
		methodName:  "GetInfo",
		methodParam: []reflect.Value{reflect.ValueOf("12")},
	})

	globalDelay.AddFunc(time.Now().Add(time.Second).Unix(), func(name string) {
		log.Printf("get name %s", name)
	}, []reflect.Value{reflect.ValueOf("jack")})

	for {

	}
}
