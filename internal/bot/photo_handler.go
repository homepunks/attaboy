package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/homepunks/attaboy/internal/config"
	"github.com/homepunks/attaboy/internal/qr"
)

func handlePhoto(upd Update, cfg config.Config) {
	chatID := upd.Message.Chat.ID

	if len(upd.Message.Photo) == 0 {
		reply(upd, cfg, "No photo found")
		return
	}

	photo := upd.Message.Photo[len(upd.Message.Photo)-1]

	photoBytes, err := downloadPhoto(photo.FileID, cfg)
	if err != nil {
		log.Printf("Could not download image: %v", err)
		reply(upd, cfg, "Failed to download image")
		return
	}

	link, err := qr.ScanQR(photoBytes)
	if errors.Is(err, qr.ErrNoQR) {
		reply(upd, cfg, "No QR found in the image")
		return
	}
	if err != nil {
		log.Printf("Error scanning QR: %v", err)
		reply(upd, cfg, "Could not read the image")
		return
	}

	if !isMoodleQR(link) {
		reply(upd, cfg, fmt.Sprintf("Unsupported QR detected: %s", link))
		return
	}

	handleMoodleQR(link, chatID, cfg)
}

func handleMoodleQR(link string, chatID int64, cfg config.Config) {
	log.Printf("Moodle QR received in chat %d", chatID)
	if err := sendMessage(chatID, cfg, "Moodle QR detected, but marking attendance is not implemented yet"); err != nil {
		log.Printf("Could not send message to chat %d: %v", chatID, err)
	}
}

func downloadPhoto(fileID string, cfg config.Config) ([]byte, error) {
	getFileURL := fmt.Sprintf("%s%s/getFile?file_id=%s",
		cfg.BaseURL, cfg.BotToken, fileID)

	resp, err := http.Get(getFileURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			FilePath string `json:"file_path"`
		} `json:"result"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if !result.OK {
		return nil, errors.New("Telegram API error: getFile failed")
	}

	downloadURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s",
		cfg.BotToken, result.Result.FilePath)

	fileResp, err := http.Get(downloadURL)
	if err != nil {
		return nil, err
	}
	defer fileResp.Body.Close()

	if fileResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Telegram file download failed: %s", fileResp.Status)
	}

	return io.ReadAll(fileResp.Body)
}

func isMoodleQR(link string) bool {
	return strings.Contains(link, "moodle.nu.edu.kz/login/index.php")
}
