package config

import (
	"fmt"
	"github.com/chijiajian/zstack-cli-go/pkg/clierr"
	"strings"
	"sync"
)

// Selection is process-local; the persisted default is never overwritten by it.
var selection struct {
	sync.RWMutex
	name     string
	identity *Context
}

func SetContextOverride(name string) {
	selection.Lock()
	defer selection.Unlock()
	selection.name = name
	selection.identity = nil
}
func SelectedContextName(cfg *ZStackConfig) string {
	selection.RLock()
	defer selection.RUnlock()
	if selection.name != "" {
		return selection.name
	}
	return cfg.CurrentContext
}
func ValidateContextName(name string) error {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(name) != name || strings.ContainsAny(name, "\r\n\t") {
		return fmt.Errorf("context name must be nonempty without surrounding whitespace or control characters")
	}
	return nil
}
func UseContext(cfg *ZStackConfig, name string) error {
	if _, ok := cfg.Contexts[name]; !ok {
		return fmt.Errorf("context %q not found", name)
	}
	cfg.CurrentContext = name
	return nil
}
func RenameContext(cfg *ZStackConfig, oldName, newName string) error {
	if err := ValidateContextName(newName); err != nil {
		return err
	}
	value, ok := cfg.Contexts[oldName]
	if !ok {
		return fmt.Errorf("context %q not found", oldName)
	}
	if _, ok := cfg.Contexts[newName]; ok {
		return clierr.Wrap(clierr.Conflict, fmt.Errorf("context %q already exists", newName))
	}
	cfg.Contexts[newName] = value
	delete(cfg.Contexts, oldName)
	if cfg.CurrentContext == oldName {
		cfg.CurrentContext = newName
	}
	return nil
}
func DeleteContext(cfg *ZStackConfig, name string) error {
	if _, ok := cfg.Contexts[name]; !ok {
		return fmt.Errorf("context %q not found", name)
	}
	if cfg.CurrentContext == name {
		return fmt.Errorf("cannot delete current context %q; use-context another context first", name)
	}
	delete(cfg.Contexts, name)
	return nil
}
func StoreContext(cfg *ZStackConfig, name string, value Context, activate bool) error {
	if err := ValidateContextName(name); err != nil {
		return err
	}
	if cfg.Contexts == nil {
		cfg.Contexts = map[string]Context{}
	}
	cfg.Contexts[name] = value
	if activate {
		cfg.CurrentContext = name
	}
	return nil
}

// PinContextSelection fixes the environment for all stages of a single command.
// Session tokens may rotate; endpoint/account credential changes require a retry.
func PinContextSelection(cfg *ZStackConfig) {
	selection.Lock()
	defer selection.Unlock()
	if selection.name == "" {
		selection.name = cfg.CurrentContext
	}
	if value, ok := cfg.Contexts[selection.name]; ok {
		selection.identity = &value
	}
}
func validatePinnedContext(name string, value Context) error {
	selection.RLock()
	defer selection.RUnlock()
	if selection.name != name || selection.identity == nil {
		return nil
	}
	expected := *selection.identity
	expected.SessionUUID = ""
	value.SessionUUID = ""
	if expected != value {
		return clierr.Wrap(clierr.Conflict, fmt.Errorf("context %q changed during command; retry with the updated configuration", name))
	}
	return nil
}
