package main

import (
	"html/template"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/oxtoacart/bpool"
)

var (
	templates map[string]*template.Template
	bufpool   *bpool.BufferPool

	templateFuncs = template.FuncMap{
		"humanizeTime": func(t time.Time) string {
			return humanize.Time(t)
		},
		"humanizeTimeInt": func(t int64) string {
			return humanize.Time(time.Unix(t, 0))
		},
		"unixToString": func(t int64) string {
			return time.Unix(t, 0).String()
		},
	}
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
			t := template.New(name).Funcs(templateFuncs)
			template.Must(t.Parse(string(data)))
			template.Must(t.New("base").Parse(base))

			templates[name] = t
		}
	}

	bufpool = bpool.NewBufferPool(20)
}
