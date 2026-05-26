package assembly

import "testing"

func TestVersionAssemblyUsesBuildInfo(t *testing.T) {
	assembled, err := Version{
		Version:   "v1.2.3",
		GitCommit: "abc12345",
		BuildTime: "2026-05-26T12:34:56Z",
	}.Assembly()
	if err != nil {
		t.Fatalf("Assembly() error = %v", err)
	}

	version, ok := assembled.(interface {
		GetVersion() string
		GetGitCommit() string
		GetBuildTime() string
	})
	if !ok {
		t.Fatalf("Assembly() returned %T, want version service", assembled)
	}

	if got := version.GetVersion(); got != "v1.2.3" {
		t.Fatalf("GetVersion() = %q, want %q", got, "v1.2.3")
	}
	if got := version.GetGitCommit(); got != "abc12345" {
		t.Fatalf("GetGitCommit() = %q, want %q", got, "abc12345")
	}
	if got := version.GetBuildTime(); got != "2026-05-26T12:34:56Z" {
		t.Fatalf("GetBuildTime() = %q, want %q", got, "2026-05-26T12:34:56Z")
	}
}
