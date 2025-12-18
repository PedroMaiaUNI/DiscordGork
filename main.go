package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"syscall"
	//"enconding/json"
	//"http"
	// "io"
	"bot/markov"
	"bot/utils"
	"sync/atomic"
	"time"
)

var (
	n_mensagens int64   = 199
	permitidos          = []string{"332298877665411084", "703322022494732303", "271218339311910912", "981279055414456341", "205508002394931200", "274615835019051008", "515989133840351242"}
	midiacast   string  = "31/12/2025 às 23:59"
	inf         float64 = 0.99
	mc          *markov.MarkovChain
	CSGO        string
)

const (
	imagensPath        = "imagensSexta.json"
	DND_PATH           = "do_not_disturb.json"
	WORD_COUNTER_PATH  = "word_counter.json"
	PALAVRA_MONITORADA = "hitler"
	TONGA              = "918671270885851187"
)

func isNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func handleListFrases(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Sempre carrega do Gist (fonte da verdade)
	respostas, err := utils.Load_Gist(
		os.Getenv("GIST_ID"),
		os.Getenv("GIST_FILENAME"),
		os.Getenv("GITHUB_TOKEN"),
	)
	if err != nil {
		s.ChannelMessageSendReply(
			m.ChannelID,
			"❌ Erro ao carregar frases.",
			m.Reference(),
		)
		return
	}

	// JS: if (respostas.length === 0)
	if len(respostas) == 0 {
		s.ChannelMessageSendReply(
			m.ChannelID,
			"⚠️ Nenhuma frase cadastrada ainda.",
			m.Reference(),
		)
		return
	}

	// args / termo
	args := strings.Fields(m.Content)
	termo := ""
	if len(args) > 1 {
		termo = strings.Join(args[1:], " ")
	}
	termo = strings.TrimSpace(termo)

	// 🔍 Busca por texto (não número, não autor, não menção)
	if termo != "" &&
		!isNumber(termo) &&
		!strings.Contains(termo, "#") &&
		!strings.HasPrefix(termo, "<@") {

		for _, f := range respostas {
			if strings.Contains(
				strings.ToLower(f.Texto),
				strings.ToLower(termo),
			) {
				s.ChannelMessageSendReply(
					m.ChannelID,
					fmt.Sprintf(
						"🧾 Frase encontrada:\n\"%s\"\n👤 Autor: %s",
						f.Texto,
						f.Autor,
					),
					m.Reference(),
				)
				return
			}
		}

		s.ChannelMessageSendReply(
			m.ChannelID,
			fmt.Sprintf(
				"❌ Nenhuma frase contendo \"%s\" foi encontrada.",
				termo,
			),
			m.Reference(),
		)
		return
	}

	// 👤 Filtro por autor (# ou <@>)
	if strings.Contains(termo, "#") || strings.HasPrefix(termo, "<@") {
		autorFiltro := termo

		// menção <@>
		if strings.HasPrefix(termo, "<@") && len(m.Mentions) > 0 {
			autorFiltro = m.Mentions[0].Username
		}

		if autorFiltro == "" {
			s.ChannelMessageSendReply(
				m.ChannelID,
				"❌ Não foi possível identificar o autor.",
				m.Reference(),
			)
			return
		}

		var frasesAutor []utils.Frase
		for _, f := range respostas {
			if f.Autor == autorFiltro {
				frasesAutor = append(frasesAutor, f)
			}
		}

		if len(frasesAutor) == 0 {
			s.ChannelMessageSendReply(
				m.ChannelID,
				fmt.Sprintf(
					"❌ Nenhuma frase encontrada para o autor %s",
					autorFiltro,
				),
				m.Reference(),
			)
			return
		}

		var b strings.Builder
		for i, f := range frasesAutor {
			fmt.Fprintf(&b, "%d. %s\n", i+1, f.Texto)
		}

		msg := b.String()
		if len(msg) > 1900 {
			msg = msg[:1900]
		}

		s.ChannelMessageSendReply(
			m.ChannelID,
			fmt.Sprintf(
				"📚 **Frases de %s:**\n%s",
				autorFiltro,
				msg,
			),
			m.Reference(),
		)
		return
	}

	// 🔢 Quantidade
	quantidade := 10
	if isNumber(termo) {
		q, _ := strconv.Atoi(termo)
		if q > 0 {
			quantidade = q
		}
	}

	// 🚫 Remove links
	var semLinks []utils.Frase
	for _, f := range respostas {
		if !strings.HasPrefix(strings.ToLower(f.Texto), "http://") &&
			!strings.HasPrefix(strings.ToLower(f.Texto), "https://") {
			semLinks = append(semLinks, f)
		}
	}

	if len(semLinks) == 0 {
		s.ChannelMessageSendReply(
			m.ChannelID,
			"⚠️ Todas as frases são links e foram ocultadas da listagem.",
			m.Reference(),
		)
		return
	}

	if quantidade > len(semLinks) {
		quantidade = len(semLinks)
	}

	ultimas := semLinks[len(semLinks)-quantidade:]

	var b strings.Builder
	offset := len(semLinks) - len(ultimas)
	for i, f := range ultimas {
		fmt.Fprintf(
			&b,
			"%d. %s (por %s)\n",
			offset+i+1,
			f.Texto,
			f.Autor,
		)
	}

	msg := b.String()
	if len(msg) > 1900 {
		s.ChannelMessageSendReply(
			m.ChannelID,
			"⚠️ Resultado muito longo. Tente um número menor (ex: `!listfrases 5`).",
			m.Reference(),
		)
		return
	}

	s.ChannelMessageSendReply(
		m.ChannelID,
		fmt.Sprintf(
			"📜 **Últimas %d frases (sem links):**\n%s",
			len(ultimas),
			msg,
		),
		m.Reference(),
	)
}

func handleRemoveFrase(s *discordgo.Session, m *discordgo.MessageCreate) {
	// SEMPRE recarrega do Gist (fonte da verdade)
	frases, err := utils.Load_Gist(
		os.Getenv("GIST_ID"),
		os.Getenv("GIST_FILENAME"),
		os.Getenv("GITHUB_TOKEN"),
	)
	if err != nil {
		s.ChannelMessageSendReply(
			m.ChannelID,
			"❌ Erro ao carregar frases do Gist. Operação cancelada.",
			m.Reference(),
		)
		log.Println("Load_Gist no rmfrase:", err)
		return
	}

	fraseAlvo := strings.TrimSpace(
		strings.TrimPrefix(m.Content, "!rmfrase"),
	)

	if fraseAlvo == "" {
		s.ChannelMessageSendReply(
			m.ChannelID,
			"❗ Escreva exatamente a frase que deseja remover após o comando.",
			m.Reference(),
		)
		return
	}

	index := -1
	for i, f := range frases {
		if f.Texto == fraseAlvo {
			index = i
			break
		}
	}

	if index == -1 {
		s.ChannelMessageSendReply(
			m.ChannelID,
			"❌ Frase não encontrada. Verifique se digitou exatamente igual.",
			m.Reference(),
		)
		return
	}

	// Remove da lista correta
	novaLista := append(
		frases[:index],
		frases[index+1:]...,
	)

	// Salva no Gist
	err = utils.Update_Gist(
		os.Getenv("GIST_ID"),
		os.Getenv("GIST_FILENAME"),
		os.Getenv("GITHUB_TOKEN"),
		novaLista,
	)
	if err != nil {
		s.ChannelMessageSendReply(
			m.ChannelID,
			"❌ Erro ao salvar a remoção no Gist.",
			m.Reference(),
		)
		log.Println("Update_Gist no rmfrase:", err)
		return
	}

	// Atualiza cache local
	frases = novaLista

	s.ChannelMessageSendReply(
		m.ChannelID,
		"✅ Frase removida com sucesso.",
		m.Reference(),
	)
}

func handleMarkov(s *discordgo.Session, m *discordgo.MessageCreate) {
	texto := mc.Generate("", 30)

	if texto == "" {
		s.ChannelMessageSendReply(
			m.ChannelID,
			"⚠️ Ainda não tenho dados suficientes para gerar texto.",
			m.Reference(),
		)
		return
	}

	s.ChannelMessageSend(m.ChannelID, texto)
}

func handleAutoMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	frases, err := utils.Load_Gist(
		os.Getenv("GIST_ID"),
		os.Getenv("GIST_FILENAME"),
		os.Getenv("GITHUB_TOKEN"),
	)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "ERRO AO CARREGAR O GIST")
		return
	}

	// 50% frase, 50% markov
	if len(frases) > 0 && rand.Intn(2) == 0 {
		// escolhe frase aleatória
		f := frases[rand.Intn(len(frases))]
		s.ChannelMessageSend(
			m.ChannelID,
			f.Texto,
		)
		fmt.Println("frase")
		return
	}

	// fallback ou escolha Markov
	texto := mc.Generate("", 30)
	if texto != "" {
		s.ChannelMessageSend(
			m.ChannelID,
			texto,
		)
		fmt.Println("markovgpt")
	}
}

func handleUndo(s *discordgo.Session, m *discordgo.MessageCreate) {
	frasesBackup, err := utils.LoadLastBackup()
	if err != nil {
		s.ChannelMessageSendReply(
			m.ChannelID,
			"❌ Nenhum backup encontrado para restaurar.",
			m.Reference(),
		)
		return
	}

	err = utils.Update_Gist(
		os.Getenv("GIST_ID"),
		os.Getenv("GIST_FILENAME"),
		os.Getenv("GITHUB_TOKEN"),
		frasesBackup,
	)
	if err != nil {
		s.ChannelMessageSendReply(
			m.ChannelID,
			"❌ Erro ao restaurar o backup.",
			m.Reference(),
		)
		return
	}

	s.ChannelMessageSendReply(
		m.ChannelID,
		"♻️ Backup restaurado com sucesso!",
		m.Reference(),
	)
}

func handlePalavraMonitorada(s *discordgo.Session, m *discordgo.MessageCreate, palavra string) {
	content := strings.ToLower(m.Content)
	palavra = strings.ToLower(palavra)

	if !strings.Contains(content, palavra) {
		return
	}

	counter, err := utils.Load_WordCounter(WORD_COUNTER_PATH)
	if err != nil {
		log.Println("load word counter:", err)
		return
	}

	now := time.Now().Unix()

	stat, exists := counter[palavra]

	//  PRIMEIRA VEZ
	if !exists {
		counter[palavra] = utils.WordStat{
			LastTime: now,
			Record:   0,
		}

		_ = utils.Save_WordCounter(WORD_COUNTER_PATH, counter)

		s.ChannelMessageSendReply(
			m.ChannelID,
			fmt.Sprintf(`Primeira vez que falamos "%s".`, palavra),
			m.Reference(),
		)
		return
	}

	//  JÁ EXISTE → COMPARA TEMPO
	diff := now - stat.LastTime

	tempoAtual := utils.FormatDuration(diff)
	tempoRecorde := utils.FormatDuration(stat.Record)

	msg := fmt.Sprintf(
		"Estamos há %s sem falar \"%s\".\nNosso recorde atual é de %s.",
		tempoAtual,
		palavra,
		tempoRecorde,
	)

	if diff > stat.Record {
		stat.Record = diff
		msg += "\n🎉 **Novo recorde!**"
	}

	//  ATUALIZA APENAS DEPOIS DA COMPARAÇÃO
	stat.LastTime = now
	counter[palavra] = stat

	_ = utils.Save_WordCounter(WORD_COUNTER_PATH, counter)

	s.ChannelMessageSendReply(
		m.ChannelID,
		msg,
		m.Reference(),
	)
}

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Erro ao carregar .env:", err)
	}

	Token := os.Getenv("TOKEN")
	if Token == "" {
		log.Fatal("TOKEN não definido")
	}
	dnd, err := utils.Load_DND(DND_PATH)
	if err != nil {
		log.Fatal(err)
	}

	wordCounter, err := utils.Load_WordCounter(WORD_COUNTER_PATH)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("DND e Word counter carregados:", len(dnd), " ", len(wordCounter))
	frases, err := utils.Load_Gist(
		os.Getenv("GIST_ID"),
		os.Getenv("GIST_FILENAME"),
		os.Getenv("GITHUB_TOKEN"),
	)
	//fmt.Println("GIST_ID:", os.Getenv("GIST_ID"))
	//fmt.Println("GIST_FILENAME:", os.Getenv("GIST_FILENAME"))
	//fmt.Println("GITHUB_TOKEN vazio?:", os.Getenv("GITHUB_TOKEN") == "")

	fmt.Println("Frases carregadas:", len(frases))
	if err != nil {
		log.Fatal(err)
	}
	CSGO = os.Getenv("CARGO_CSGO")

	mc, err = markov.NewMarkovChain("markov_chain.json")
	if err != nil {
		log.Fatal("Erro ao carregar Markov:", err)
	}
	fmt.Println("Cadeia de Markov carregada")
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Erro ao criar sessão:", err)
		return
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("Erro ao conectar:", err)
		return
	}
	fmt.Println("Bot está rodando. Pressione CTRL+C para sair.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	dg.Close()
}

func bot_mencionado(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	for _, user := range m.Mentions {
		if user.ID == s.State.User.ID {
			return true
		}
	}
	return false
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}
	utils.HandleFixEmbeds(s, m)
	if slices.Contains(m.MentionRoles, CSGO) {
		s.ChannelMessageSendReply(
			m.ChannelID,
			"OXALAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA \nhttps://cdn.discordapp.com/attachments/1319356140198428794/1365829109909028914/481771641_122126717024407019_8394426687156425162_n.png?ex=680ebafb&is=680d697b&hm=d51ae7095668e9fa508ff67fb69ab4923f34dba30b2658cdd802e5f0d20e1062&",
			m.Reference(),
		)
		return
	}
	
	if !strings.HasPrefix(m.Content, "!") {
		_ = mc.AddMessage(m.Content)
	}

	if n_mensagens == 200 || bot_mencionado(s, m) {
		handleAutoMessage(s, m)
		atomic.StoreInt64(&n_mensagens, 0)
	}

	atomic.AddInt64(&n_mensagens, 1)

	if m.Content == "!ping" {
		s.ChannelMessageSend(m.ChannelID, "pong 🏓")
	}

	if m.Content == "!teste" {
		if slices.Contains(permitidos, m.Author.ID) {
			s.ChannelMessageSendReply(m.ChannelID, "oiiii", m.Reference())
		}
	}

	if after, ok := strings.CutPrefix(m.Content, "!say "); ok {
		msg := after
		if msg == "" || strings.Contains(msg, "@everyone") {
			s.ChannelMessageSendReply(m.ChannelID, "Mensagem inválida", m.Reference())
			return
		}
		s.ChannelMessageSend(m.ChannelID, msg)
	}

	if m.Content == "!num" {
		s.ChannelMessageSend(m.ChannelID, strconv.FormatInt(atomic.LoadInt64(&n_mensagens), 10))
	}

	if m.Content == "!tabela" {
		s.ChannelMessageSend(m.ChannelID, "https://cdn.discordapp.com/attachments/1319356140198428794/1445147334286770357/image.png?ex=692f49d6&is=692df856&hm=003297ec638848ada09a99b0a64f18e01530c931a189467783a47db6d3ab7523&")
	}

	if after, ok := strings.CutPrefix(m.Content, "!addfrase "); ok {
		msg := after
		if msg == "" || strings.Contains(msg, "@everyone") {
			s.ChannelMessageSendReply(
				m.ChannelID,
				"Mensagem inválida",
				m.Reference(),
			)
			return
		}

		//  SEMPRE recarrega do Gist (fonte da verdade)
		frases, err := utils.Load_Gist(
			os.Getenv("GIST_ID"),
			os.Getenv("GIST_FILENAME"),
			os.Getenv("GITHUB_TOKEN"),
		)
		if err != nil {
			s.ChannelMessageSendReply(
				m.ChannelID,
				"❌ Erro ao carregar frases do Gist. Operação cancelada.",
				m.Reference(),
			)
			log.Println("Load_Gist no addfrase:", err)
			return
		}
		if len(frases) == 0 {
			s.ChannelMessageSendReply(
				m.ChannelID,
				"⚠️ Gist retornou vazio. Salvamento bloqueado por segurança.",
				m.Reference(),
			)
			return
		}

		nova := utils.Frase{
			Texto: msg,
			Autor: m.Author.Username,
		}

		frases = append(frases, nova)

		err = utils.Update_Gist(
			os.Getenv("GIST_ID"),
			os.Getenv("GIST_FILENAME"),
			os.Getenv("GITHUB_TOKEN"),
			frases,
		)
		if err != nil {
			s.ChannelMessageSendReply(
				m.ChannelID,
				"❌ Erro ao salvar a frase.",
				m.Reference(),
			)
			log.Println("Update_Gist no addfrase:", err)
			return
		}

		s.ChannelMessageSendReply(
			m.ChannelID,
			"✅ Frase adicionada com sucesso!",
			m.Reference(),
		)
	}

	if strings.HasPrefix(m.Content, "!listfrases") {
		handleListFrases(s, m)
	}

	if strings.HasPrefix(m.Content, "!rmfrase ") {
		frasesAtuais, err := utils.Load_Gist(
			os.Getenv("GIST_ID"),
			os.Getenv("GIST_FILENAME"),
			os.Getenv("GITHUB_TOKEN"),
		)
		if err == nil {
			_ = utils.BackupFrases(frasesAtuais)
		}
		handleRemoveFrase(s, m)

	}

	if m.Content == "!markov" {
		handleMarkov(s, m)
	}

	if m.Content == "!undo" {
		handleUndo(s, m)
	}

	if strings.Contains(m.Content, "hitler") {
		handlePalavraMonitorada(s, m, "hitler")
	}

	if strings.Contains(m.Content, "!inf ") {
		if slices.Contains(permitidos, m.Author.ID) {
			msg := strings.TrimPrefix(m.Content, "!inf ")
			if msg == "" || !isNumber(msg) {
				s.ChannelMessageSendReply(
					m.ChannelID,
					"Mensagem inválida",
					m.Reference(),
				)
				return
			}
			inf, _ = strconv.ParseFloat(msg, 64)
			s.ChannelMessageSendReply(
				m.ChannelID,
				"Probabilidade alterada.",
				m.Reference(),
			)
			return
		}
		
	}

	if strings.Contains(strings.ToLower(m.Content), "quando") &&
		strings.Contains(strings.ToLower(m.Content), "acaba") &&
		strings.Contains(strings.ToLower(m.Content), "midiacast") {
		s.ChannelMessageSendReply(m.ChannelID, midiacast, m.Reference())
		return
	}
	
	if strings.Contains(m.Content, "!attdata "){
		if slices.Contains(permitidos, m.Author.ID){
			nova_data := strings.TrimPrefix(m.Content, "!attdata ")
			midiacast = nova_data
			s.ChannelMessageSendReply(m.ChannelID,"data alerada para " + nova_data, m.Reference())
			return
		}
	}
}
