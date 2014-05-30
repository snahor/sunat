package main

import (
	"log"
	"os/exec"
)

func isTesseractInstalled() bool {
	output, _ := exec.Command("which", "tesseract").Output()
	return len(output) > 0
}

func captchaToText(imagePath string) string {
	output, err := exec.Command(
		"tesseract",
		imagePath,
		"stdout",
		"-psm", "7",
		"-c", "tessedit_char_whitelist=ABCDEFGHIJKLMNOPQRSTUVWXYZ",
	).Output()
	if err != nil {
		log.Print(err)
		return ""
	}
	return string(output[0:4])
}
