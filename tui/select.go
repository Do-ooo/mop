package tui

import (
	"fmt"
	"mop/cleaner"
	"mop/config"
	"mop/scanner"
	"mop/update"
	"mop/whitelist"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#C2410C", Dark: "#F97316"}).
			MarginBottom(1)

	toolHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#EA580C", Dark: "#FB923C"})

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#EA580C", Dark: "#F97316"})

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#C2410C", Dark: "#F97316"}).
			Bold(true)

	totalStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#FBBF24"}).
			MarginTop(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#6B7280"}).
			MarginTop(1)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#EA580C", Dark: "#F97316"}).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#EF4444"})

	menuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#374151", Dark: "#D1D5DB"})

	menuCursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#C2410C", Dark: "#F97316"}).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#6B7280"})

	whitelistStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#4B5563"})

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#EF4444"}).
			Bold(true)
)

var menuItems = []struct {
	label string
	desc  string
}{
	{"Analyze", "Scan safe caches (Regular mode)"},
	{"Deep Analyze", "Scan caches + session history (Deep mode)"},
	{"Manage Tools", "Enable or disable scanners"},
	{"About", "About mop"},
}

func viewMenu(m Model) string {
	var b strings.Builder
	b.WriteString(logoStyle.Render(asciiLogo))
	b.WriteString("\n")

	labelWidth := 10
	for i, item := range menuItems {
		cursor := "  "
		if m.menuCursor == i {
			cursor = menuCursorStyle.Render("> ")
		}
		paddedLabel := fmt.Sprintf("%-*s", labelWidth, item.label)
		var label string
		if m.menuCursor == i {
			label = menuCursorStyle.Render(paddedLabel)
		} else {
			label = menuItemStyle.Render(paddedLabel)
		}
		b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, label, dimStyle.Render(item.desc)))
	}

	b.WriteString("\n")

	if info := update.GetCachedUpdate(); info != nil && info.Available {
		b.WriteString(selectedStyle.Render(fmt.Sprintf("  ↑ Update available: v%s → v%s (run 'mop update')\n\n", info.CurrentVersion, info.LatestVersion)))
	}

	b.WriteString(helpStyle.Render("[↑/↓] Navigate  [Enter] Select  [q] Quit"))

	return b.String()
}

func updateMenu(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.menuCursor > 0 {
				m.menuCursor--
			}
		case "down", "j":
			if m.menuCursor < len(menuItems)-1 {
				m.menuCursor++
			}
		case "enter":
			switch menuItem(m.menuCursor) {
			case menuAnalyzeRegular:
				m.deepMode = false
				m.screen = screenScanning
				return m, scanCmdWithFilter(m.timeFilter, m.deepMode)
			case menuAnalyzeDeep:
				m.deepMode = true
				m.screen = screenScanning
				return m, scanCmdWithFilter(m.timeFilter, m.deepMode)
			case menuManageTools:
				m.screen = screenManageTools
				m.cursor = 0
			case menuAbout:
				m.screen = screenAbout
			}
		}
	}
	return m, nil
}

func viewScanning(m Model) string {
	var b strings.Builder
	b.WriteString("\n\n")

	spinner := spinnerFrames[m.spinnerIdx%len(spinnerFrames)]
	b.WriteString(fmt.Sprintf("%s Scanning AI tools...\n\n", spinner))

	for _, s := range scanner.Scanners {
		if s.Available() {
			b.WriteString(fmt.Sprintf("  %s  %s\n", dimStyle.Render("•"), s.Name()))
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Please wait..."))
	return b.String()
}

func updateSelect(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			m.screen = screenMenu
		case "r", "R":
			m.screen = screenScanning
			m.spinnerIdx = 0
			return m, scanCmdWithFilter(m.timeFilter, m.deepMode)
		case "d", "D":
			m.trashMode = !m.trashMode
			appConfig := &config.AppConfig{TrashMode: m.trashMode}
			config.Save(appConfig)
		case "t", "T":
			m.timeFilter = (m.timeFilter + 1) % 4
			m.screen = screenScanning
			m.spinnerIdx = 0
			return m, scanCmdWithFilter(m.timeFilter, m.deepMode)
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}
		}

		if m.deepMode && m.confirmClean {
			switch msg.String() {
			case "enter", "y":
				m.confirmClean = false
				m.screen = screenCleaning
				m.currentIdx = 0
				m.cleanedSize = 0
				m.spinnerIdx = 0
				m.cleanResults = nil
				m.cleanItems = m.collectSelectedItems()
				m.cleanStartTime = time.Now()
				return m, tea.Batch(cleanStepCmd(m), tickCmd())
			case "n", "esc", "q":
				m.confirmClean = false
			}
			return m, nil
		}

		switch msg.String() {
		case " ":
			entry := m.entries[m.cursor]
			if entry.isHeader {
				allSelected := true
				g := m.groups[entry.groupIdx]
				for ii := range g.Items {
					if whitelist.IsWhitelisted(m.whitelist, g.Items[ii].Path) {
						continue
					}
					if !m.selected[[2]int{entry.groupIdx, ii}] {
						allSelected = false
						break
					}
				}
				for ii := range g.Items {
					if whitelist.IsWhitelisted(m.whitelist, g.Items[ii].Path) {
						continue
					}
					m.selected[[2]int{entry.groupIdx, ii}] = !allSelected
				}
			} else {
				key := [2]int{entry.groupIdx, entry.itemIdx}
				m.selected[key] = !m.selected[key]
			}
			m.totalSize = m.calcTotal()
		case "w":
			entry := m.entries[m.cursor]
			if !entry.isHeader {
				item := m.groups[entry.groupIdx].Items[entry.itemIdx]
				whitelist.Toggle(m.whitelist, item.Path)
				whitelist.Save(m.whitelist)
				key := [2]int{entry.groupIdx, entry.itemIdx}
				if whitelist.IsWhitelisted(m.whitelist, item.Path) {
					m.selected[key] = false
				}
				m.totalSize = m.calcTotal()
			}
		case "a":
			allSelected := true
			for gi, g := range m.groups {
				for ii := range g.Items {
					if whitelist.IsWhitelisted(m.whitelist, g.Items[ii].Path) {
						continue
					}
					if !m.selected[[2]int{gi, ii}] {
						allSelected = false
						break
					}
				}
				if !allSelected {
					break
				}
			}
			for gi, g := range m.groups {
				for ii := range g.Items {
					if whitelist.IsWhitelisted(m.whitelist, g.Items[ii].Path) {
						continue
					}
					m.selected[[2]int{gi, ii}] = !allSelected
				}
			}
			m.totalSize = m.calcTotal()
		case "i":
			for gi, g := range m.groups {
				for ii := range g.Items {
					if whitelist.IsWhitelisted(m.whitelist, g.Items[ii].Path) {
						continue
					}
					m.selected[[2]int{gi, ii}] = !m.selected[[2]int{gi, ii}]
				}
			}
			m.totalSize = m.calcTotal()
		case "enter":
			if m.deepMode && !m.confirmClean {
				m.confirmClean = true
				return m, nil
			}
			m.confirmClean = false
			m.screen = screenCleaning
			m.currentIdx = 0
			m.cleanedSize = 0
			m.spinnerIdx = 0
			m.cleanResults = nil
			m.cleanItems = m.collectSelectedItems()
			m.cleanStartTime = time.Now()
			return m, tea.Batch(cleanStepCmd(m), tickCmd())
		}
	}

	m.updateScrollOffset()

	return m, nil
}

func (m Model) cursorLine() int {
	if len(m.entries) == 0 || m.cursor >= len(m.entries) {
		return 0
	}
	// m.entries is built in group order (header then its items), so the entry
	// index advances by one per rendered row; no inner search is needed.
	line := 0
	entryIdx := 0
	for gi, g := range m.groups {
		if gi > 0 {
			line++
		}
		if m.cursor == entryIdx {
			return line
		}
		entryIdx++
		line++
		for range g.Items {
			if m.cursor == entryIdx {
				return line
			}
			entryIdx++
			line++
		}
	}
	return line
}

func (m *Model) updateScrollOffset() {
	contentLines := buildContentLines(*m)
	totalContent := len(contentLines)
	cursorLine := m.cursorLine()

	visibleHeight := calcVisibleHeight(*m)

	if totalContent <= visibleHeight {
		m.scrollOffset = 0
		return
	}

	maxOffset := totalContent - visibleHeight
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
	if cursorLine < m.scrollOffset {
		m.scrollOffset = cursorLine
	} else if cursorLine >= m.scrollOffset+visibleHeight {
		m.scrollOffset = cursorLine - visibleHeight + 1
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

func buildContentLines(m Model) []string {
	var lines []string

	if len(m.groups) == 0 {
		lines = append(lines, "No AI tool caches found.")
		return lines
	}

	descColWidth := 40
	sizeColWidth := 12

	// m.entries is ordered as [header, items...] per group, so we can track the
	// current entry index with a running counter instead of searching.
	entryIdx := 0
	for gi, g := range m.groups {
		headerIdx := entryIdx
		entryIdx++

		isCursorOnHeader := m.cursor == headerIdx
		allSelected := true
		hasNonWhitelisted := false
		var groupActiveSize int64
		for ii := range g.Items {
			if whitelist.IsWhitelisted(m.whitelist, g.Items[ii].Path) {
				continue
			}
			hasNonWhitelisted = true
			groupActiveSize += g.Items[ii].Size
			if !m.selected[[2]int{gi, ii}] {
				allSelected = false
			}
		}
		if !hasNonWhitelisted {
			allSelected = false
		}

		check := " "
		if allSelected {
			check = selectedStyle.Render("x")
		}

		cursor := " "
		if isCursorOnHeader {
			cursor = cursorStyle.Render(">")
		}

		headerPlain := fmt.Sprintf("◆ %s (%s)", g.Name, g.Type)
		headerPlainPadded := fmt.Sprintf("%-*s", descColWidth, headerPlain)
		headerColored := toolHeaderStyle.Render(headerPlainPadded)

		sizeStr := scanner.FormatSize(groupActiveSize)
		sizePadded := fmt.Sprintf(" %*s", sizeColWidth-1, sizeStr)

		headerLine := fmt.Sprintf("%s [%s] %s  %s",
			cursor, check,
			headerColored,
			sizePadded,
		)
		lines = append(lines, headerLine)

		if len(g.Items) == 0 {
			lines = append(lines, dimStyle.Render("    (No caches found)"))
		} else {
			for ii, item := range g.Items {
				thisEntryIdx := entryIdx
				entryIdx++

				isWL := whitelist.IsWhitelisted(m.whitelist, item.Path)
				cursor := "  "
				if thisEntryIdx == m.cursor {
					cursor = cursorStyle.Render("> ")
				}

				check := " "
				if m.selected[[2]int{gi, ii}] {
					check = selectedStyle.Render("x")
				}

				var descColored string
				var sizeColored string

				if isWL {
					descPlain := fmt.Sprintf("⊘ %s", item.Description)
					descPadded := fmt.Sprintf("%-*s", descColWidth-2, descPlain)
					descColored = whitelistStyle.Render(descPadded)
					sizeColored = whitelistStyle.Render(fmt.Sprintf(" %*s", sizeColWidth-1, "--"))
				} else {
					var prefix string
					if item.Risk == scanner.RiskDeep {
						prefix = "⚠ "
					} else {
						prefix = "  "
					}
					descPlain := fmt.Sprintf("%s%s", prefix, item.Description)
					descPadded := fmt.Sprintf("%-*s", descColWidth-2, descPlain)
					if item.Risk == scanner.RiskDeep {
						descColored = warningStyle.Render(descPadded)
					} else {
						descColored = descPadded
					}
					sizePlain := scanner.FormatSize(item.Size)
					if item.Size >= 100*1024*1024 {
						sizeColored = selectedStyle.Render(fmt.Sprintf("🔥%*s", sizeColWidth-2, sizePlain))
					} else {
						sizeColored = fmt.Sprintf(" %*s", sizeColWidth-1, sizePlain)
					}
				}

				line := fmt.Sprintf("%s [%s] %s  %s",
					cursor, check,
					descColored,
					sizeColored,
				)

				if thisEntryIdx == m.cursor && !isWL {
					line = cursorStyle.Render(line)
				}

				lines = append(lines, line)
			}
		}

		lines = append(lines, "")
	}

	return lines
}

func calcVisibleHeight(m Model) int {
	topPad := 4
	bottomPad := 4
	visibleHeight := m.height - topPad - bottomPad - 1
	if visibleHeight < 1 {
		visibleHeight = 1
	}
	return visibleHeight
}

func viewSelect(m Model) string {
	var b strings.Builder

	b.WriteString("\n\n")

	filterLabels := []string{"3 days", "7 days", "30 days", "All time"}
	filterLabel := filterLabels[m.timeFilter%len(filterLabels)]
	b.WriteString(dimStyle.Render("Filter: " + filterLabel + " [t]"))
	b.WriteString("\n\n")

	contentLines := buildContentLines(m)
	totalContent := len(contentLines)

	visibleHeight := calcVisibleHeight(m)

	scrollOffset := m.scrollOffset
	if totalContent <= visibleHeight {
		scrollOffset = 0
	} else {
		maxOffset := totalContent - visibleHeight
		if scrollOffset > maxOffset {
			scrollOffset = maxOffset
		}
		if scrollOffset < 0 {
			scrollOffset = 0
		}
	}

	endIdx := scrollOffset + visibleHeight
	if endIdx > totalContent {
		endIdx = totalContent
	}
	for i := scrollOffset; i < endIdx; i++ {
		b.WriteString(contentLines[i] + "\n")
	}
	for i := endIdx - scrollOffset; i < visibleHeight; i++ {
		b.WriteString("\n")
	}

	b.WriteString(totalStyle.Render(fmt.Sprintf("Selected: %s", scanner.FormatSize(m.totalSize))))
	b.WriteString(dimStyle.Render(fmt.Sprintf("  │  Total: %s", scanner.FormatSize(m.calcAllSize()))))
	if m.deepMode {
		b.WriteString("  ")
		b.WriteString(warningStyle.Render("⚠ Deep Mode"))
	}
	b.WriteString("\n\n")

	trashLabel := "Trash"
	if !m.trashMode {
		trashLabel = "Delete"
	}
	if m.deepMode && m.confirmClean {
		b.WriteString(warningStyle.Render("⚠ Deep clean will delete unrecoverable data! [Enter/y] Confirm  [n] Cancel"))
	} else {
		b.WriteString(helpStyle.Render(fmt.Sprintf("[Space] Select  [a] All  [i] Invert  [w] Whitelist  [r] Refresh  [d] %s  [t] Time  [Enter] Clean  [q] Back", trashLabel)))
	}
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("Error: " + m.err.Error()))
	}

	return b.String()
}

func cleanStepCmd(m Model) tea.Cmd {
	idx := m.currentIdx
	trashMode := m.trashMode
	if idx >= len(m.cleanItems) {
		return func() tea.Msg {
			return cleanProgressMsg{done: true}
		}
	}
	item := m.cleanItems[idx]
	return func() tea.Msg {
		var fileCleaner cleaner.Cleaner
		if trashMode {
			fileCleaner = cleaner.NewTrashCleaner()
		} else {
			fileCleaner = cleaner.NewFileCleaner()
		}
		result, _ := fileCleaner.Clean(item)
		var size int64
		if result.Success {
			size = result.Size
		}
		return cleanProgressMsg{
			idx:       idx + 1,
			size:      size,
			result:    result,
			hasResult: true,
		}
	}
}

func viewCleaning(m Model) string {
	var b strings.Builder
	b.WriteString("\n\n")

	spinner := spinnerFrames[m.spinnerIdx%len(spinnerFrames)]
	b.WriteString(fmt.Sprintf("%s Cleaning selected items...\n\n", spinner))

	totalSelected := 0
	for gi, g := range m.groups {
		for ii := range g.Items {
			if m.selected[[2]int{gi, ii}] {
				totalSelected++
			}
		}
	}

	if totalSelected > 0 {
		progress := float64(m.currentIdx) / float64(totalSelected) * 100
		barWidth := 40
		filled := int(progress / 100 * float64(barWidth))
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		b.WriteString(fmt.Sprintf("[%s] %.0f%%\n\n", bar, progress))
	}

	b.WriteString(helpStyle.Render("Please wait..."))
	return b.String()
}

func viewDone(m Model) string {
	var b strings.Builder
	b.WriteString("\n\n")
	b.WriteString(successStyle.Render("Clean complete!"))
	b.WriteString("\n\n")

	var successCount, failCount int
	var successSize int64
	var failedResults []cleaner.CleanResult
	for _, r := range m.cleanResults {
		if r.Success {
			successCount++
			successSize += r.Size
		} else {
			failCount++
			failedResults = append(failedResults, r)
		}
	}

	totalCount := successCount + failCount

	b.WriteString(fmt.Sprintf("%s %d items (%s)\n\n",
		successStyle.Render("Freed"),
		successCount,
		successStyle.Render(scanner.FormatSize(successSize)),
	))

	if totalCount > 0 {
		b.WriteString(fmt.Sprintf("  %s  %d of %d\n", dimStyle.Render("Success:"), successCount, totalCount))
		if failCount > 0 {
			b.WriteString(fmt.Sprintf("  %s  %d\n", errorStyle.Render("Failed:"), failCount))
		}
		if m.cleanElapsed > 0 {
			elapsedStr := fmt.Sprintf("%.1fs", m.cleanElapsed.Seconds())
			b.WriteString(fmt.Sprintf("  %s  %s\n", dimStyle.Render("Time:"), elapsedStr))
		}
		b.WriteString("\n")
	}

	if len(failedResults) > 0 {
		b.WriteString(errorStyle.Render("Failed items:\n"))
		for _, r := range failedResults {
			b.WriteString(fmt.Sprintf("  %s\n", r.Path))
			b.WriteString(fmt.Sprintf("    %s\n", errorStyle.Render(r.Error)))
		}
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render("[Enter] Back to list  [q] Main menu"))
	return b.String()
}

func updateDone(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			m.screen = screenMenu
		case "enter":
			m.screen = screenScanning
			m.spinnerIdx = 0
			return m, scanCmdWithFilter(m.timeFilter, m.deepMode)
		}
	}
	return m, nil
}

func viewAbout(m Model) string {
	var b strings.Builder
	b.WriteString("\n\n")
	b.WriteString("mop - AI Tool Cache Cleaner\n\n")
	b.WriteString("Lightweight, fast TUI tool for cleaning up\n")
	b.WriteString("cache and session data from AI coding tools.\n\n")
	b.WriteString(dimStyle.Render(fmt.Sprintf("v%s\n", update.Version)))
	b.WriteString(dimStyle.Render("Built with Go + Bubble Tea\n\n"))
	b.WriteString(helpStyle.Render("[q/esc] Back to menu"))
	return b.String()
}

func updateAbout(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "enter":
			m.screen = screenMenu
		}
	}
	return m, nil
}

func viewManageTools(m Model) string {
	var b strings.Builder
	b.WriteString("\n\n")
	b.WriteString(toolHeaderStyle.Render("Manage Tools"))
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("Toggle tools to include or exclude from scan\n"))

	for i, s := range m.toolList {
		enabled := s.Enabled()
		available := s.Available()

		var check, checkStr string
		if enabled {
			check = selectedStyle.Render("[x]")
			checkStr = "[x]"
		} else {
			check = dimStyle.Render("[ ]")
			checkStr = "[ ]"
		}

		name := s.Name()
		typeStr := s.Type()

		var status, statusStr string
		if available {
			status = successStyle.Render("found")
			statusStr = "found"
		} else {
			status = dimStyle.Render("not installed")
			statusStr = "not installed"
		}

		// 构建行内容
		var line string
		if i == m.cursor {
			cursor := cursorStyle.Render("> ")
			line = fmt.Sprintf("%s%s %s %s (%s)", cursor, check, name, typeStr, status)
		} else {
			line = fmt.Sprintf("  %s %s %s (%s)", checkStr, name, typeStr, statusStr)
		}

		b.WriteString(line + "\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("[Space] Toggle  [a] Enable all  [n] Disable all  [q] Back to menu"))
	return b.String()
}

func updateManageTools(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			scanner.SaveEnabled(m.enabledScanners)
			m.screen = screenMenu
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.toolList)-1 {
				m.cursor++
			}
		case " ":
			s := m.toolList[m.cursor]
			name := s.Name()
			currentEnabled := s.Enabled()
			newEnabled := !currentEnabled

			scanner.SetEnabled(name, newEnabled)
			m.enabledScanners[name] = newEnabled
		case "a":
			for _, s := range m.toolList {
				name := s.Name()
				scanner.SetEnabled(name, true)
				m.enabledScanners[name] = true
			}
		case "n":
			for _, s := range m.toolList {
				name := s.Name()
				scanner.SetEnabled(name, false)
				m.enabledScanners[name] = false
			}
		}
	}
	return m, nil
}


