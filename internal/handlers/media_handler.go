package handlers

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/minio/minio-go/v7"
	"github.com/noteduco342/OMMessenger-backend/internal/httpx"
	"github.com/noteduco342/OMMessenger-backend/internal/storage"
)

type MediaHandler struct {
	s3 *storage.S3Storage
}

func NewMediaHandler(s3 *storage.S3Storage) *MediaHandler {
	return &MediaHandler{s3: s3}
}

func normalizeETag(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "W/")
	v = strings.Trim(v, "\"")
	return v
}

func (h *MediaHandler) GetAvatar(c *fiber.Ctx) error {
	if h.s3 == nil {
		return httpx.Error(c, fiber.StatusServiceUnavailable, "storage_not_configured", "Storage not configured")
	}

	keyParam := strings.TrimSpace(c.Params("*"))
	key, err := storage.SafeJoinAvatarPath("", keyParam)
	if err != nil {
		return httpx.Error(c, fiber.StatusNotFound, "not_found", "Not found")
	}

	obj, st, err := h.s3.GetObject(c.Context(), key)
	if err != nil {
		// Hide details.
		var resp minio.ErrorResponse
		if errors.As(err, &resp) {
			if resp.StatusCode == 404 || resp.Code == "NoSuchKey" || resp.Code == "NoSuchObject" {
				return httpx.Error(c, fiber.StatusNotFound, "not_found", "Not found")
			}
		}
		return httpx.Internal(c, "media_fetch_failed")
	}
	defer obj.Close()

	etag := st.ETag
	if etag != "" {
		c.Set("ETag", "\""+etag+"\"")
		if inm := normalizeETag(c.Get("If-None-Match")); inm != "" && inm == normalizeETag(etag) {
			return c.SendStatus(fiber.StatusNotModified)
		}
	}
	if !st.LastModified.IsZero() {
		c.Set("Last-Modified", st.LastModified.UTC().Format(time.RFC1123))
	}

	c.Set("Cache-Control", "private, max-age=31536000, immutable")
	if st.ContentType != "" {
		c.Type(st.ContentType)
	} else {
		c.Type("image/jpeg")
	}

	// Stream object.
	return c.SendStream(obj, int(st.Size))
}
