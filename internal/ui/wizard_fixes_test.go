package ui

import (
	"reflect"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// The wizard reuses one text input for every field. A prefilled field must open
// with the cursor at the end, whatever the previous field left behind.
func TestWizardPrefilledFieldOpensWithCursorAtEnd(t *testing.T) {
	m := NewModelWithWizard()
	m.wizardStep = WizardStepBasicInfo
	m.wizardData.CurrentField = FieldProfileName
	m.configureTextInputForField()
	m.wizardTextInput.SetValue("dummytoken") // previous field: cursor ends at 10

	m.wizardData.CurrentField = FieldDomain
	m.wizardData.Domain = "angelsomething.com"
	m.configureTextInputForField()

	if got, want := m.wizardTextInput.Position(), len("angelsomething.com"); got != want {
		t.Errorf("cursor at %d, want %d (end of %q)", got, want, m.wizardTextInput.Value())
	}
}

// "b: Go back and edit" and "n: Cancel" are shown on the review screen.
func TestWizardReviewScreenBackAndCancelKeys(t *testing.T) {
	review := func() Model {
		m := NewModelWithWizard()
		m.wizardStep = WizardStepSummary
		return m
	}

	m, _ := review().handleKeyMsg(tea.KeyPressMsg{Code: 'b', Text: "b"})
	if m.wizardStep == WizardStepSummary {
		t.Error("b did not leave the review screen")
	}
	if m.currentView != ViewWizard {
		t.Errorf("b left the wizard: view = %v", m.currentView)
	}

	n := review()
	n.profile.Available = []string{"existing"}
	n, _ = n.handleKeyMsg(tea.KeyPressMsg{Code: 'n', Text: "n"})
	if n.currentView != ViewProfileSelector {
		t.Errorf("n: view = %v, want the profile selector", n.currentView)
	}
}

// Editing an entry must not reorder its imports from run to run.
func TestSelectedSnippetNamesAreSorted(t *testing.T) {
	sel := map[string]bool{"performance": true, "log_access": true, "ip_restricted": true, "off": false, "security_headers": true}
	want := []string{"ip_restricted", "log_access", "performance", "security_headers"}
	for i := 0; i < 20; i++ {
		if got := getSelectedSnippetNames(sel); !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
