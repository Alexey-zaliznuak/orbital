package main

import (
	"bytes"
	"log"
	"net/http"
	"sync"
)

func main() {
	url := "http://localhost:8081/api/v1/message"
	client := http.Client{}
	cnt := 1

	body := bytes.NewBufferString(`{
		"routing_key": "notifications.email",
		"payload": "dGVzdA==",
		"metadata": {
			"priority": "high",
			"source": "api"
		}
	}`)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	statuses := make(map[int]int, cnt)
	wg := sync.WaitGroup{}
	wg.Add(cnt)
	for i := range cnt {
		go func(i int) {
			defer wg.Done()
			resp, err := client.Do(req)
			if err != nil {
				log.Println(err)
				statuses[500]++
			}
			defer resp.Body.Close()
			statuses[resp.StatusCode]++
		}(i)
	}

	wg.Wait()
	for i, s := range statuses {
		log.Printf("Status %d: %d\n", i, s)
	}

	log.Println("All messages sent successfully")
}
