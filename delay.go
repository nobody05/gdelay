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
	fun       any
	funcParam []reflect.Value

	methodName     string
	obj            any
	methodParam    []reflect.Value
	methodCallback func(any)
}

type Delay struct {
	delayEntries   map[int64][]*DelayParam
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
	if d.tk != nil {
		d.tk.Stop()
	}
	d.delayEntries = nil
}

func (d *Delay) process() {
	d.mu.Lock()
	defer func() {
		d.mu.Unlock()
	}()
	for callTime, delayList := range d.delayEntries {
		if time.Now().Unix() >= callTime {
			delete(d.delayEntries, callTime)
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

					if currDelay.obj != nil && len(currDelay.methodName) > 0 {
						obj := reflect.ValueOf(currDelay.obj)
						m := obj.MethodByName(currDelay.methodName)
						if !m.IsValid() {
							log.Printf("method %s invalid", currDelay.methodName)
							return
						}
						m.Call(currDelay.methodParam)
					} else {
						f := reflect.ValueOf(currDelay.fun)
						if !f.IsValid() {
							log.Printf("func %s invalid", currDelay.methodName)
							return
						}
						f.Call(currDelay.funcParam)
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

	if d.delayEntries == nil {
		d.delayEntries = make(map[int64][]*DelayParam)
	}
	_, ok := d.delayEntries[param.Duration]
	if ok {
		d.delayEntries[param.Duration] = append(d.delayEntries[param.Duration], param)
	} else {
		d.delayEntries[param.Duration] = []*DelayParam{param}
	}
	log.Printf("delay entry : %+v", d.delayEntries)
}

func (d *Delay) AddFunc(duration int64, fun any, funcParam []reflect.Value) {
	d.DelayAdd(&DelayParam{
		Duration:  duration,
		fun:       fun,
		funcParam: funcParam,
	})
}

func (d *Delay) AddMethod(duration int64, obj any, methodName string, methodParam []reflect.Value) {
	d.DelayAdd(&DelayParam{
		Duration:    duration,
		obj:         obj,
		methodName:  methodName,
		methodParam: methodParam,
	})
}

func (d *Delay) clearTicker() {
	if d.tk != nil {
		d.tk.Stop()
		d.tk = nil
	}
}
