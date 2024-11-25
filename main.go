package main

import (
	"bytes"
	"github.com/google/uuid"
	"html/template"
	"log"
	"net/http"
	"time"
)

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	getTemplate := func(id string) []byte {
		tmpl, err := template.ParseFiles("templates/index.html")
		if err != nil {
			log.Fatalf("template parsing: %s", err)
		}

		var renderedTemplate bytes.Buffer
		err = tmpl.Execute(&renderedTemplate, map[string]interface{}{
			"ID": id,
		})
		if err != nil {
			log.Fatalf("template parsing: %s", err)
		}

		return renderedTemplate.Bytes()
	}

	id, _ := uuid.NewUUID()
	http.ServeContent(w, r, "templates/index.html", time.Now(), bytes.NewReader(getTemplate(id.String())))
}

func main() {
	hub := NewHub()
	go hub.Run()

	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	log.Fatal(http.ListenAndServe(":3000", nil))
}
