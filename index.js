const express = require("express");
require('dotenv').config();
const fs = require('fs');
const cron = require('node-cron');
const { Client, GatewayIntentBits, Partials } = require("discord.js");
const MarkovChain = require('./markov.js');

const TOKEN = process.env.TOKEN;
const GITHUB_TOKEN = process.env.GITHUB_TOKEN;
const GIST_ID = process.env.GIST_ID;
const GIST_FILENAME = process.env.GIST_FILENAME;
const canalTeste = process.env.CANAL_TESTE;
const cargoTeste = process.env.CARGO_TESTE;
const cargoCSGO = process.env.CARGO_CSGO;
const jogos = process.env.JOGOS;
const WORD_COUNTER_PATH = 'word_counter.json';
const markov = new MarkovChain();
let inf = 0.95;

// Celeron --- Hokster --- Outachi --- Gull --- Marsh --- Maia
const permitidos = ["332298877665411084","703322022494732303","271218339311910912","981279055414456341", "205508002394931200", "274615835019051008", "515989133840351242"];

const DND_PATH = 'do_not_disturb.json';

function carregarDoNotDisturb() {
  try {
    if (fs.existsSync(DND_PATH)) {
      return JSON.parse(fs.readFileSync(DND_PATH, 'utf8'));
    }
  } catch (e) {
    console.error('Erro ao carregar DO_NOT_DISTURB:', e);
  }
  return [];
}

function salvarDoNotDisturb(lista) {
  try {
    fs.writeFileSync(DND_PATH, JSON.stringify(lista, null, 2), 'utf8');
  } catch (e) {
    console.error('Erro ao salvar DO_NOT_DISTURB:', e);
  }
}

let DO_NOT_DISTURB = carregarDoNotDisturb();

function carregarWordCounter() {
  try {
    if (fs.existsSync(WORD_COUNTER_PATH)) {
      return JSON.parse(fs.readFileSync(WORD_COUNTER_PATH, 'utf8'));
    }
  } catch (e) {
    console.error('Erro ao carregar word_counter:', e);
  }
  return {};
}

function salvarWordCounter(data) {
  try {
    fs.writeFileSync(WORD_COUNTER_PATH, JSON.stringify(data, null, 2), 'utf8');
  } catch (e) {
    console.error('Erro ao salvar word_counter:', e);
  }
}

let wordCounter = carregarWordCounter();


const app = express();
app.get("/", (req, res) => res.send("Bot está vivo!"));
app.listen(3000, () => console.log("Servidor web rodando"));

const client = new Client({
  intents: [
    GatewayIntentBits.Guilds,
    GatewayIntentBits.GuildMessages,
    GatewayIntentBits.MessageContent,
    GatewayIntentBits.DirectMessages,
  ],
  partials: [Partials.Channel], // Necessário para receber DMs
});

// --- Utilitários para imagens de sexta-feira ---
const imagensPath = 'imagensSexta.json';
let imagensSexta = [];
function carregarImagensSexta() {
  try {
    imagensSexta = JSON.parse(fs.readFileSync(imagensPath));
  } catch (e) {
    imagensSexta = [];
  }
}
function salvarImagensSexta() {
  try {
    fs.writeFileSync(imagensPath, JSON.stringify(imagensSexta, null, 2));
  } catch (e) {
    console.error('Erro ao salvar imagens de sexta:', e);
  }
}
carregarImagensSexta();

// --- Utilitários para frases ---
let respostas = [];
async function carregarRespostas() {
  const res = await fetch(`https://api.github.com/gists/${GIST_ID}`, {
    headers: { Authorization: `token ${GITHUB_TOKEN}` }
  });
  const data = await res.json();
  try {
    respostas = JSON.parse(data.files[GIST_FILENAME].content);
    console.log('✅ Frases carregadas do Gist');
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

// --- Eventos do bot ---
client.once('ready', async () => {
  console.log(`🤖 Bot está online como ${client.user.tag}`);
  await carregarRespostas();
  setInterval(carregarRespostas, 1000 * 60 * 10); // Atualiza as respostas a cada 10 minutos
});

client.on("ready", () => {
  console.log(`Bot está online como ${client.user.tag}`);
  try {
    const guild = client.guilds.cache.get('918671270885851187'); // Tonga
    const emotesArray = [...guild.emojis.cache.map(e => e), '🫃'];
    if (emotesArray) console.log("emotes obtidos");
    global.emotesArray = emotesArray;
  } catch (e) {
    console.log("Not tonga");
  }
});

const PALAVRA_MONITORADA = "hitler";
// --- Handler principal de mensagens ---
client.on("messageCreate", async (message) => {
  if (message.author.bot) return;
  
  const tracker = message.content.toLowerCase();
  // Alimenta a cadeia de markov com as mensagens lidas
  try {
    markov.addMessage(message.content);
  } catch (e) {
    console.error('❌ Ocorreu um erro ao ler esta mensagem: ', e);
  }

  if (message.content.startsWith('!markov')) {
  const resposta = markov.generate();
  await message.channel.send(resposta);
  return;
  }
  
  // Salva imagens recebidas por DM
  if (message.channel.type === 1 && message.attachments.size > 0) {
    try {
      for (const [, attachment] of message.attachments) {
        if (attachment.contentType && attachment.contentType.startsWith('image/')) {
          imagensSexta.push(attachment.url);
          salvarImagensSexta();
          await message.reply('✅ Imagem recebida e salva para o sorteio de sexta-feira!');
        }
      }
    } catch (e) {
      console.error('Erro ao salvar imagem recebida por DM:', e);
      await message.reply('❌ Ocorreu um erro ao salvar sua imagem.');
    }
    return;
  }
  

  // Reage com emote aleatório
  if (global.emotesArray && Math.random() < 0.005) {
    const emote = global.emotesArray[Math.floor(Math.random() * global.emotesArray.length)];
    try {
      await message.react(emote.id ? emote.id : emote);
    } catch (e) {
      console.log('Falha ao reagir com emote:', e);
    }
  }

  // Mensagem automática a cada N mensagens
  let messageCount = 0;
  const N = 200;
  messageCount++;
  if (messageCount >= N) {
    messageCount = 0;
    let texto;
    if (Math.random() < 0.5 && respostas.length > 0) {
      texto = respostas[Math.floor(Math.random() * respostas.length)].texto;
    } else {
      texto = markov.generate();
    }
    message.channel.send(texto);
  }

  // Comando para mencionar cargo CSGO
  const mencoes = message.mentions.roles;
  if (mencoes.has(cargoCSGO)) {
    message.channel.send('OXALÁAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA');
    message.channel.send('https://cdn.discordapp.com/attachments/1319356140198428794/1365829109909028914/481771641_122126717024407019_8394426687156425162_n.png?ex=680ebafb&is=680d697b&hm=d51ae7095668e9fa508ff67fb69ab4923f34dba30b2658cdd802e5f0d20e1062&');
  }

  // --- Comandos de frases ---
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

  // Função utilitária para remover acentos
  function removerAcentos(str) {
    return str.normalize("NFD").replace(/[\u0300-\u036f]/g, "");
  }
  const conteudo = removerAcentos(message.content.toLowerCase());

  // Respostas automáticas
  if (
    (conteudo.includes("acaba") && conteudo.includes("midiacast")) ||
    (conteudo.includes("termina") && conteudo.includes("midiacast"))
  ) {
    message.reply("14/12/2025 as 23:59");
  }
  if (conteudo.includes("qual o minimo")) {
    message.react('🫃');
  }
  if (conteudo.includes("is this true")) {
    if (Math.random() < 0.2) {
      message.reply('https://tenor.com/view/morgan-freeman-true-morgan-freeman-true-nodding-gif-13973817878387504960');
    } else if (Math.random() < 0.4){
      message.reply('https://tenor.com/view/anon-chihaya-chihaya-anon-anon-chihaya-mygo-gif-14775622618894457051')
    } else if (Math.random() < 0.6) {
      message.reply('https://tenor.com/view/its-peak-its-mid-fight-morgan-freeman-gif-6564041502742593422');
    } else if (Math.random < 0.8) {
      message.reply('https://tenor.com/view/chihaya-anon-anon-chihaya-anon-true-mygo-true-gif-11063547078262177235')
    } else {
      message.reply('https://cdn.discordapp.com/attachments/1362454934997696642/1374740964790243399/images373.jpg?ex=682f26cb&is=682dd54b&hm=b6230e85ddd3e2ce9eb9c2bfd8dbab0d3936cac158462cac60f06a9f7fe149ca&');
    }
    return;
  }

  // --- Comando !listfrases ---
  if (message.content.startsWith('!listfrases')) {
    if (respostas.length === 0) {
      return message.reply('⚠️ Nenhuma frase cadastrada ainda.');
    }
    const args = message.content.trim().split(' ').slice(1);
    const termo = args.join(' ').trim();
    if (termo && !termo.match(/^\d+$/) && !termo.includes('#') && !termo.startsWith('<@')) {
      const encontrada = respostas.find(f => f.texto.toLowerCase().includes(termo.toLowerCase()));
      if (encontrada) {
        return message.reply(`🧾 Frase encontrada:\n"${encontrada.texto}"\n👤 Autor: ${encontrada.autor}`);
      } else {
        return message.reply(`❌ Nenhuma frase contendo "${termo}" foi encontrada.`);
      }
    }
    if (termo.includes('#') || termo.startsWith('<@')) {
      const autorFiltro = termo.startsWith('<@')
        ? message.mentions.users.first()?.tag
        : termo;
      if (!autorFiltro) return message.reply('❌ Não foi possível identificar o autor.');
      const frasesAutor = respostas.filter(f => f.autor === autorFiltro);
      if (frasesAutor.length === 0) {
        return message.reply(`❌ Nenhuma frase encontrada para o autor ${autorFiltro}`);
      }
      const listagem = frasesAutor.map((f, i) => `${i + 1}. ${f.texto}`).join('\n');
      const respostaFinal = `📚 **Frases de ${autorFiltro}:**\n${listagem.slice(0, 1900)}`;
      return message.reply(respostaFinal);
    }
    let quantidade = parseInt(termo);
    if (isNaN(quantidade) || quantidade <= 0) quantidade = 10;
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

  // --- Comando !rmfrase ---
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

  // --- Menciona o bot ---
 if (message.mentions.has(client.user) && !message.author.bot && !message.content.startsWith('!')) {
    let texto;
    try {
      if (Math.random() < 0.5 && respostas.length > 0) {
        texto = respostas[Math.floor(Math.random() * respostas.length)].texto;
      } else {
        texto = markov.generate();
      }
      message.reply(texto);
    } catch(e) {
        await message.send("erro ao responder essa mensagem.");
    }
  }
  
  // --- Comando !gozei ---
  let ultimoGozado = null;
  function shuffle(array) {
    for (let i = array.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [array[i], array[j]] = [array[j], array[i]];
    }
    return array;
  }
  if (message.content.startsWith('!gozei')) {
    const user = message.author.id
    DO_NOT_DISTURB = carregarDoNotDisturb();
    if (DO_NOT_DISTURB.includes(user)) {
        return message.reply('❌ Você está no modo "do not disturb" e não pode usar este comando.');
    }
    try {
      if (!message.member.permissions.has('ManageMessages') ) {
        return message.reply('❌ Você não tem permissão para usar este comando.');
      }
      const guild = message.guild;
      let galaRole = guild.roles.cache.find(role => role.name === 'gozado');
      let members = guild.members.cache.filter(m => 
        !m.user.bot && !carregarDoNotDisturb().includes(m.id)
      );
      if (members.size === 0) return message.reply('Não há membros humanos no servidor!');
      let pool = members;
      if (ultimoGozado && members.has(ultimoGozado)) {
        pool = members.filter(m => m.id !== ultimoGozado);
        if (pool.size === 0) pool = members;
      }
      const shuffled = shuffle(Array.from(pool.values()));
      let victim = shuffled[0];
      if (!galaRole) {
        await message.channel.send('Criando cargo...');
        galaRole = await guild.roles.create({ name: 'gozado', color: 0xFFFFFF });
      }
      for (const [_, membro] of galaRole.members) {
        await membro.roles.remove(galaRole);
      }
      await victim.roles.add(galaRole);
      ultimoGozado = victim.id;
      await message.channel.send(`gozei no ${victim.toString()}`);
    } catch (error) {
      console.error('Erro no comando !gozei:', error);
      message.reply('❌ Ocorreu um erro ao executar o comando !gozei.');
    }
  } 

  // --- Comando !limpagala ---
  if (message.content.startsWith('!limpagala')) {
    try {
      if (!message.member.permissions.has('ManageMessages')) {
        return message.reply('❌ Você não tem permissão para usar este comando.');
      }
      const guild = message.guild;
      const galaRole = guild.roles.cache.find(role => role.name === 'gozado');
      if (!galaRole) {
        return message.reply('O cargo "gozado" não existe.');
      }
      const membros = galaRole.members;
      if (membros.size === 0) {
        return message.reply('Não tem ninguém melado.');
      }
      for (const [_, membro] of membros) {
        await membro.roles.remove(galaRole);
      }
      await message.channel.send('Todo mundo limpinho.');
    } catch (error) {
      console.error('Erro no comando !limpagala:', error);
      message.reply('❌ Ocorreu um erro ao executar o comando !limpagala.');
    }
  }

  // --- Do not disturb momento ---
  if (message.content.startsWith('!consent')) {
    const user = message.author.id;
    DO_NOT_DISTURB = carregarDoNotDisturb();
    if (!DO_NOT_DISTURB.includes(user)) {
      DO_NOT_DISTURB.push(user);
      salvarDoNotDisturb(DO_NOT_DISTURB);
      await message.react("👎");
      await message.reply("Removido da diversão. BUUUXA");
    } else {
      await message.reply("Você já foi removido da diversão.");
    }
  }

  // --- Remove o user da do not disturb ---
  if (message.content.startsWith("!reinclude")){
    
    if (!permitidos.includes(message.author.id)) {
      await message.reply("Você não tem permissão pra usar esse comando. 🐀🐀🐀");
      return;
    }

    const args = message.content.trim().split(' ').slice(1);
    const user = args.join(' ').trim();
    
    const index = DO_NOT_DISTURB.indexOf(user);
    if (index !== -1) {
      DO_NOT_DISTURB.splice(index, 1);
      salvarDoNotDisturb(DO_NOT_DISTURB);
      await message.reply(`Usuário removido da lista DND ✅`);
    } else {
      await message.reply(`Usuário não estava na lista DND ❌`);
    }
}
// --- [Redacted] Counter ---
if (tracker.includes(PALAVRA_MONITORADA)) {
  const agora = Date.now();
  if (!wordCounter[PALAVRA_MONITORADA]) {
    wordCounter[PALAVRA_MONITORADA] = {
      last: agora,
      record: 0,
      recordDate: agora
    };
    salvarWordCounter(wordCounter);
    await message.channel.send(`Primeira vez que falamos "${PALAVRA_MONITORADA}".`);
    return;
  }
  const tempoSemFalar = agora - wordCounter[PALAVRA_MONITORADA].last;
  let recorde = wordCounter[PALAVRA_MONITORADA].record;
  let recordeData = wordCounter[PALAVRA_MONITORADA].recordDate;
  let novoRecorde = false;
  if (tempoSemFalar > recorde) {
    recorde = tempoSemFalar;
    recordeData = agora;
    wordCounter[PALAVRA_MONITORADA].record = recorde;
    wordCounter[PALAVRA_MONITORADA].recordDate = recordeData;
    novoRecorde = true;
  }
  wordCounter[PALAVRA_MONITORADA].last = agora;
  salvarWordCounter(wordCounter);

  function formatarTempo(ms) {
    const seg = Math.floor(ms / 1000) % 60;
    const min = Math.floor(ms / 1000 / 60) % 60;
    const horas = Math.floor(ms / 1000 / 60 / 60) % 24;
    const dias = Math.floor(ms / 1000 / 60 / 60 / 24);
    let partes = [];
    if (dias) partes.push(`${dias}d`);
    if (horas) partes.push(`${horas}h`);
    if (min) partes.push(`${min}m`);
    if (seg) partes.push(`${seg}s`);
    return partes.length ? partes.join(' ') : 'menos de 1s';
  }

  await message.channel.send(
    `Estamos há ${formatarTempo(tempoSemFalar)} sem falar "${PALAVRA_MONITORADA}".\n` +
    `Nosso recorde atual é de ${formatarTempo(recorde)}.` +
    (novoRecorde ? " 🏆 Novo recorde!" : "")
  );
}


  if (message.content.startsWith('!leite')) {
  await message.channel.send(`**LEITE
  ingredientes
  meu pau

  ferramentas
  sua mão

  instruções
  1. bate uma pra mim**`);
  




}
  // ---- conserta os embedding ----
  if(
  message.content.includes("https://x.com/") ||
  message.content.includes("https://twitter.com/") ||
  message.content.includes("https://instagram.com/") ||
  message.content.includes("https://www.instagram.com/")
  ) {
    let msg = message.content;
    let autor = message.author.username;
    if (message.author.bot) return;
    await message.delete();

    if (message.content.includes("https://x.com/")) {
      // troca x.com por vx.com
      msg = msg.replace(/https:\/\/x\.com\//g, "https://fixvx.com/");
    } else if (message.content.includes("https://twitter.com/")) {
      // troca twitter.com por vx.com
      msg = msg.replace(/https:\/\/twitter\.com\//g, "https://fixvx.com/");
    } else {
      // troca instagram.com por vxinstagram.com
      msg = msg.replace(/https:\/\/(www\.)?instagram\.com\//g, "https://www.vxinstagram.com/");
      // remove endpoints extras e parâmetros do shareID para /reel/ ou /p/
      msg = msg.replace(/(https:\/\/www\.vxinstagram\.com\/(reel|p)\/[^\/]+)\/?.*/, "$1/");
    }
  await message.channel.send(`Mensagem enviada por **${autor}** \n` + msg);
}

if (message.author.id == "270893747045269505") {
  let result = Math.random();
  if (result >= inf) {
    await message.channel.send(">Terres");
  }
}

if (message.content.startsWith("!inf")){
  if (!permitidos.includes(message.author.id)) await message.reply("você não tem permissão para usar esse comando.");
  const novo_min = message.content.replace('!inf', '').trim();
  inf = parseFloat(novo_min);
  await message.reply("probabilidade alterada.");
}

});

/*
// --- Cron para enviar imagem toda sexta-feira às 13h ---
cron.schedule('00 13 * * 5', async () => {
  if (imagensSexta.length === 0) return;
  const sorteada = imagensSexta[Math.floor(Math.random() * imagensSexta.length)];
  try {
    const canal = await client.channels.fetch('919309611885015140'); 
    if (canal) {
      await canal.send({ content: 'Imagem da sexta-feira:', files: [sorteada] });
    }
  } catch (e) {
    console.error('Erro ao enviar imagem da sexta:', e);
  }
});
*/
client.login(TOKEN);
