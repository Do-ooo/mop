package main

import (
	"flag"
	"fmt"
	"mop/cleaner"
	"mop/config"
	"mop/scanner"
	"mop/tui"
	"mop/whitelist"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	_ "mop/scanner"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "clean":
			runClean()
			return
		case "scan":
			runScan()
			return
		case "--dry-run":
			runDryRun()
			return
		}
	}

	p := tea.NewProgram(tui.InitialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func runClean() {
	fs := flag.NewFlagSet("clean", flag.ExitOnError)
	deleteMode := fs.Bool("delete", false, "Permanently delete instead of moving to Trash")
	dryRun := fs.Bool("dry", false, "Preview only, don't actually delete")
	fs.Parse(os.Args[2:])

	appConfig, _ := config.Load()
	trashMode := true
	if appConfig != nil {
		trashMode = appConfig.TrashMode
	}
	if *deleteMode {
		trashMode = false
	}

	fmt.Println("Scanning...")
	groups, err := scanner.ScanAll()
	if err != nil {
		fmt.Printf("Scan error: %v\n", err)
		os.Exit(1)
	}

	wl, _ := whitelist.Load()
	if wl == nil {
		wl = make(map[string]bool)
	}

	var totalItems int
	var totalSize int64
	for _, g := range groups {
		for _, item := range g.Items {
			if whitelist.IsWhitelisted(wl, item.Path) {
				continue
			}
			totalItems++
			totalSize += item.Size
		}
	}

	if totalItems == 0 {
		fmt.Println("Nothing to clean.")
		return
	}

	fmt.Printf("Found %d items (%s)\n", totalItems, scanner.FormatSize(totalSize))

	if *dryRun {
		fmt.Println("\n(Dry run - no files were deleted)")
		for _, g := range groups {
			for _, item := range g.Items {
				if whitelist.IsWhitelisted(wl, item.Path) {
					continue
				}
				fmt.Printf("  %-40s %10s\n", item.Description, scanner.FormatSize(item.Size))
			}
		}
		return
	}

	var c cleaner.Cleaner
	if trashMode {
		c = cleaner.NewTrashCleaner()
		fmt.Println("\nMoving to Trash...")
	} else {
		c = cleaner.NewFileCleaner()
		fmt.Println("\nDeleting...")
	}

	start := time.Now()
	var success, failed int
	var freed int64

	for _, g := range groups {
		for _, item := range g.Items {
			if whitelist.IsWhitelisted(wl, item.Path) {
				continue
			}
			result, err := c.Clean(item)
			if err != nil || !result.Success {
				failed++
				fmt.Printf("  FAIL %s\n", item.Path)
			} else {
				success++
				freed += result.Size
			}
		}
	}

	elapsed := time.Since(start)

	fmt.Println()
	fmt.Printf("Freed %s (%d items)\n", scanner.FormatSize(freed), success)
	if failed > 0 {
		fmt.Printf("Failed: %d\n", failed)
	}
	fmt.Printf("Time: %.1fs\n", elapsed.Seconds())
}

func runScan() {
	fmt.Println("Scanning...")
	groups, err := scanner.ScanAll()
	if err != nil {
		fmt.Printf("Scan error: %v\n", err)
		os.Exit(1)
	}

	var totalItems int
	var totalSize int64

	for _, g := range groups {
		fmt.Printf("\n%s (%s)\n", g.Name, g.Type)
		fmt.Printf("  %-40s %10s\n", "Path", "Size")
		fmt.Printf("  %s\n", "─")
		for _, item := range g.Items {
			fmt.Printf("  %-40s %10s\n", item.Description, scanner.FormatSize(item.Size))
			totalItems++
			totalSize += item.Size
		}
	}

	fmt.Printf("\nTotal: %d items, %s\n", totalItems, scanner.FormatSize(totalSize))
}

func runDryRun() {
	fmt.Println("Scanning (dry run)...")
	groups, err := scanner.ScanAll()
	if err != nil {
		fmt.Printf("Scan error: %v\n", err)
		os.Exit(1)
	}

	wl, _ := whitelist.Load()
	if wl == nil {
		wl = make(map[string]bool)
	}

	var totalItems int
	var totalSize int64
	for _, g := range groups {
		for _, item := range g.Items {
			if whitelist.IsWhitelisted(wl, item.Path) {
				continue
			}
			totalItems++
			totalSize += item.Size
		}
	}

	if totalItems == 0 {
		fmt.Println("Nothing to clean.")
		return
	}

	fmt.Printf("Found %d items (%s)\n\n", totalItems, scanner.FormatSize(totalSize))
	for _, g := range groups {
		fmt.Printf("%s (%s)\n", g.Name, g.Type)
		for _, item := range g.Items {
			if whitelist.IsWhitelisted(wl, item.Path) {
				continue
			}
			riskTag := ""
			if item.Risk == scanner.RiskDeep {
				riskTag = " [Deep]"
			}
			fmt.Printf("  %-40s %10s%s\n", item.Description, scanner.FormatSize(item.Size), riskTag)
		}
		fmt.Println()
	}
	fmt.Println("(Dry run - no files were deleted)")
}
