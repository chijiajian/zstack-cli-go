package config

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigLockProcessHelper(t *testing.T) {
	if os.Getenv("ZSTACK_LOCK_HELPER") != "1" {
		return
	}
	err := UpdateConfig(context.Background(), func(*ZStackConfig) error {
		fmt.Println("locked")
		select {}
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestConfigLockCancellationAndProcessExit(t *testing.T) {
	t.Setenv("ZSTACK_CONFIG", filepath.Join(t.TempDir(), "config.yaml"))
	cmd := exec.Command(os.Args[0], "-test.run=^TestConfigLockProcessHelper$")
	cmd.Env = append(os.Environ(), "ZSTACK_LOCK_HELPER=1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()
	ready := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		if scanner.Scan() {
			ready <- scanner.Text()
		} else {
			ready <- ""
		}
	}()
	select {
	case line := <-ready:
		if line != "locked" {
			t.Fatalf("helper: %s", line)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("lock helper did not start")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	called := false
	err = UpdateConfig(ctx, func(*ZStackConfig) error { called = true; return nil })
	if !errors.Is(err, context.DeadlineExceeded) || called {
		t.Fatalf("lock ignored deadline: %v called=%v", err, called)
	}
	if err = cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = cmd.Wait()
	if err = UpdateConfig(context.Background(), func(c *ZStackConfig) error { return StoreContext(c, "a", Context{Endpoint: "a"}, true) }); err != nil {
		t.Fatal(err)
	}
	sentinel := errors.New("abort")
	if err = UpdateConfig(context.Background(), func(c *ZStackConfig) error { c.CurrentContext = "bad"; return sentinel }); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	saved, err := LoadConfig()
	if err != nil || saved.CurrentContext != "a" {
		t.Fatalf("failed transaction committed: %+v %v", saved, err)
	}
}

func TestCanceledTransactionHasNoFilesystemSideEffects(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "not-created")
	t.Setenv("ZSTACK_CONFIG", filepath.Join(directory, "config.yaml"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	err := UpdateConfig(ctx, func(*ZStackConfig) error { called = true; return nil })
	if !errors.Is(err, context.Canceled) || called {
		t.Fatalf("err=%v called=%v", err, called)
	}
	if _, err = os.Stat(directory); !os.IsNotExist(err) {
		t.Fatalf("canceled operation created directory: %v", err)
	}
}

func TestConfigSymlinkLockIdentity(t *testing.T) {
	directory := t.TempDir()
	target := filepath.Join(directory, "config.yaml")
	alias := filepath.Join(directory, "alias.yaml")
	t.Setenv("ZSTACK_CONFIG", target)
	if err := SaveConfig(&ZStackConfig{Contexts: map[string]Context{}}); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, alias); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	// Hold the real path's lock, then attempt a transaction through its alias.
	lock, err := os.OpenFile(target+".lock", os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	held, err := tryConfigLock(lock)
	if err != nil || !held {
		t.Fatalf("%v %v", held, err)
	}
	defer unlockConfig(lock)
	t.Setenv("ZSTACK_CONFIG", alias)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err = UpdateConfig(ctx, func(*ZStackConfig) error { t.Error("alias bypassed lock"); return nil })
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	unlockConfig(lock)
	if err = UpdateConfig(context.Background(), func(cfg *ZStackConfig) error { return StoreContext(cfg, "a", Context{Endpoint: "a"}, true) }); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Lstat(alias); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("transaction replaced symlink")
	}
	saved, err := loadConfigFile(target)
	if err != nil || saved.CurrentContext != "a" {
		t.Fatalf("target not updated: %v", err)
	}
}

func TestDanglingConfigSymlinkIsNotReplaced(t *testing.T) {
	directory := t.TempDir()
	alias := filepath.Join(directory, "alias.yaml")
	target := filepath.Join(directory, "missing.yaml")
	if err := os.Symlink(target, alias); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	t.Setenv("ZSTACK_CONFIG", alias)
	called := false
	err := UpdateConfig(context.Background(), func(*ZStackConfig) error { called = true; return nil })
	if err == nil || called {
		t.Fatalf("err=%v called=%v", err, called)
	}
	if info, err := os.Lstat(alias); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("dangling link replaced")
	}
	if _, err := os.Stat(alias + ".lock"); !os.IsNotExist(err) {
		t.Fatal("created separate alias lock")
	}
}
