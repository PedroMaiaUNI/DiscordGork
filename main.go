package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"

	//"enconding/json"
	//"http"
	// "io"
	"bot/markov"
	"bot/utils"
	"github.com/robfig/cron/v3"
	"sync"
	"sync/atomic"
	"time"
)

var (
	n_mensagens        int64   = 195
	permitidos                 = []string{"332298877665411084", "703322022494732303", "271218339311910912", "981279055414456341", "205508002394931200", "274615835019051008", "515989133840351242"}
	midiacast          string  = "31/12/2025 às 23:59"
	inf                float64 = 0.99
	mc                 *markov.MarkovChain
	CSGO               string
	Emojis             []*discordgo.Emoji
	frasesCache        []utils.Frase
	frasesMu           sync.RWMutex
	Servers_permitidos = map[string]bool{
		"715343022363246642":  true, //selerom
		"918671270885851187":  true, //tonga
		"828746329093177374":  true, //maia
		"1235684622810222753": true, //ruan
		"1452723817825964137" : true, //gork 2

	}
)

const (
	DND_PATH           = "do_not_disturb.json"
	WORD_COUNTER_PATH  = "word_counter.json"
	PALAVRA_MONITORADA = "hitler"
	Tonga              = "918671270885851187"
)

func guildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	if g.Guild.Unavailable {
		return
	}

	if Servers_permitidos[g.Guild.ID] {
		return
	}

	channels, err := s.GuildChannels(g.Guild.ID)
	if err == nil {
		for _, ch := range channels {
			if ch.Type == discordgo.ChannelTypeGuildText {
				_, _ = s.ChannelMessageSend(
					ch.ID,
					"https://media.discordapp.net/attachments/490286224754475008/1209105780876247062/SPOILER_1708325187840346.gif",
				)
				break
			}
		}
	}

	_ = s.GuildLeave(g.Guild.ID)
}

func isNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func iniciarAgendador(s *discordgo.Session) {

	loc, _ := time.LoadLocation("America/Sao_Paulo")
	c := cron.New(cron.WithLocation(loc))

	_, err := c.AddFunc("0 13 * * *", func() {
		utils.Load_ImgSexta(s)
	})

	if err != nil {
		return
	}

	c.Start()
}

func handleAddFrase(s *discordgo.Session, m *discordgo.MessageCreate) {
	msg := strings.TrimSpace(strings.TrimPrefix(m.Content, "!addfrase"))
	if msg == "" || strings.Contains(msg, "@everyone") {
		return
	}

	frasesMu.Lock()
	frasesCache = append(frasesCache, utils.Frase{
		Texto: msg,
		Autor: m.Author.Username,
	})
	utils.BackupFrases(frasesCache)
	frasesMu.Unlock()

	saveFrasesAsync()

	s.ChannelMessageSendReply(m.ChannelID, "✅ Frase adicionada!", m.Reference())
}

func HandleListFrases(s *discordgo.Session, m *discordgo.MessageCreate) {
	frasesMu.RLock()
	defer frasesMu.RUnlock()

	if len(frasesCache) == 0 {
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

		for _, f := range frasesCache {
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
		for _, f := range frasesCache {
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
	for _, f := range frasesCache {
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

func saveFrasesAsync() {
	go func() {
		frasesMu.RLock()
		defer frasesMu.RUnlock()

		_ = utils.Update_Gist(
			os.Getenv("GIST_ID"),
			os.Getenv("GIST_FILENAME"),
			os.Getenv("GITHUB_TOKEN"),
			frasesCache,
		)
	}()
}

func HandleRemoveFrase(s *discordgo.Session, m *discordgo.MessageCreate) {
	alvo := strings.TrimSpace(strings.TrimPrefix(m.Content, "!rmfrase"))

	frasesMu.Lock()
	defer frasesMu.Unlock()

	for i, f := range frasesCache {
		if f.Texto == alvo {
			frasesCache = append(frasesCache[:i], frasesCache[i+1:]...)
			utils.BackupFrases(frasesCache)
			saveFrasesAsync()
			s.ChannelMessageSendReply(m.ChannelID, "✅ Frase removida.", m.Reference())
			return
		}
	}

	s.ChannelMessageSendReply(m.ChannelID, "❌ Frase não encontrada.", m.Reference())

}

func HandleMarkov(s *discordgo.Session, m *discordgo.MessageCreate) {
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

func HandleAutoMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	frasesMu.RLock()
	defer frasesMu.RUnlock()

	if len(frasesCache) > 0 && rand.Intn(2) == 0 {
		f := frasesCache[rand.Intn(len(frasesCache))]
		s.ChannelMessageSend(m.ChannelID, f.Texto)
		return
	}

	s.ChannelMessageSend(m.ChannelID, mc.Generate("", 30))
}

func HandleUndo(s *discordgo.Session, m *discordgo.MessageCreate) {
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
	frasesMu.Lock()
	frasesCache = frasesBackup
	frasesMu.Unlock()

	s.ChannelMessageSendReply(
		m.ChannelID,
		"♻️ Backup restaurado com sucesso!",
		m.Reference(),
	)
}

func HandlePalavraMonitorada(s *discordgo.Session, m *discordgo.MessageCreate, palavra string) {
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

	nowMs := time.Now().UnixMilli()

	stat, exists := counter[palavra]

	// PRIMEIRA VEZ
	if !exists {
		counter[palavra] = utils.WordStat{
			Last:       nowMs,
			Record:     0,
			RecordDate: 0,
		}

		_ = utils.Save_WordCounter(WORD_COUNTER_PATH, counter)

		s.ChannelMessageSendReply(
			m.ChannelID,
			fmt.Sprintf(`Primeira vez que falamos "%s".`, palavra),
			m.Reference(),
		)
		return
	}

	// diferença em SEGUNDOS
	diffSeconds := (nowMs - stat.Last) / 1000

	tempoAtual := utils.FormatDuration(diffSeconds)
	tempoRecorde := utils.FormatDuration(stat.Record)

	msg := fmt.Sprintf(
		"Estamos há %s sem falar \"%s\".\nNosso recorde atual é de %s.",
		tempoAtual,
		palavra,
		tempoRecorde,
	)

	// NOVO RECORDE
	if diffSeconds > stat.Record {
		stat.Record = diffSeconds
		stat.RecordDate = nowMs
		msg += "\n🎉 **Novo recorde!**"
	}

	// atualiza última ocorrência
	stat.Last = nowMs
	counter[palavra] = stat

	_ = utils.Save_WordCounter(WORD_COUNTER_PATH, counter)

	s.ChannelMessageSendReply(
		m.ChannelID,
		msg,
		m.Reference(),
	)
}

func main() {
	_ = godotenv.Load()

	token := os.Getenv("TOKEN")
	CSGO = os.Getenv("CARGO_CSGO")
	if token == "" {
		log.Fatal("TOKEN ausente")
	}

	// 🔹 Load inicial do Gist (UMA VEZ)
	frases, err := utils.Load_Gist(
		os.Getenv("GIST_ID"),
		os.Getenv("GIST_FILENAME"),
		os.Getenv("GITHUB_TOKEN"),
	)
	if err != nil {
		log.Fatal("Erro ao carregar Gist:", err)
	}
	frasesCache = frases
	fmt.Println("Frases carregadas:", len(frasesCache))

	// 🔹 Markov
	mc, err = markov.NewMarkovChain("markov_chain.json")
	if err != nil {
		log.Fatal(err)
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}

	Emojis, _ = dg.GuildEmojis(Tonga)
	Emojis = append(Emojis, &discordgo.Emoji{Name: "🫃"})

	dg.AddHandler(messageCreate)
	dg.AddHandler(guildCreate)
	if err := dg.Open(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Bot online 🚀")
	go iniciarAgendador(dg)
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
	HandlePalavraMonitorada(s, m, PALAVRA_MONITORADA)
	utils.MaybeReact(s, m, Emojis)
	utils.HandleFixEmbeds(s, m)
	utils.Handler_ImgSexta(s, m)
	if m.Author.ID == "271218339311910912" && strings.Contains(m.Content, "mygo") {
		s.MessageReactionAdd(m.ChannelID, m.Reference().MessageID, "🧩")
		s.MessageReactionAdd(m.ChannelID, m.Reference().MessageID, "🦖")
		//return
	}
	if slices.Contains(m.MentionRoles, CSGO) {
		s.ChannelMessageSendReply(
			m.ChannelID,
			"OXALAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA \nhttps://cdn.discordapp.com/attachments/1319356140198428794/1365829109909028914/481771641_122126717024407019_8394426687156425162_n.png?ex=680ebafb&is=680d697b&hm=d51ae7095668e9fa508ff67fb69ab4923f34dba30b2658cdd802e5f0d20e1062&",
			m.Reference(),
		)
		//return
	}

	if !strings.HasPrefix(m.Content, "!") {
		_ = mc.AddMessage(m.Content)
		//return
	}

	if atomic.AddInt64(&n_mensagens, 1) >= 200 || bot_mencionado(s, m) {
		HandleAutoMessage(s, m)
		atomic.StoreInt64(&n_mensagens, 0)
	}

	if m.Content == "!ping" {
		s.ChannelMessageSend(m.ChannelID, "pong 🏓")
		return
	}

	if m.Content == "!teste" {
		if slices.Contains(permitidos, m.Author.ID) {
			s.ChannelMessageSendReply(m.ChannelID, "oiiii", m.Reference())
		}
		return
	}

	if after, ok := strings.CutPrefix(m.Content, "!say "); ok {
		msg := after
		if msg == "" || strings.Contains(msg, "@everyone") {
			s.ChannelMessageSendReply(m.ChannelID, "Mensagem inválida", m.Reference())
			return
		}
		s.ChannelMessageSend(m.ChannelID, msg)
		return
	}

	if m.Content == "!num" {
		s.ChannelMessageSend(m.ChannelID, strconv.FormatInt(atomic.LoadInt64(&n_mensagens), 10))
		return
	}

	if m.Content == "!tabela" {
		s.ChannelMessageSend(m.ChannelID, "https://cdn.discordapp.com/attachments/919309611885015140/1424473155061285025/image.png?ex=694a3fc1&is=6948ee41&hm=3fa4e96cce2b043b6f42e0fb5e7e405c191f6576c63773f8854d92a5838e908d&")
	}

	switch {
	case strings.HasPrefix(m.Content, "!addfrase "):
		handleAddFrase(s, m)
		return

	case strings.HasPrefix(m.Content, "!rmfrase "):
		HandleRemoveFrase(s, m)
		return

	case strings.HasPrefix(m.Content, "!listfrases"):
		HandleListFrases(s, m)
		return

	case m.Content == "!undo":
		HandleUndo(s, m)
		return

	case m.Content == "!markov":
		HandleMarkov(s, m)
		return
		
	case strings.HasPrefix(m.Content, "!anime "):
		termo := strings.TrimSpace(m.Content[len("!anime "):])
		if termo == "" {
			s.ChannelMessageSendReply(m.ChannelID, "Use: !anime <nome>", m.Reference())
			return
		}
		title, link, err := utils.SearchMedia("ANIME", termo)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Erro: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**%s**: %s", title, link))
		return
		
	case strings.HasPrefix(m.Content, "!manga "):
		termo := strings.TrimSpace(m.Content[len("!manga "):])
		if termo == "" {
			s.ChannelMessageSendReply(m.ChannelID, "Use: !manga <nome>", m.Reference())
			return
		}
		title, link, err := utils.SearchMedia("MANGA", termo)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Erro: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**%s**: %s", title, link))
		return
		
	case strings.HasPrefix(m.Content, "!user "):
		name := strings.TrimSpace(m.Content[len("!user "):])
		if name == "" {
			s.ChannelMessageSend(m.ChannelID, "Use: !user <nome>")
			return
		}
		user, err := utils.GetUserProfile(name)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Erro: "+err.Error())
			return
		}
		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Perfil de %s", user.Name),
			URL:         user.SiteUrl,
			Description: user.About,
			Color:       0x2E51A2,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: user.Avatar.Large,
			},
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)

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
	if strings.Contains(strings.ToLower(m.Content), "qual o minimo") || strings.Contains(strings.ToLower(m.Content), "qual o mínimo") {
		s.MessageReactionAdd(m.ChannelID, m.Reference().MessageID, "🫃")
	}
	if strings.Contains(m.Content, "!attdata ") {
		if slices.Contains(permitidos, m.Author.ID) {
			nova_data := strings.TrimPrefix(m.Content, "!attdata ")
			midiacast = nova_data
			s.ChannelMessageSendReply(m.ChannelID, "data alerada para "+nova_data, m.Reference())
			return
		}
	}
	if m.Content == "!leite" {
		s.ChannelMessageSend(m.ChannelID, `
			**LEITE
ingredientes
meu pau

ferramentas
sua mão

instruções
   	1. bate uma pra mim**`)
	}

	if strings.Contains(strings.ToLower(m.Content), "is this true") {
		r := rand.Float32() // número entre 0.0 e 1.0

		var resposta string

		switch {
		case r < 0.2:
			resposta = "https://tenor.com/view/morgan-freeman-true-morgan-freeman-true-nodding-gif-13973817878387504960"
		case r < 0.4:
			resposta = "https://tenor.com/view/anon-chihaya-chihaya-anon-anon-chihaya-mygo-gif-14775622618894457051"
		case r < 0.6:
			resposta = "https://tenor.com/view/its-peak-its-mid-fight-morgan-freeman-gif-6564041502742593422"
		case r < 0.8:
			resposta = "https://tenor.com/view/chihaya-anon-anon-chihaya-anon-true-mygo-true-gif-11063547078262177235"
		default:
			resposta = "https://cdn.discordapp.com/attachments/1362454934997696642/1374740964790243399/images373.jpg?ex=682f26cb&is=682dd54b&hm=b6230e85ddd3e2ce9eb9c2bfd8dbab0d3936cac158462cac60f06a9f7fe149ca&"

		}

		s.ChannelMessageSendReply(
			m.ChannelID,
			resposta,
			m.Reference(),
		)

		return
	}

}
