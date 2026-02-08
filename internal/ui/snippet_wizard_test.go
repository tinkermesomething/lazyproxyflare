package ui

import (
	"strings"
	"testing"

	snippet_wizard "lazyproxyflare/internal/ui/snippet_wizard"
)

func TestRenderSnippetWizardTemplateParams(t *testing.T) {
	tests := []struct {
		name              string
		selectedTemplates map[string]bool
		snippetConfigs    map[string]snippet_wizard.SnippetConfig
		wantContains      string
	}{
		{
			name: "CORS headers template selected",
			selectedTemplates: map[string]bool{
				"cors_headers": true,
			},
			snippetConfigs: map[string]snippet_wizard.SnippetConfig{
				"cors_headers": {
					Parameters: map[string]interface{}{
						"allowed_origins": "*",
						"allowed_methods": "GET, POST",
					},
				},
			},
			wantContains: "Configure: CORS Headers",
		},
		{
			name: "Rate limiting template selected",
			selectedTemplates: map[string]bool{
				"rate_limiting": true,
			},
			snippetConfigs: map[string]snippet_wizard.SnippetConfig{
				"rate_limiting": {
					Parameters: map[string]interface{}{
						"requests_per_second": 100,
						"burst_size":          50,
					},
				},
			},
			wantContains: "Configure: Rate Limiting",
		},
		{
			name: "No template selected",
			selectedTemplates: map[string]bool{
				"cors_headers": false,
			},
			snippetConfigs: map[string]snippet_wizard.SnippetConfig{},
			wantContains:   "No template selected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				snippetWizardData: SnippetWizardData{
					SelectedTemplates: tt.selectedTemplates,
					SnippetConfigs:    tt.snippetConfigs,
				},
				wizardCursor: 0,
			}

			result := m.renderSnippetWizardTemplateParams()

			if result == "" {
				t.Errorf("Expected non-empty render result")
			}

			if tt.wantContains != "" {
				if !strings.Contains(result, tt.wantContains) {
					t.Errorf("Expected result to contain %q, got:\n%s", tt.wantContains, result)
				}
			}
		})
	}
}
