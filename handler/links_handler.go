package handler

import (
	"net/http"
	"time"

	"github.com/sealsurlaw/gouvre/errs"
	"github.com/sealsurlaw/gouvre/helper"
	"github.com/sealsurlaw/gouvre/request"
	"github.com/sealsurlaw/gouvre/response"
)

func (h *Handler) Links(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		h.createLink(w, r)
		return
	} else if r.Method == "GET" {
		h.getImageFromToken(w, r)
		return
	} else {
		response.SendMethodNotFound(w)
		return
	}
}

func (h *Handler) createLink(w http.ResponseWriter, r *http.Request) {
	if !h.hasWhitelistedToken(r) {
		response.SendInvalidAuthToken(w)
		return
	}

	if !h.hasWhitelistedIpAddress(r) {
		response.SendError(w, 401, "Not on ip whitelist.", errs.ErrNotAuthorized)
		return
	}

	// filename
	req := request.LinkRequest{}
	err := request.ParseJson(r, &req)
	if err != nil {
		response.SendError(w, 400, "Could not parse json request.", err)
		return
	}

	// optional queries
	expiresAt := request.ParseExpires(r)

	filename, err := h.checkFileExists(req.Filename, req.EncryptionSecret)
	if err != nil {
		response.SendCouldntFindImage(w, err)
		return
	}

	token, err := h.tokenizer.CreateToken(filename, expiresAt, req.EncryptionSecret)
	if err != nil {
		response.SendError(w, 500, "Couldn't create token.", err)
		return
	}

	response.SendJson(w, &response.GetLinkResponse{
		Url:       h.makeTokenUrl(token),
		ExpiresAt: expiresAt,
	}, 200)
}

func (h *Handler) getImageFromToken(w http.ResponseWriter, r *http.Request) {
	// token
	token, err := request.ParseTokenFromUrl(r)
	if err != nil {
		response.SendBadRequest(w, "token")
	}

	filename, expiresAt, encryptionSecret, err := h.tokenizer.ParseToken(token)
	if err != nil {
		response.SendError(w, 500, "Couldn't create token.", err)
		return
	}

	if time.Now().After(*expiresAt) {
		response.SendError(w, 400, "Bad token.", errs.ErrTokenExpired)
		return
	}

	// open file
	fullFilePath := h.makeFullFilePath(filename)
	fileData, err := helper.OpenFile(fullFilePath)
	if err != nil {
		response.SendError(w, 500, "Couldn't read file data.", err)
		return
	}

	if h.tryDecryptFile(&fileData, encryptionSecret) != nil {
		response.SendError(w, 500, "Couldn't decrypt file.", err)
		return
	}

	response.SendImage(w, fileData, expiresAt)
}