package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Baixar_audio(url string, caminhoArquivo string) error {
	// Faz a requisição para pegar o áudio
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status HTTP %d", resp.StatusCode)
	}

	// Cria o arquivo no caminho específico (ex: "audios/ID_audio.ogg")
	out, err := os.Create(caminhoArquivo)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copia os bytes para o arquivo
	_, erro := io.Copy(out, resp.Body)
	return erro
}

func Transcrever(path string) (string, error) {
	
	arqConvertido := strings.TrimSuffix(path, filepath.Ext(path)) + ".wav"
	defer os.Remove(arqConvertido) // Remove o .wav ao finalizar

	//  Converte OGG Opus do Discord para WAV PCM 16kHz mono
	cmdFFmpeg := exec.Command("ffmpeg", "-y", "-i", path, "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le", arqConvertido)
	if output, err := cmdFFmpeg.CombinedOutput(); err != nil {
		return "", fmt.Errorf("erro no ffmpeg: %v - %s", err, string(output))
	}

	
	info, err := os.Stat(arqConvertido)
	if err != nil || info.Size() == 0 {
		return "", fmt.Errorf("falha ao converter áudio: arquivo WAV de tamanho 0")
	}

	whisper := "./whisper-cli"
	model := "models/ggml-small.bin"

	//  whisper-cli
	cmdWhisper := exec.Command(
		whisper,
		"-m", model,
		"-f", arqConvertido,
		"-l", "pt",
		"-nt",
		"-t", "2", 
	)

	out, err := cmdWhisper.Output()
	if err != nil {
		return "", fmt.Errorf("erro no whisper: %v - %s", err, string(out))
	}

	textoTranscrito := strings.TrimSpace(string(out))
	if textoTranscrito == "" {
		textoTranscrito = "Não foi possível compreender o áudio."
	}

	// Apaga o arquivo .ogg original
	if errRm := os.Remove(path); errRm != nil {
		fmt.Printf("Aviso: erro ao remover %s: %v\n", path, errRm)
	}

	return textoTranscrito, nil
}
