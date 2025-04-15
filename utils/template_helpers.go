package utils

import (
	"text/template"
	"time"
)

// TemplateFuncs tüm template fonksiyonlarını döner
func TemplateHelpers() template.FuncMap {
	return template.FuncMap{
		"CurrentYear": func() int {
			return time.Now().Year()
		},
		// başka fonksiyonlar da buraya eklenebilir
	}
}
