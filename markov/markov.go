package markov

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

// MarkovChain é uma implementação simples de cadeia de Markov.
// - chain: mapa de palavra -> lista de palavras possíveis seguintes
// - path: arquivo onde a cadeia é persistida (JSON)
// - mu: protege acesso concorrente à chain
// - rnd: gerador de números aleatórios (testável/encapsulado)
type MarkovChain struct {
	mu    sync.RWMutex
	chain map[string][]string
	path  string
	rnd   *rand.Rand
}

// NewMarkovChain cria uma nova cadeia e tenta carregar dados do arquivo path.
// Se path for vazio, usa "markov_chain.json".
func NewMarkovChain(path string) (*MarkovChain, error) {
	if path == "" {
		path = "markov_chain.json"
	}
	mc := &MarkovChain{
		chain: make(map[string][]string),
		path:  path,
		rnd:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	if err := mc.load(); err != nil {
		// Se o arquivo não existir, load retorna nil; somente erros reais são retornados.
		return mc, err
	}
	return mc, nil
}

// AddMessage adiciona as transições de uma mensagem à cadeia e salva em disco.
// Divide a mensagem por whitespace, converte para minúsculas e registra pares (word -> next).
func (m *MarkovChain) AddMessage(msg string) error {
	words := strings.Fields(msg)
	if len(words) < 2 {
		// nada a adicionar se tiver menos de 2 palavras
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for i := 0; i < len(words)-1; i++ {
		w := strings.ToLower(words[i])
		next := strings.ToLower(words[i+1])
		m.chain[w] = append(m.chain[w], next)
	}
	// Salva após cada adição (como no exemplo JS)
	return m.save()
}

// Generate gera uma sequência a partir de startWord (se startWord == "", escolhe aleatoriamente).
// maxLen limita o número de palavras (incluindo a inicial).
func (m *MarkovChain) Generate(startWord string, maxLen int) string {
	if maxLen <= 0 {
		maxLen = 30
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.chain) == 0 {
		return ""
	}

	word := ""
	if startWord == "" {
		word = m.randomWordLocked()
	} else {
		word = strings.ToLower(startWord)
	}

	result := []string{word}
	for i := 1; i < maxLen; i++ {
		nexts := m.chain[word]
		if len(nexts) == 0 {
			break
		}
		word = nexts[m.rnd.Intn(len(nexts))]
		result = append(result, word)
	}
	return strings.Join(result, " ")
}

// randomWordLocked escolhe uma chave aleatória da cadeia.
// Deve ser chamado com read lock ou lock mantido pelo chamador.
func (m *MarkovChain) randomWordLocked() string {
	// construir slice de chaves
	keys := make([]string, 0, len(m.chain))
	for k := range m.chain {
		keys = append(keys, k)
	}
	if len(keys) == 0 {
		return ""
	}
	return keys[m.rnd.Intn(len(keys))]
}

// save escreve a cadeia para o arquivo JSON de forma atômica (temp + rename).
// Deve ser chamado com o lock apropriado pelo chamador.
func (m *MarkovChain) save() error {
	data, err := json.MarshalIndent(m.chain, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	tmp := m.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	if err := os.Rename(tmp, m.path); err != nil {
		return fmt.Errorf("rename tmp: %w", err)
	}
	return nil
}

// load tenta carregar a cadeia do arquivo JSON.
// Se o arquivo não existir, retorna nil sem erro.
func (m *MarkovChain) load() error {
	if _, err := os.Stat(m.path); os.IsNotExist(err) {
		// sem arquivo — começar vazio
		return nil
	} else if err != nil {
		return fmt.Errorf("stat: %w", err)
	}

	data, err := os.ReadFile(m.path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := json.Unmarshal(data, &m.chain); err != nil {
		// se falhar, zera a cadeia para evitar estado inconsistente
		m.chain = make(map[string][]string)
		return fmt.Errorf("unmarshal: %w", err)
	}
	return nil
}
