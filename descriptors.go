package main

type AssetDescriptor struct {
	Path string
	Mime string
}

func AssetDescriptors() []AssetDescriptor {
	return []AssetDescriptor{
		{"form.tmpl", "None"},
		{"success.tmpl", "None"},
		{"index.tmpl", "None"},
		{"base.tmpl", "None"},
		{"css/page.css", "text/css"},
		{"css/bootstrap.min.css", "text/css"},
		{"fonts/glyphicons-halflings-regular.ttf", "application/x-font-ttf"},
		{"fonts/glyphicons-halflings-regular.woff", "application/x-font-woff"},
		{"fonts/glyphicons-halflings-regular.svg", "image/svg+xml"},
		{"fonts/glyphicons-halflings-regular.eot", "application/vnd.ms-fontobject"},
	}
}