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
	Last       int64 `json:"last"`
	Record     int64 `json:"record"`
	RecordDate int64 `json:"recordDate"`
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

type Img_Sexta struct {
	Autor    string `json:"autor"`
	Conteudo string `json:"imagem"`
}

const (
	filename    = "imagensSexta.json"
	Comunicados = "1494165588120178798"
)

var cooldowns = make(map[string]time.Time)

// Do not disturb
/* func Load_DND(path string) ([]string, error) {
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
*/
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

func FormatSince(lastMillis int64) string {
	if lastMillis <= 0 {
		return "nunca"
	}

	diffSeconds := (time.Now().UnixMilli() - lastMillis) / 1000
	if diffSeconds < 0 {
		diffSeconds = 0
	}

	return FormatDuration(diffSeconds)
}

func FormatDuration(seconds int64) string {
	if seconds <= 0 {
		return "0s"
	}

	d := seconds / 86400
	h := (seconds % 86400) / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60

	var parts []string

	if d > 0 {
		parts = append(parts, fmt.Sprintf("%dd", d))
	}
	if h > 0 {
		parts = append(parts, fmt.Sprintf("%dh", h))
	}
	if m > 0 {
		parts = append(parts, fmt.Sprintf("%dm", m))
	}
	if s > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%ds", s))
	}

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

	counter := make(map[string]WordStat)
	if err := json.Unmarshal(data, &counter); err != nil {
		return nil, err
	}

	// saneamento básico
	now := time.Now().UnixMilli()
	for k, v := range counter {
		if v.Last > now {
			v.Last = now
		}
		if v.Record < 0 {
			v.Record = 0
		}
		if v.RecordDate > now {
			v.RecordDate = 0
		}
		counter[k] = v
	}

	return counter, nil
}

func Save_WordCounter(path string, data map[string]WordStat) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, bytes, 0644)
}

// imgs de sexta (um dia vai funcionar, eu confio)

func Handler_ImgSexta(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID == "" {
		if m.Content == "" {
			// eh anexo
			if ultimaVez, ok := cooldowns[m.Author.ID]; ok {
				// Calcula quanto tempo passou
				tempoPassado := time.Since(ultimaVez)

				if tempoPassado < time.Hour {
					s.ChannelMessageSend(m.ChannelID, "🖕, espere o cooldown de uma hora acabar fuleiro")
					return
				}
			}
			for _, img := range m.Attachments {
				nova_img := Img_Sexta{
					Autor:    m.Author.Username,
					Conteudo: img.URL,
				}
				Save_ImgSexta(nova_img)
			}
			cooldowns[m.Author.ID] = time.Now()
			s.ChannelMessageSend(m.ChannelID, "Recebido.")
		} /* else if strings.Contains(m.Content, "attchments") {
		// eh link
		for img := range strings.SplitSeq(m.Content, " ") {
			nova_img := Img_Sexta{
				Autor:    m.Author.Username,
				Conteudo: img,
			}
			Save_ImgSexta(nova_img)
			cooldowns[m.Author.ID] = time.Now()
		}
		} */
	}
}

func Load_ImgSexta(s *discordgo.Session) {
	var lista []Img_Sexta

	dados, err := os.ReadFile(filename)
	if err != nil || len(dados) == 0 {
		return
	}

	if err := json.Unmarshal(dados, &lista); err != nil {
		return
	}

	if len(lista) == 0 {
		return
	}

	index := rand.Intn(len(lista))
	imagemEscolhida := lista[index]

	mensagem := "Imagem do dia" + " enviado por: " + imagemEscolhida.Autor
	s.ChannelMessageSend(Comunicados, mensagem)
	s.ChannelMessageSend(Comunicados, imagemEscolhida.Conteudo)
	lista = append(lista[:index], lista[index+1:]...)
	dadosEscrita, _ := json.MarshalIndent(lista, "", "  ")
	err = os.WriteFile(filename, dadosEscrita, 0644)
	if err != nil {
		log.Printf("Erro ao atualizar arquivo após remoção: %v", err)
	}

}

func Save_ImgSexta(img Img_Sexta) {

	var lista []Img_Sexta

	dados, err := os.ReadFile(filename)
	if err == nil {
		json.Unmarshal(dados, &lista)
	}

	lista = append(lista, img)

	dadosEscrita, _ := json.MarshalIndent(lista, "", "  ")

	os.WriteFile(filename, dadosEscrita, 0644)
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
var (
	reAllLinks       = regexp.MustCompile(`https?://[^\s]+`)
	reTwitter        = regexp.MustCompile(`https://(x|twitter)\.com/`)
	reInstagramBase  = regexp.MustCompile(`https://(www\.)?instagram\.com/`)
	reInstagramClean = regexp.MustCompile(`(https://www\.vxinstagram\.com/(reel|p)/[^/?]+)/?`)
	rePixiv = regexp.MustCompile(`https://www\.pixiv\.net/`)
)

func fixLink(link string) (string, bool) {
	// Ignora já corrigidos
	if strings.Contains(link, "fixvx.com")||strings.Contains(link, "vxtwitter.com") || strings.Contains(link, "vxinstagram.com") || strings.Contains(link, "phixiv.net"){
		return "", false
	}

	// Twitter / X
	if strings.Contains(link, "x.com") || strings.Contains(link, "twitter.com") {
		fixed := reTwitter.ReplaceAllString(link, "https://fixvx.com/")
		return fixed, true
	}

	// Instagram
	if strings.Contains(link, "instagram.com") {
		fixed := reInstagramBase.ReplaceAllString(link, "https://www.vxinstagram.com/")
		fixed = reInstagramClean.ReplaceAllString(fixed, `$1/`)
		return fixed, true
	}
	
	// Pixiv
	if strings.Contains(link, "pixiv.net") {
		fixed := rePixiv.ReplaceAllString(link, "https://www.phixiv.net/")
		return fixed, true
		
	}

	return "", false
}

func HandleFixEmbeds(s *discordgo.Session, m *discordgo.MessageCreate) {
	content := m.Content

	// Pega todos os links da mensagem
	links := reAllLinks.FindAllString(content, -1)
	if len(links) == 0 {
		return
	}

	var fixedLinks []string

	for _, link := range links {
		if fixed, ok := fixLink(link); ok {
			fixedLinks = append(fixedLinks, fixed)
		}
	}

	if len(fixedLinks) == 0 {
		return
	}

	// Suprime embed da mensagem original
	flags := discordgo.MessageFlagsSuppressEmbeds
	_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel: m.ChannelID,
		ID:      m.ID,
		Flags:   flags,
	})
	if err != nil {
		log.Println("Erro ao suprimir embed:", err)
	}

	// Envia apenas os links corrigidos
	msg := strings.Join(fixedLinks, "\n")

	_, err = s.ChannelMessageSend(m.ChannelID, msg)
	if err != nil {
		log.Println("Erro ao enviar mensagem:", err)
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
	if len(emojis) == 0 {
		return
	}

	// 1% de chance
	if rand.Intn(100) >= 1 {
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
