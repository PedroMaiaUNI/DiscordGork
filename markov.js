const fs = require('fs');

const CHAIN_PATH = 'markov_chain.json';

class MarkovChain {
    constructor() {
        this.chain = {};
        this._loadChain();
    }

    addMessage(msg) {
        const words = msg.split(/\s+/);
        try {
            for (let i = 0; i < words.length - 1; i++) {
                const word = words[i].toLowerCase();
                const next_word = words[i + 1] ? words[i + 1].toLowerCase() : null;
                if (!this.chain[word]) this.chain[word] = [];
                if (next_word) this.chain[word].push(next_word);
            }
            this._saveChain();
        } catch (e) {
            console.error("erro na leitura das mensagens", e);
        }
    }

    generate(startWord = null, maxLength = 30) {
        let word = startWord || this._randomWord();
        let result = [word];
        for (let i = 0; i < maxLength - 1; i++) {
            const nextWords = this.chain[word];
            if (!nextWords || nextWords.length === 0) break;
            word = nextWords[Math.floor(Math.random() * nextWords.length)];
            result.push(word);
        }
        return result.join(' ');
    }

    _randomWord() {
        const keys = Object.keys(this.chain);
        return keys[Math.floor(Math.random() * keys.length)];
    }

    _saveChain() {
        try {
            fs.writeFileSync(CHAIN_PATH, JSON.stringify(this.chain, null, 2), 'utf8');
        } catch (e) {
            console.error('Erro ao salvar a cadeia Markov:', e);
        }
    }

    _loadChain() {
        try {
            if (fs.existsSync(CHAIN_PATH)) {
                this.chain = JSON.parse(fs.readFileSync(CHAIN_PATH, 'utf8'));
            }
        } catch (e) {
            console.error('Erro ao carregar a cadeia Markov:', e);
            this.chain = {};
        }
    }
}

module.exports = MarkovChain;