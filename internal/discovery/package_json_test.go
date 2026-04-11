package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindPackageJSON_InCurrentDir(t *testing.T) {
	dir := t.TempDir()
	pkgFile := filepath.Join(dir, "package.json")
	os.WriteFile(pkgFile, []byte(`{}`), 0644)

	result, err := FindPackageJSON(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != pkgFile {
		t.Errorf("expected %s, got %s", pkgFile, result)
	}
}

func TestFindPackageJSON_InParentDir(t *testing.T) {
	parent := t.TempDir()
	child := filepath.Join(parent, "subdir")
	os.Mkdir(child, 0755)
	pkgFile := filepath.Join(parent, "package.json")
	os.WriteFile(pkgFile, []byte(`{}`), 0644)

	result, err := FindPackageJSON(child)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != pkgFile {
		t.Errorf("expected %s, got %s", pkgFile, result)
	}
}

func TestFindPackageJSON_NotFound(t *testing.T) {
	dir := t.TempDir()

	result, err := FindPackageJSON(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got %s", result)
	}
}

func TestIsMakeHelpDependency_InDeps(t *testing.T) {
	dir := t.TempDir()
	pkgFile := filepath.Join(dir, "package.json")
	os.WriteFile(pkgFile, []byte(`{"dependencies":{"make-help":"^1.0.0"}}`), 0644)

	isDep, err := IsMakeHelpDependency(pkgFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isDep {
		t.Error("expected make-help to be detected as dependency")
	}
}

func TestIsMakeHelpDependency_InDevDeps(t *testing.T) {
	dir := t.TempDir()
	pkgFile := filepath.Join(dir, "package.json")
	os.WriteFile(pkgFile, []byte(`{"devDependencies":{"make-help":"^1.0.0"}}`), 0644)

	isDep, err := IsMakeHelpDependency(pkgFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isDep {
		t.Error("expected make-help to be detected as devDependency")
	}
}

func TestIsMakeHelpDependency_ScopedName(t *testing.T) {
	dir := t.TempDir()
	pkgFile := filepath.Join(dir, "package.json")
	os.WriteFile(pkgFile, []byte(`{"devDependencies":{"@sdlcforge/make-help":"^1.0.0"}}`), 0644)

	isDep, err := IsMakeHelpDependency(pkgFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isDep {
		t.Error("expected @sdlcforge/make-help to be detected as devDependency")
	}
}

func TestIsMakeHelpDependency_NotPresent(t *testing.T) {
	dir := t.TempDir()
	pkgFile := filepath.Join(dir, "package.json")
	os.WriteFile(pkgFile, []byte(`{"dependencies":{"express":"^4.0.0"}}`), 0644)

	isDep, err := IsMakeHelpDependency(pkgFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isDep {
		t.Error("expected make-help NOT to be detected as dependency")
	}
}

func TestDetectDynamicMode_WithDependency(t *testing.T) {
	dir := t.TempDir()
	pkgFile := filepath.Join(dir, "package.json")
	os.WriteFile(pkgFile, []byte(`{"devDependencies":{"@sdlcforge/make-help":"^1.0.0"}}`), 0644)

	if !DetectDynamicMode(dir) {
		t.Error("expected dynamic mode to be detected")
	}
}

func TestDetectDynamicMode_WithoutDependency(t *testing.T) {
	dir := t.TempDir()
	pkgFile := filepath.Join(dir, "package.json")
	os.WriteFile(pkgFile, []byte(`{"dependencies":{"express":"^4.0.0"}}`), 0644)

	if DetectDynamicMode(dir) {
		t.Error("expected static mode (no make-help dependency)")
	}
}

func TestDetectDynamicMode_NoPackageJSON(t *testing.T) {
	dir := t.TempDir()

	if DetectDynamicMode(dir) {
		t.Error("expected static mode (no package.json)")
	}
}
