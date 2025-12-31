package gwt

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Add(name string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	config, err := LoadConfig(cwd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	dirName := strings.ReplaceAll(name, "/", "-")
	worktreePath := filepath.Join(cwd, "..", dirName)

	if err := createWorktree(name, worktreePath); err != nil {
		return err
	}

	if err := createSymlinks(cwd, worktreePath, config.Include); err != nil {
		return err
	}

	fmt.Printf("Created worktree at %s\n", worktreePath)
	return nil
}

func createWorktree(branch, path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("directory already exists: %s", path)
	}

	var cmd *exec.Cmd
	if branchExists(branch) {
		if branchInUse(branch) {
			return fmt.Errorf("branch %s is already checked out in another worktree", branch)
		}
		cmd = exec.Command("git", "worktree", "add", path, branch)
	} else {
		cmd = exec.Command("git", "worktree", "add", "-b", branch, path)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	return nil
}

func branchExists(name string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", "refs/heads/"+name)
	return cmd.Run() == nil
}

func branchInUse(name string) bool {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "branch refs/heads/"+name) {
			return true
		}
	}
	return false
}

func createSymlinks(srcDir, dstDir string, targets []string) error {
	for _, target := range targets {
		srcPath := filepath.Join(srcDir, target)
		dstPath := filepath.Join(dstDir, target)

		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "warning: %s does not exist, skipping\n", target)
			continue
		}

		if err := os.Symlink(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to create symlink for %s: %w", target, err)
		}

		fmt.Printf("Created symlink: %s -> %s\n", dstPath, srcPath)
	}

	return nil
}
