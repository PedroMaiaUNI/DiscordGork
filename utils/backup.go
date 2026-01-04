package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	BackupDir   = "backups"
	Max_Backups = 3
)

func BackupFrases(frases []Frase) error {
	if err := os.MkdirAll(BackupDir, 0755); err != nil {
		return err
	}

	filename := fmt.Sprintf(
		"frases_%d.json",
		time.Now().Unix(),
	)

	path := filepath.Join(BackupDir, filename)

	data, err := json.MarshalIndent(frases, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	return cleanupOldBackups()
}

func LoadLastBackup() ([]Frase, error) {
	files, err := filepath.Glob(filepath.Join(BackupDir, "frases_*.json"))
	if err != nil || len(files) == 0 {
		return nil, fmt.Errorf("nenhum backup encontrado")
	}

	sort.Slice(files, func(i, j int) bool {
		fi, _ := os.Stat(files[i])
		fj, _ := os.Stat(files[j])
		return fi.ModTime().After(fj.ModTime())
	})

	data, err := os.ReadFile(files[0])
	if err != nil {
		return nil, err
	}

	var frases []Frase
	if err := json.Unmarshal(data, &frases); err != nil {
		return nil, err
	}

	return frases, nil
}

func cleanupOldBackups() error {
	files, err := filepath.Glob(filepath.Join(BackupDir, "frases_*.json"))
	if err != nil {
		return err
	}

	if len(files) <= Max_Backups {
		return nil
	}

	sort.Slice(files, func(i, j int) bool {
		fi, _ := os.Stat(files[i])
		fj, _ := os.Stat(files[j])
		return fi.ModTime().Before(fj.ModTime())
	})

	// remove os mais antigos
	for _, f := range files[:len(files)-Max_Backups] {
		_ = os.Remove(f)
	}

	return nil
}
