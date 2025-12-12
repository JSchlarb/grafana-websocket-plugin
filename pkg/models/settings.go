package models

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type PluginSettings struct {
	Path            string                `json:"path"`
	Headers         map[string]string     `json:"headers,omitempty"`
	QueryParameters map[string]string     `json:"queryParameters,omitempty"`
	Secrets         *SecretPluginSettings `json:"-"`
}

type SecretPluginSettings struct {
	ApiKey string            `json:"apiKey"`
	secure map[string]string `json:"-"`
}

// LoadPluginSettings reads JSONData/SecureJSONData from Grafana into a typed struct.
// It also supports the legacy headerName*/queryParamName* pattern by merging those keys.
func LoadPluginSettings(source backend.DataSourceInstanceSettings) (*PluginSettings, error) {
	raw := map[string]any{}
	if err := json.Unmarshal(source.JSONData, &raw); err != nil {
		return nil, fmt.Errorf("could not unmarshal PluginSettings json: %w", err)
	}

	settings := &PluginSettings{
		Path:            toString(raw["path"]),
		Headers:         map[string]string{},
		QueryParameters: map[string]string{},
		Secrets:         loadSecretPluginSettings(source.DecryptedSecureJSONData),
	}

	// Prefer structured headers/queryParameters if provided
	if hdrs, ok := raw["headers"].(map[string]any); ok {
		for k, v := range hdrs {
			if s := toString(v); s != "" {
				settings.Headers[k] = s
			}
		}
	}
	if qps, ok := raw["queryParameters"].(map[string]any); ok {
		for k, v := range qps {
			if s := toString(v); s != "" {
				settings.QueryParameters[k] = s
			}
		}
	}

	// Fallback to legacy headerName*/queryParamName* pattern if present
	for key, val := range raw {
		if !hasPrefix(key, "headerName") && !hasPrefix(key, "queryParamName") {
			continue
		}
		name := toString(val)
		if name == "" {
			continue
		}
		switch {
		case hasPrefix(key, "headerName"):
			settings.Headers[name] = settings.Secrets.HeaderValue(key)
		case hasPrefix(key, "queryParamName"):
			settings.QueryParameters[name] = settings.Secrets.QueryParamValue(key)
		}
	}

	return settings, nil
}

func loadSecretPluginSettings(source map[string]string) *SecretPluginSettings {
	return &SecretPluginSettings{
		ApiKey: source["apiKey"],
		secure: source,
	}
}

func (s *SecretPluginSettings) HeaderValue(headerNameKey string) string {
	return s.valueFor(prefixFromName(headerNameKey), suffixFromName(headerNameKey))
}

func (s *SecretPluginSettings) QueryParamValue(paramNameKey string) string {
	return s.valueFor(prefixFromName(paramNameKey), suffixFromName(paramNameKey))
}

func (s *SecretPluginSettings) valueFor(prefix, suffix string) string {
	if s == nil {
		return ""
	}
	// Stored as prefix + "Value" + suffix
	return s.secure[prefix+"Value"+suffix]
}

// prefixFromName splits "headerName1" -> "header", suffix "1"
func prefixFromName(setting string) string {
	if parts := splitOnName(setting); len(parts) == 2 {
		return parts[0]
	}
	return ""
}

// suffixFromName splits "headerName1" -> "1"
func suffixFromName(setting string) string {
	if parts := splitOnName(setting); len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// splitOnName splits the first occurrence of "Name" in a legacy key.
func splitOnName(setting string) []string {
	const marker = "Name"
	if idx := indexOf(setting, marker); idx >= 0 {
		return []string{setting[:idx], setting[idx+len(marker):]}
	}
	return []string{setting}
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return fmt.Sprint(t)
	}
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
