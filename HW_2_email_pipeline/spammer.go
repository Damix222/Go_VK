package main

import (
	"fmt"
	"sort"
	"sync"
)

func RunPipeline(cmds ...cmd) {
	var wg sync.WaitGroup

	in := make(chan interface{})

	for _, command := range cmds {
		out := make(chan interface{})

		wg.Add(1)

		go func(in, out chan interface{}) {
			defer wg.Done()
			command(in, out)
			close(out)
		}(in, out)

		in = out
	}

	wg.Wait()
}

func SelectUsers(in, out chan interface{}) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	seen := make(map[uint64]struct{})

	for email := range in {
		emailStr := email.(string)

		wg.Add(1)

		go func(emailStr string) {
			defer wg.Done()

			user := GetUser(emailStr)

			mu.Lock()
			_, ok := seen[user.ID]

			if !ok {
				seen[user.ID] = struct{}{}
			}
			mu.Unlock()

			if !ok {
				out <- user
			}
		}(emailStr)
	}

	wg.Wait()
}

func SelectMessages(in, out chan interface{}) {
	var wg sync.WaitGroup

	users := make([]User, 0, 2)

	processBatch := func(users []User) {
		batch := make([]User, len(users))
		copy(batch, users)

		wg.Add(1)

		go func(batch []User) {
			defer wg.Done()

			messages, err := GetMessages(batch...)

			if err == nil {
				for _, message := range messages {
					out <- message
				}
			}
		}(batch)
	}

	for user := range in {
		userData := user.(User)
		users = append(users, userData)

		if len(users) == 2 {
			processBatch(users)
			users = users[:0]
		}
	}

	if len(users) > 0 {
		processBatch(users)
	}

	wg.Wait()
}

func CheckSpam(in, out chan interface{}) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)

	for message := range in {
		messageData := message.(MsgID)

		wg.Add(1)

		go func(messageData MsgID) {
			defer wg.Done()

			sem <- struct{}{}

			res, err := HasSpam(messageData)

			<-sem

			if err == nil {
				msgData := MsgData{
					ID:      messageData,
					HasSpam: res,
				}

				out <- msgData
			}

		}(messageData)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	results := make([]MsgData, 0)

	for message := range in {
		msgData := message.(MsgData)
		results = append(results, msgData)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].HasSpam != results[j].HasSpam {
			return results[i].HasSpam
		}

		return results[i].ID < results[j].ID
	})

	for _, msg := range results {
		text := fmt.Sprintf("%t %d", msg.HasSpam, msg.ID)

		out <- text
	}
}
