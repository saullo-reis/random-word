package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type Estado struct {
	LastDate   string `json:"last_date"`
	RandomWord string `json:"random_word"`
}

type Palavras struct {
	Palavras []string `json:"palavras"`
}

func carregarPalavras() ([]string, error) {

	file, err := ioutil.ReadFile("palavras.json")
	if err != nil {
		return nil, err
	}

	
	var p Palavras
	err = json.Unmarshal(file, &p)
	if err != nil {
		return nil, err
	}

	return p.Palavras, nil
}

func carregarEstado() (*Estado, error) {
	
	file, err := ioutil.ReadFile("estado.json")
	if err != nil {
		if os.IsNotExist(err) {
			return &Estado{}, nil 
		}
		return nil, err
	}

	
	var e Estado
	err = json.Unmarshal(file, &e)
	if err != nil {
		return nil, err
	}

	return &e, nil
}

func salvarEstado(estado *Estado) error {
	
	data, err := json.MarshalIndent(estado, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("estado.json", data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func obterPalavraAleatoria(palavras []string) string {
	rand.Seed(time.Now().UnixNano())
	return palavras[rand.Intn(len(palavras))]
}

func main() {
	
	palavras, err := carregarPalavras()
	if err != nil {
		fmt.Println("Erro ao carregar palavras:", err)
		return
	}

	
	estado, err := carregarEstado()
	if err != nil {
		fmt.Println("Erro ao carregar estado:", err)
		return
	}

	
	dataAtual := time.Now().Format("2006-01-02")

	
	if estado.LastDate != dataAtual {
		
		estado.RandomWord = obterPalavraAleatoria(palavras)
		estado.LastDate = dataAtual

		
		err = salvarEstado(estado)
		if err != nil {
			fmt.Println("Erro ao salvar estado:", err)
			return
		}
	}

	ehPalavraValida := false
	for _, palavra := range palavras {
		if palavra == estado.RandomWord {
			ehPalavraValida = true
			break
		}
	}
	if !ehPalavraValida {
		estado.RandomWord = obterPalavraAleatoria(palavras)
	}

	estado.LastDate = dataAtual

	
	err = salvarEstado(estado)
	if err != nil {
		fmt.Println("Erro ao salvar estado:", err)
		return
	}

	router := gin.Default()
	router.Static("/static", "./src")

	router.LoadHTMLGlob("templates/index.html")

	router.GET("/", func(c *gin.Context) {

		c.HTML(http.StatusOK, "index.html", gin.H{
			"Palavra": estado.RandomWord,
		})
	})

	router.Run(":8080")
}
