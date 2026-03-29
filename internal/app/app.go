package app

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spinchange/yanp-tui/internal/config"
	"github.com/spinchange/yanp-tui/internal/ui"
	"github.com/spinchange/yanp-tui/internal/vault"
)

func Run(args []string, cfg config.Config) error {
	command := "tui"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		command = args[0]
		args = args[1:]
	}

	switch command {
	case "tui":
		return runTUI(args, cfg)
	case "publish":
		return runPublish(args, cfg)
	case "capture":
		return runCapture(args, cfg)
	case "rename":
		return runRename(args, cfg)
	default:
		return fmt.Errorf("unknown command %q", command)
	}
}

func runTUI(args []string, cfg config.Config) error {
	fs := flag.NewFlagSet("tui", flag.ContinueOnError)
	defaultVault := strings.TrimSpace(cfg.Vault)
	vaultPath := fs.String("vault", defaultVault, "path to YANP vault")
	if err := fs.Parse(args); err != nil {
		return err
	}

	model, err := ui.New(*vaultPath, cfg)
	if err != nil {
		return err
	}
	_, err = tea.NewProgram(model, tea.WithAltScreen()).Run()
	return err
}

func runPublish(args []string, cfg config.Config) error {
	fs := flag.NewFlagSet("publish", flag.ContinueOnError)
	vaultPath := fs.String("vault", resolvedVault(cfg), "path to YANP vault")
	outputDir := fs.String("out", filepath.Join(resolvedVault(cfg), "_published"), "publish output directory")
	skipDrafts := fs.Bool("skip-drafts", false, "skip notes with draft status")
	markUnresolved := fs.Bool("mark-unresolved", true, "mark unresolved links in output")
	if err := fs.Parse(args); err != nil {
		return err
	}

	v, err := vault.Load(*vaultPath)
	if err != nil {
		return err
	}
	warnings, err := v.Publish(vault.PublishOptions{
		OutputDir:           *outputDir,
		SkipDrafts:          *skipDrafts,
		MarkUnresolved:      *markUnresolved,
		PreserveFrontmatter: true,
	})
	for _, warning := range warnings {
		fmt.Fprintln(os.Stderr, "warning:", warning)
	}
	return err
}

func runCapture(args []string, cfg config.Config) error {
	fs := flag.NewFlagSet("capture", flag.ContinueOnError)
	vaultPath := fs.String("vault", resolvedVault(cfg), "path to YANP vault")
	text := fs.String("text", "", "capture text")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*text) == "" {
		return errors.New("capture text is required")
	}
	v, err := vault.Load(*vaultPath)
	if err != nil {
		return err
	}
	return v.Capture(*text)
}

func runRename(args []string, cfg config.Config) error {
	fs := flag.NewFlagSet("rename", flag.ContinueOnError)
	vaultPath := fs.String("vault", resolvedVault(cfg), "path to YANP vault")
	notePath := fs.String("note", "", "relative note path to rename")
	title := fs.String("title", "", "new title")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*notePath) == "" || strings.TrimSpace(*title) == "" {
		return errors.New("both -note and -title are required")
	}
	v, err := vault.Load(*vaultPath)
	if err != nil {
		return err
	}
	newPath, warnings, err := v.RenameNote(*notePath, *title)
	for _, warning := range warnings {
		fmt.Fprintln(os.Stderr, "warning:", warning)
	}
	if err != nil {
		return err
	}
	fmt.Println(newPath)
	return nil
}

func resolvedVault(cfg config.Config) string {
	if strings.TrimSpace(cfg.Vault) != "" {
		return cfg.Vault
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}
