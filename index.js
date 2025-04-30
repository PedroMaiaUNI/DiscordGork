const express = require("express");
require('dotenv').config();
const TOKEN = process.env.TOKEN;
const GITHUB_TOKEN = process.env.GITHUB_TOKEN;
const app = express();
const fs = require('fs');
app.get("/", (req, res) => res.send("Bot está vivo!"));
app.listen(3000, () => console.log("Servidor web rodando"));

const { Client, GatewayIntentBits } = require("discord.js");
const gistjson = "https://gist.githubusercontent.com/UltiMaia/92b22c76e0aef88be92e444716420398/raw/59936d89cdf0c53fbd8488a3c6390d0c17a2c4cd/tongagorkfrases.json"

const client = new Client({
  intents: [
    GatewayIntentBits.Guilds,
    GatewayIntentBits.GuildMessages,
    GatewayIntentBits.MessageContent,
  ],
});

const GIST_ID = process.env.GIST_ID;
const GIST_FILENAME = process.env.GIST_FILENAME;

const canalTeste = process.env.CANAL_TESTE;
const cargoTeste = process.env.CARGO_TESTE;

const cargoCSGO = process.env.CARGO_CSGO;
const jogos = process.env.JOGOS;

if (!TOKEN || !GITHUB_TOKEN || !GIST_ID) {
  console.error('❌ Variáveis de ambiente não carregadas corretamente.');
  process.exit(1);
}

let messageCount = 0;
const N = 200;

let respostas = [];
async function carregarRespostas() {
  const res = await fetch(`https://api.github.com/gists/${GIST_ID}`, {
    headers: { Authorization: `token ${GITHUB_TOKEN}` }
  });
  const data = await res.json();

  try {
    respostas = JSON.parse(data.files[GIST_FILENAME].content);
    console.log('✅ Frases carregadas do Gist');
    console.log(respostas)
  } catch (e) {
    console.error('❌ Erro ao ler conteúdo JSON do Gist:', e);
  }
}

async function atualizarGist(novasFrases) {
  const body = {
    files: {
      [GIST_FILENAME]: {
        content: JSON.stringify(novasFrases, null, 2)
      }
    }
  };

  const res = await fetch(`https://api.github.com/gists/${GIST_ID}`, {
    method: 'PATCH',
    headers: {
      Authorization: `token ${GITHUB_TOKEN}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(body)
  });

  if (res.ok) {
    console.log('✅ Gist atualizado com nova frase');
    respostas = novasFrases;
  } else {
    console.error('❌ Falha ao atualizar o Gist:', await res.text());
  }
}

client.once('ready', async () => {
  console.log(`🤖 Bot está online como ${client.user.tag}`);
  await carregarRespostas();

  // Atualiza as respostas a cada 10 minutos
  setInterval(carregarRespostas, 1000 * 60 * 10);
});

client.on("ready", () => {
  console.log(`Bot está online como ${client.user.tag}`);
});


client.on("messageCreate", async (message) => {

  if (message.author.bot) return;
  
  messageCount++;
  if (messageCount >= N) {
    messageCount = 0;
    const autoMsg = respostas[Math.floor(Math.random() * respostas.length)];
    message.channel.send(autoMsg.texto);
  }

  const mencoes = message.mentions.roles;
  if (mencoes.has(cargoCSGO)) {
    message.channel.send('OXALÁAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA');
    message.channel.send('https://cdn.discordapp.com/attachments/1319356140198428794/1365829109909028914/481771641_122126717024407019_8394426687156425162_n.png?ex=680ebafb&is=680d697b&hm=d51ae7095668e9fa508ff67fb69ab4923f34dba30b2658cdd802e5f0d20e1062&');
  }

  if (message.content.startsWith('!addfrase')) {
    const novaFraseTexto = message.content.replace('!addfrase', '').trim();
    if (!novaFraseTexto) return message.reply('❗ Escreva uma frase após o comando.');

    const novaFrase = {
      texto: novaFraseTexto,
      autor: `${message.author.tag}`
    };

    const novas = [...respostas, novaFrase];
    await atualizarGist(novas);
    message.reply('✅ Frase adicionada com sucesso!');
    return;
  }

  if (message.content.startsWith('!listfrases')) {
    if (respostas.length === 0) {
      return message.reply('⚠️ Nenhuma frase cadastrada ainda.');
    }

    // Extrai argumento opcional: !listfrases 5
    const partes = message.content.trim().split(' ');
    let quantidade = parseInt(partes[1]);

    if (isNaN(quantidade) || quantidade <= 0) {
      quantidade = 10; // valor padrão
    }

    // filtra links
    const semLinks = respostas.filter(f => !f.texto.match(/^https?:\/\/\S+/i));

    if (semLinks.length === 0) {
      return message.reply('⚠️ Todas as frases são links e foram ocultadas da listagem.');
    }

    const ultimas = semLinks.slice(-quantidade);
    const listagem = ultimas.map((f, i) =>
      `${semLinks.length - ultimas.length + i + 1}. ${f.texto} (por ${f.autor})`
    ).join('\n');

    if (listagem.length > 1900) {
      return message.reply('⚠️ Resultado muito longo. Tente um número menor (ex: `!listfrases 5`).');
    }

    message.reply(`📜 **Últimas ${ultimas.length} frases (sem links):**\n${listagem}`);
    return;
  }

  // comando !rmfrase
  if (message.content.startsWith('!rmfrase')) {
    const fraseAlvo = message.content.replace('!rmfrase', '').trim();
    if (!fraseAlvo) {
      return message.reply('❗ Escreva exatamente a frase que deseja remover após o comando.');
    }

    const index = respostas.findIndex(f => f.texto === fraseAlvo);
    if (index === -1) {
      return message.reply('❌ Frase não encontrada. Verifique se digitou exatamente igual.');
    }

    const novas = [...respostas];
    novas.splice(index, 1);

    await atualizarGist(novas);
    message.reply(`✅ Frase removida com sucesso:\n> ${fraseAlvo}`);
    return;
  }
  
  if (message.mentions.has(client.user) && !message.author.bot) {
    const random = respostas[Math.floor(Math.random() * respostas.length)];

    const resposta = random.texto;
    message.reply(resposta);
  }
});

client.login(TOKEN);