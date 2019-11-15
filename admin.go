package main

import (
	"time"
	"log"
	"google.golang.org/grpc"
	"context"
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
	c := time.Tick(time.Second * time.Duration(in.IntervalSeconds))

	for {
		select {
		case <-c:

			m.mu.Lock()
			if err := stream.Send(m.Stat); err != nil {
				log.Println("send error")
				return err
			}
			log.Println("send interval ", in.IntervalSeconds ,m.ByService)
			as.publisher.RemoveObserver(&m)
			m.mu.Unlock()

			m = NewStatistic()
			as.publisher.AddObserver(&m)

			break
		case str := <-as.Manager.Counter:
			as.publisher.NotyfiObserver(str)
		default:

		}
	}
	return nil
}

func (as *AdminServerManager) AdminStreamInterceptor1(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (error) {

	if  _, ok := as.publisher.(*publish); !ok {
		as.publisher = &publish{}
	}

	err := handler(srv, ss)

	return err
}

func (as *AdminServerManager) autoCheck(ctx context.Context)  {
	defer as.manageWait.Done()

	ctx1, cancel := context.WithCancel(context.Background())
	c := time.Tick(time.Second * time.Duration(60))
LOOP:
	for {
		select {
		case <-c:
			check(ctx1, as.Manager, as.manageWait)
		case <-ctx.Done():
			log.Println("Получил отмену из server")
			cancel()
			log.Println("Отправил отмену из autoCheck")

			break LOOP
		default:


		}
	}
}

