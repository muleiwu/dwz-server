package controller

import (
	"bytes"
	"html/template"
	"strings"
	"testing"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/service"
)

func TestInstallPageDatabaseOptions(t *testing.T) {
	tpl, err := template.ParseGlob("../../templates/*.html")
	if err != nil {
		t.Fatal(err)
	}
	for _, driver := range []string{"mysql", "mariadb", "postgresql", "sqlite"} {
		t.Run(driver, func(t *testing.T) {
			var buf bytes.Buffer
			err := tpl.ExecuteTemplate(&buf, "install.html", installPageData{
				SiteName: "DWZ",
				DatabaseConfig: service.InstallDatabaseConfig{
					Driver: driver, Host: "localhost", Port: 3307,
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			page := buf.String()
			if !strings.Contains(page, `<option value="`+driver+`" selected>`) {
				t.Fatalf("database %q was not selected in the rendered page", driver)
			}
			if !strings.Contains(page, `<option value="mariadb"`) || !strings.Contains(page, `>MariaDB</option>`) {
				t.Fatal("MariaDB option is missing")
			}
			if !strings.Contains(page, `id="dbPort" name="dbPort" value="3307"`) {
				t.Fatal("configured database port was not preserved")
			}
		})
	}
}

func TestInstallPageDataSupportsHeaderLogoURL(t *testing.T) {
	tpl := template.Must(template.New("header").Parse(`
{{define "icon_logo"}}default-logo{{end}}
{{define "header"}}
{{if .LogoURL}}<img src="{{.LogoURL}}" alt="">{{else}}{{template "icon_logo" .}}{{end}}
{{end}}
`))

	var buf bytes.Buffer
	err := tpl.ExecuteTemplate(&buf, "header", installPageData{
		SiteName: "DWZ",
		LogoURL:  "/uploads/branding/logo.png",
	})
	if err != nil {
		t.Fatalf("execute header template: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, `/uploads/branding/logo.png`) {
		t.Fatalf("expected rendered logo URL, got %q", got)
	}
}
