package utils

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"os"
)

var links []string

func BaoBao(s *discordgo.Session) {
	filename := "kino.json"
	conteudo, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("erro ao ler o JSON: ", err)
		return
	}

	err = json.Unmarshal(conteudo, &links)
	if err != nil {
		fmt.Println("erro ao converter conteudo do json: ", err)
		return
	}

	if len(links) > 0 {
		index := rand.Intn(len(links))
		msg := "Daily baobao: " + links[index]
		s.ChannelMessageSend("1450181687119184014", msg)
		links = append(links[:index], links[index+1:]...)
		dadosEscrita, _ := json.MarshalIndent(links, "", "  ")
		err = os.WriteFile(filename, dadosEscrita, 0644)
		if err != nil {
			fmt.Printf("Erro ao atualizar arquivo após remoção: %v", err)
		}
		return
	} else {
		fmt.Println("sem conteudo")
		return
	}

}
