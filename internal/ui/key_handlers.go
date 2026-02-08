package ui

import (
	tea "github.com/charmbracelet/bubbletea"

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
			return m.handleTabKey()

		case "shift+tab":
			return m.handleShiftTabKey()

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
			return m.handleEditProfile()

		case "w", "ctrl+s":
			return m.handleOpenSnippetWizard()

		case "d":
			return m.handleDeleteAction()

		case "D":
			return m.handleOpenBulkDeleteMenu()

		case "b", "ctrl+b":
			return m.handleOpenBackupManager()

		case "l", "L":
			return m.handleOpenAuditLog()

		case "p", "ctrl+p":
			return m.handleProfileOrPreview()

		case "R":
			return m.handleRestoreBackup()

		case "right":
			return m.handleNavigateRight()

		case "left":
			return m.handleNavigateLeft()

		case "x":
			return m.handleDeleteBackup()

		case "c":
			return m.handleCleanupBackups()

		case "X":
			return m.handleBatchDeleteSelected()

		case "S":
			return m.handleBatchSyncSelected()

		case "enter":
			return m.handleEnterKey()

		case "s":
			return m.handleSyncEntry()

		case "r":
			return m.handleRefreshData()

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
			return m.handleAddProfile()

		case "n":
			return m.handleCancelConfirmation()

		case " ":
			return m.handleSpaceKey()

		case "backspace":
			return m.handleBackspaceKey()

		// Navigation keys - j/k for dashboard only (single keys)
		case "j":
			return m.handleListDown()

		case "k":
			return m.handleListUp()

		case "down":
			return m.handleNavigateDown()

		case "up":
			return m.handleNavigateUp()

		case "g":
			return m.handleGoToTop()

		case "G":
			return m.handleGoToBottom()

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
