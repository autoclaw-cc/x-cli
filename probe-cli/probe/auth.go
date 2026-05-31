package probe

import (
	"encoding/json"
	"fmt"
	"strings"

	"probe-cli/browser"
)

// authDetectJS scans localStorage, sessionStorage, and cookies for auth-related keys.
const authDetectJS = `(() => {
  const AUTH_PATTERNS = [
    'token', 'access_token', 'auth_token', 'jwt', 'bearer',
    'session', 'user_id', 'uid', 'api_key', 'csrf', 'ct0',
    'xt', 'sid', 'refresh', 'id_token', 'identity'
  ];

  const result = {
    localStorage: {},
    sessionStorage: {},
    cookies: {},
    globals: {},
    detected_type: null
  };

  // Helper: check if key matches any auth pattern
  function isAuthKey(key) {
    const lower = key.toLowerCase();
    return AUTH_PATTERNS.some(p => lower.includes(p));
  }

  // Scan localStorage
  try {
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (isAuthKey(key)) {
        result.localStorage[key] = localStorage.getItem(key).substring(0, 80);
      }
    }
  } catch(e) {}

  // Scan sessionStorage
  try {
    for (let i = 0; i < sessionStorage.length; i++) {
      const key = sessionStorage.key(i);
      if (isAuthKey(key)) {
        result.sessionStorage[key] = sessionStorage.getItem(key).substring(0, 80);
      }
    }
  } catch(e) {}

  // Scan cookies
  try {
    document.cookie.split(';').forEach(c => {
      const eq = c.indexOf('=');
      if (eq < 0) return;
      const name = c.substring(0, eq).trim();
      if (isAuthKey(name)) {
        result.cookies[name] = c.substring(eq + 1).trim().substring(0, 80);
      }
    });
  } catch(e) {}

  // Check framework globals
  if (window.__NEXT_DATA__) result.globals['__NEXT_DATA__'] = true;
  if (window.__NUXT__) result.globals['__NUXT__'] = true;

  // Detect auth type
  const lsKeys = Object.keys(result.localStorage);
  const ssKeys = Object.keys(result.sessionStorage);
  const ckKeys = Object.keys(result.cookies);

  if (lsKeys.some(k => k.toLowerCase().includes('token'))) {
    result.detected_type = 'bearer-localstorage';
  } else if (ckKeys.some(k => k.toLowerCase().includes('csrf') || k.toLowerCase().includes('ct0'))) {
    result.detected_type = 'csrf-cookie';
  } else if (ckKeys.length > 0 && lsKeys.length === 0) {
    result.detected_type = 'cookie-only';
  } else if (lsKeys.length > 0) {
    result.detected_type = 'localstorage-other';
  } else if (ssKeys.length > 0) {
    result.detected_type = 'session-storage';
  }

  return result;
})()`

// rawAuthResult mirrors the JS return value.
type rawAuthResult struct {
	LocalStorage map[string]string `json:"localStorage"`
	SessionStorage map[string]string `json:"sessionStorage"`
	Cookies      map[string]string `json:"cookies"`
	Globals      map[string]bool   `json:"globals"`
	DetectedType *string           `json:"detected_type"`
}

// DetectAuth probes the page for authentication mechanisms.
func DetectAuth(client *browser.Client) (*AuthResult, []string) {
	var warnings []string

	raw, err := client.Evaluate(authDetectJS)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("auth detection failed: %v", err))
		return &AuthResult{Detected: false}, warnings
	}

	var parsed rawAuthResult
	if err := json.Unmarshal(raw, &parsed); err != nil {
		warnings = append(warnings, fmt.Sprintf("auth result parse failed: %v", err))
		return &AuthResult{Detected: false}, warnings
	}

	result := &AuthResult{}

	// Collect all auth keys found
	var allKeys []string
	for k := range parsed.LocalStorage {
		allKeys = append(allKeys, "localStorage:"+k)
	}
	for k := range parsed.SessionStorage {
		allKeys = append(allKeys, "sessionStorage:"+k)
	}
	for k := range parsed.Cookies {
		allKeys = append(allKeys, "cookie:"+k)
	}
	if len(allKeys) > 0 {
		allJSON, _ := json.Marshal(allKeys)
		result.AllKeys = allJSON
	}

	// Determine method
	if parsed.DetectedType != nil && *parsed.DetectedType != "" {
		result.Detected = true
		result.Method = *parsed.DetectedType

		// Extract token keys
		switch {
		case strings.Contains(result.Method, "localstorage"):
			result.StorageType = "localStorage"
			for k := range parsed.LocalStorage {
				if strings.Contains(strings.ToLower(k), "token") {
					result.TokenKeys = append(result.TokenKeys, k)
				}
			}
		case strings.Contains(result.Method, "session"):
			result.StorageType = "sessionStorage"
			for k := range parsed.SessionStorage {
				result.TokenKeys = append(result.TokenKeys, k)
			}
		case strings.Contains(result.Method, "cookie"):
			result.StorageType = "cookie"
			for k := range parsed.Cookies {
				result.TokenKeys = append(result.TokenKeys, k)
			}
		}
	}

	return result, warnings
}
