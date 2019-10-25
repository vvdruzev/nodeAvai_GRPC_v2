package main

import (
	"context"
	"net"
	"log"
	"google.golang.org/grpc"
	"sync"
	"time"
	"sort"
	"net/http"
	"fmt"
	"net/url"
)

type AdminServerManager struct {
	*Manager
}

func NewAdminServerManager(manager *Manager) (*AdminServerManager) {
	return &AdminServerManager{Manager: manager}
}

func (as *AdminServerManager) Statistics(in *StatInterval, stream Admin_StatisticsServer) error {

	m := NewStatistic()
	as.publisher.AddObserver(&m)
	as.Manager.Statistics = append(as.Manager.Statistics,m)
	c := time.Tick(time.Second * time.Duration(in.IntervalSeconds))
	for {
		select {
		case <-c:
			m.mu.Lock()
			if err := stream.Send(m.Stat); err != nil {
				log.Println("senderror")
				return err
			}
			log.Println("send interval ", in.IntervalSeconds ,m.ByService)
			as.publisher.RemoveObserver(&m)
			m.mu.Unlock()

			m = NewStatistic()
			as.publisher.AddObserver(&m)

			break
		case str := <-as.Manager.Counter:
			for _,v := range as.ObserverList{
				v.add(str)
			}
			//log.Println("receive from counter",str, "interval " , in.IntervalSeconds)
			//m.add(str)
		default:

		}
	}
	return nil
}
type Statistic struct {
	*Stat
	mu 	*sync.Mutex
}

func NewStatistic() Statistic {
	return Statistic{Stat:&Stat{ByService: map[string]uint64{}},mu:&sync.Mutex{}}
}

func (s *Statistic) add(str string)  {
	s.mu.Lock()
	s.Stat.ByService[str]++
	s.mu.Unlock()
}


func (as *AdminServerManager) AdminStreamInterceptor1(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (error) {
	as.once.Do(func() {
		as.Counter = make(chan string)
	})
	err := handler(srv, ss)

	return err
}

type service struct {
	name string
	reaction time.Duration
	status int
}
const https_proxy="http://vvdruze1:Kiasoul$0898@bproxy.ug.mts.ru:3131/"

func setProxy(c *http.Client)  {
	proxyURL, err := url.Parse(https_proxy)

	if err != nil {
		log.Println(err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	c.Transport = transport
	//adding the Transport object to the http Client
	c.Timeout = 1*time.Second
}

func check(as *Manager)  {
	cli := &http.Client{}
	setProxy(cli)

	out := make(chan service, len(as.ServicesStatus.ServicesStatus))
	defer close(out)
	for k,_ := range as.ServicesStatus.ServicesStatus{
		go func(k string) {
			var s service
			s.name = k
			start := time.Now()
			resp, _ := cli.Get(k)
			if resp == nil {
				s.status = http.StatusNotFound
				s.reaction = 0
				out <- s
				return
			}
			s.reaction = time.Since(start)
			s.status = resp.StatusCode
			out <- s
		}(k)
	}

	for i :=0 ; i < len(as.ServicesStatus.ServicesStatus); i++{
		as.mu.Lock()
		s := <- out
		as.ServicesStatus.ServicesStatus[s.name] = s
		as.mu.Unlock()
	}
	as.mu.Lock()
	as.min()
	as.max()
	fmt.Println(as.ServicesStatus.ServicesStatus)
	fmt.Printf("max = %s :  %s ,   min = %s : %s\n\n", as.GetMax().name,as.GetMax().reaction, as.GetMin().name, as.GetMin().reaction)
	as.mu.Unlock()
}

func (as *AdminServerManager) autoCheck()  {
	c := time.Tick(time.Second * time.Duration(60))
	for {
		select {
		case <-c:
			check(as.Manager)
		case <-as.done:
			return
		default:


		}
	}
}

type BizServerManager struct {
	*Manager
}

func NewBizServer(manager *Manager) *BizServerManager {
	return &BizServerManager{Manager: manager}
}

func (bz *BizServerManager) Check(ctx context.Context, in *Request) (*Response, error) {
	if bz.Counter != nil {
		bz.Counter <- in.Service
	}

	bz.mu.Lock()
	s := bz.ServicesStatus.ServicesStatus[in.Service]
	bz.mu.Unlock()

	m := make(map[string]int64)
	m[s.name] = s.reaction.Nanoseconds()
	return &Response{Service:m}, nil
}

func (bz *BizServerManager) Min(ctx context.Context, in *Nothing) (*Response, error) {
	bz.mu.Lock()
	s := bz.ServicesStatus.GetMin()
	bz.mu.Unlock()
	m := make(map[string]int64)
	m[s.name] = s.reaction.Nanoseconds()

	return &Response{Service:m}, nil
}

func (bz *BizServerManager) Max(ctx context.Context, in *Nothing) (*Response, error) {
	bz.mu.Lock()
	s := bz.ServicesStatus.GetMax()
	bz.mu.Unlock()
	m := make(map[string]int64)
	m[s.name] = s.reaction.Nanoseconds()
	return &Response{Service:m}, nil
}

func (bz *BizServerManager) BizUnaryInterceptor1(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler, ) (interface{}, error) {
	reply, err := handler(ctx, req)

	return reply, err
}

type Services []string

type serviceChan chan Services

func StartMyMicroservice(ctx context.Context, listenAddr string, serviceChan2 serviceChan) error {

	services := <-serviceChan2

	manager, err := NewManager(services)
	if err != nil {
		return err
	}





	bizServer := NewBizServer(manager)
	adminServer := NewAdminServerManager(manager)
	check(adminServer.Manager)
	//go adminServer.autoCheck()
	go func() {
		lis, err := net.Listen("tcp", listenAddr)
		if err != nil {
			log.Fatalln("cant listet port", err)
		}
		server := grpc.NewServer(grpc.UnaryInterceptor(bizServer.BizUnaryInterceptor1), grpc.StreamInterceptor(adminServer.AdminStreamInterceptor1))
		RegisterBizServer(server, bizServer)
		RegisterAdminServer(server, adminServer)
		go server.Serve(lis)
		//fmt.Println("starting server at " + listenAddr)
		defer server.Stop()
		defer lis.Close()

		select {
		case <-ctx.Done():
			adminServer.done <- struct{}{}
			server.Stop()
		}
	}()
	return nil
}



type Manager struct {
	Event      *Event
	ServicesStatus
	mu         *sync.Mutex
	count      int32
	done chan struct{}
	Counter chan string
	Statistics []Statistic
	*publisher
	once sync.Once
}

type publisher struct {
	ObserverList []Observer
}

type Observer interface {
	add(string)
}

func (s *publisher) AddObserver(o Observer)  {
	s.ObserverList = append(s.ObserverList, o)
}

func (s *publisher) RemoveObserver(o Observer)  {
	var indexToRemove  int

	for i, observer := range s.ObserverList {
		if observer == o {
			indexToRemove = i
			break
		}
	}

	s.ObserverList = append(s.ObserverList[:indexToRemove], s.ObserverList[indexToRemove+1:]...)
}

func (s *publisher) NotyfiOservber(m string)  {

}

type ServicesStatus struct {
	ServicesStatus map[string]service
	mintime service
	maxtime service
}

func (s *ServicesStatus) min() () {
	var values []time.Duration
	for _,v := range s.ServicesStatus {
		if v.status == 200 {
		values = append(values,v.reaction)
	}
	}
	sort.SliceStable(values, func(i, j int) bool {
		return values[i] < values[j]
	})
	max := values[0]
	for _,v := range s.ServicesStatus {
		if v.reaction == max && v.status == 200 {
			s.mintime = v
			break
		}
	}

}

func (s *ServicesStatus) max() () {
	var values []time.Duration
	for _,v := range s.ServicesStatus {
		if v.status == 200 {
			values = append(values,v.reaction)
		}
	}

	sort.SliceStable(values, func(i, j int) bool {
		return values[i] < values[j]
	})
	min := values[len(values)-1]
	for _,v := range s.ServicesStatus {
		if v.reaction == min && v.status == 200 {
			s.maxtime = v
			break
		}
	}
}

func (s *ServicesStatus) GetMin() (service) {
	return s.mintime
}

func (s *ServicesStatus) GetMax() (service) {
	return s.maxtime
}



func NewManager(services Services) (*Manager, error) {
	m := make(map[string]service)
	for _,serv := range services{
		m[serv] = service{name:serv,reaction:0,status:404}
	}
	return &Manager{ServicesStatus:ServicesStatus{ServicesStatus:m,maxtime:service{}, mintime:service{}},mu:&sync.Mutex{},done:make(chan struct{}), publisher:&publisher{}}, nil
}

