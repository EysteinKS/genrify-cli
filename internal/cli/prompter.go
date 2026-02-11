package cli

import (
	"fmt"
	"strconv"

	"github.com/manifoldco/promptui"
)

// promptuiPrompter wraps promptui for production interactive mode.
type promptuiPrompter struct{}

// NewPrompter returns a Prompter implementation using promptui.
func NewPrompter() Prompter {
	return &promptuiPrompter{}
}

func (p *promptuiPrompter) PromptString(label, defaultValue string) (string, error) {
	prompt := promptui.Prompt{
		Label:   label,
		Default: defaultValue,
	}
	result, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrCancelled, err)
	}
	return result, nil
}

func (p *promptuiPrompter) PromptInt(label string, defaultValue int) (int, error) {
	prompt := promptui.Prompt{
		Label:   label,
		Default: strconv.Itoa(defaultValue),
		Validate: func(s string) error {
			if _, err := strconv.Atoi(s); err != nil {
				return fmt.Errorf("must be a number")
			}
			return nil
		},
	}
	result, err := prompt.Run()
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrCancelled, err)
	}
	val, _ := strconv.Atoi(result)
	return val, nil
}

func (p *promptuiPrompter) PromptSelect(label string, items []string) (int, string, error) {
	prompt := promptui.Select{
		Label: label,
		Items: items,
	}
	idx, choice, err := prompt.Run()
	if err != nil {
		return 0, "", fmt.Errorf("%w: %v", ErrCancelled, err)
	}
	return idx, choice, nil
}
