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
	twilio "github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
)

type DailyWord struct {
	mu   sync.Mutex
	word string
}

var dailyWord DailyWord

func main() {
	_, err := refreshWord(); 
	if err != nil {
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

func refreshWord() (string, error) {
	newWord, err := chooseTheWord()
	if err != nil {
		return "", err
	}

	dailyWord.mu.Lock()
	dailyWord.word = newWord
	dailyWord.mu.Unlock()

	return newWord, nil
}

func scheduleDailyReset() {
	for {
		now := time.Now()
		nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		durationUntilMidnight := nextMidnight.Sub(now)

		<-time.After(durationUntilMidnight)
		

		wordGenerated, err := refreshWord()
		if err != nil {
			log.Printf("Erro ao atualizar palavra: %v", err)
		} else {
			log.Println("Palavra do dia atualizada com sucesso")
		}

		from := "+18458511016"
		client := twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: os.Getenv("TWILIO_ACCOUNT_SID"),
			Password: os.Getenv("TWILIO_AUTH_TOKEN"),
		})
		numbersPhone := []string{
			"+5521994625997",
		}
		body := wordGenerated

		for _, to := range numbersPhone {
			// Configura os parâmetros da mensagem
			params := &api.CreateMessageParams{
				To:   &to,
				From: &from,
				Body: &body,
			}
	
			resp, err := client.Api.CreateMessage(params)
			if err != nil {
				log.Printf("Erro ao enviar mensagem para %s: %v\n", to, err)
				continue
			}

			if resp.Sid != nil {
				fmt.Printf("Mensagem enviada para %s, SID: %s\n", to, *resp.Sid)
			} else {
				fmt.Printf("Mensagem enviada para %s, sem SID retornado\n", to)
			}
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