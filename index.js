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
  const guild = client.guilds.cache.get('918671270885851187'); // Tonga
  
  const emotesArray = [...guild.emojis.cache.map(e => e), '🫃'];
  if (emotesArray) console.log("emotes obtidos");
  global.emotesArray = emotesArray;
  
});


client.on("messageCreate", async (message) => {

  if (message.author.bot) return;
  
  // SE FOR DO COMUNICADOS, ELE NAO VAI MAIS RESPONDER.
  if (message.channel.parentId === '919309359916388372') {
    console.log(`Mensagem ignorada no canal ${message.channel.name} da categoria proibida.`);
    return;
  }

  if (global.emotesArray && Math.random() < 0.005) {
    const emote = global.emotesArray[Math.floor(Math.random() * global.emotesArray.length)];
    try {
      await message.react(emote.id ? emote.id : emote); 
    } catch (e) {
      console.log('Falha ao reagir com emote:', e);
    }
  }

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

  // Função utilitária para remover acentos
  function removerAcentos(str) {
    return str.normalize("NFD").replace(/[\u0300-\u036f]/g, "");
  }

  const conteudo = removerAcentos(message.content.toLowerCase());

  // Verifica se todas as palavras-chave estão presentes
  if (
    conteudo.includes("acaba") &&
    conteudo.includes("midiacast")
  ) {
    message.reply("09/08/2025 às 23:59");
  }

  if (conteudo.includes("qual o minimo")) {
    message.react('🫃');
  }

  if (conteudo.includes("is this true")) {
    if(Math.random() < 0.45){
      message.reply('https://tenor.com/view/morgan-freeman-true-morgan-freeman-true-nodding-gif-13973817878387504960');
    }else if(Math.random() < 0.95){
      message.reply('https://tenor.com/view/its-peak-its-mid-fight-morgan-freeman-gif-6564041502742593422');
    }else{
      message.reply('https://cdn.discordapp.com/attachments/1362454934997696642/1374740964790243399/images373.jpg?ex=682f26cb&is=682dd54b&hm=b6230e85ddd3e2ce9eb9c2bfd8dbab0d3936cac158462cac60f06a9f7fe149ca&');
    }
    return;
  }
  

  //comando !listfrases
  if (message.content.startsWith('!listfrases')) {
    if (respostas.length === 0) {
      return message.reply('⚠️ Nenhuma frase cadastrada ainda.');
    }
  
    const args = message.content.trim().split(' ').slice(1);
    const termo = args.join(' ').trim();
  
    // 🔍 Buscar frase exata ou trecho de frase
    if (termo && !termo.match(/^\d+$/) && !termo.includes('#') && !termo.startsWith('<@')) {
      const encontrada = respostas.find(f => f.texto.toLowerCase().includes(termo.toLowerCase()));
      if (encontrada) {
        return message.reply(`🧾 Frase encontrada:\n"${encontrada.texto}"\n👤 Autor: ${encontrada.autor}`);
      } else {
        return message.reply(`❌ Nenhuma frase contendo "${termo}" foi encontrada.`);
      }
    }
  
    // 👤 Buscar por autor
    if (termo.includes('#') || termo.startsWith('<@')) {
      // Se for menção, extrai o username do objeto de usuário
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
  
    // 🔟 Listar últimas N frases (sem links)
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

  let ultimoGozado = null;

  // Função para embaralhar um array (Fisher-Yates)
  function shuffle(array) {
    for (let i = array.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [array[i], array[j]] = [array[j], array[i]];
    }
    return array;
  }

  if (message.content.startsWith('!gozei')) {
    try {
      if (!message.member.permissions.has('ManageMessages')) {
        return message.reply('❌ Você não tem permissão para usar este comando.');
      }

      const guild = message.guild;
      let galaRole = guild.roles.cache.find(role => role.name === 'gozado');

      let members = guild.members.cache.filter(m => !m.user.bot);
      if (members.size === 0) return message.reply('Não há membros humanos no servidor!');

      // Remove o último sorteado da lista, se possível
      let pool = members;
      if (ultimoGozado && members.has(ultimoGozado)) {
        pool = members.filter(m => m.id !== ultimoGozado);
        if (pool.size === 0) pool = members;
      }

      // Embaralha o pool e pega o primeiro
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

  if (message.content.startsWith('!limpagala')) {
    try {
      // Só permite se o autor tiver permissão de gerenciar mensagens
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
});

client.login(TOKEN);