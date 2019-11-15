package main

import (
	"sync"
	"time"
	"sort"
)

type Manager struct {
	ServicesStatus
	mu      *sync.Mutex
	count   int32
	Counter chan string
	publisher
	manageWait *sync.WaitGroup
}

func NewManager(services Services) (*Manager, error) {
	m := make(map[string]service)
	for _, serv := range services {
		m[serv] = service{name: serv, reaction: 0, status: 404}
	}
	return &Manager{
		ServicesStatus: ServicesStatus{ServicesStatus: m,
			maxtime: service{},
			mintime: service{},
		},
		mu:        &sync.Mutex{},
		publisher: &emptyPublish{},
		manageWait:&sync.WaitGroup{},
		Counter:make(chan string),
	},
		nil
}

type ServicesStatus struct {
	ServicesStatus map[string]service
	mintime        service
	maxtime        service
}

func (s *ServicesStatus) min() () {
	var values []time.Duration
	if len(s.ServicesStatus) == 0 {
		return
	}
	for _, v := range s.ServicesStatus {
		if v.status == 200 {
			values = append(values, v.reaction)
		}
	}
	sort.SliceStable(values, func(i, j int) bool {
		return values[i] < values[j]
	})
	if len(values) == 0 {
		return
	}

	max := values[0]
	for _, v := range s.ServicesStatus {
		if v.reaction == max && v.status == 200 {
			s.mintime = v
			break
		}
	}

}

func (s *ServicesStatus) max() () {
	var values []time.Duration
	if len(s.ServicesStatus) == 0 {
		return
	}

	for _, v := range s.ServicesStatus {
		if v.status == 200 {
			values = append(values, v.reaction)
		}
	}

	sort.SliceStable(values, func(i, j int) bool {
		return values[i] < values[j]
	})
	if len(values) == 0 {
		return
	}

	min := values[len(values)-1]
	for _, v := range s.ServicesStatus {
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

type Services []string

type serviceChan chan Services

type Statistic struct {
	*Stat
	mu *sync.Mutex
}

func NewStatistic() Statistic {
	return Statistic{Stat: &Stat{ByService: map[string]uint64{}}, mu: &sync.Mutex{}}
}

func (s *Statistic) Notify(str string) {
	s.mu.Lock()
	s.Stat.ByService[str]++
	s.mu.Unlock()
}

type service struct {
	name     string
	reaction time.Duration
	status   int
}
