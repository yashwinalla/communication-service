package form

import "html/template"

type Email struct {
	Name    string `json:"name"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Type    string `json:"type"`
	Url     string `json:"url"`
	Message string
	Data    map[string]string `json:"data"`
	List    interface{}       `json:"list"`
	BaseUrl string
}

type EmailTemplate struct {
	Name     string
	Path     string
	Title    string
	Template *template.Template
}

type EmailConfig struct {
	Provider  string
	FromName  string
	From      string
	Sandbox   bool
	Key       string
	ServerUrl string
}
