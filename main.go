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
	mu         sync.Mutex
	word       string
	lastUpdate time.Time
}

var dailyWord DailyWord

func main() {
	if err := refreshWordIfNeeded(); err != nil {
		log.Fatalf("Erro ao carregar palavra inicial: %v", err)
	}
	go scheduleDailyReset()

	router := gin.Default()
	router.Static("/static", "./src")

	router.LoadHTMLGlob("templates/index.html")

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

func refreshWordIfNeeded() error {
	now := time.Now().UTC()
	lastMidnight := getLastMidnight(now)

	dailyWord.mu.Lock()
	defer dailyWord.mu.Unlock()

	if dailyWord.lastUpdate.After(lastMidnight) {
		return nil
	}

	newWord, err := chooseTheWord(lastMidnight)
	if err != nil {
		return err
	}

	dailyWord.word = newWord
	dailyWord.lastUpdate = lastMidnight
	return nil
}

func scheduleDailyReset() {
	for {
		now := time.Now().UTC()
		nextMidnight := getLastMidnight(now).Add(24 * time.Hour)
		durationUntilMidnight := nextMidnight.Sub(now)

		<-time.After(durationUntilMidnight)

		if err := refreshWordIfNeeded(); err != nil {
			log.Printf("Erro ao atualizar palavra: %v", err)
		} else {
			log.Println("Palavra do dia atualizada com sucesso")
		}
	}
}

func chooseTheWord(referenceTime time.Time) (string, error) {
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

	seed := referenceTime.Unix()
	source := rand.NewSource(seed)
	r := rand.New(source)
	randomIndex := r.Intn(len(words))

	return words[randomIndex], nil
}

func getLastMidnight(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}