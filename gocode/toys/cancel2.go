package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			fmt.Fprintf(os.Stderr, "signal %s received, preparing to exit\n", sig)
			cancel()
		}
	}()
	<-server(ctx)
}

func ts() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

var rnd = rand.New(rand.NewSource(42))

func server(ctx context.Context) chan interface{} {
	completed := make(chan interface{})
	go func() {
		bctx, cancel := context.WithCancel(context.Background())
		_ = bctx
		bc1 := batch(bctx, "1")
		bc2 := batch(bctx, "2")
		bc3 := batch(ctx, "3")
		var wg sync.WaitGroup
		wg.Add(3)
		defer func() {
			time.Sleep(time.Duration(rnd.Int63n(2000) * int64(time.Millisecond)))
			time.Sleep(time.Duration(rnd.Int63n(2000) * int64(time.Millisecond)))
			fmt.Printf("%s server shutting down\n", ts())
			cancel()
			wg.Wait()
			fmt.Printf("%s server completed\n", ts())
			close(completed)
		}()
		go func() {
			<-bc1
			fmt.Printf("%s server detected batch1 completion\n", ts())
			wg.Done()
		}()
		go func() {
			<-bc2
			fmt.Printf("%s server detected batch2 completion\n", ts())
			wg.Done()
		}()
		go func() {
			<-bc3
			fmt.Printf("%s server detected batch3 completion\n", ts())
			wg.Done()
		}()

		time.Sleep(time.Duration(rnd.Int63n(2000) * int64(time.Millisecond)))
		fmt.Printf("%s server started\n", ts())
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("%s server detected cancel\n", ts())
				return
			case <-time.After(5 * time.Second):
				fmt.Printf("%s server is running\n", ts())
			}
		}
	}()
	return completed
}

func batch(ctx context.Context, name string) chan interface{} {
	completed := make(chan interface{})

	go func() {
		defer func() {
			fmt.Printf("%s batch %s shutting down\n", ts(), name)
			time.Sleep(time.Duration(rnd.Int63n(2000) * int64(time.Millisecond)))
			time.Sleep(time.Duration(rnd.Int63n(2000) * int64(time.Millisecond)))
			fmt.Printf("%s batch %s completed\n", ts(), name)
			close(completed)
		}()

		time.Sleep(time.Duration(rnd.Int63n(2000) * int64(time.Millisecond)))
		fmt.Printf("%s batch %s started\n", ts(), name)

		for {
			select {
			case <-ctx.Done():
				fmt.Printf("%s batch %s detected cancel\n", ts(), name)
				return
			case <-time.After(10 * time.Second):
				fmt.Printf("%s batch %s is running\n", ts(), name)
			}
		}
	}()
	return completed
}
