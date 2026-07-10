package resolver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const markerName = ".envmoat"

// MarkerContent represents the parsed content of a .envmoat marker file.
type MarkerContent int

const (
	// MarkerDefault means the marker file is empty — use the default auto bundle.
	MarkerDefault MarkerContent = iota
	// MarkerDisabled means the marker file contains "disabled" — stop, no bundle.
	MarkerDisabled
	// MarkerProfile means the marker file contains "profile: <name>" — use named profile.
	MarkerProfile
	// MarkerUnknown is returned when the marker file could not be parsed.
	MarkerUnknown
)

// ResolveResult holds the result of resolving a marker from a directory.
type ResolveResult struct {
	// MarkerDir is the absolute path of the directory containing the .envmoat marker.
	MarkerDir string
	// Marker is the parsed marker content type.
	Marker MarkerContent
	// ProfileName is the profile name when Marker is MarkerProfile; empty otherwise.
	ProfileName string
}

// Resolve walks up from dir looking for a .envmoat marker file.
// Returns the directory containing the marker and its parsed content.
// Returns nil, nil if no marker is found (not an error).
func Resolve(dir string) (*ResolveResult, error) {
	dir = filepath.Clean(dir)
	walkRoot := FindWalkRoot()

	for {
		markerPath := filepath.Join(dir, markerName)
		if info, err := os.Stat(markerPath); err == nil && !info.IsDir() {
			content, profileName, err := ParseMarker(markerPath)
			if err != nil {
				return nil, err
			}
			absDir, err := filepath.Abs(dir)
			if err != nil {
				return nil, fmt.Errorf("resolve absolute path for %s: %w", dir, err)
			}
			debug("found marker at %s (content: %s)", markerPath, describeContent(content))
			return &ResolveResult{
				MarkerDir:   absDir,
				Marker:      content,
				ProfileName: profileName,
			}, nil
		}

		debug("checking %s — no %s", dir, markerName)
		if dir == walkRoot || dir == "/" {
			debug("reached walk root %s", walkRoot)
			return nil, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Safety: should not happen if walkRoot is correct, but guard against infinite loop.
			break
		}
		dir = parent
	}

	return nil, nil
}

// ResolveFromPWD resolves from the current working directory.
func ResolveFromPWD() (*ResolveResult, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}
	return Resolve(pwd)
}

// ParseMarker reads and parses a .envmoat marker file.
// Returns the marker content type, profile name (if applicable), and any error.
func ParseMarker(path string) (MarkerContent, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return MarkerUnknown, "", fmt.Errorf("read marker %s: %w", path, err)
	}

	// Trim whitespace and trailing newline.
	content := strings.TrimSpace(string(data))

	if content == "" {
		return MarkerDefault, "", nil
	}

	if content == "disabled" {
		return MarkerDisabled, "", nil
	}

	if strings.HasPrefix(content, "profile: ") {
		profileName := strings.TrimSpace(content[len("profile: "):])
		if profileName == "" {
			return MarkerUnknown, "", fmt.Errorf("marker %s: empty profile name after 'profile: '", path)
		}
		return MarkerProfile, profileName, nil
	}

	return MarkerUnknown, "", fmt.Errorf("marker %s: unrecognized content (expected empty, 'disabled', or 'profile: <name>')", path)
}

// FindWalkRoot returns the walk boundary directory.
// Returns the value of ENVMOAT_WALK_ROOT env var, or "/" if not set.
func FindWalkRoot() string {
	root := os.Getenv("ENVMOAT_WALK_ROOT")
	if root == "" {
		return "/"
	}
	root = filepath.Clean(root)
	abs, err := filepath.Abs(root)
	if err != nil {
		return "/"
	}
	return abs
}

func debug(format string, args ...any) {
	if os.Getenv("ENVMOAT_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "envmoat: resolver: "+format+"\n", args...)
	}
}

func describeContent(c MarkerContent) string {
	switch c {
	case MarkerDefault:
		return "default"
	case MarkerDisabled:
		return "disabled"
	case MarkerProfile:
		return "profile"
	case MarkerUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}
