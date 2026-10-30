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
	arq_convertido := strings.TrimSuffix(path, filepath.Ext(path)) + ".wav"
	defer os.Remove(arq_convertido)

	cmd := exec.Command("ffmpeg", "-y", "-i", path, "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le", arq_convertido)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("erro ao converter audio para Wav: %v - %s", err, string(output))
	}

	whisper := "./whisper-cli"
	model := "models/ggml-base.bin"
	cmdWhisper := exec.Command(
		whisper,
		"-m", model,
		"-f", arq_convertido,
		"-l", "pt", // pt-br
		"-nt", // (apenas o texto limpo)
	)
	out, err := cmdWhisper.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("erro no whisper: %v - %s", err, string(out))
	}
	textoTranscrito := strings.TrimSpace(string(out))
	if textoTranscrito == "" {
		return "Não foi possível compreender o áudio.", nil
	}
	if err_rm := os.Remove(path); err_rm != nil {
		return textoTranscrito, fmt.Errorf("erro ao remover arquivo baixado: %s", err_rm)
	}
	return textoTranscrito, nil
}
