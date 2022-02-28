package handler

import (
	"net/http"

	"github.com/sealsurlaw/ImageServer/errs"
	"github.com/sealsurlaw/ImageServer/request"
	"github.com/sealsurlaw/ImageServer/response"
)

func (h *Handler) ThumbnailsBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		h.createThumbnailLinks(w, r)
		return
	} else {
		response.SendMethodNotFound(w)
		return
	}
}

func (h *Handler) createThumbnailLinks(w http.ResponseWriter, r *http.Request) {
	if !h.hasWhitelistedToken(r) {
		response.SendInvalidAuthToken(w)
		return
	}

	if !h.hasWhitelistedIpAddress(r) {
		response.SendError(w, 401, "Not on ip whitelist.", errs.ErrNotAuthorized)
		return
	}

	// request json
	req := request.ThumbnailsRequest{}
	err := request.ParseJson(r, &req)
	if err != nil {
		response.SendError(w, 400, "Could not parse json request.", err)
		return
	}

	// optional queries
	square := request.ParseSquare(r)
	expiresAt := request.ParseExpires(r)

	filenameToUrls := make(map[string]string)
	for _, filename := range req.Filenames {
		thumbnailParameters := &ThumbnailParameters{filename, req.Resolution, square}
		fullFilename, err := h.checkOrCreateThumbnailFile(thumbnailParameters)
		if err != nil {
			continue
		}

		// create and add link to link store
		token, err := h.tryToAddLink(fullFilename, expiresAt)
		if err != nil {
			response.SendError(w, 500, err.Error(), err)
			return
		}

		url := h.makeTokenUrl(token)
		filenameToUrls[filename] = url
	}

	response.SendJson(w, &response.GetThumbnailLinksResponse{
		ExpiresAt:     expiresAt,
		FilenameToUrl: filenameToUrls,
	}, 200)
}