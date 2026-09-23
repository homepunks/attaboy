package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/homepunks/attaboy/internal/config"
	"github.com/homepunks/attaboy/internal/moodle"
	"github.com/homepunks/attaboy/internal/qr"
)

func handlePhoto(upd Update, cfg config.Config) {
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

	client := moodle.Client{
		BaseURL:  cfg.MoodleURL,
		Username: cfg.MoodleUsername,
		Password: cfg.MoodlePassword,
	}

	if !client.IsMoodleLink(link) {
		reply(upd, cfg, fmt.Sprintf("Unsupported QR detected: %s", link))
		return
	}

	handleMoodleQR(upd, cfg, client, link)
}

func handleMoodleQR(upd Update, cfg config.Config, client moodle.Client, link string) {
	log.Printf("Marking attendance for %s (@%s)", upd.Message.Chat.Name, upd.Message.Chat.Username)

	result, err := client.MarkAttendance(link)
	if err != nil {
		log.Printf("Could not mark attendance: %v", err)
		reply(upd, cfg, fmt.Sprintf("Could not mark attendance: %v", err))
		return
	}

	if (result.OK || alreadyMarked(result.Message)) && firstCheckIn(link) {
		log.Printf("Attendance marked for %s (@%s)",
			upd.Message.Chat.Name, upd.Message.Chat.Username)
		reply(upd, cfg, "Attendance marked")
		return
	}

	log.Printf("Moodle replied (ok=%t): %s", result.OK, result.Message)
	reply(upd, cfg, "Moodle says: "+result.Message)
}

var checkedIn sync.Map

func firstCheckIn(link string) bool {
	_, seen := checkedIn.LoadOrStore(sessionKey(link), time.Now())
	return !seen
}

func sessionKey(link string) string {
	u, err := url.Parse(link)
	if err != nil {
		return link
	}
	if sessid := u.Query().Get("sessid"); sessid != "" {
		return sessid
	}
	return link
}

func alreadyMarked(message string) bool {
	return strings.Contains(strings.ToLower(message), "already")
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
