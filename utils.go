package main

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func getRandomString() string {
	return RandStringRunes(10)
}

func generateRandomQueueName() string {
	return fmt.Sprintf("queue-%v", RandStringRunes(10))
}

func getRandomUUID() string {
	id := uuid.New()
	uuidStr := id.String()
	return uuidStr
}

func elapsed(what string) func() {
	start := time.Now()
	return func() {
		log.Printf("(%s) took (%v) to execute", what, time.Since(start))
	}
}
