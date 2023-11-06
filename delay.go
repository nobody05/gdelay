package gdelay

import (
	"log"
	"reflect"
	"sync"
	"time"
)

const MaxConcurrentG = 3

type DelayParam struct {
	Duration  int64
	Fun       any
	FuncParam []reflect.Value

	MethodName     string
	Obj            any
	MethodParam    []reflect.Value
	methodCallback func(any)
}

type Delay struct {
	delayEntities  map[int64][]*DelayParam
	tk             *time.Ticker
	mu             sync.Mutex
	stopChan       chan struct{}
	concurrentChan chan int
}

func NewDelay(concurrentG int) *Delay {
	if concurrentG > MaxConcurrentG {
		concurrentG = MaxConcurrentG
	}
	return &Delay{
		concurrentChan: make(chan int, concurrentG),
		stopChan:       make(chan struct{}),
	}
}

func (d *Delay) Start() {
	go func() {
		d.clearTicker()
		d.tk = time.NewTicker(time.Second)
		for {
			select {
			case <-d.stopChan:
				log.Println("ticker stop")
				return
			case <-d.tk.C:
				log.Printf("ticker")
				d.process()
			}
		}
	}()
}

func (d *Delay) Stop() {
	d.clearTicker()
	d.delayEntities = nil
}

func (d *Delay) process() {
	d.mu.Lock()
	defer func() {
		d.mu.Unlock()
	}()
	for callTime, delayList := range d.delayEntities {
		if time.Now().Unix() >= callTime {
			delete(d.delayEntities, callTime)
			curr := delayList
			for _, delay := range curr {
				currDelay := delay
				d.concurrentChan <- 1
				go func() {
					defer func() {
						if err := recover(); err != nil {
							log.Printf("%+v", err)

							d.clearTicker()
							d.stopChan <- struct{}{}
							d.Start()
						}
						<-d.concurrentChan
					}()

					if currDelay.Obj != nil && len(currDelay.MethodName) > 0 {
						obj := reflect.ValueOf(currDelay.Obj)
						m := obj.MethodByName(currDelay.MethodName)
						if !m.IsValid() {
							log.Printf("method %s invalid", currDelay.MethodName)
							return
						}
						m.Call(currDelay.MethodParam)
					} else {
						f := reflect.ValueOf(currDelay.Fun)
						if !f.IsValid() {
							log.Printf("func %+v invalid", currDelay.Fun)
							return
						}
						f.Call(currDelay.FuncParam)
					}
				}()
			}
		}
	}
}

func (d *Delay) DelayAdd(param *DelayParam) {
	d.mu.Lock()
	defer func() {
		d.mu.Unlock()
	}()

	if d.delayEntities == nil {
		d.delayEntities = make(map[int64][]*DelayParam)
	}
	_, ok := d.delayEntities[param.Duration]
	if ok {
		d.delayEntities[param.Duration] = append(d.delayEntities[param.Duration], param)
	} else {
		d.delayEntities[param.Duration] = []*DelayParam{param}
	}
	log.Printf("delay entity : %+v", d.delayEntities)
}

func (d *Delay) AddFunc(duration int64, fun any, funcParam []reflect.Value) {
	d.DelayAdd(&DelayParam{
		Duration:  duration,
		Fun:       fun,
		FuncParam: funcParam,
	})
}

func (d *Delay) AddMethod(duration int64, obj any, methodName string, methodParam []reflect.Value) {
	d.DelayAdd(&DelayParam{
		Duration:    duration,
		Obj:         obj,
		MethodName:  methodName,
		MethodParam: methodParam,
	})
}

func (d *Delay) clearTicker() {
	if d.tk != nil {
		d.tk.Stop()
		d.tk = nil
	}
}
