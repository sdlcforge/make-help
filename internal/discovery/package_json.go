package discovery

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// packageJSON is a minimal representation of package.json for dependency checking.
type packageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// FindPackageJSON walks up from startDir looking for package.json.
// Returns the path if found, or empty string if not found.
func FindPackageJSON(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	for {
		candidate := filepath.Join(dir, "package.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return "", nil
		}
		dir = parent
	}
}

// IsMakeHelpDependency checks whether make-help is listed as a dependency
// in the given package.json file. Checks both dependencies and devDependencies,
// and looks for both "make-help" and "@sdlcforge/make-help" package names.
func IsMakeHelpDependency(packageJSONPath string) (bool, error) {
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return false, err
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false, err
	}

	for _, deps := range []map[string]string{pkg.Dependencies, pkg.DevDependencies} {
		if _, ok := deps["make-help"]; ok {
			return true, nil
		}
		if _, ok := deps["@sdlcforge/make-help"]; ok {
			return true, nil
		}
	}

	return false, nil
}

// DetectDynamicMode checks whether dynamic help mode should be used
// based on the presence of make-help as a dependency in a nearby package.json.
// Returns true if make-help is found as a dependency, false otherwise.
func DetectDynamicMode(makefileDir string) bool {
	pkgPath, err := FindPackageJSON(makefileDir)
	if err != nil || pkgPath == "" {
		return false
	}

	isDep, err := IsMakeHelpDependency(pkgPath)
	if err != nil {
		return false
	}

	return isDep
}
