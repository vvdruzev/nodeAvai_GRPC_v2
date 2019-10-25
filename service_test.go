package main

import (
	"context"
	"testing"
	"google.golang.org/grpc"
	"time"
	"runtime"
	"fmt"
	"sync"
	"io"
	"reflect"
	"log"
)

const (
	// какой адрес-порт слушать серверу
	listenAddr string = "127.0.0.1:8083"
)

func getGrpcConn(t *testing.T) *grpc.ClientConn {
	grcpConn, err := grpc.Dial(
		listenAddr,
		grpc.WithInsecure(),
	)
	if err != nil {
		t.Fatalf("cant connect to grpc: %v", err)
	}
	return grcpConn
}

// чтобы не было сюрпризов когда где-то не успела преключиться горутина и не успело что-то стортовать
func wait(amout int) {
	time.Sleep(time.Duration(amout) * 10 * time.Millisecond)
}

var s Services



func TestServerStartStop(t *testing.T) {
	s := Services{"http://www.google.com","http://youtube.com"}
	services := make(chan Services,len(s))
	services <- s
	ctx, finish := context.WithCancel(context.Background())
	err := StartMyMicroservice(ctx, listenAddr, services)
	if err != nil {
		t.Fatalf("cant start server initial: %v", err)
	}
	wait(1)
	finish() // при вызове этой функции ваш сервер должен остановиться и освободить порт
	wait(1)

	// теперь проверим что вы освободили порт и мы можем стартовать сервер ещё раз
	ctx, finish = context.WithCancel(context.Background())
	services <- s
	err = StartMyMicroservice(ctx, listenAddr,services)
	if err != nil {
		t.Fatalf("cant start server again: %v", err)
	}
	wait(1)
	finish()
	wait(1)
}


func TestServerLeak(t *testing.T) {
	//return
	goroutinesStart := runtime.NumGoroutine()
	TestServerStartStop(t)
	goroutinesPerTwoIterations := runtime.NumGoroutine() - goroutinesStart
	log.Println("NumGoroutine ",runtime.NumGoroutine(), "goroutinesStart ", goroutinesStart,  "goroutinesPerTwoIterations ", goroutinesPerTwoIterations)
	goroutinesStart = runtime.NumGoroutine()
	goroutinesStat := []int{}
	for i := 0; i <= 25; i++ {
		TestServerStartStop(t)
		goroutinesStat = append(goroutinesStat, runtime.NumGoroutine())
	}
	goroutinesPerFiftyIterations := runtime.NumGoroutine() - goroutinesStart
	if goroutinesPerFiftyIterations > goroutinesPerTwoIterations*5 {
		t.Fatalf("looks like you have goroutines leak: %+v", goroutinesStat)
	}
}



func TestServerStartStop2(t *testing.T) {
	s := Services{"http://www.google.com","http://ya.ru","http://facebook.com","http://baidu.com", "http://wikipedia.org", "http://qq.com"}
	services := make(chan Services, len(s))
	services <- s
	ctx, finish := context.WithCancel(context.Background())
	err := StartMyMicroservice(ctx, listenAddr, services)
	if err != nil {
		t.Fatalf("cant start server initial: %v", err)
	}
	wait(700)
	finish() // при вызове этой функции ваш сервер должен остановиться и освободить порт

}

func TestBizClient_Check(t *testing.T) {
	s := Services{"http://www.google.com","http://ya.ru","http://facebook.com","http://baidu.com", "http://wikipedia.org", "http://qq.com"}
	services := make(chan Services, len(s))
	services <- s
	ctx, finish := context.WithCancel(context.Background())
	defer finish()
	err := StartMyMicroservice(ctx, listenAddr, services)
	if err != nil {
		t.Fatalf("cant start server initial: %v", err)
	}

	conn := getGrpcConn(t)
	defer conn.Close()

	biz := NewBizClient(conn)
	adm := NewAdminClient(conn)
	log.Println(adm)
	ctx = context.Background()
	request := &Request{Service:"http://www.google.com"}

	wait(200)
	resp,err := biz.Check(ctx,request)
	if resp == nil {
		t.Fatalf("Error1. Expected : %s , got : %v ", "http://www.google.com", nil)
	}

	if key,ok := resp.Service["http://www.google.com"]; !ok {
		t.Fatalf("Error2. Expected : %s , got : %v ", "http://www.google.com", key)
	}
	fmt.Println(resp.Service)
	finish() // при вызове этой функции ваш сервер должен остановиться и освободить порт

}

func TestBizServerManager_Max(t *testing.T) {
	s := Services{"http://www.google.com","http://ya.ru","http://facebook.com","http://baidu.com", "http://wikipedia.org", "http://qq.com"}
	services := make(chan Services, len(s))
	services <- s
	ctx, finish := context.WithCancel(context.Background())
	defer finish()
	err := StartMyMicroservice(ctx, listenAddr, services)
	if err != nil {
		t.Fatalf("cant start server initial: %v", err)
	}

	conn := getGrpcConn(t)
	defer conn.Close()

	biz := NewBizClient(conn)
	//adm := NewAdminClient(conn)

	ctx = context.Background()
	//request := &Request{Service:"http://www.google.com"}

	wait(200)
	resp,err := biz.Max(ctx,&Nothing{})

	fmt.Println("1",resp.Service)

	resp,err = biz.Max(ctx,&Nothing{})

	fmt.Println("2",resp.Service)


	resp,err = biz.Max(ctx,&Nothing{})

	fmt.Println("3",resp.Service)

	wait(700)
	finish() // при вызове этой функции ваш сервер должен остановиться и освободить порт

}

func TestBizServerManager_Min(t *testing.T) {
	s := Services{"http://www.google.com","http://ya.ru","http://facebook.com","http://baidu.com", "http://wikipedia.org", "http://qq.com"}
	services := make(chan Services, len(s))
	services <- s
	ctx, finish := context.WithCancel(context.Background())
	defer finish()
	err := StartMyMicroservice(ctx, listenAddr, services)
	if err != nil {
		t.Fatalf("cant start server initial: %v", err)
	}

	conn := getGrpcConn(t)
	defer conn.Close()

	biz := NewBizClient(conn)
	//adm := NewAdminClient(conn)

	ctx = context.Background()
	//request := &Request{Service:"http://www.google.com"}

	wait(200)
	resp,err := biz.Min(ctx,&Nothing{})

	fmt.Println("1",resp.Service)

	resp,err = biz.Min(ctx,&Nothing{})

	fmt.Println("2",resp.Service)


	resp,err = biz.Min(ctx,&Nothing{})

	fmt.Println("3",resp.Service)

	wait(700)
	finish() // при вызове этой функции ваш сервер должен остановиться и освободить порт
}

func TestAdminServerManager_Statistics(t *testing.T) {
	s := Services{"http://www.google.com","http://ya.ru","http://facebook.com","http://baidu.com", "http://wikipedia.org", "http://qq.com"}
	services := make(chan Services, len(s))
	services <- s
	ctx, finish := context.WithCancel(context.Background())
	defer finish()
	err := StartMyMicroservice(ctx, listenAddr, services)
	if err != nil {
		t.Fatalf("cant start server initial: %v", err)
	}

	conn := getGrpcConn(t)
	defer conn.Close()

	biz := NewBizClient(conn)
	adm := NewAdminClient(conn)
	wait(100)

	ctx = context.Background()
	//request := &Request{Service:"http://www.google.com"}
	statStream1, err := adm.Statistics(ctx, &StatInterval{IntervalSeconds: 2})
	wait(1)
	statStream2, err := adm.Statistics(ctx, &StatInterval{IntervalSeconds: 3})

	mu := &sync.Mutex{}
	stat1 := &Stat{}
	stat2 := &Stat{}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			stat, err := statStream1.Recv()
			if err != nil && err != io.EOF {
				//fmt.Printf("unexpected error %v\n", err)
				return
			} else if err == io.EOF {
				break
			}
			//log.Println("stat1", stat, err)
			mu.Lock()
			stat1 = stat
			stat1.Timestamp = 0
			mu.Unlock()
		}
	}()
	go func() {
		for {
			stat, err := statStream2.Recv()
			if err != nil && err != io.EOF {
				// fmt.Printf("unexpected error %v\n", err)
				return
			} else if err == io.EOF {
				break
			}
			//log.Println("stat2", stat, err)
			mu.Lock()
			stat2 = stat
			stat2.Timestamp = 0
			mu.Unlock()
		}
	}()

	wait(1)
	request := &Request{Service:"http://www.google.com"}


	biz.Check(ctx,request)
	biz.Check(ctx,request)
	biz.Check(ctx,request)

	wait(200) // 2 sec
	expectedStat1 := &Stat{
		Timestamp: 0,
		ByService: map[string]uint64{
			"http://www.google.com":        3,
		},
	}

	mu.Lock()
	if !reflect.DeepEqual(stat1, expectedStat1) {
		t.Fatalf("stat1-1 dont match\nhave %+v\nwant %+v", stat1, expectedStat1)
	}
	mu.Unlock()

	request = &Request{Service:"http://ya.ru"}
	biz.Check(ctx,request)

	wait(220) // 2+ sec
	expectedStat1 = &Stat{
		Timestamp: 0,
		ByService: map[string]uint64{
			"http://ya.ru":1,
		},
	}
	expectedStat2 := &Stat{
		Timestamp: 0,
		ByService: map[string]uint64{
			"http://www.google.com":        3,
			"http://ya.ru":1,
		},
	}

	mu.Lock()
	if !reflect.DeepEqual(stat1, expectedStat1) {
		t.Fatalf("stat1-2 dont match\nhave %+v\nwant %+v", stat1, expectedStat1)
	}
	if !reflect.DeepEqual(stat2, expectedStat2) {
		t.Fatalf("stat2 dont match\nhave %+v\nwant %+v", stat2, expectedStat2)
	}
	mu.Unlock()

	finish()


}
