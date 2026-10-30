package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// UpdateConfig serializes read-modify-write transactions across CLI processes.
// The callback must not perform network calls or re-enter configuration writes.
func UpdateConfig(ctx context.Context, mutate func(*ZStackConfig) error) error {
	return withConfigLock(ctx, func(file string) error {
		cfg, err := loadConfigFile(file)
		if err != nil {
			return err
		}
		if err = mutate(cfg); err != nil {
			return err
		}
		if err = ctx.Err(); err != nil {
			return err
		}
		return saveConfigFile(file, cfg)
	})
}

func withConfigLock(ctx context.Context, action func(string) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	file, err := filepath.Abs(getDefaultConfigFile())
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		return err
	}
	// Resolve existing aliases so symlink paths cooperate on the same lock inode.
	if resolved, e := filepath.EvalSymlinks(file); e == nil {
		file = resolved
	} else if !os.IsNotExist(e) {
		return e
	} else {
		// A dangling leaf symlink is not an absent config file. Replacing it
		// would split aliases across different configs and lock files.
		if info, err := os.Lstat(file); err == nil && info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("config path is a dangling symlink: %s", file)
		}
		parent, e := filepath.EvalSymlinks(filepath.Dir(file))
		if e != nil {
			return e
		}
		file = filepath.Join(parent, filepath.Base(file))
	}
	lock, err := os.OpenFile(file+".lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("open config lock: %w", err)
	}
	defer lock.Close() // OS releases the lock even when the process exits unexpectedly.
	waitCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	for {
		if err = waitCtx.Err(); err != nil {
			return fmt.Errorf("wait for config lock: %w", err)
		}
		acquired, err := tryConfigLock(lock)
		if err != nil {
			return fmt.Errorf("lock config: %w", err)
		}
		if acquired {
			defer unlockConfig(lock)
			if err := waitCtx.Err(); err != nil {
				return fmt.Errorf("wait for config lock: %w", err)
			}
			return action(file)
		}
		timer := time.NewTimer(10 * time.Millisecond)
		select {
		case <-waitCtx.Done():
			timer.Stop()
		case <-timer.C:
		}
	}
	// Never unlink the sidecar: replacing it would allow two independent locks.
}
