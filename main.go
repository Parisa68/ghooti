package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			if _, err := os.Stat(dstPath); os.IsNotExist(err) {
				if err := os.MkdirAll(dstPath, 0755); err != nil {
					return err
				}
			}
			return nil
		}

		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return err
			}
			if err := os.RemoveAll(dstPath); err != nil {
				return err
			}
			if err := os.Symlink(linkTarget, dstPath); err != nil {
				return err
			}
			return nil
		}
		sourceFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		destinationFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer destinationFile.Close()

		if _, err := io.Copy(destinationFile, sourceFile); err != nil {
			return err
		}

		if err := os.Chmod(dstPath, info.Mode()); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func prepareChroot(newroot string) error {
	// List of directories to copy
	dirsToCopy := []string{"/bin", "/lib", "/lib64", "/usr"}
	for _, dir := range dirsToCopy {
		srcDir := dir
		dstDir := filepath.Join(newroot, dir)
		if err := copyDir(srcDir, dstDir); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	newroot := "/home/pge005/ghooti"

	if err := prepareChroot(newroot); err != nil {
		fmt.Printf("Failed to set up chroot environment: %v", err)
		os.Exit(1)
	}

	cmd := exec.Command("/usr/bin/unshare", "--mount", "--uts", "--ipc", "--net", "--pid", "--fork", "--user", "--map-root-user", "/bin/busybox", "sh", "-c", "ls /")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Chroot:     newroot,
		Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWNET | syscall.CLONE_NEWPID | syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getgid(), Size: 1},
		},
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to run: %v", err)
		os.Exit(1)
	}
}
