# gdelay
## Delayed execution function

### Install

```go
go get github.com/nobody05/gdelay
```

### Used

```go
delay := NewDelay(3)
delay.Start()


delay.DelayAdd(&DelayParam{
    Duration: time.Now().Add(time.Second * time.Duration(r)).Unix(),
    Fun: func() {
        log.Println("hello world")
    },
})

// wait for delay execute 
for {

}

```
