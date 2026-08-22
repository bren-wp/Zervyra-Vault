package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func AcquireLock(vaultPath string) (*VaultLock, error) {
	lockPath := filepath.Clean(vaultPath) + ".lock"
	if err := os.MkdirAll(filepath.Dir(lockPath), 0700); err != nil {
		return nil, err
	}
	if st, err := os.Stat(lockPath); err == nil {
		pid := readLockPID(lockPath)
		if pid > 0 && processAlive(pid) {
			return nil, errors.New("vault is already open in another Zervyra Vault instance")
		}
		// A dead PID is a definitive stale lock. A malformed lock is removed only
		// after the heartbeat timeout to avoid racing a just-starting instance.
		if pid > 0 || time.Since(st.ModTime()) > LockStaleAfter {
			if err := os.Remove(lockPath); err != nil && !os.IsNotExist(err) {
				return nil, err
			}
		} else {
			return nil, errors.New("vault lock is active or too new to recover safely")
		}
	}

	token := fmt.Sprintf("%d-%s", os.Getpid(), RandomID())
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		if os.IsExist(err) {
			return nil, errors.New("vault is already open in another Zervyra Vault instance")
		}
		return nil, err
	}
	if _, err := fmt.Fprintf(f, "pid=%d\ntoken=%s\ncreated=%s\n", os.Getpid(), token, nowRFC3339()); err != nil {
		_ = f.Close()
		_ = os.Remove(lockPath)
		return nil, err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(lockPath)
		return nil, err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(lockPath)
		return nil, err
	}
	return &VaultLock{path: lockPath, token: token}, nil
}

func (l *VaultLock) ownsCurrentLock() bool {
	return l != nil && l.path != "" && l.token != "" && readLockToken(l.path) == l.token
}

func (l *VaultLock) Touch() {
	if !l.ownsCurrentLock() {
		return
	}
	now := time.Now()
	_ = os.Chtimes(l.path, now, now)
}

func (l *VaultLock) Release() {
	if l == nil || l.path == "" {
		return
	}
	if l.ownsCurrentLock() {
		_ = os.Remove(l.path)
	}
	l.path = ""
	l.token = ""
}

func ExportVaultBackup(dst, password string, v Vault) error {
	dst = filepath.Clean(strings.TrimSpace(dst))
	if dst == "" || dst == "." {
		return errors.New("invalid backup destination")
	}
	raw, err := encodeVault(password, v)
	if err != nil {
		return fmt.Errorf("backup encryption failed: %w", err)
	}
	if err := writeAtomicFile(dst, raw); err != nil {
		return fmt.Errorf("backup write failed: %w", err)
	}
	loaded, err := Load(dst, password)
	if err != nil {
		return fmt.Errorf("backup verification failed: %w", err)
	}
	if loaded.UpdatedAt == "" {
		return errors.New("backup verification returned invalid vault")
	}
	return nil
}

func CreateNew(path, password string, v Vault) error {
	path = filepath.Clean(strings.TrimSpace(path))
	if path == "" || path == "." {
		return errors.New("invalid vault path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	raw, err := encodeVault(password, v)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		if os.IsExist(err) {
			return errors.New("vault destination already exists")
		}
		return err
	}
	mainVerified := false
	defer func() {
		// Once the main vault was fully written and verified it is valuable data.
		// Never delete that verified file merely because creating the redundant
		// recovery generation failed afterwards.
		if !mainVerified {
			_ = os.Remove(path)
		}
	}()
	if _, err := f.Write(raw); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if _, err := Load(path, password); err != nil {
		return fmt.Errorf("new vault verification failed: %w", err)
	}
	mainVerified = true

	// Establish a second complete encrypted generation immediately. If this
	// redundant copy fails, the already verified main vault remains intact.
	if err := writeAtomicFile(path+".recovery", raw); err != nil {
		return fmt.Errorf("vault was created and verified, but recovery copy failed; main vault was preserved: %w", err)
	}
	if _, err := Load(path+".recovery", password); err != nil {
		return fmt.Errorf("vault was created and verified, but recovery verification failed; main vault was preserved: %w", err)
	}
	return nil
}
