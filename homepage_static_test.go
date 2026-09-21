package main

import (
	"bytes"
	"embed"
	"io"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	appConfig "cnb.cool/mliev/dwz/dwz-server/v2/config"
	"cnb.cool/mliev/open/go-web/pkg/container"
	"cnb.cool/mliev/open/go-web/pkg/interfaces"
	configImpl "cnb.cool/mliev/open/go-web/pkg/server/config/impl"
	httpImpl "cnb.cool/mliev/open/go-web/pkg/server/http_server/impl"
	httpInterfaces "cnb.cool/mliev/open/go-web/pkg/server/http_server/interfaces"
	"github.com/muleiwu/gsr"
)

type homepageTestAssembly struct {
	typ   reflect.Type
	value any
}

func (a homepageTestAssembly) Type() reflect.Type      { return a.typ }
func (homepageTestAssembly) DependsOn() []reflect.Type { return nil }
func (a homepageTestAssembly) Assembly() (any, error)  { return a.value, nil }

type homepageTestLogger struct{}

func (homepageTestLogger) Debug(string, ...gsr.LoggerField)  {}
func (homepageTestLogger) Info(string, ...gsr.LoggerField)   {}
func (homepageTestLogger) Notice(string, ...gsr.LoggerField) {}
func (homepageTestLogger) Warn(string, ...gsr.LoggerField)   {}
func (homepageTestLogger) Error(string, ...gsr.LoggerField)  {}
func (homepageTestLogger) Fatal(string, ...gsr.LoggerField)  {}

// Exercise the real application config assembly and go-web HTTP server, not a
// standalone file server: embedding an asset does not automatically expose it.
func TestHomepageStaticResources(t *testing.T) {
	cases := []struct {
		name     string
		dirs     []string
		disabled bool
	}{
		{name: "default"},
		{name: "existing_admin_config", dirs: []string{"admin"}},
		{name: "explicit_assets", dirs: []string{"admin", "assets"}},
		{name: "nested_home_mount", dirs: []string{"admin", "assets/home"}},
		{name: "static_disabled", dirs: []string{"admin"}, disabled: true},
	}
	for _, mode := range []string{"embed", "disk"} {
		for _, tc := range cases {
			caseName := mode + "/" + tc.name
			if selected := os.Getenv("DWZ_TEST_HOME_STATIC_CASE"); selected != "" && selected != caseName {
				continue
			}
			t.Run(caseName, func(t *testing.T) {
				// go-web retains providers on ResetAll; isolate each mode/config
				// in its own process so a previous case cannot mask a failure.
				if os.Getenv("DWZ_TEST_HOME_STATIC_CASE") == "" {
					cmd := exec.Command(os.Args[0], "-test.run=^TestHomepageStaticResources$")
					cmd.Env = append(os.Environ(), "DWZ_TEST_HOME_STATIC_CASE="+caseName)
					if output, err := cmd.CombinedOutput(); err != nil {
						t.Fatalf("%v\n%s", err, output)
					}
					return
				}
				t.Chdir(t.TempDir())
				// Disk mode uses the same release assets as embed mode, copied
				// to an isolated static tree without any real DB/config files.
				if mode == "disk" {
					err := fs.WalkDir(staticFS, "static/assets/home", func(path string, entry fs.DirEntry, err error) error {
						if err != nil {
							return err
						}
						if entry.IsDir() {
							return os.MkdirAll(path, 0700)
						}
						data, err := staticFS.ReadFile(path)
						if err != nil {
							return err
						}
						return os.WriteFile(path, data, 0600)
					})
					if err != nil {
						t.Fatal(err)
					}
				}
				container.ResetAll()
				t.Cleanup(container.ResetAll)
				env := configImpl.NewConfig()
				env.Set("jwt.secret", "homepage-static-test-only")
				env.Set("http.mode", "release")
				env.Set("http.static_mode", mode)
				env.Set("http.load_static", !tc.disabled)
				if tc.dirs != nil {
					env.Set("http.static_dir", tc.dirs)
				}
				assemblies := []interfaces.AssemblyInterface{
					homepageTestAssembly{reflect.TypeFor[gsr.Enver](), env},
					homepageTestAssembly{reflect.TypeFor[gsr.Logger](), homepageTestLogger{}},
				}
				for _, assembly := range (appConfig.App{}).Assemblies() {
					if assembly.Type() == reflect.TypeFor[gsr.Provider]() {
						assemblies = append(assemblies, assembly)
					}
				}
				if err := container.RegisterAssemblies(assemblies); err != nil {
					t.Fatal(err)
				}
				cfg := container.MustGet[gsr.Provider]()
				if cfg.GetString("http.static_mode", "") != mode || cfg.GetBool("http.load_static", false) == tc.disabled {
					t.Fatal("test configuration was not applied")
				}
				// Reserve a local ephemeral port; RunHttp owns its listener.
				listener, err := net.Listen("tcp", "127.0.0.1:0")
				if err != nil {
					t.Fatal(err)
				}
				addr := listener.Addr().String()
				if err := listener.Close(); err != nil {
					t.Fatal(err)
				}
				cfg.Set("http.addr", addr)
				cfg.Set("static.fs", map[string]embed.FS{"templates": templateFS, "web.static": staticFS})
				cfg.Set("http.middleware", []httpInterfaces.HandlerFunc{})
				cfg.Set("http.router", func(router httpInterfaces.RouterInterface) {
					router.GET("/", func(c httpInterfaces.RouterContextInterface) {
						c.HTML(http.StatusOK, "index.html", map[string]any{
							"SiteName": "Homepage test", "LogoURL": "", "Copyright": "",
							"CopyrightEnabled": false, "ICPNumber": "", "PoliceNumber": "",
						})
					})
				})
				server := httpImpl.NewHttpServer()
				server.RunHttp()
				t.Cleanup(func() {
					if err := server.Stop(); err != nil {
						t.Error(err)
					}
				})
				client := &http.Client{Timeout: 2 * time.Second}
				t.Cleanup(client.CloseIdleConnections)
				base := "http://" + addr
				var response *http.Response
				deadline := time.Now().Add(3 * time.Second)
				for time.Now().Before(deadline) {
					response, err = client.Get(base + "/")
					if err == nil {
						break
					}
					time.Sleep(10 * time.Millisecond)
				}
				if err != nil {
					t.Fatal(err)
				}
				html, err := io.ReadAll(response.Body)
				response.Body.Close()
				if err != nil {
					t.Fatal(err)
				}
				if response.StatusCode != http.StatusOK || !bytes.Contains(html, []byte("/assets/home/hero.webp")) {
					t.Fatalf("homepage did not render: status=%d", response.StatusCode)
				}
				entries, err := staticFS.ReadDir("static/assets/home")
				if err != nil {
					t.Fatal(err)
				}
				if len(entries) == 0 {
					t.Fatal("no homepage assets were embedded")
				}
				for _, entry := range entries {
					if entry.IsDir() {
						continue
					}
					path := "/assets/home/" + entry.Name()
					response, err := client.Get(base + path)
					if err != nil {
						t.Fatal(err)
					}
					body, err := io.ReadAll(response.Body)
					response.Body.Close()
					if err != nil {
						t.Fatal(err)
					}
					if tc.disabled {
						if response.StatusCode != http.StatusNotFound {
							t.Fatalf("disabled static route %s: %d", path, response.StatusCode)
						}
						continue
					}
					expected, err := staticFS.ReadFile("static" + path)
					if err != nil {
						t.Fatal(err)
					}
					if response.StatusCode != http.StatusOK || !bytes.Equal(body, expected) {
						t.Fatalf("%s: status=%d, asset bytes match=%t (static_dir=%v)", path, response.StatusCode, bytes.Equal(body, expected), cfg.GetStringSlice("http.static_dir", nil))
					}
					mediaType, _, _ := mime.ParseMediaType(response.Header.Get("Content-Type"))
					switch filepath.Ext(path) {
					case ".webp":
						if mediaType != "image/webp" {
							t.Errorf("%s Content-Type=%s", path, mediaType)
						}
					case ".svg":
						if mediaType != "image/svg+xml" {
							t.Errorf("%s Content-Type=%s", path, mediaType)
						}
					case ".css":
						if mediaType != "text/css" {
							t.Errorf("%s Content-Type=%s", path, mediaType)
						}
					case ".js":
						if !strings.Contains(mediaType, "javascript") {
							t.Errorf("%s Content-Type=%s", path, mediaType)
						}
					}
					head, err := client.Head(base + path)
					if err != nil {
						t.Fatal(err)
					}
					head.Body.Close()
					if head.StatusCode != http.StatusOK {
						t.Errorf("HEAD %s: %d", path, head.StatusCode)
					}
				}
				missing, err := client.Get(base + "/assets/home/does-not-exist.webp")
				if err != nil {
					t.Fatal(err)
				}
				missing.Body.Close()
				if missing.StatusCode != http.StatusNotFound {
					t.Fatalf("missing asset should return 404, got %d", missing.StatusCode)
				}
			})
		}
	}
}
