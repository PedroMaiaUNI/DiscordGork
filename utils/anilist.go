package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func SearchMedia(mediaType, search string) (title, link string, err error) {
	query := `
		query ($id: Int, $page: Int, $perPage: Int, $search: String) {
		  Page (page: $page, perPage: $perPage) {
			media (id: $id, search: $search, type: ` + mediaType + `) {
			  id
			  title {
				romaji
			  }
			}
		  }
		}
		`
	vars := map[string]any{
		"search":  search,
		"page":    1,
		"perPage": 3,
	}
	b, err := Query_AniList(query, vars)
	if err != nil {
		return "", "", err
	}
	var resp mediaResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return "", "", err
	}
	if len(resp.Data.Page.Media) == 0 {
		return "", "", fmt.Errorf("nenhum resultado para %q", search)
	}
	m := resp.Data.Page.Media[0]
	title = m.Title.Romaji
	if mediaType == "ANIME" {
		link = fmt.Sprintf("https://anilist.co/anime/%d/", m.ID)
	} else {
		link = fmt.Sprintf("https://anilist.co/manga/%d/", m.ID)
	}

	return title, link, nil

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
