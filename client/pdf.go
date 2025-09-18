package main

import (
	"embed"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

var pdfContent embed.FS

func OpenEmbeddedPDF() {
	data, err := pdfContent.ReadFile("letter.pdf")
	if err != nil {
		return
	}
	tmpFile, err := os.CreateTemp("", "*.pdf")
	if err != nil {
		return
	}
	defer func() {
		time.Sleep(60 * time.Second)
		os.Remove(tmpFile.Name())
	}()
	if _, err := tmpFile.Write(data); err != nil {
		return
	}

	if err := tmpFile.Close(); err != nil {
		return
	}

	tmpFilePath, err := filepath.Abs(tmpFile.Name())
	if err != nil {
		return
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", tmpFilePath)
	case "darwin":
		cmd = exec.Command("open", tmpFilePath)
	case "linux":
		cmd = exec.Command("xdg-open", tmpFilePath)
	default:
		log.Println("不支持的操作系统:", runtime.GOOS)
		return
	}

	if err := cmd.Run(); err != nil {
		return
	}
}
