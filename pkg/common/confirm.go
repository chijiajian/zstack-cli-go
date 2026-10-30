package common

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

// Confirm honors explicit --yes and never authorizes an action after an I/O error.
func Confirm(cmd *cobra.Command, prompt string) (bool, error) {
	yes, _ := cmd.Flags().GetBool("yes")
	if yes {
		return true, nil
	}
	if _, err := fmt.Fprint(cmd.OutOrStdout(), prompt); err != nil {
		return false, fmt.Errorf("write confirmation: %w", err)
	}
	input, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("read confirmation: %w", err)
	}
	answer := strings.ToLower(strings.TrimSpace(input))
	return answer == "yes" || answer == "y", nil
}
