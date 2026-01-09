package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// structs para resposta GraphQL de media (anime/manga)
type mediaTitle struct {
	Romaji string `json:"romaji"`
}

type mediaItem struct {
	ID    int        `json:"id"`
	Title mediaTitle `json:"title"`
}

type pageData struct {
	Media []mediaItem `json:"media"`
}

type pageWrapper struct {
	Page pageData `json:"Page"`
}

type mediaResponse struct {
	Data pageWrapper `json:"data"`
}

// structs para resposta GraphQL de user
type avatar struct {
	Large string `json:"large"`
}

type userItem struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	About   string `json:"about"`
	Avatar  avatar `json:"avatar"`
	SiteUrl string `json:"siteUrl"`
}

type userWrapper struct {
	User userItem `json:"User"`
}

type userResponse struct {
	Data userWrapper `json:"data"`
}

const aniListURL = "https://graphql.anilist.co"

const (
	Query = `query ($id: Int, $page: Int, $perPage: Int, $search: String) {
        Page (page: $page, perPage: $perPage) {
            pageInfo {
                currentPage
                hasNextPage
                perPage
            }
            media (id: $id, search: $search, type: ANIME) {
                id
                title {
                    romaji
                }
            }
        }
    }`
	url = "https://graphql.anilist.co"
)

//

type AniUserResponse struct {
	Data struct {
		Viewer struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"Viewer"`
	} `json:"data"`
}

type UserToken struct {
	AccessToken string `json:"access_token"`
	AniName     string `json:"ani_name"`
}

// -----------------------------------------------structs pro embed
type MediaListEntry struct {
	Status   string  `json:"status"`
	Score    float64 `json:"score"`
	Progress int     `json:"progress"`
}

type MediaDetailed struct {
	ID          int        `json:"id"`
	Title       mediaTitle `json:"title"`
	Genres      []string   `json:"genres"`
	Description string     `json:"description"`
	CoverImage  struct {
		Large string `json:"large"`
		Color string `json:"color"`
	} `json:"coverImage"`
	AverageScore   int             `json:"averageScore"`
	Episodes       int             `json:"episodes"` // ou Chapters para manga
	Chapters       int             `jason:"chapters"`
	SiteUrl        string          `json:"siteUrl"`
	MediaListEntry *MediaListEntry `json:"mediaListEntry"` // Null se não logado
}

type mediaDetailedResponse struct {
	Data struct {
		Media MediaDetailed `json:"Media"`
	} `json:"data"`
}

// ----------------------------------------------------------funcoes de post pro anilist
func Query_AniList(query string, variables map[string]any) ([]byte, error) {
	// aqui cria um hashmap pra guardar query e variaveis
	body := map[string]any{
		"query":     query,
		"variables": variables,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	// faz o post pra pegar
	resp, err := http.Post(aniListURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		all, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(all))
	}

	return io.ReadAll(resp.Body)
}

func SearchMedia(mediaType, search, token string) (*MediaDetailed, error) {
	query := `
    query ($search: String, $type: MediaType) {
      Media(search: $search, type: $type) {
        id
        title { romaji }
        genres
        description(asHtml: false) 
        coverImage { large color }
        averageScore
        episodes
        chapters
        siteUrl
        mediaListEntry {
            status
            score
            progress
        }
      }
    }
    `
	vars := map[string]any{
		"search": search,
		"type":   mediaType,
	}

	// Prepara o JSON
	bodyMap := map[string]any{"query": query, "variables": vars}
	bodyBytes, _ := json.Marshal(bodyMap)

	req, _ := http.NewRequest("POST", "https://graphql.anilist.co", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// SE tiver token, adiciona no header. O AniList usa isso para preencher o 'mediaListEntry'
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("erro API: %d", resp.StatusCode)
	}

	var result mediaDetailedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Data.Media.ID == 0 {
		return nil, fmt.Errorf("nenhum resultado encontrado")
	}

	return &result.Data.Media, nil
}

// Função auxiliar para limpar a descrição (HTML -> Texto simples)
func CleanDescription(desc string) string {
	desc = strings.ReplaceAll(desc, "<br>", "\n")
	desc = strings.ReplaceAll(desc, "<i>", "*")
	desc = strings.ReplaceAll(desc, "</i>", "*")
	// Limita tamanho para não quebrar o Discord (max 4096, mas vamos por segurança em 300)
	if len(desc) > 300 {
		return desc[:300] + "..."
	}
	return desc
}
func GetUserProfile(name string) (*userItem, error) {
	query := `
	query ($name: String) {
	  User(name: $name) {
		id
		name
		about
		avatar {
		  large
		}
		siteUrl
	  }
	}
	`
	vars := map[string]any{
		"name": name,
	}
	b, err := Query_AniList(query, vars)
	if err != nil {
		return nil, err
	}
	var resp userResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}
	// Se user vazio, User zerado
	if resp.Data.User.ID == 0 && resp.Data.User.Name == "" {
		return nil, fmt.Errorf("usuário %q não encontrado", name)
	}
	return &resp.Data.User, nil
}

//
//

func GetAniListUser(token string) (string, error) {
	jsonData := map[string]string{
		"query": "{ Viewer { id name } }",
	}
	body, _ := json.Marshal(jsonData)

	req, _ := http.NewRequest("POST", "https://graphql.anilist.co", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result AniUserResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Data.Viewer.Name, nil
}
