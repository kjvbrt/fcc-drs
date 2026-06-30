package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"

	"dataset-tracker/internal/middleware"
	"dataset-tracker/internal/models"
)

const maxGeneratorCardSize = 1 << 20 // 1 MB

// saveGeneratorCardFromForm reads "generator_card" from a parsed multipart
// form and persists it. Silently skips if no file was provided.
func (h *Handler) saveGeneratorCardFromForm(r *http.Request, requestID, userID int) {
	file, header, err := r.FormFile("generator_card")
	if err != nil {
		return
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(file, maxGeneratorCardSize+1))
	if err != nil || int64(len(content)) > maxGeneratorCardSize || bytes.IndexByte(content, 0) >= 0 {
		return
	}
	card := &models.GeneratorCard{
		RequestID:  requestID,
		Filename:   filepath.Base(header.Filename),
		Size:       int64(len(content)),
		Content:    content,
		UploadedBy: userID,
	}
	if _, err := h.generatorCards.Add(card); err != nil {
		slog.Error("save generator card from form", "error", err)
	}
}

func (h *Handler) UploadGeneratorCard(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	req, err := h.requests.GetByID(id)
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	user := middleware.GetUser(r)
	if !canEdit(user, req) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := r.ParseMultipartForm(maxGeneratorCardSize + 512); err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}
	file, header, err := r.FormFile("generator_card")
	if err != nil {
		http.Error(w, "no file provided", 400)
		return
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(file, maxGeneratorCardSize+1))
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	if int64(len(content)) > maxGeneratorCardSize {
		http.Error(w, "file too large (max 1 MB)", 400)
		return
	}
	if bytes.IndexByte(content, 0) >= 0 {
		http.Error(w, "only plain-text files are accepted", 400)
		return
	}

	userID := 0
	if user != nil {
		userID = user.ID
	}
	card := &models.GeneratorCard{
		RequestID:  id,
		Filename:   filepath.Base(header.Filename),
		Size:       int64(len(content)),
		Content:    content,
		UploadedBy: userID,
	}
	if _, err := h.generatorCards.Add(card); err != nil {
		slog.Error("upload generator card", "error", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	cards, _ := h.generatorCards.GetByRequestID(id)
	h.renderPartial(w, r, "generator_cards", PageData{Request: req, GeneratorCards: cards})
}

func (h *Handler) DownloadGeneratorCard(w http.ResponseWriter, r *http.Request) {
	cardID, err := strconv.Atoi(r.PathValue("card_id"))
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	card, err := h.generatorCards.GetByID(cardID)
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}

	ct := mime.TypeByExtension(filepath.Ext(card.Filename))
	if ct == "" {
		ct = "application/octet-stream"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, card.Filename))
	w.Header().Set("Content-Length", strconv.FormatInt(card.Size, 10))
	w.Write(card.Content) //nolint:errcheck
}

func (h *Handler) DeleteGeneratorCard(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	cardID, err := strconv.Atoi(r.PathValue("card_id"))
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	card, err := h.generatorCards.GetByID(cardID)
	if err != nil || card.RequestID != id {
		http.Error(w, "Not Found", 404)
		return
	}
	req, err := h.requests.GetByID(id)
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	user := middleware.GetUser(r)
	ownerCanDelete := user != nil && card.UploadedBy == user.ID && canEdit(user, req)
	if !user.IsManager() && !ownerCanDelete {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.generatorCards.Delete(cardID); err != nil {
		slog.Error("delete generator card", "error", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	cards, _ := h.generatorCards.GetByRequestID(id)
	h.renderPartial(w, r, "generator_cards", PageData{Request: req, GeneratorCards: cards})
}
