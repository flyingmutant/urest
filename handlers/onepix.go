package handlers

import (
	"encoding/base64"
	"log"
	"net/http"
	"strconv"
)

const (
	onePxBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABAQMAAAAl21bKAAAAA1BMVEUAAACnej3aAAAAAXRSTlMAQObYZgAAAApJREFUCNdjYAAAAAIAAeIhvDMAAAAASUVORK5CYII="
)

var (
	onePxBytes []byte
)

func init() {
	var err error
	onePxBytes, err = base64.StdEncoding.DecodeString(onePxBase64)
	if err != nil {
		log.Fatalf("Failed to decode 1x1px transparent PNG: %v", err)
	}
}

func OnePxTransparentImageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(onePxBytes)))
	w.Write(onePxBytes)
}
