package bot

import (
	"log"
	"net/http"
	"fmt"
	"io"
	"strings"
	"encoding/json"

	"github.com/attaboy/internal/config"	
	"github.com/attaboy/internal/qr"
)

func handlePhoto(upd Update, cfg config.Config) {
	chatID := upd.Message.Chat.ID

	if len(upd.Message.Photo) == 0 {
		if err := sendMessage(chatID, cfg, "No photo found"); err != nil {
			log.Printf("Could not send message to %s (@%s)",
				upd.Message.Chat.Name, upd.Message.Chat.Username)
			return
		}
	}

	photo := upd.Message.Photo[len(upd.Message.Photo) - 1]
	
	photoBytes, err := downloadPhoto(photo.FileID, cfg)
	if err != nil {
		log.Printf("Could not download image: %v", err)
		if err := sendMessage(chatID, cfg, "Failed to download image"); err != nil {
			log.Printf("Could not send message to %s (@%s)",
				upd.Message.Chat.Name, upd.Message.Chat.Username)
		}		
		return
	}
}

func downloadImage(fileID string, cfg config.Config) ([]byte, error) {
	getFileURL := fmt.Sprintf("%s%s/getFile?file_id=%s",
		cfg.BaseURL, cfg.BotToken, fileID)

	resp, err := http.Get(getFileURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		OK bool `json:"ok"`
		Result struct {
			FilePath string `json:"file_path"`
		} `json:"result"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		continue
	}
	json.Unmarshal(body, &result)

	if !result.OK {
		return nil, errors.New("Telegram API error: getFile failed")
	}

	downloadURL := fmt.Sprintf("%s/file/bot%s/%s",
		cfg.BaseURL, cfg.BotToken, result.Result.FilePath)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
	
}

func isMoodleQR(link string) bool {
	return strings.Contains(link, "moodle.nu.edu.kz/login/index.php")
}
