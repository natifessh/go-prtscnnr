package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

type Port struct {
	Port int    `json:"port"`
	Type string `json:"type"`
	Time string `json:"scannedAt"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <host>")
		os.Exit(1)
	}
	target := os.Args[1]
	ch := make(chan Port, 100)
	var wg sync.WaitGroup
	var results []Port
	go func() {
		for p := range ch {
			results = append(results, p)
			fmt.Printf("Found open port: %d (%s)\n", p.Port, p.Type)
		}
	}()
	for port := 1; port <= 100; port++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			scanPort(ch, p, target)
		}(port)
	}
	wg.Wait()
	close(ch)
	if len(results) > 0 {
		file, _ := os.Create("ports.json")
		defer file.Close()
		file.WriteString("[\n")
		for i, p := range results {
			jsonData, _ := json.Marshal(p)
			file.Write(jsonData)
			if i < len(results)-1 {
				file.WriteString(",\n")
			}
		}
		file.WriteString("\n]")
		fmt.Println("Results saved to ports.json")
	} else {
		fmt.Println("No open ports found")
	}
}

func scanPort(ch chan<- Port, port int, host string) {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		return
	}
	defer conn.Close()
	ch <- Port{port, getService(port), time.Now().Format("2006-01-02 15:04:05")}
}

func getService(port int) string {
	switch port {
	case 22:
		return "ssh"
	case 80:
		return "http"
	case 443:
		return "https"
	default:
		return "unknown"
	}
}
