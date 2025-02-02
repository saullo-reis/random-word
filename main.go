package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func chargeWords(path string) ([]string, error){
	archive, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir o arquivo: %v", err)
	}
	defer archive.Close()

	var words []string
	decode := json.NewDecoder(archive)
	if err := decode.Decode(&words); err != nil{
		return nil, fmt.Errorf("erro ao decodificar: %v", err)
	}

	return words, nil
}

func wordOfTheDay(words []string) (string, error){
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return "", fmt.Errorf("erro ao carregar fuso: %v", err)
	}
	now := time.Now().In(loc)
	dataBase := time.Date(now.Year(), now.Month(), now.Day(), 0,0,0,0, loc)

	seed := dataBase.UnixNano()
	r := rand.New(rand.NewSource(seed))

	indice := r.Intn(len(words))
	fmt.Println(words[indice])
	return words[indice], nil
}

func main() {

	router := gin.Default()
	router.Static("/static", "./src")

	router.LoadHTMLGlob("templates/index.html")

	router.GET("/", func(c *gin.Context) {
		fmt.Println("Request INFO")
		fmt.Printf("ip: %v \n", c.ClientIP())
		fmt.Printf("method: %v \n", c.Request.Method)

		words, err := chargeWords("palavras.json")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": 500,
				"message": err,
			})
		}

		word, err := wordOfTheDay(words)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": 500,
				"message": err,
			})
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"Palavra": word,
		})
	})

	router.Run(":8080")
}
