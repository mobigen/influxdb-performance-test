package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

// Config influxdb sim config
type Config struct {
	TagLength      int  `yaml:"tagLength"`
	NumTags        int  `yaml:"numTags"`
	NumFields      int  `yaml:"numField"`
	NumGoRoutine   int  `yaml:"numGoroutine"`
	WritePerSecond int  `yaml:"writePerSecond"`
	BatchSize      uint `yaml:"batchSize"`
}

var (
	kill          bool
	syncWaitGroup sync.WaitGroup
	numTry        []uint64
	numErr        []uint64
	numSuccess    uint64
)

func writeRoutine(id int, config Config) {
	fmt.Printf("Start Write Routine[ %d ]\n", id)

	client := influxdb2.NewClientWithOptions("http://localhost:8086",
		"BZTSZQByjQwjTa8ynfGlR5Y0N3v5MdsjARfE53y9qdPpAsHugN4DhPY4GaMN1Vy9XL04n13bP_uNg6oBrMHO7Q==",
		influxdb2.DefaultOptions().SetBatchSize(config.BatchSize))
	writeAPI := client.WriteAPI("mobigen", "testbuck")

	errCh := writeAPI.Errors()
	go func() {
		select {
		case <-errCh:
			atomic.AddUint64(&numErr[id], 1)
		}
	}()

	fmt.Printf("Make Write Data\n")
	writeData := write.NewPointWithMeasurement("performance-test")

	// Tag
	var tagKeys []string
	var tagValues []string
	tagKeys = make([]string, config.NumTags)
	tagValues = make([]string, config.NumTags)
	for idx := 0; idx < config.NumTags; idx++ {
		tagKeys[idx] = fmt.Sprintf("%dabcdefg", idx)
		tagValues[idx] = fmt.Sprintf("%d", idx)
		for tagLen := 0; tagLen < config.TagLength; tagLen++ {
			tagValues[idx] = fmt.Sprintf("%s%d", tagValues[idx], tagLen)
		}
	}
	for idx := 0; idx < config.NumTags; idx++ {
		writeData.AddTag(tagKeys[idx], tagValues[idx])
	}
	writeData.SortTags()

	// Field
	FieldKeys := []string{
		"utime",
		"stime",
		"cutime",
		"cstime",
		"rxByte",
		"rxPacket",
		"txByte",
		"txPacket",
		"vmsize",
		"vmrss",
		"rssfile",
	}
	for idx := 0; idx < config.NumFields; idx++ {
		writeData.AddField(FieldKeys[idx], idx)
	}
	writeData.SortFields()

	var now, next, remain int64
	period := int64(time.Second) / int64(config.WritePerSecond)
	fmt.Printf("period[ %d ]microsecond\n", period)
	fmt.Printf("Go Routine[ %d ] Wating..\n", id)

	syncWaitGroup.Wait()

	now = time.Now().UnixNano()
	remain = period - (now % period)
	next = (now + remain)

	//fmt.Printf("now[ %d ] remain[ %d ] next[ %d ]\n", now, remain, next)

	for {
		now = time.Now().UnixNano()
		if now >= next {
			atomic.AddUint64(&numTry[id], 1)
			writeData.SetTime(time.Now())
			writeAPI.WritePoint(writeData)

			// Calc Next Runtime
			now = time.Now().UnixNano()
			remain = period - (now % period)
			next = (now + remain)
		}
		if kill {
			fmt.Printf("stop[ %d ]\n", id)
			break
		}
		time.Sleep(100 * time.Microsecond)
	}
	writeAPI.Flush()
	client.Close()
}

func main() {
	configPath := GetAbsPath("./config.yaml")
	var config Config
	if err := ReadYaml(configPath, &config); err != nil {
		fmt.Printf("error[ %s ]", err.Error())
		os.Exit(0)
	}
	fmt.Printf("Read Config\n")
	PrintYaml(&config)
	numTry = make([]uint64, config.NumGoRoutine)
	numErr = make([]uint64, config.NumGoRoutine)

	kill = false
	syncWaitGroup.Add(1)
	for id := 0; id < config.NumGoRoutine; id++ {
		go writeRoutine(id, config)
	}

	time.Sleep(1 * time.Second)
	syncWaitGroup.Done()

	signalch := make(chan os.Signal, 1)
	signal.Notify(signalch, os.Interrupt, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM)

	var bTry, cTry uint64
	var bError, cError uint64
	for {
		select {
		case <-signalch:
			fmt.Printf("Shutting down\n")
			kill = true
			time.Sleep(1 * time.Second)
			os.Exit(0)
		case <-time.After(time.Second):

			for idx := 0; idx < config.NumGoRoutine; idx++ {
				cTry += atomic.LoadUint64(&numTry[idx])
				cError += atomic.LoadUint64(&numErr[idx])
			}

			fmt.Printf("\n")
			fmt.Printf("Total    : Try[ %10d ] Error[ %10d ]\n", cTry, cError)
			if bTry == 0 && bError == 0 {
				fmt.Printf("1 Second : Try[ %10d ] Error[ %10d ]\n", cTry, cError)
			} else {
				fmt.Printf("1 Second : Try[ %10d ] Error[ %10d ]\n", cTry-bTry, cError-bError)
			}
			bTry = cTry
			bError = cError
			cTry = 0
			cError = 0
		}
	}
}
