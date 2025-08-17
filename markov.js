class MarkovChain {
    constructor() {
        this.chain = {};
    }

    addMessage(msg) {
        const words = msg.split(/\s+/);
        try {
            for (let i = 0; i < words.length; i++) {
                const word = words[i].toLowerCase();
                const next_word = words[i+1] ? words[i + 1].toLowerCase() : null;;
                if (!this.chain[word]) this.chain[word] = [];
                this.chain[word].push(next_word);
                
            }
            
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

}
module.exports = MarkovChain;