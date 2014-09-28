package main

import (
	"crypto/rand"
	"fmt"
	"net/http"
)

type M map[string]interface{}

// renderTemplate is a wrapper around template.ExecuteTemplate.
// It writes into a bytes.Buffer before writing to the http.ResponseWriter to catch
// any errors resulting from populating the template.
func renderTemplate(w http.ResponseWriter, name string, data M) error {
	// Ensure the template exists in the map.
	tmpl, ok := templates[name]
	if !ok {
		return fmt.Errorf("The template %s does not exist", name)
	}

	// Create a buffer to temporarily write to and check if any errors were encounted.
	buf := bufpool.Get()
	defer bufpool.Put(buf)

	err := tmpl.ExecuteTemplate(buf, "base", data)
	if err != nil {
		return err
	}

	// Set the header and write the buffer to the http.ResponseWriter
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
	return nil
}

func renderError(w http.ResponseWriter, message string, code int) {
	// Error data
	data := M{
		"message": message,
		"code":    code,
	}

	// Create a buffer to temporarily write to and check if any errors were encounted.
	buf := bufpool.Get()
	defer bufpool.Put(buf)

	err := templates["error"].ExecuteTemplate(buf, "base", data)
	if err != nil {
		http.Error(w, "error rendering error template o_0", 500)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	buf.WriteTo(w)
}

func ServeAsset(name, mime string) http.Handler {
	// Assert that the asset exists.
	_, err := Asset(name)
	if err != nil {
		panic(fmt.Sprintf("asset named '%s' does not exist", name))
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		asset, _ := Asset(name)
		w.Header().Set("Content-Type", mime)
		w.Write(asset)
	}

	return http.HandlerFunc(handler)
}

// Note: slightly biased towards first 8 characters of the alphabet, since 255
// isn't a multiple of 62 (length of alphanum).  We don't really care that
// much, though.
func randString(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}
