package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/unicode/norm"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"
)

type WordStat struct {
	LastTime int64 `json:"last_time"`
	Record   int64 `json:"record"`
}

type Frase struct {
	Texto string `json:"texto"`
	Autor string `json:"autor"`
}

type gistResponse struct {
	Files map[string]struct {
		Content string `json:"content"`
	} `json:"files"`
}

// Do not disturb
func Load_DND(path string) ([]string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []string{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result []string

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func Save_DND(path string, data []string) error {
	bytes, err := json.MarshalIndent(data, "", " ")

	if err != nil {
		return err
	}

	return os.WriteFile(path, bytes, 0644)
}

// REDACTED
func Remover_Acentos(s string) string {
	t := norm.NFD.String(s)
	var b strings.Builder
	for _, r := range t {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func FormatarTempo(ms int64) string {
	seg := (ms / 1000) % 60
	min := (ms / 1000 / 60) % 60
	horas := (ms / 1000 / 60 / 60) % 24
	dias := ms / 1000 / 60 / 60 / 24

	var partes []string
	if dias > 0 {
		partes = append(partes, fmt.Sprintf("%dd", dias))
	}
	if horas > 0 {
		partes = append(partes, fmt.Sprintf("%dh", horas))
	}
	if min > 0 {
		partes = append(partes, fmt.Sprintf("%dm", min))
	}
	if seg > 0 {
		partes = append(partes, fmt.Sprintf("%ds", seg))
	}

	if len(partes) == 0 {
		return "menos de 1s"
	}
	return strings.Join(partes, " ")
}

func FormatDuration(seconds int64) string {
	if seconds < 0 {
		seconds = 0
	}

	d := seconds / 86400
	h := (seconds % 86400) / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60

	var parts []string

	if d > 0 {
		parts = append(parts, fmt.Sprintf("%dd", d))
	}
	if h > 0 || d > 0 {
		parts = append(parts, fmt.Sprintf("%dh", h))
	}
	if m > 0 || h > 0 || d > 0 {
		parts = append(parts, fmt.Sprintf("%dm", m))
	}
	parts = append(parts, fmt.Sprintf("%ds", s))

	return strings.Join(parts, " ")
}

func Load_WordCounter(path string) (map[string]WordStat, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return make(map[string]WordStat), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var counter map[string]WordStat
	if err := json.Unmarshal(data, &counter); err != nil {
		return nil, err
	}

	return counter, nil
}

func Save_WordCounter(path string, data map[string]WordStat) error {
	bytes, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, bytes, 0644)
}

// imgs de sexta (um dia vai funcionar, eu confio)
func Load_ImgSexta() {

}

func Save_ImgSexta() {

}

// para o gist
func Load_Gist(gistID string, filename string, token string) ([]Frase, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("https://api.github.com/gists/%s", gistID),
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"erro ao buscar gist: %d - %s",
			resp.StatusCode,
			string(body),
		)
	}

	var data gistResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	file, ok := data.Files[filename]
	if !ok {
		return nil, fmt.Errorf("arquivo %s não encontrado no gist", filename)
	}

	var frases []Frase
	if err := json.Unmarshal([]byte(file.Content), &frases); err != nil {
		return nil, err
	}
	//fmt.Println("Conteúdo bruto do Gist:")
	//fmt.Println(file.Content)
	return frases, nil
}

func mustJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func Update_Gist(gistID string, filename string, token string, frases []Frase) error {
	payload := map[string]any{
		"files": map[string]map[string]string{
			filename: {
				"content": mustJSON(frases),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"PATCH",
		fmt.Sprintf("https://api.github.com/gists/%s", gistID),
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf(
			"erro ao atualizar gist: %d - %s",
			resp.StatusCode,
			string(respBody),
		)
	}

	return nil
}

// conserta os embedding

func HandleFixEmbeds(s *discordgo.Session, m *discordgo.MessageCreate) {
	content := m.Content

	// Detecta links monitorados
	if !strings.Contains(content, "https://x.com/") &&
		!strings.Contains(content, "https://twitter.com/") &&
		!strings.Contains(content, "https://instagram.com/") &&
		!strings.Contains(content, "https://www.instagram.com/") {
		return
	}

	msg := content
	autor := m.Author.Username

	// Twitter / X
	if strings.Contains(content, "https://x.com/") {
		re := regexp.MustCompile(`https://x\.com/`)
		msg = re.ReplaceAllString(msg, "https://fixvx.com/")
	} else if strings.Contains(content, "https://twitter.com/") {
		re := regexp.MustCompile(`https://twitter\.com/`)
		msg = re.ReplaceAllString(msg, "https://fixvx.com/")
	} else {
		// Instagram
		reBase := regexp.MustCompile(`https://(www\.)?instagram\.com/`)
		msg = reBase.ReplaceAllString(msg, "https://www.vxinstagram.com/")

		// Remove parâmetros extras (reel / p)
		reClean := regexp.MustCompile(`(https://www\.vxinstagram\.com/(reel|p)/[^/]+)/?.*`)
		msg = reClean.ReplaceAllString(msg, `$1/`)
	}

	// Apaga mensagem original
	err := s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		log.Println("Erro ao deletar mensagem:", err)
		return
	}

	// Reenvia mensagem corrigida
	_, err = s.ChannelMessageSend(
		m.ChannelID,
		fmt.Sprintf(
			"Mensagem enviada por **%s**\n%s",
			autor,
			msg,
		),
	)
	if err != nil {
		log.Println("Erro ao reenviar mensagem:", err)
	}
}
func EmojiToReaction(e discordgo.Emoji) string {
	if e.ID != "" {
		// emoji de guilda
		return e.Name + ":" + e.ID
	}
	// emoji unicode
	return e.Name
}
func MaybeReact(s *discordgo.Session, m *discordgo.MessageCreate, emojis []*discordgo.Emoji) {

	// 5% de chance
	if rand.Intn(100) >= 5 {
		return
	}

	// escolhe emoji aleatório
	e := emojis[rand.Intn(len(emojis))]

	err := s.MessageReactionAdd(
		m.ChannelID,
		m.ID,
		EmojiToReaction(*e),
	)
	if err != nil {
		fmt.Println("Erro ao reagir:", err)
	}
}
