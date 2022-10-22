package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"

	"black-hat-go/logger"
)

type input struct {
	url  string
	port int
}

type output struct {
	port   int
	isOpen bool
}

func main() {
	logger.Init()
	log := zap.S()

	threads := 512
	url := "scanme.nmap.org"
	startPort := 1
	endPort := 1024
	connectTimeout := 5 * time.Second
	noOfPorts := endPort - startPort + 1

	log.Info("Port scanner")
	log.Infof("scanning from %d to %d (%d)", startPort, endPort, noOfPorts)

	ctx, cancel := context.WithCancel(context.Background())

	inputCh := make(chan input, 1)
	outputCh := make(chan output, 1)

	// start worker
	log.Debugf("start %d workers", threads)
	for i := 0; i < threads; i++ {
		go worker(ctx, i, inputCh, outputCh, connectTimeout)
	}

	// generate work
	log.Debug("generating work...")
	go func() {
		for port := startPort; port <= endPort; port++ {
			inputCh <- input{
				url:  url,
				port: port,
			}
		}
	}()

	result := make(map[int]bool, endPort-startPort+1)

	func() {
		currentPercentage := 10
		nextTick := noOfPorts * currentPercentage / 100
		fmt.Printf("----------\n")
		for out := range outputCh {
			result[out.port] = out.isOpen
			if len(result) == noOfPorts {
				fmt.Printf("|\n----------\n")
				return
			}
			if len(result) >= nextTick {
				fmt.Printf("|")
				currentPercentage += 10
				nextTick = noOfPorts * currentPercentage / 100
			}
		}
	}()

	cancel()

	fmt.Printf("Scanning Result\n")
	for port := startPort; port <= endPort; port++ {
		if !result[port] {
			continue
		}
		fmt.Printf("port %5d: open\n", port)
	}
}

func worker(ctx context.Context, id int, inputCh <-chan input, outputCh chan output, timeout time.Duration) {
	for {
		select {
		case <-ctx.Done():
			// fmt.Printf("worker id %d exited\n", id)
			return
		case in := <-inputCh:
			result := func(in input) bool {
				target := fmt.Sprintf("%s:%d", in.url, in.port)
				conn, err := net.DialTimeout("tcp", target, timeout)
				if err != nil {
					return false
				}
				defer conn.Close()
				return true
			}(in)

			outputCh <- output{
				port:   in.port,
				isOpen: result,
			}
		}
	}
}
