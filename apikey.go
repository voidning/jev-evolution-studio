package main

import (
	"os"
	"path/filepath"
	"strings"
)

func loadAPIKey() string {
	if value := strings.TrimSpace(os.Getenv("TYPESAFE_API_KEY")); value != "" {
		return value
	}
	paths := apiKeyPaths("")
	if executable, err := os.Executable(); err == nil {
		paths = apiKeyPaths(executable)
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			parts := strings.SplitN(strings.TrimSpace(line), "=", 2)
			if len(parts) == 2 && parts[0] == "TYPESAFE_API_KEY" {
				return strings.Trim(strings.TrimSpace(parts[1]), "\"'")
			}
		}
	}
	return ""
}

func apiKeyPaths(executable string) []string {
	developmentPaths := []string{".env.local", "web-prototype/.env.local"}
	if executable == "" {
		return developmentPaths
	}
	executableDir := filepath.Dir(executable)
	contentsDir := filepath.Dir(executableDir)
	appBundle := filepath.Dir(contentsDir)
	if filepath.Base(executableDir) != "MacOS" || filepath.Base(contentsDir) != "Contents" || !strings.HasSuffix(filepath.Base(appBundle), ".app") {
		return append([]string{filepath.Join(executableDir, ".env.local")}, developmentPaths...)
	}
	projectRoot := filepath.Clean(filepath.Join(executableDir, "..", "..", "..", "..", ".."))
	return []string{
		filepath.Join(projectRoot, ".env.local"),
		filepath.Join(projectRoot, "web-prototype", ".env.local"),
		filepath.Join(executableDir, ".env.local"),
		developmentPaths[0],
		developmentPaths[1],
	}
}
