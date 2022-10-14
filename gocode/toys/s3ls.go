package main

import (
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"os"
	"sync"
)

func getOC() cabridss.ObsConfig {
	return cabridss.ObsConfig{Container: os.Getenv("OVHCT"), Endpoint: os.Getenv("OVHEP"), Region: os.Getenv("OVHRG"), AccessKey: os.Getenv("OVHAK"), SecretKey: os.Getenv("OVHSK")}
}

func main() {
	s3s := cabridss.NewS3Session(getOC(), nil)
	s3s.Initialize()
	cs, err := s3s.List("meta-")
	fmt.Println(len(cs), err, cs[0], cs[len(cs)-1])
	MAX := 80
	var wg sync.WaitGroup
	for c := 0; c < MAX; c++ {
		wg.Add(1)
		go func(cc int) {
			defer wg.Done()
			for i := c; i < len(cs); i += MAX {
				//bs, err := s3s.Get(cs[i])
				bs, err := s3s.List(cs[i])
				_ = bs
				if err != nil {
					fmt.Println(err)
				}
				//fmt.Println(cs[i], len(bs), err)
			}
		}(c)
	}
	wg.Wait()
}
