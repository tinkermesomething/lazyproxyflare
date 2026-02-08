package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/config"
	"lazyproxyflare/internal/diff"
	snippet_wizard "lazyproxyflare/internal/ui/snippet_wizard"
)

// handleKeyMsg handles all keyboard input for the application.
// This method was extracted from Update() to improve maintainability.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
		// CRITICAL: Handle wizard text input FIRST (before keybindings)
		// When in wizard text input mode, forward ALL keys to textinput component
		// EXCEPT escape (cancel), enter (submit), 'b' (back), and navigation keys (j/k)
		if m.currentView == ViewWizard && m.isWizardTextInputStep() {
			key := msg.String()
			// Let navigation keys through to wizard handler
			if key != "esc" && key != "enter" && key != "tab" && key != "shift+tab" && key != "down" && key != "up" {
				var cmd tea.Cmd
				m.wizardTextInput, cmd = m.wizardTextInput.Update(msg)
				return m, cmd
			}
		}		

		// Handle text input in add/edit form - Subdomain field (field 0, multi-line)
		if (m.currentView == ViewAdd || m.currentView == ViewEdit) && m.addForm.FocusedField == 0 {
			key := msg.String()
			// Allow Enter for newlines to support multiple subdomains
			if key == "enter" {
				m.addForm.Subdomain += "\n"
				return m, nil
			}
			// Allow single character input for subdomain field
			if len(key) == 1 {
				// Allow alphanumeric, hyphen, underscore, period (for FQDN entry)
				if (key >= "a" && key <= "z") || (key >= "A" && key <= "Z") ||
					(key >= "0" && key <= "9") || key == "-" || key == "_" || key == "." {
					m.addForm.Subdomain += key
					return m, nil
				}
			}
		}

		// Handle text input in add/edit form - Other text fields
		// Text fields are: 2 (DNS Target), 4 (Reverse Proxy), 5 (Service Port)
		if (m.currentView == ViewAdd || m.currentView == ViewEdit) &&
			(m.addForm.FocusedField == 2 || m.addForm.FocusedField == 4 || m.addForm.FocusedField == 5) {
			// Only handle single character input for text fields
			if len(msg.String()) == 1 {
				char := msg.String()
				switch m.addForm.FocusedField {
				case 2: // DNS Target - validation depends on DNS type
					if m.addForm.DNSType == "A" {
						// A record: allow IP address characters (numbers and dots)
						if (char >= "0" && char <= "9") || char == "." {
							m.addForm.DNSTarget += char
							return m, nil
						}
					} else {
						// CNAME: allow domain characters
						if (char >= "a" && char <= "z") || (char >= "A" && char <= "Z") ||
							(char >= "0" && char <= "9") || char == "-" || char == "." {
							m.addForm.DNSTarget += char
							return m, nil
						}
					}
				case 4: // Reverse Proxy Target
					// Allow domain/IP characters
					if (char >= "a" && char <= "z") || (char >= "A" && char <= "Z") ||
						(char >= "0" && char <= "9") || char == "-" || char == "." {
						m.addForm.ReverseProxyTarget += char
						return m, nil
					}
				case 5: // Service Port
					// Allow only numbers
					if char >= "0" && char <= "9" {
						m.addForm.ServicePort += char
						return m, nil
					}
				}
			}
		}

		// Handle text input for custom Caddy config field (multi-line)
		// Field index: 8 + len(snippets)
		customConfigFieldIndex := 8 + len(m.snippets)
		if (m.currentView == ViewAdd || m.currentView == ViewEdit) &&
			m.addForm.FocusedField == customConfigFieldIndex {
			key := msg.String()
			// Ctrl+M for newlines (Enter alone proceeds to preview)
			// Note: Ctrl+Enter sends "ctrl+m" in bubbletea (both are ASCII 13)
			if key == "ctrl+m" {
				m.addForm.CustomCaddyConfig += "\n"
				return m, nil
			}
			// Allow all printable characters
			if len(key) == 1 {
				m.addForm.CustomCaddyConfig += key
				return m, nil
			}
		}

		// Handle text input in profile edit mode
		if m.currentView == ViewProfileEdit && len(msg.String()) == 1 {
			char := msg.String()
			// Allow printable ASCII characters for text fields
			if char[0] >= 32 && char[0] <= 126 {
				return m.handleProfileEditKeyPress(char)
			}
		}

		// Handle text input in search mode
		if m.searching && len(msg.String()) == 1 {
			m.searchQuery += msg.String()
			m.cursor = 0
			m.scrollOffset = 0
			return m, nil
		}

		// Handle text input in snippet wizard IP restriction step
		if m.currentView == ViewSnippetWizard && m.snippetWizardStep == SnippetWizardIPRestriction && m.snippetWizardData.CreateIPRestriction {
			if len(msg.String()) == 1 {
				char := msg.String()
				// Allow CIDR notation characters: numbers, dots, slashes
				if (char >= "0" && char <= "9") || char == "." || char == "/" {
					if m.wizardCursor == 0 {
						// LAN Subnet field
						m.snippetWizardData.LANSubnet += char
						return m, nil
					} else if m.wizardCursor == 1 {
						// External IP field
						m.snippetWizardData.AllowedExternalIP += char
						return m, nil
					}
				}
			}
		}

		// Handle text input in custom snippet wizard step using textinput component
		if m.currentView == ViewSnippetWizard && m.snippetWizardStep == snippet_wizard.StepCustomSnippet {
			key := msg.String()
			if key == "tab" || key == "shift+tab" || key == "enter" || key == "esc" || key == "up" || key == "down" {
				// Let these fall through to navigation/submit/cancel handlers
			} else {
				var cmd tea.Cmd
				if m.wizardCursor == 0 {
					m.snippetWizardData.CustomNameInput, cmd = m.snippetWizardData.CustomNameInput.Update(msg)
					m.snippetWizardData.CustomSnippetName = m.snippetWizardData.CustomNameInput.Value()
				} else {
					m.snippetWizardData.CustomContentInput, cmd = m.snippetWizardData.CustomContentInput.Update(msg)
					m.snippetWizardData.CustomSnippetContent = m.snippetWizardData.CustomContentInput.Value()
				}
				return m, cmd
			}
		}

		// Handle text input in snippet edit mode using textarea component
		if m.currentView == ViewSnippetDetail && m.snippetPanel.Editing {
			key := msg.String()
			if key == "y" || key == "enter" || key == "d" || key == "esc" {
				// Let these fall through to save/delete/cancel handlers
			} else {
				var cmd tea.Cmd
				m.snippetPanel.EditTextarea, cmd = m.snippetPanel.EditTextarea.Update(msg)
				return m, cmd
			}
		}

		// NOTE: Wizard text input is handled AFTER the switch statement
		// to ensure ESC, backspace, and other special keys work properly

		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "ctrl+q":
			// In snippet wizard: Ctrl+Q closes the modal
			if m.currentView == ViewSnippetWizard {
				m.currentView = ViewList
				m.snippetWizardData = SnippetWizardData{}
				m.err = nil
				return m, nil
			}
			// Otherwise quit app
			m.quitting = true
			return m, tea.Quit

		case "esc", "ctrl+w":
			return m.handleDismiss()

		case "tab":
			// In profile edit: navigate fields
			if m.currentView == ViewProfileEdit {
				return m.handleProfileEditKeyPress("tab")
			}
			// In setup wizard: navigate down through selections
			if m.currentView == ViewWizard {
				return m.handleWizardKeyPress("tab")
			}
			// In snippet wizard: navigate forward through fields/options
			if m.currentView == ViewSnippetWizard {
				switch m.snippetWizardStep {
				case SnippetWizardWelcome:
					// Navigate wizard mode selection (3 modes)
					if m.wizardCursor < 2 {
						m.wizardCursor++
					}
					return m, nil
				case snippet_wizard.StepTemplateSelection:
					// Navigate template list (15 templates)
					if m.wizardCursor < 14 {
						m.wizardCursor++
					}
					return m, nil
				case snippet_wizard.StepCustomSnippet:
					// Custom mode: cycle between name (0) and content (1)
					m.wizardCursor = (m.wizardCursor + 1) % 2
					// Update focus on textinput components
					if m.wizardCursor == 0 {
						m.snippetWizardData.CustomNameInput.Focus()
						m.snippetWizardData.CustomContentInput.Blur()
					} else {
						m.snippetWizardData.CustomNameInput.Blur()
						m.snippetWizardData.CustomContentInput.Focus()
					}
					return m, nil
				case snippet_wizard.StepIPRestriction:
					// IP Restriction: cycle between LAN subnet (0) and External IP (1)
					m.wizardCursor = (m.wizardCursor + 1) % 2
					return m, nil
				case snippet_wizard.StepTemplateParams:
					// Template params: navigate forward to next field
					maxCursor := m.getTemplateParamsCursorMax()
					if m.wizardCursor < maxCursor {
						m.wizardCursor++
					}
					return m, nil
				}
			}
			// In list view: switch between Cloudflare/Caddy tabs
			if m.currentView == ViewList {
				if m.activeTab == TabCloudflare {
					m.activeTab = TabCaddy
				} else {
					m.activeTab = TabCloudflare
				}
				// Reset cursor and scroll for new filtered list
				m.cursor = 0
				m.scrollOffset = 0
				return m, nil
			}
			// Navigate forward through form fields (in add/edit form)
			if m.currentView == ViewAdd || m.currentView == ViewEdit {
				// Smart navigation: skip Caddy fields if DNS-only mode
				if m.addForm.DNSOnly {
					// DNS-only mode: cycle through 0->1->2->3->6
					switch m.addForm.FocusedField {
					case 0:
						m.addForm.FocusedField = 1
					case 1:
						m.addForm.FocusedField = 2
					case 2:
						m.addForm.FocusedField = 3
					case 3:
						m.addForm.FocusedField = 6
					case 6:
						m.addForm.FocusedField = 0
					default:
						m.addForm.FocusedField = 0
					}
				} else {
					// Full mode: cycle through all fields (0-7 + snippets + custom config)
					maxFields := 8 + len(m.snippets) + 1
					m.addForm.FocusedField = (m.addForm.FocusedField + 1) % maxFields
				}
				return m, nil
			}

		case "shift+tab":
			// In setup wizard: navigate up through selections
			if m.currentView == ViewWizard {
				return m.handleWizardKeyPress("shift+tab")
			}
			// In snippet wizard: navigate backward through fields/options
			if m.currentView == ViewSnippetWizard {
				switch m.snippetWizardStep {
				case SnippetWizardWelcome:
					// Navigate wizard mode selection backward
					if m.wizardCursor > 0 {
						m.wizardCursor--
					}
					return m, nil
				case snippet_wizard.StepTemplateSelection:
					// Navigate template list backward
					if m.wizardCursor > 0 {
						m.wizardCursor--
					}
					return m, nil
				case snippet_wizard.StepCustomSnippet:
					// Custom mode: cycle backward between content (1) and name (0)
					m.wizardCursor = (m.wizardCursor - 1 + 2) % 2
					// Update focus on textinput components
					if m.wizardCursor == 0 {
						m.snippetWizardData.CustomNameInput.Focus()
						m.snippetWizardData.CustomContentInput.Blur()
					} else {
						m.snippetWizardData.CustomNameInput.Blur()
						m.snippetWizardData.CustomContentInput.Focus()
					}
					return m, nil
				case snippet_wizard.StepIPRestriction:
					// IP Restriction: cycle backward between External IP (1) and LAN subnet (0)
					m.wizardCursor = (m.wizardCursor - 1 + 2) % 2
					return m, nil
				case snippet_wizard.StepTemplateParams:
					// Template params: navigate backward to previous field
					if m.wizardCursor > 0 {
						m.wizardCursor--
					}
					return m, nil
				}
			}
			// In list view: cycle backward through panels
			if m.currentView == ViewList {
				switch m.panelFocus {
				case PanelFocusLeft:
					m.panelFocus = PanelFocusSnippets
				case PanelFocusSnippets:
					m.panelFocus = PanelFocusDetails
				case PanelFocusDetails:
					m.panelFocus = PanelFocusLeft
				}
				return m, nil
			}
			// Navigate backward through form fields (in add/edit form)
			if m.currentView == ViewAdd || m.currentView == ViewEdit {
				// Smart navigation: skip Caddy fields if DNS-only mode
				if m.addForm.DNSOnly {
					// DNS-only mode: cycle backward through 6->3->2->1->0
					switch m.addForm.FocusedField {
					case 0:
						m.addForm.FocusedField = 6
					case 1:
						m.addForm.FocusedField = 0
					case 2:
						m.addForm.FocusedField = 1
					case 3:
						m.addForm.FocusedField = 2
					case 6:
						m.addForm.FocusedField = 3
					default:
						m.addForm.FocusedField = 6
					}
				} else {
					// Full mode: cycle backward through all fields (0-7 + snippets + custom config)
					maxFields := 8 + len(m.snippets) + 1
					m.addForm.FocusedField = (m.addForm.FocusedField - 1 + maxFields) % maxFields
				}
				return m, nil
			}

		case "/":
			return m.handleSearchStart()

		case "f":
			return m.handleStatusFilterCycle()

		case "t":
			return m.handleDNSTypeFilterCycle()

		case "o":
			return m.handleSortModeCycle()

		case "a":
			return m.handleAddEntry()

		case "e":
			// Edit profile in profile selector
			if m.currentView == ViewProfileSelector {
				return m.handleProfileSelectorKeyPress("e")
			}

		case "w", "ctrl+s":
			// Open snippet wizard (only from list view, not while searching or loading)
			if m.currentView == ViewList && !m.searching && !m.loading {
				// Run auto-detection on Caddyfile
				detectedPatterns, err := snippet_wizard.DetectPatternsFromFile(m.config.Caddy.CaddyfilePath)
				if err != nil {
					// If detection fails, continue with empty patterns
					detectedPatterns = []snippet_wizard.DetectedPattern{}
				}

				// Initialize snippet wizard
				m.snippetWizardStep = SnippetWizardWelcome
				m.snippetWizardData = SnippetWizardData{
					// Set defaults from config
					LANSubnet:         m.config.Defaults.LANSubnet,
					AllowedExternalIP: m.config.Defaults.AllowedExternalIP,
					SecurityPreset:    "basic", // default to basic

					// Auto-detection data
					DetectedPatterns: detectedPatterns,
					SelectedPatterns: make(map[string]bool),

					// Template mode data
					SelectedTemplates: make(map[string]bool),
					SnippetConfigs:    make(map[string]snippet_wizard.SnippetConfig),

					// Custom mode data
					CustomSnippetName:    "",
					CustomSnippetContent: "",
				}
				m.wizardCursor = 0
				m.currentView = ViewSnippetWizard
				return m, nil
			}

		case "d":
			// Delete snippet when in edit mode
			if m.currentView == ViewSnippetDetail && m.snippetPanel.Editing {
				return m.deleteSnippet()
			}
			// Delete backup from preview mode
			if m.currentView == ViewBackupPreview && !m.loading {
				backups, err := caddy.ListBackups(m.config.Caddy.CaddyfilePath)
				if err == nil && m.backup.Cursor < len(backups) {
					m.loading = true
					m.currentView = ViewBackupManager // Return to manager after deletion
					return m, deleteBackupCmd(backups[m.backup.Cursor].Path)
				}
				return m, nil
			}
			// Delete selected entry (from list view)
			if m.currentView == ViewList && !m.searching && !m.loading && len(m.getFilteredEntries()) > 0 {
				m.delete.EntryIndex = m.cursor
				m.delete.ScopeCursor = 0         // Reset cursor
				m.delete.Scope = DeleteAll       // Default to delete all
				m.currentView = ViewDeleteScope // Go to scope selection first
				return m, nil
			}

		case "D":
			return m.handleOpenBulkDeleteMenu()

		case "b", "ctrl+b":
			return m.handleOpenBackupManager()

		case "l", "L":
			return m.handleOpenAuditLog()

		case "p", "ctrl+p":
			// Handle in profile selector
			if m.currentView == ViewProfileSelector {
				return m.handleProfileSelectorKeyPress("p")
			}
			// Open profile selector from main view
			if m.currentView == ViewList {
				m.cursor = 0
				// Load available profiles
				profiles, err := config.ListProfiles()
				if err == nil {
					m.profile.Available = profiles
				}
				m.currentView = ViewProfileSelector
				return m, nil
			}
			// Preview backup (from backup manager)
			if m.currentView == ViewBackupManager && !m.loading {
				backups, err := caddy.ListBackups(m.config.Caddy.CaddyfilePath)
				if err == nil && m.backup.Cursor < len(backups) {
					m.backup.PreviewPath = backups[m.backup.Cursor].Path
					m.backup.PreviewScroll = 0 // Reset scroll position
					m.currentView = ViewBackupPreview
				}
				return m, nil
			}

		case "R":
			// Restore backup (from backup manager or preview)
			if (m.currentView == ViewBackupManager || m.currentView == ViewBackupPreview) && !m.loading {
				backups, err := caddy.ListBackups(m.config.Caddy.CaddyfilePath)
				if err == nil && m.backup.Cursor < len(backups) {
					m.backup.PreviewPath = backups[m.backup.Cursor].Path
					m.backup.RestoreScopeCursor = 0    // Reset cursor
					m.backup.RestoreScope = RestoreAll // Default to restore all
					m.currentView = ViewRestoreScope
				}
				return m, nil
			}

		case "right":
			// Navigate to next help page
			if m.currentView == ViewHelp && m.helpPage < 4 {
				m.helpPage++
				return m, nil
			}
			// Navigate to next backup in preview
			if m.currentView == ViewBackupPreview && !m.loading {
				backups, err := caddy.ListBackups(m.config.Caddy.CaddyfilePath)
				if err == nil && m.backup.Cursor < len(backups)-1 {
					m.backup.Cursor++
					m.backup.PreviewPath = backups[m.backup.Cursor].Path
					m.backup.PreviewScroll = 0 // Reset scroll position
				}
				return m, nil
			}

		case "left":
			// Navigate to previous help page
			if m.currentView == ViewHelp && m.helpPage > 0 {
				m.helpPage--
				return m, nil
			}
			// Navigate to previous backup in preview
			if m.currentView == ViewBackupPreview && !m.loading {
				backups, err := caddy.ListBackups(m.config.Caddy.CaddyfilePath)
				if err == nil && m.backup.Cursor > 0 {
					m.backup.Cursor--
					m.backup.PreviewPath = backups[m.backup.Cursor].Path
					m.backup.PreviewScroll = 0 // Reset scroll position
				}
				return m, nil
			}

		case "x":
			// Delete backup (from backup manager only)
			if m.currentView == ViewBackupManager && !m.loading {
				backups, err := caddy.ListBackups(m.config.Caddy.CaddyfilePath)
				if err == nil && m.backup.Cursor < len(backups) {
					m.loading = true
					return m, deleteBackupCmd(backups[m.backup.Cursor].Path)
				}
				return m, nil
			}

		case "c":
			// Cleanup old backups (from backup manager only)
			if m.currentView == ViewBackupManager && !m.loading {
				m.currentView = ViewConfirmCleanup
				return m, nil
			}

		case "X":
			return m.handleBatchDeleteSelected()

		case "S":
			return m.handleBatchSyncSelected()

		case "enter":
			// Edit selected entry (from list view)
			if m.currentView == ViewList && !m.searching && !m.loading {
				// Skip if snippets panel is focused (no action needed)
				if m.panelFocus == PanelFocusSnippets {
					return m, nil
				}

				// Edit DNS entry
				filteredEntries := m.getFilteredEntries()
				if m.cursor < len(filteredEntries) {
					entry := filteredEntries[m.cursor]

					// Pre-populate form with existing entry values
					// For multi-domain entries, populate all domains as newline-separated
					subdomain := ""
					if entry.Caddy != nil && len(entry.Caddy.Domains) > 0 {
						// Multi-domain entry - format all domains as newline-separated subdomains
						subdomain = GetSubdomainsTextareaValue(entry.Caddy.Domains, m.config.Domain)
					} else if len(entry.Domain) > len(m.config.Domain)+1 {
						// Single domain (backwards compatibility)
						subdomain = entry.Domain[:len(entry.Domain)-len(m.config.Domain)-1]
					}

					dnsType := "CNAME"
					dnsTarget := m.config.Defaults.CNAMETarget
					proxied := m.config.Defaults.Proxied
					if entry.DNS != nil {
						dnsType = entry.DNS.Type
						dnsTarget = entry.DNS.Content
						proxied = entry.DNS.Proxied
					}

					dnsOnly := entry.Caddy == nil // If no Caddy config, it's DNS-only

					reverseProxyTarget := "localhost"
					servicePort := fmt.Sprintf("%d", m.config.Defaults.Port)
					ssl := m.config.Defaults.SSL
					lanOnly := false
					oauth := false
					webSocket := false
					if entry.Caddy != nil {
						reverseProxyTarget = entry.Caddy.Target
						servicePort = fmt.Sprintf("%d", entry.Caddy.Port)
						ssl = entry.Caddy.SSL
						lanOnly = entry.Caddy.IPRestricted
						oauth = entry.Caddy.OAuthHeaders
						webSocket = entry.Caddy.WebSocket
					}

					// Pre-populate selected snippets from entry's imports
					selectedSnippets := make(map[string]bool)
					if entry.Caddy != nil {
						for _, importName := range entry.Caddy.Imports {
							selectedSnippets[importName] = true
						}
					}

					m.addForm = AddFormData{
						Subdomain:          subdomain,
						DNSType:            dnsType,
						DNSTarget:          dnsTarget,
						DNSOnly:            dnsOnly,
						ReverseProxyTarget: reverseProxyTarget,
						ServicePort:        servicePort,
						Proxied:            proxied,
						LANOnly:            lanOnly,
						SSL:                ssl,
						OAuth:              oauth,
						WebSocket:          webSocket,
						SelectedSnippets:   selectedSnippets,
						FocusedField:       0,
					}
					m.editingEntry = &entry
					m.currentView = ViewEdit
					return m, nil
				}
			}

			// If in snippet detail view mode, enter edit mode
			if m.currentView == ViewSnippetDetail && !m.snippetPanel.Editing {
				if m.snippetPanel.Cursor < len(m.snippets) {
					// Initialize textarea with snippet content
					ta := textarea.New()
					ta.SetWidth(70)
					ta.SetHeight(15)
					ta.SetValue(m.snippets[m.snippetPanel.Cursor].Content)
					ta.Focus()
					m.snippetPanel.EditTextarea = ta
					m.snippetPanel.Editing = true
					m.snippetPanel.EditingIndex = m.snippetPanel.Cursor
					return m, nil
				}
			}


		// Dismiss error modal
		if m.currentView == ViewError {
			m.currentView = m.previousView
			m.err = nil
			return m, nil
		}
		// Save profile edit
		if m.currentView == ViewProfileEdit {
			return m.handleProfileEditKeyPress("enter")
		}
		// Handle Enter in wizard
		if m.currentView == ViewWizard {
			return m.handleWizardKeyPress("enter")
		}
		// Handle Enter in migration wizard
		if m.currentView == ViewMigrationWizard {
			return m.handleMigrationWizardKeyPress("enter")
		}
		// Handle Enter in profile selector
		if m.currentView == ViewProfileSelector {
			return m.handleProfileSelectorKeyPress("enter")
		}
		// Handle Enter in snippet wizard
		if m.currentView == ViewSnippetWizard {
			switch m.snippetWizardStep {
			case snippet_wizard.StepWelcome:
				// Set mode based on cursor selection and advance
				switch m.wizardCursor {
				case 0: // Templated - Advanced
					m.snippetWizardData.Mode = snippet_wizard.ModeTemplated
					m.snippetWizardStep = snippet_wizard.StepTemplateSelection
				case 1: // Custom - Paste Your Own
					m.snippetWizardData.Mode = snippet_wizard.ModeCustom
					m.snippetWizardStep = snippet_wizard.StepCustomSnippet
					// Initialize text input components for custom mode
					m.snippetWizardData.CustomNameInput = textinput.New()
					m.snippetWizardData.CustomNameInput.Placeholder = "snippet_name"
					m.snippetWizardData.CustomNameInput.Focus()
					m.snippetWizardData.CustomContentInput = textarea.New()
					m.snippetWizardData.CustomContentInput.Placeholder = "Paste your Caddy snippet here..."
				case 2: // Guided - Step by Step
					m.snippetWizardData.Mode = snippet_wizard.ModeGuided
					// Guided mode goes through legacy flow
					m.snippetWizardStep = snippet_wizard.StepIPRestriction
				}
				m.wizardCursor = 0
				return m, nil
			case snippet_wizard.StepCustomSnippet:
				// Validate custom snippet has name and content
				if m.snippetWizardData.CustomSnippetName == "" {
					m.err = fmt.Errorf("snippet name is required")
					return m, nil
				}
				if m.snippetWizardData.CustomSnippetContent == "" {
					m.err = fmt.Errorf("snippet content is required")
					return m, nil
				}
				// Set mode to custom and advance to summary
				m.snippetWizardData.Mode = snippet_wizard.ModeCustom
				m.snippetWizardStep = snippet_wizard.StepSummary
				m.wizardCursor = 0
				return m, nil
			case snippet_wizard.StepTemplateSelection:
				// Check if any templates selected
				hasSelection := false
				for _, selected := range m.snippetWizardData.SelectedTemplates {
					if selected {
						hasSelection = true
						break
					}
				}
				if !hasSelection {
					m.err = fmt.Errorf("select at least one template")
					return m, nil
				}
				// Advance to template params or summary
				m.snippetWizardData.Mode = snippet_wizard.ModeTemplated
				m.snippetWizardStep = snippet_wizard.StepTemplateParams
				m.wizardCursor = 0
				return m, nil
			case snippet_wizard.StepTemplateParams:
				// Advance to summary
				m.snippetWizardStep = snippet_wizard.StepSummary
				m.wizardCursor = 0
				return m, nil
			case snippet_wizard.StepIPRestriction:
				// Guided mode: Advance to security headers
				m.snippetWizardStep = snippet_wizard.StepSecurityHeaders
				m.wizardCursor = 0
				// Set default security preset if not set
				if m.snippetWizardData.SecurityPreset == "" {
					m.snippetWizardData.SecurityPreset = "basic"
				}
				return m, nil
			case snippet_wizard.StepSecurityHeaders:
				// Guided mode: Advance to performance
				m.snippetWizardStep = snippet_wizard.StepPerformance
				m.wizardCursor = 0
				return m, nil
			case snippet_wizard.StepPerformance:
				// Guided mode: Advance to summary
				m.snippetWizardStep = snippet_wizard.StepSummary
				m.wizardCursor = 0
				return m, nil
			case snippet_wizard.StepSummary:
				// Create the snippets
				return m.createSnippetsFromWizard()
			}
			return m, nil
		}
		// Handle Enter in delete scope selection: proceed to confirm
		if m.currentView == ViewDeleteScope && !m.loading {
			m.currentView = ViewConfirmDelete
			return m, nil
		}
		// Handle Enter in restore scope selection: proceed to confirm
		if m.currentView == ViewRestoreScope && !m.loading {
			m.currentView = ViewConfirmRestore
			return m, nil
		}
		// In bulk delete menu: select option
		if m.currentView == ViewBulkDeleteMenu {
			// Collect entries based on selection
			if m.bulkDelete.MenuCursor == 0 {
				// Delete all orphaned DNS
				m.bulkDelete.Type = "dns"
				m.bulkDelete.Entries = []diff.SyncedEntry{}
				for _, entry := range m.entries {
					if entry.Status == diff.StatusOrphanedDNS {
						m.bulkDelete.Entries = append(m.bulkDelete.Entries, entry)
					}
				}
			} else if m.bulkDelete.MenuCursor == 1 {
				// Delete all orphaned Caddy
				m.bulkDelete.Type = "caddy"
				m.bulkDelete.Entries = []diff.SyncedEntry{}
				for _, entry := range m.entries {
					if entry.Status == diff.StatusOrphanedCaddy {
						m.bulkDelete.Entries = append(m.bulkDelete.Entries, entry)
					}
				}
			}

			// Only proceed if there are entries to delete
			if len(m.bulkDelete.Entries) > 0 {
				m.currentView = ViewConfirmBulkDelete
			} else {
				// No entries to delete, show error and return to list
				m.err = fmt.Errorf("No %s entries to delete", m.bulkDelete.Type)
				m.currentView = ViewList
			}
			return m, nil
		}

		// In add/edit form: validate and go to preview
		if m.currentView == ViewAdd || m.currentView == ViewEdit {
			// Validate required fields
			if m.addForm.Subdomain == "" {
				m.err = fmt.Errorf("Subdomain is required")
				return m, nil
			}
			if m.addForm.DNSTarget == "" {
				m.err = fmt.Errorf("DNS Target is required")
				return m, nil
			}

			// Validate A record IP address format
			if m.addForm.DNSType == "A" {
				if !isValidIPAddress(m.addForm.DNSTarget) {
					m.err = fmt.Errorf("Invalid IP address format for A record")
					return m, nil
				}
			}

			// Validate Caddy fields if not DNS-only
			if !m.addForm.DNSOnly {
				if m.addForm.ReverseProxyTarget == "" {
					m.err = fmt.Errorf("Reverse Proxy Target is required (or enable DNS Only)")
					return m, nil
				}
			}

			// Clear any previous errors and go to preview
			m.err = nil
			m.currentView = ViewPreview
			return m, nil
		}

		// In search mode, enter accepts the search
		if m.searching {
			m.searching = false
			return m, nil
		}
		// Confirm restore backup (only in confirm restore screen)
		if m.currentView == ViewConfirmRestore && !m.loading {
			m.loading = true
			apiToken, err := m.config.GetAPIToken()
			if err != nil {
				m.err = fmt.Errorf("failed to get API token: %w", err)
				return m, nil
			}
			return m, restoreBackupCmd(m.config, m.backup.PreviewPath, m.backup.RestoreScope, apiToken)
		}
		// Confirm cleanup old backups (only in confirm cleanup screen)
		if m.currentView == ViewConfirmCleanup && !m.loading {
			m.loading = true
			return m, cleanupBackupsCmd(m.config.Caddy.CaddyfilePath, m.backup.RetentionDays)
		}
		// Preview backup with Enter key (from backup manager)
		if m.currentView == ViewBackupManager && !m.loading {
			backups, err := caddy.ListBackups(m.config.Caddy.CaddyfilePath)
			if err == nil && m.backup.Cursor < len(backups) {
				m.backup.PreviewPath = backups[m.backup.Cursor].Path
				m.backup.PreviewScroll = 0 // Reset scroll position
				m.currentView = ViewBackupPreview
			}
			return m, nil
		}
		return m, nil

		case "s":
			return m.handleSyncEntry()

		case "r":
			// If showing error modal, retry by clearing error and returning to previous view
			if m.currentView == ViewError {
				m.currentView = m.previousView
				m.err = nil
				return m, nil
			}
			// Refresh data (only from list view, not while searching or loading)
			if m.currentView == ViewList && !m.searching && !m.loading {
				m.loading = true
				return m, refreshDataCmd(m.config)
			}

		case "m":
			return m.handleOpenMigrationWizard()

		case "?", "h", "ctrl+h":
			return m.handleOpenHelp()

		case "1", "2", "3", "4", "5":
			return m.handleHelpPageJump(msg.String())

		case "q":
			return m.handleQuit()

		case "y":
			return m.handleConfirmAction()

		case "+":
			// Handle '+' (new profile) in profile selector
			if m.currentView == ViewProfileSelector {
				return m.handleProfileSelectorKeyPress("+")
			}

		case "n":
			return m.handleCancelConfirmation()

		case " ":
			// In profile edit: toggle boolean fields
			if m.currentView == ViewProfileEdit {
				return m.handleProfileEditKeyPress(" ")
			}
			// In snippet wizard: space toggles checkboxes
			if m.currentView == ViewSnippetWizard {
				switch m.snippetWizardStep {
				case snippet_wizard.StepTemplateSelection:
					// Toggle template selection at cursor
					// Template order matches template_selection.go view
					templateKeys := []string{
						// Security
						"cors_headers", "rate_limiting", "auth_headers", "ip_restricted", "security_headers",
						// Performance
						"static_caching", "compression_advanced", "performance",
						// Backend Integration
						"websocket_advanced", "extended_timeouts", "https_backend",
						// Content Control
						"large_uploads", "custom_headers_inject", "frame_embedding", "rewrite_rules",
					}
					if m.wizardCursor >= 0 && m.wizardCursor < len(templateKeys) {
						key := templateKeys[m.wizardCursor]
						m.snippetWizardData.SelectedTemplates[key] = !m.snippetWizardData.SelectedTemplates[key]
					}
					return m, nil
				case SnippetWizardAutoDetect:
					// Toggle pattern selection at cursor
					if m.wizardCursor >= 0 && m.wizardCursor < len(m.snippetWizardData.DetectedPatterns) {
						pattern := m.snippetWizardData.DetectedPatterns[m.wizardCursor]
						m.snippetWizardData.SelectedPatterns[pattern.SuggestedName] = !m.snippetWizardData.SelectedPatterns[pattern.SuggestedName]
					}
					return m, nil
				case SnippetWizardIPRestriction:
					m.snippetWizardData.CreateIPRestriction = !m.snippetWizardData.CreateIPRestriction
					return m, nil
				case SnippetWizardSecurityHeaders:
					m.snippetWizardData.CreateSecurityHeaders = !m.snippetWizardData.CreateSecurityHeaders
					return m, nil
				case SnippetWizardPerformance:
					m.snippetWizardData.CreatePerformance = !m.snippetWizardData.CreatePerformance
					return m, nil
				case snippet_wizard.StepTemplateParams:
					// Toggle checkbox parameters in template config
					m.toggleSnippetTemplateParamCheckbox()
					return m, nil
				}
			}
			// In list view: space toggles selection
			if m.currentView == ViewList && !m.searching && !m.loading {
				filteredEntries := m.getFilteredEntries()
				if m.cursor < len(filteredEntries) {
					domain := filteredEntries[m.cursor].Domain
					// Toggle selection
					if m.selectedEntries[domain] {
						delete(m.selectedEntries, domain)
					} else {
						m.selectedEntries[domain] = true
					}
				}
				return m, nil
			}
			// In add/edit form: space toggles DNS type and checkboxes
			if m.currentView == ViewAdd || m.currentView == ViewEdit {
				switch m.addForm.FocusedField {
				case 1: // DNS Type toggle
					if m.addForm.DNSType == "CNAME" {
						m.addForm.DNSType = "A"
					} else {
						m.addForm.DNSType = "CNAME"
					}
				case 3: // DNS Only checkbox
					m.addForm.DNSOnly = !m.addForm.DNSOnly
				case 6: // Proxied checkbox
					m.addForm.Proxied = !m.addForm.Proxied
				case 7: // SSL checkbox
					m.addForm.SSL = !m.addForm.SSL
				default:
					// Fields 8+ are snippet checkboxes
					if m.addForm.FocusedField >= 8 {
						snippetIndex := m.addForm.FocusedField - 8
						if snippetIndex < len(m.snippets) {
							snippetName := m.snippets[snippetIndex].Name
							// Initialize map if nil
							if m.addForm.SelectedSnippets == nil {
								m.addForm.SelectedSnippets = make(map[string]bool)
							}
							// Toggle snippet selection
							m.addForm.SelectedSnippets[snippetName] = !m.addForm.SelectedSnippets[snippetName]
						}
					}
				}
				return m, nil
			}

		case "backspace":
			// Handle backspace in profile edit
			if m.currentView == ViewProfileEdit {
				return m.handleProfileEditKeyPress("backspace")
			}
			// Handle backspace in add/edit form
			if m.currentView == ViewAdd || m.currentView == ViewEdit {
				// Check if we're in the custom config field
				customConfigFieldIndex := 8 + len(m.snippets)
				if m.addForm.FocusedField == customConfigFieldIndex {
					// Custom Caddy Config field
					if len(m.addForm.CustomCaddyConfig) > 0 {
						m.addForm.CustomCaddyConfig = m.addForm.CustomCaddyConfig[:len(m.addForm.CustomCaddyConfig)-1]
					}
					return m, nil
				}

				// Handle other fields
				switch m.addForm.FocusedField {
				case 0: // Subdomain
					if len(m.addForm.Subdomain) > 0 {
						m.addForm.Subdomain = m.addForm.Subdomain[:len(m.addForm.Subdomain)-1]
					}
				case 2: // DNS Target
					if len(m.addForm.DNSTarget) > 0 {
						m.addForm.DNSTarget = m.addForm.DNSTarget[:len(m.addForm.DNSTarget)-1]
					}
				case 4: // Reverse Proxy Target
					if len(m.addForm.ReverseProxyTarget) > 0 {
						m.addForm.ReverseProxyTarget = m.addForm.ReverseProxyTarget[:len(m.addForm.ReverseProxyTarget)-1]
					}
				case 5: // Service Port
					if len(m.addForm.ServicePort) > 0 {
						m.addForm.ServicePort = m.addForm.ServicePort[:len(m.addForm.ServicePort)-1]
					}
				}
				return m, nil
			}

			// Handle backspace in search mode
			if m.searching && len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.cursor = 0
				m.scrollOffset = 0
				return m, nil
			}

			// Handle backspace in snippet wizard IP restriction step
			if m.currentView == ViewSnippetWizard && m.snippetWizardStep == SnippetWizardIPRestriction && m.snippetWizardData.CreateIPRestriction {
				if m.wizardCursor == 0 {
					// LAN Subnet field
					if len(m.snippetWizardData.LANSubnet) > 0 {
						m.snippetWizardData.LANSubnet = m.snippetWizardData.LANSubnet[:len(m.snippetWizardData.LANSubnet)-1]
					}
					return m, nil
				} else if m.wizardCursor == 1 {
					// External IP field
					if len(m.snippetWizardData.AllowedExternalIP) > 0 {
						m.snippetWizardData.AllowedExternalIP = m.snippetWizardData.AllowedExternalIP[:len(m.snippetWizardData.AllowedExternalIP)-1]
					}
					return m, nil
				}
			}

		// Navigation keys - j/k for dashboard only (single keys)
		case "j":
			return m.handleListDown()

		case "k":
			return m.handleListUp()

		case "down":
			return m.handleNavigateDown()

		case "up":
			return m.handleNavigateUp()

		case "g": // Go to top
			if m.currentView == ViewList && !m.searching {
				if m.panelFocus == PanelFocusSnippets {
					m.snippetPanel.Cursor = 0
					m.snippetPanel.ScrollOffset = 0
				} else {
					m.cursor = 0
					m.scrollOffset = 0
				}
			}
			// In backup preview: scroll to top
			if m.currentView == ViewBackupPreview && !m.loading {
				m.backup.PreviewScroll = 0
			}

		case "G": // Go to bottom
			if m.currentView == ViewList && !m.searching {
				if m.panelFocus == PanelFocusSnippets {
					if len(m.snippets) > 0 {
						m.snippetPanel.Cursor = len(m.snippets) - 1
						// Calculate scroll offset (if needed in future)
					}
				} else {
					filtered := m.getFilteredEntries()
					m.cursor = len(filtered) - 1
					if len(filtered) > m.height-5 {
						m.scrollOffset = len(filtered) - (m.height - 5)
					}
				}
			}
			// In backup preview: scroll to bottom
			if m.currentView == ViewBackupPreview && !m.loading {
				content, err := os.ReadFile(m.backup.PreviewPath)
				if err == nil {
					lines := strings.Split(string(content), "\n")
					visibleHeight := m.height - 12
					if visibleHeight < 5 {
						visibleHeight = 5
					}
					// Scroll to show the last page
					maxScroll := len(lines) - visibleHeight
					if maxScroll > 0 {
						m.backup.PreviewScroll = maxScroll
					} else {
						m.backup.PreviewScroll = 0
					}
				}
			}

		case "pgup":
			return m.handleBackupPageUp()

		case "pgdown":
			return m.handleBackupPageDown()

		case "home":
			return m.handleListHome()

		case "end":
			return m.handleListEnd()

		default:
			// No specific keybinding matched
		}

		// NOTE: Wizard text input is now handled at the TOP of this function
		// to prevent keybindings from intercepting characters like 'q', 'k', etc.


	return m, nil
}
