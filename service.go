package main

import (
	"context"
	"net"
	"log"
	"google.golang.org/grpc"
)


func StartMyMicroservice(ctx context.Context, listenAddr string, serviceChan2 serviceChan) error {

	services := <-serviceChan2

	manager, err := NewManager(services)
	if err != nil {
		return err
	}

	bizServer := NewBizServer(manager)
	adminServer := NewAdminServerManager(manager)

	ctx1,cancel := context.WithCancel(context.Background())

	check(ctx1,adminServer.Manager, manager.manageWait)
	manager.manageWait.Add(1)
	go adminServer.autoCheck(ctx1)

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
			cancel()
			log.Println("Отправил отмену из server")
			manager.manageWait.Wait()
			server.Stop()
			log.Println("stop server")

		}
	}()
	return nil
}









