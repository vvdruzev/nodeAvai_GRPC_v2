package main

import (
	"google.golang.org/grpc"
	"context"
)

type BizServerManager struct {
	*Manager
}

func NewBizServer(manager *Manager) *BizServerManager {
	return &BizServerManager{Manager: manager}
}

func (bz *BizServerManager) Check(ctx context.Context, in *Request) (*Response, error) {

	bz.publisher.sendEvent(in.Service,bz.Counter)

	bz.mu.Lock()
	s := bz.ServicesStatus.ServicesStatus[in.Service]
	bz.mu.Unlock()

	m := make(map[string]int64)
	m[s.name] = s.reaction.Nanoseconds()

	return &Response{Service: m}, nil
}

func (bz *BizServerManager) Min(ctx context.Context, in *Nothing) (*Response, error) {

	bz.mu.Lock()
	s := bz.ServicesStatus.GetMin()
	bz.mu.Unlock()

	m := make(map[string]int64)
	m[s.name] = s.reaction.Nanoseconds()

	return &Response{Service: m}, nil
}

func (bz *BizServerManager) Max(ctx context.Context, in *Nothing) (*Response, error) {

	bz.mu.Lock()
	s := bz.ServicesStatus.GetMax()
	bz.mu.Unlock()

	m := make(map[string]int64)
	m[s.name] = s.reaction.Nanoseconds()

	return &Response{Service: m}, nil
}

func (bz *BizServerManager) BizUnaryInterceptor1(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler, ) (interface{}, error) {

	reply, err := handler(ctx, req)

	return reply, err
}
