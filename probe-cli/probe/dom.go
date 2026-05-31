package probe

import (
	"encoding/json"
	"fmt"

	"probe-cli/browser"
)

// domExtractJS scans the page for forms, inputs, and buttons.
const domExtractJS = `(() => {
  const result = { forms: [], standalone_inputs: [], buttons: [] };

  // Forms
  document.querySelectorAll('form').forEach((f, i) => {
    const inputs = Array.from(f.querySelectorAll('input, textarea, select')).map(el => ({
      tag: el.tagName.toLowerCase(),
      type: el.getAttribute('type') || null,
      name: el.getAttribute('name') || null,
      placeholder: el.getAttribute('placeholder') || null,
      id: el.id || null,
      content_editable: el.contentEditable === 'true',
      role: el.getAttribute('role') || null
    }));
    result.forms.push({
      index: i,
      action: f.action || null,
      method: (f.method || 'GET').toUpperCase(),
      id: f.id || null,
      input_count: inputs.length,
      inputs: inputs
    });
  });

  // Standalone inputs (not inside a form)
  document.querySelectorAll('input, textarea, select, [contenteditable="true"]').forEach(el => {
    if (!el.closest('form')) {
      result.standalone_inputs.push({
        tag: el.tagName.toLowerCase(),
        type: el.getAttribute('type') || 'text',
        name: el.getAttribute('name') || null,
        placeholder: el.getAttribute('placeholder') || null,
        id: el.id || null,
        content_editable: el.contentEditable === 'true',
        role: el.getAttribute('role') || null
      });
    }
  });

  // Buttons (limit to 30 to avoid huge output)
  const buttons = document.querySelectorAll('button, [role="button"], input[type="submit"]');
  const limit = Math.min(buttons.length, 30);
  for (let i = 0; i < limit; i++) {
    const el = buttons[i];
    result.buttons.push({
      text: (el.textContent || '').trim().substring(0, 80) || null,
      type: el.getAttribute('type') || null,
      id: el.id || null,
      aria_label: el.getAttribute('aria-label') || null,
      disabled: !!el.disabled
    });
  }

  return result;
})()`

type rawDOMResult struct {
	Forms            []rawFormInfo   `json:"forms"`
	StandaloneInputs []rawInputInfo  `json:"standalone_inputs"`
	Buttons          []rawButtonInfo `json:"buttons"`
}

type rawFormInfo struct {
	Index      int            `json:"index"`
	Action     string         `json:"action"`
	Method     string         `json:"method"`
	ID         string         `json:"id"`
	InputCount int            `json:"input_count"`
	Inputs     []rawInputInfo `json:"inputs"`
}

type rawInputInfo struct {
	Tag             string `json:"tag"`
	Type            string `json:"type"`
	Name            string `json:"name"`
	Placeholder     string `json:"placeholder"`
	ID              string `json:"id"`
	ContentEditable bool   `json:"content_editable"`
	Role            string `json:"role"`
}

type rawButtonInfo struct {
	Text      string `json:"text"`
	Type      string `json:"type"`
	ID        string `json:"id"`
	AriaLabel string `json:"aria_label"`
	Disabled  bool   `json:"disabled"`
}

// ExtractDOM probes the page for interactive elements.
func ExtractDOM(client *browser.Client) (*DOMResult, []string) {
	var warnings []string

	raw, err := client.Evaluate(domExtractJS)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("DOM extraction failed: %v", err))
		return &DOMResult{}, warnings
	}

	var parsed rawDOMResult
	if err := json.Unmarshal(raw, &parsed); err != nil {
		warnings = append(warnings, fmt.Sprintf("DOM result parse failed: %v", err))
		return &DOMResult{}, warnings
	}

	result := &DOMResult{}

	// Convert forms
	for _, f := range parsed.Forms {
		form := FormInfo{
			Index:      f.Index,
			Action:     f.Action,
			Method:     f.Method,
			ID:         f.ID,
			InputCount: f.InputCount,
		}
		for _, inp := range f.Inputs {
			form.Inputs = append(form.Inputs, InputInfo{
				Tag:             inp.Tag,
				Type:            inp.Type,
				Name:            inp.Name,
				Placeholder:     inp.Placeholder,
				ID:              inp.ID,
				ContentEditable: inp.ContentEditable,
				Role:            inp.Role,
			})
		}
		result.Forms = append(result.Forms, form)
	}

	// Convert standalone inputs
	for _, inp := range parsed.StandaloneInputs {
		result.StandaloneInputs = append(result.StandaloneInputs, InputInfo{
			Tag:             inp.Tag,
			Type:            inp.Type,
			Name:            inp.Name,
			Placeholder:     inp.Placeholder,
			ID:              inp.ID,
			ContentEditable: inp.ContentEditable,
			Role:            inp.Role,
		})
	}

	// Convert buttons
	for _, b := range parsed.Buttons {
		result.Buttons = append(result.Buttons, ButtonInfo{
			Text:      b.Text,
			Type:      b.Type,
			ID:        b.ID,
			AriaLabel: b.AriaLabel,
			Disabled:  b.Disabled,
		})
	}

	return result, warnings
}
