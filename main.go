package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func RenderIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := renderTemplate(w, "index", nil)
	if err != nil {
		log.Printf("Error rendering template: %s", err)
		http.Error(w, "error rendering template", 500)
	}
}

func RenderForm(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm()

	token := r.Form.Get("token")
	if len(token) < 1 {
		http.Error(w, "no token found", 403)
		return
	}

	err := renderTemplate(w, "form", map[string]interface{}{
		"token": token,
	})
	if err != nil {
		log.Printf("Error rendering template: %s", err)
		http.Error(w, "error rendering template", 500)
	}
}

func HandleSubmit(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm()

	token := r.Form.Get("token")
	if len(token) < 1 {
		http.Error(w, "no token found", 403)
		return
	}

	err := renderTemplate(w, "success", nil)
	if err != nil {
		log.Printf("Error rendering template: %s", err)
		http.Error(w, "error rendering template", 500)
	}
}

func main() {
	log.Println("Started")

	router := httprouter.New()
	router.GET("/", RenderIndex)
	router.GET("/form", RenderForm)
	router.POST("/submit", HandleSubmit)

	for _, asset := range AssetDescriptors() {
		if !strings.HasSuffix(asset.Path, ".tmpl") {
			log.Printf("Serving asset: %s", asset.Path)
			router.GET("/"+asset.Path, ServeAsset(asset.Path, asset.Mime))
		}
	}

	log.Println("Serving HTTP on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
