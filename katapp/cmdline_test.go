package katapp

import (
	"bytes"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"testing"
)

func TestCmdlineExecute(t *testing.T) {
	handlerCalled := false

	handler := &CmdlineHandler{
		Run: func(deployment string) {
			if deployment != "test-deployment" {
				t.Errorf("Expected deployment 'test-deployment', got '%s'", deployment)
			}
			handlerCalled = true
		},
	}

	// Mock arguments
	os.Args = []string{"cmd", "run", "--deployment=test-deployment"}

	// Capture output
	output := &bytes.Buffer{}
	CmdlineExecute("test", "short desc", "long desc", handler)

	if !handlerCalled {
		t.Errorf("Expected handler to be called")
	}

	if output.String() != "" {
		t.Errorf("Unexpected output: %s", output.String())
	}
}

func TestNewCobraCmdlineCommand(t *testing.T) {
	cmd := newCobraCmdlineCommand("test", "short desc", "long desc")
	if cmd.Use != "test" {
		t.Errorf("Expected Use to be 'test', got '%s'", cmd.Use)
	}
	if cmd.Short != "short desc" {
		t.Errorf("Expected Short to be 'short desc', got '%s'", cmd.Short)
	}
	if cmd.Long != "long desc" {
		t.Errorf("Expected Long to be 'long desc', got '%s'", cmd.Long)
	}
}

func TestBindViperToCobraCommands(t *testing.T) {
	viper.Set("test-flag", "test-value")
	cmd := &cobra.Command{
		Use: "test",
	}
	cmd.Flags().String("test-flag", "", "test flag description")
	bindViperToCobraCommands([]*cobra.Command{cmd})

	flagValue, _ := cmd.Flags().GetString("test-flag")
	if flagValue != "test-value" {
		t.Errorf("Expected flag value 'test-value', got '%s'", flagValue)
	}
}

func TestCopyViperToCobraFlags(t *testing.T) {
	viper.Set("test-flag", "test-value")
	cmd := &cobra.Command{
		Use: "test",
	}
	cmd.Flags().String("test-flag", "", "test flag description")
	copyViperToCobraFlags(cmd)

	flagValue, _ := cmd.Flags().GetString("test-flag")
	if flagValue != "test-value" {
		t.Errorf("Expected flag value 'test-value', got '%s'", flagValue)
	}
}
