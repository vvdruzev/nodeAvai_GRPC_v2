package main

import (
	"fmt"
	"net/http"
	"net/url"
	"log"
	"time"
	"context"
	"sync"
	"strings"
)

const https_proxy="http://vvdruze1:Kiasoul$0898@bproxy.ug.mts.ru:3131/"

func setProxy(c *http.Client)  error {
	proxyURL, err := url.Parse(https_proxy)

	if err != nil {
		log.Println(err)
		return err
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	c.Transport = transport
	//adding the Transport object to the http Client
	c.Timeout = 1*time.Second
	return nil
}

func check(c context.Context,as *Manager, wg *sync.WaitGroup)  {
	//cli := &http.Client{}
	//setProxy(cli)

	out := make(chan service, len(as.ServicesStatus.ServicesStatus))
	defer close(out)
	ctx, cancel1 := context.WithCancel(context.Background())

	for k,_ := range as.ServicesStatus.ServicesStatus{
		wg.Add(1)

		go func(c context.Context,k string) error {
			defer wg.Done()
			var s service
			s.name = k
			start := time.Now()
			req, _ := http.NewRequest("GET", k,nil)
			err := httpDo(ctx, req, func(response *http.Response, e error) error {
				if e != nil && strings.Contains(e.Error(),"canceled"){
					s.status = http.StatusRequestTimeout
					s.reaction = 0
					out <- s
					return nil
				}
				if response == nil {
					s.status = http.StatusNotFound
					s.reaction = 0
					out <- s
					return nil
				}
				s.reaction = time.Since(start)
				s.status = response.StatusCode
				out <- s
				return nil

			})

			return err

		}(ctx,k)

	}
	i:=0
LOOP:
	for {
		select {
		case <-c.Done():
			log.Println("получил cancel от autoCheck")
			cancel1()
			break LOOP
		case s := <-out:
			i++
			as.mu.Lock()
			as.ServicesStatus.ServicesStatus[s.name] = s
			as.mu.Unlock()
			if i >= len(as.ServicesStatus.ServicesStatus){
				break LOOP
			}

		}
	}
	as.mu.Lock()
	as.min()
	as.max()
	fmt.Println(as.ServicesStatus.ServicesStatus)
	fmt.Printf("max = %s :  %s ,   min = %s : %s\n\n", as.GetMax().name,as.GetMax().reaction, as.GetMin().name, as.GetMin().reaction)
	as.mu.Unlock()


}

func httpDo(ctx context.Context, req *http.Request, f func(*http.Response, error) error) error  {
	cli := &http.Client{}
	err := setProxy(cli)
	if err != nil {
		return err
	}
	ctx1, cancel := context.WithCancel(context.Background())
	req.WithContext(ctx1)
	c := make(chan error,1)
	go func() {c <- f(cli.Do(req))}()
	select {
	case <-ctx.Done():
		log.Println("получил cancel от check")
		//tr.CancelRequest(req)
		cancel()
		<-c // Wait for f to return.
		return ctx.Err()
	case err := <-c:
		return err
	}

}
