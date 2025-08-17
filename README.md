# 🤖 Bot Discord – Frases Aleatórias com Gist

> @gork isso eh verdade ?

Bot paródia do Grok, a IA do Twitter (atualmente X). 

Um bot do Discord que responde automaticamente com frases aleatórias. As frases são armazenadas em um **Gist secreto do GitHub**, permitindo que sejam atualizadas remotamente sem reiniciar o bot.

Tenha em mente que esse projeto foi feito apenas para entretenimento, e boa parte do código escrito teve auxílio do ChatGPT. Eu não tenho intenção de demonstrar minha habilidades com esse projeto, mas apenas manter ele como uma memória de uma piada engraçadinha que fiz com meus amigos no nosso servidor do Discord. 

---

## 🚀 Funcionalidades

- 🎲 Responde automaticamente com frases ao ser mencionado
- ➕ `!addfrase` – Adiciona nova frase ao Gist
- 📜 `!listfrases "frase"` – Lista uma frase especifica armazenada no Gist
- 📜 `!listfrases [n]` – Lista as últimas frases (sem links), padrão: 10
- 📜 `!listfrases @[username]` – Lista as últimas frases de um autor específico
- ❌ `!rmfrase` – Remove uma frase exata
- ⏰ Envia mensagens automáticas a cada N mensagens no servidor
- 🔁 Atualiza frases do Gist periodicamente
- 🧠 Gera mensagens novas a partir de uma [Cadeia de Markov](https://pt.wikipedia.org/wiki/Cadeias_de_Markov), que armazena o conteudo enviado no servidor.
