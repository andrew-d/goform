package main

import (
	"html/template"
	"path/filepath"

	"github.com/oxtoacart/bpool"
)

var (
	templates map[string]*template.Template
	bufpool   *bpool.BufferPool
)

func init() {
	templates = make(map[string]*template.Template)

	// Read the base template.
	baseb, err := Asset("base.tmpl")
	if err != nil {
		panic(err)
	}
	base := string(baseb)

	// Parse all other templates.
	for _, asset := range AssetNames() {
		ext := filepath.Ext(asset)

		if ext == ".tmpl" && asset != "base.tmpl" {
			name := asset[0 : len(asset)-5]
			data, _ := Asset(asset)

			// Mimic the ParseFiles function manually here
			t := template.New(name)
			template.Must(t.Parse(string(data)))
			template.Must(t.New("base").Parse(base))

			templates[name] = t
		}
	}

	bufpool = bpool.NewBufferPool(20)
}
