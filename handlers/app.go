package handlers

import (
	"html/template"
	"net/http"
	"path"
	"strings"
	"time"
)

type (
	appHandler struct {
		templatePath     string
		tpl              *template.Template
		templateDataFunc func(*http.Request) interface{}
	}
)

func NewAppHandler(templatePath string, reloadTemplate bool, templateDataFunc func(*http.Request) interface{}) http.Handler {
	var t *template.Template
	if !reloadTemplate {
		t = parseTemplate(templatePath)
	}

	return NewGzipHandler(false, &appHandler{
		templatePath:     templatePath,
		tpl:              t,
		templateDataFunc: templateDataFunc,
	})
}

func (h *appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t := h.tpl
	if t == nil {
		t = parseTemplate(h.templatePath)
	}

	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Date", time.Now().Format(http.TimeFormat))
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")

	if isFromMSIE(r) {
		w.Header().Set("X-UA-Compatible", "IE=edge")
	}

	t.Execute(w, h.templateDataFunc(r))
}

func parseTemplate(templatePath string) *template.Template {
	t := template.Must(template.New("").Delims("{{{", "}}}").ParseFiles(templatePath))
	return t.Lookup(path.Base(templatePath))
}

func isFromMSIE(r *http.Request) bool {
	return strings.Index(r.Header.Get("User-Agent"), "MSIE") != -1
}
