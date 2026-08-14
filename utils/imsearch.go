package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"sync"
	"time"
)

// structs necessarias
type SerperImageItem struct {
	Title    string `json:"title"`
	ImageURL string `json:"imageUrl"`
	Link     string `json:"link"`
}

type SerperResponse struct {
	Images []SerperImageItem `json:"images"`
}

// struct p manter o estado da navegacao de cada mensagem
type SessionState struct {
	Results []SerperImageItem
	Index   int
}

var (
	sessions = make(map[string]*SessionState)
	mu       sync.Mutex
)

func FetchSerperImages(query, SerperAPIKey string) ([]SerperImageItem, error) {
	url := "https://google.serper.dev/images"

	payload := map[string]interface{}{
		"q":   query,
		"gl":  "br",
		"hl":  "pt-br",
		"num": 10,
	}
	jsonBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-KEY", SerperAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req) // faz o request
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var serperRes SerperResponse
	if err := json.NewDecoder(resp.Body).Decode(&serperRes); err != nil {
		return nil, err
	}

	return serperRes.Images, nil
}

// Constrói o Embed estilizado
func buildEmbed(item SerperImageItem, index, total int) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       item.Title,
		URL:         item.Link,
		Description: fmt.Sprintf("[Acesse a página original](%s)", item.Link),
		Color:       0x5865F2, // Cor do Discord
		Image: &discordgo.MessageEmbedImage{
			URL: item.ImageURL,
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    fmt.Sprintf("Página %d/%d dos resultados", index+1, total),
			IconURL: "https://pbs.twimg.com/media/Gdw2S0DbEAAxtjk?format=jpg&name=360x360",
		},
	}
}

// Constrói a linha de botões interativos
func buildButtons(index, total int) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Emoji:    &discordgo.ComponentEmoji{Name: "◀️"},
					CustomID: "img_prev",
					Style:    discordgo.PrimaryButton,
					Disabled: index == 0,
				},
				discordgo.Button{
					Emoji:    &discordgo.ComponentEmoji{Name: "▶️"},
					CustomID: "img_next",
					Style:    discordgo.PrimaryButton,
					Disabled: index >= total-1,
				},
				discordgo.Button{
					Emoji:    &discordgo.ComponentEmoji{Name: "🔀"},
					CustomID: "img_rand",
					Style:    discordgo.SecondaryButton,
				},
				discordgo.Button{
					Emoji:    &discordgo.ComponentEmoji{Name: "🗑️"},
					CustomID: "img_delete",
					Style:    discordgo.DangerButton,
				},
			},
		},
	}
}

// handler para quando os botoes forem clicados
func InteractionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	customID := i.MessageComponentData().CustomID
	msgID := i.Message.ID

	mu.Lock()
	state, exists := sessions[msgID]
	mu.Unlock()

	if !exists && customID != "img_delete" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "⚠️ Essa sessão de busca expirou.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	switch customID {
	case "img_prev":
		if state.Index > 0 {
			state.Index--
		}
	case "img_next":
		if state.Index < len(state.Results)-1 {
			state.Index++
		}
	case "img_rand":
		// Pega um índice aleatório simples
		state.Index = (state.Index + 3) % len(state.Results)
	case "img_delete":
		s.ChannelMessageDelete(i.ChannelID, msgID)
		mu.Lock()
		delete(sessions, msgID)
		mu.Unlock()
		return
	}

	// Atualiza a mensagem existente com a nova imagem/página
	currentItem := state.Results[state.Index]
	embed := buildEmbed(currentItem, state.Index, len(state.Results))
	components := buildButtons(state.Index, len(state.Results))

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}

func HandleSearchCommand(s *discordgo.Session, m *discordgo.MessageCreate, query, SerperAPIKey string) {
	results, err := FetchSerperImages(query, SerperAPIKey)
	if err != nil || len(results) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Nenhuma imagem encontrada.")
		return
	}

	// cria o Embed e botoes
	embed := buildEmbed(results[0], 0, len(results))
	components := buildButtons(0, len(results))

	msg, err := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
	if err != nil {
		return
	}

	// salva o estado da busca no mapa global associado ao ID da mensagem
	mu.Lock()
	sessions[msg.ID] = &SessionState{
		Results: results,
		Index:   0,
	}
	mu.Unlock()

	go func(msgID, channelID string) {
		time.Sleep(20 * time.Second)

		// verifica se a sessão ainda existe em memória
		mu.Lock()
		_, exists := sessions[msgID]
		if exists {
			// Remove o estado da memória para liberar RAM
			delete(sessions, msgID)
		}
		mu.Unlock()

		// Se a sessão ainda existia (o usuário NÃO apertou o botão de deletar antes)
		if exists {
			// Edita a mensagem no Discord passando uma lista vazia de botões
			emptyComponents := []discordgo.MessageComponent{}
			_, _ = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				ID:         msgID,
				Channel:    channelID,
				Components: &emptyComponents, // Remove a linha de botões
			})
		}
	}(msg.ID, msg.ChannelID)
}
