package main

import (
	"github.com/samber/lo"
	"github.com/samber/mo"

	"github.com/hivemindd/communication-service/internal/form"
)

var templates = []form.EmailTemplate{
	{Name: "set_password", Path: "templates/set_password.html", Title: "Verify Your Email", Template: nil},
	{Name: "forgot_password", Path: "templates/forgot_password.html", Title: "Hivemindd Password Reset", Template: nil},
	{Name: "account_lock", Path: "templates/account_lock.html", Title: "Hivemindd Account Locked", Template: nil},
}

func findTemplateByType(name string) mo.Option[form.EmailTemplate] {
	t, ok := lo.Find(templates, func(t form.EmailTemplate) bool {
		return t.Name == name
	})
	if ok {
		return mo.Some(t)
	}

	return mo.None[form.EmailTemplate]()
}
