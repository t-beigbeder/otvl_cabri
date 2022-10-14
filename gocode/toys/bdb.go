package main

import (
	"fmt"
	"github.com/tidwall/buntdb"
	"sync"
	"time"
)

func BdbRun(db *buntdb.DB, key string) error {
	err := db.Update(func(tx *buntdb.Tx) error {
		fmt.Printf("%s: set value...\n", key)
		_, _, err := tx.Set("mykey", "myvalue-"+key, nil)
		if err != nil {
			return err
		}
		time.Sleep(2250 * time.Millisecond)
		fmt.Printf("%s: set value done\n", key)
		return nil
	})
	if err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)
	err = db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get("mykey")
		if err != nil {
			return err
		}
		fmt.Printf("%s: value is %s\n", key, val)
		return nil
	})
	return nil
}

func main() {
	db, err := buntdb.Open("/tmp/bdb-tbe.dat")
	if err != nil {
		return
	}
	defer db.Close()
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		fmt.Printf("first BdbRun\n")
		if err := BdbRun(db, "1"); err != nil {
			fmt.Printf("error %v\n", err)
		}
		fmt.Printf("first BdbRun finished\n")
		wg.Done()
	}()
	go func() {
		time.Sleep(200 * time.Millisecond)
		fmt.Printf("second BdbRun\n")
		if err := BdbRun(db, "2"); err != nil {
			fmt.Printf("error %v\n", err)
		}
		fmt.Printf("second BdbRun finished\n")
		wg.Done()
	}()
	go func() {
		time.Sleep(400 * time.Millisecond)
		fmt.Printf("third BdbRun\n")
		if err := BdbRun(db, "3"); err != nil {
			fmt.Printf("error %v\n", err)
		}
		fmt.Printf("third BdbRun finished\n")
		wg.Done()
	}()
	wg.Wait()
}
