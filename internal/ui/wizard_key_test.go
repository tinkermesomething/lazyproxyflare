package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestWizardEnterKeyOnWelcome tests that Enter key advances from Welcome step
func TestWizardEnterKeyOnWelcome(t *testing.T) {
	// Create model in wizard mode
	m := NewModelWithWizard()

	// Verify we're on welcome step
	if m.currentView != ViewWizard {
		t.Fatalf("Expected ViewWizard (%d), got %d", ViewWizard, m.currentView)
	}
	if m.wizardStep != WizardStepWelcome {
		t.Fatalf("Expected WizardStepWelcome (%d), got %d", WizardStepWelcome, m.wizardStep)
	}

	// Simulate Enter key press
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.handleKeyMsg(enterKey)
	m = newModel

	// Verify we advanced to BasicInfo step
	if m.wizardStep != WizardStepBasicInfo {
		t.Errorf("Expected WizardStepBasicInfo (%d) after Enter, got %d", WizardStepBasicInfo, m.wizardStep)
	}
}

// TestWizardTabKeyOnWelcome tests that Tab key advances from Welcome step
func TestWizardTabKeyOnWelcome(t *testing.T) {
	// Create model in wizard mode
	m := NewModelWithWizard()

	// Verify we're on welcome step
	if m.currentView != ViewWizard {
		t.Fatalf("Expected ViewWizard (%d), got %d", ViewWizard, m.currentView)
	}
	if m.wizardStep != WizardStepWelcome {
		t.Fatalf("Expected WizardStepWelcome (%d), got %d", WizardStepWelcome, m.wizardStep)
	}

	// Simulate Tab key press
	tabKey := tea.KeyMsg{Type: tea.KeyTab}
	newModel, _ := m.handleKeyMsg(tabKey)
	m = newModel

	// Verify we advanced to BasicInfo step
	if m.wizardStep != WizardStepBasicInfo {
		t.Errorf("Expected WizardStepBasicInfo (%d) after Tab, got %d", WizardStepBasicInfo, m.wizardStep)
	}
}

// TestIsWizardTextInputStep tests that Welcome step returns false
func TestIsWizardTextInputStep(t *testing.T) {
	m := NewModelWithWizard()

	// Welcome step should NOT be a text input step
	if m.isWizardTextInputStep() {
		t.Error("WizardStepWelcome should NOT be a text input step")
	}

	// Advance to BasicInfo
	m.wizardStep = WizardStepBasicInfo
	if !m.isWizardTextInputStep() {
		t.Error("WizardStepBasicInfo SHOULD be a text input step")
	}
}

// TestWizardStartFromProfileSelector tests wizard launched from profile selector
func TestWizardStartFromProfileSelector(t *testing.T) {
	// Create model in profile selector mode
	m := NewModelWithProfileSelector([]string{"existing-profile"}, "")

	// Move cursor to "Add New" option
	m.cursor = 1 // Past the existing profiles

	// Verify we're in profile selector
	if m.currentView != ViewProfileSelector {
		t.Fatalf("Expected ViewProfileSelector (%d), got %d", ViewProfileSelector, m.currentView)
	}

	// Simulate Enter to launch wizard
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.handleKeyMsg(enterKey)
	m = newModel

	// Verify we're now in wizard view on welcome step
	if m.currentView != ViewWizard {
		t.Errorf("Expected ViewWizard (%d), got %d", ViewWizard, m.currentView)
	}
	if m.wizardStep != WizardStepWelcome {
		t.Errorf("Expected WizardStepWelcome (%d), got %d", WizardStepWelcome, m.wizardStep)
	}

	// Now simulate Enter on welcome screen
	newModel, _ = m.handleKeyMsg(enterKey)
	m = newModel

	// Verify we advanced to BasicInfo
	if m.wizardStep != WizardStepBasicInfo {
		t.Errorf("Expected WizardStepBasicInfo (%d) after Enter on welcome, got %d", WizardStepBasicInfo, m.wizardStep)
	}
}

// TestWizardEnterKeyFullPath tests Enter key through the full Update() path
func TestWizardEnterKeyFullPath(t *testing.T) {
	// Create model in wizard mode (same as runtime)
	m := NewModelWithWizard()
	
	// Verify initial state matches what we expect at runtime
	if m.currentView != ViewWizard {
		t.Fatalf("Expected ViewWizard, got %d", m.currentView)
	}
	if m.wizardStep != WizardStepWelcome {
		t.Fatalf("Expected WizardStepWelcome, got %d", m.wizardStep)
	}
	
	// Call Init() like bubbletea does
	m.Init()
	
	// Verify isWizardTextInputStep returns false for WelcomeStep
	if m.isWizardTextInputStep() {
		t.Fatal("WizardStepWelcome should NOT be a text input step")
	}
	
	// Simulate Enter key through Update() (the real path)
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.Update(enterKey)
	newModel := result.(Model)
	
	// Verify we advanced
	if newModel.wizardStep == WizardStepWelcome {
		t.Errorf("Expected to advance from WizardStepWelcome, but still on welcome step")
		t.Logf("currentView: %d", newModel.currentView)
		t.Logf("wizardStep: %d", newModel.wizardStep)
	}
}
