package game

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"io"
	"os"
	"path/filepath"
)

type ResetPlayerHealSource struct{}

func (ResetPlayerHealSource) HealingSource() {}
func (ResetPlayerHealSource) ResetPlayer()   {}

// resetPlayer resets the player passed to the default State.
func resetPlayer(p *player.Player) {
	p.Inventory().Clear()
	p.Armour().Clear()
	p.SetHeldItems(item.Stack{}, item.Stack{})
	for _, effect := range p.Effects() {
		p.RemoveEffect(effect.Type())
	}
	p.EnderChestInventory().Clear()
	p.SetMaxHealth(20)
	p.Heal(20, ResetPlayerHealSource{})
	p.SetFood(20)
	p.SetExperienceProgress(0)
	p.SetExperienceLevel(0)
	p.SetScale(1)
}

// copyDir copies the contents of a directory from `path` to `to` recursively.
func copyDir(path, to string) error {
	// Ensure the source directory exists
	srcInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat source directory: %w", err)
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source path is not a directory")
	}

	// Create the destination directory if it doesn't exist
	err = os.MkdirAll(to, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Walk through the source directory
	return filepath.Walk(path, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking through source directory: %w", err)
		}

		// Determine the target path
		relPath, err := filepath.Rel(path, srcPath)
		if err != nil {
			return fmt.Errorf("failed to compute relative path: %w", err)
		}
		destPath := filepath.Join(to, relPath)

		if info.IsDir() {
			// Create subdirectory
			return os.MkdirAll(destPath, info.Mode())
		}

		// Copy file
		return copyFile(srcPath, destPath, info)
	})
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string, info os.FileInfo) error {
	// Open the source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy the contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	return nil
}
