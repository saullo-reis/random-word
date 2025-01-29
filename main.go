package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type DailyWord struct {
	mu   sync.Mutex
	word string
}

var dailyWord DailyWord

func main() {
	if err := refreshWord(); err != nil {
		log.Fatalf("Erro ao carregar palavra inicial: %v", err)
	}
	go scheduleDailyReset()

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	router.GET("/", func(c *gin.Context) {
		dailyWord.mu.Lock()
		currentWord := dailyWord.word
		dailyWord.mu.Unlock()

		c.HTML(http.StatusOK, "index.html", gin.H{
			"Palavra": currentWord,
		})
	})

	router.Run(":8080")
}

func refreshWord() error {
	newWord, err := chooseTheWord()
	if err != nil {
		return err
	}

	dailyWord.mu.Lock()
	dailyWord.word = newWord
	dailyWord.mu.Unlock()

	return nil
}

func scheduleDailyReset() {
	for {
		now := time.Now()
		nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		durationUntilMidnight := nextMidnight.Sub(now)

		<-time.After(durationUntilMidnight)

		if err := refreshWord(); err != nil {
			log.Printf("Erro ao atualizar palavra: %v", err)
		} else {
			log.Println("Palavra do dia atualizada com sucesso")
		}
	}
}

func chooseTheWord() (string, error) {
	file, err := os.Open("words.json")
	if err != nil {
		return "", fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer file.Close()

	var words []string
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&words); err != nil {
		return "", fmt.Errorf("erro ao decodificar JSON: %w", err)
	}

	if len(words) == 0 {
		return "", fmt.Errorf("nenhuma palavra disponível no arquivo")
	}

	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	randomIndex := r.Intn(len(words))

	return words[randomIndex], nil
}