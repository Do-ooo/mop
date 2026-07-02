package tui

import (
	"mop/cleaner"
	"mop/config"
	"mop/scanner"
	"mop/whitelist"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenMenu screen = iota
	screenScanning
	screenSelect
	screenCleaning
	screenDone
	screenAbout
	screenManageTools
)

type menuItem int

const (
	menuAnalyzeRegular menuItem = iota
	menuAnalyzeDeep
	menuManageTools
	menuAbout
)

type listEntry struct {
	groupIdx int
	itemIdx  int
	isHeader bool
}

type Model struct {
	screen           screen
	menuCursor       int
	groups           []scanner.ToolGroup
	entries          []listEntry
	selected         map[[2]int]bool
	whitelist        map[string]bool
	enabledScanners  map[string]bool
	toolList         []scanner.ToolScanner
	cursor           int
	totalSize        int64
	cleanedSize      int64
	currentIdx       int
	err              error
	spinnerIdx       int
	scanIdx          int
	cleanResults     []cleaner.CleanResult
	cleanStartTime   time.Time
	cleanElapsed     time.Duration
	trashMode        bool
	timeFilter       int
	deepMode         bool
	confirmClean     bool
	width            int
	height           int
	scrollOffset     int
}

type scanResultMsg struct {
	groups []scanner.ToolGroup
	err    error
}

type scanProgressMsg struct {
	group scanner.ToolGroup
}

type cleanProgressMsg struct {
	idx     int
	size    int64
	done    bool
	err     error
	results []cleaner.CleanResult
}

type tickMsg struct{}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

const asciiLogo = `
███╗   ███╗   ██████╗   ██████╗  
████╗ ████║  ██╔═══██╗  ██╔══██╗ 
██╔████╔██║  ██║   ██║  ██████╔╝ 
██║╚██╔╝██║  ██║   ██║  ██╔═══╝  
██║ ╚═╝ ██║  ╚██████╔╝  ██║      
╚═╝     ╚═╝   ╚═════╝   ╚═╝      
`

func InitialModel() Model {
	wl, _ := whitelist.Load()
	if wl == nil {
		wl = make(map[string]bool)
	}
	enabled, _ := scanner.LoadEnabled()
	if enabled == nil {
		enabled = make(map[string]bool)
	}
	scanner.SetEnabledFromMap(enabled)

	appConfig, _ := config.Load()
	trashMode := true
	if appConfig != nil {
		trashMode = appConfig.TrashMode
	}

	return Model{
		screen:          screenMenu,
		selected:        make(map[[2]int]bool),
		whitelist:       wl,
		enabledScanners: enabled,
		toolList:        scanner.Scanners,
		trashMode:       trashMode,
		timeFilter:      3,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.WindowSize()
}

func scanCmd() tea.Cmd {
	return scanCmdWithFilter(0, false)
}

func scanCmdWithFilter(timeFilter int, deepMode bool) tea.Cmd {
	return tea.Batch(
		tickCmd(),
		func() tea.Msg {
			scanner.IncrementScanCount()
			var groups []scanner.ToolGroup
			now := time.Now()
			for _, s := range scanner.Scanners {
				if !s.Enabled() {
					continue
				}
				if !s.Available() {
					continue
				}
				items, err := s.Scan()
				if err != nil {
					continue
				}

				var filtered []scanner.CacheItem
				for _, item := range items {
					if !deepMode && item.Risk == scanner.RiskDeep {
						continue
					}
					age := now.Sub(item.ModTime)
					days := int(age.Hours() / 24)
					switch timeFilter {
					case 0:
						if days <= 3 {
							filtered = append(filtered, item)
						}
					case 1:
						if days <= 7 {
							filtered = append(filtered, item)
						}
					case 2:
						if days <= 30 {
							filtered = append(filtered, item)
						}
					case 3:
						filtered = append(filtered, item)
					}
				}

				sort.Slice(filtered, func(i, j int) bool {
					return filtered[i].Size > filtered[j].Size
				})
				var total int64
				for _, item := range filtered {
					total += item.Size
				}
				groups = append(groups, scanner.ToolGroup{
					Name:      s.Name(),
					Type:      s.Type(),
					Items:     filtered,
					TotalSize: total,
				})
			}
			return scanResultMsg{groups: groups}
		},
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(80e6, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

	case tickMsg:
		if m.screen == screenScanning || m.screen == screenCleaning {
			m.spinnerIdx = (m.spinnerIdx + 1) % len(spinnerFrames)
			return m, tickCmd()
		}

	case scanResultMsg:
		m.groups = msg.groups
		m.err = msg.err
		m.screen = screenSelect
		m.cursor = 0
		m.scrollOffset = 0
		m.buildEntries()
		for gi, g := range m.groups {
			for ii := range g.Items {
				if !whitelist.IsWhitelisted(m.whitelist, g.Items[ii].Path) {
					m.selected[[2]int{gi, ii}] = true
				}
			}
		}
		m.totalSize = m.calcTotal()
		m.updateScrollOffset()
		return m, nil

	case cleanProgressMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		if msg.done {
			m.cleanResults = msg.results
			m.cleanElapsed = time.Since(m.cleanStartTime)
			m.screen = screenDone
			return m, nil
		}
		m.currentIdx = msg.idx
		m.cleanedSize += msg.size
		return m, nil
	}

	switch m.screen {
	case screenMenu:
		return updateMenu(m, msg)
	case screenSelect:
		return updateSelect(m, msg)
	case screenCleaning:
		return m, nil
	case screenDone:
		return updateDone(m, msg)
	case screenAbout:
		return updateAbout(m, msg)
	case screenManageTools:
		return updateManageTools(m, msg)
	}

	return m, nil
}

func (m Model) View() string {
	switch m.screen {
	case screenMenu:
		return viewMenu(m)
	case screenScanning:
		return viewScanning(m)
	case screenSelect:
		return viewSelect(m)
	case screenCleaning:
		return viewCleaning(m)
	case screenDone:
		return viewDone(m)
	case screenAbout:
		return viewAbout(m)
	case screenManageTools:
		return viewManageTools(m)
	}
	return ""
}

func (m *Model) buildEntries() {
	m.entries = nil
	for gi, g := range m.groups {
		m.entries = append(m.entries, listEntry{groupIdx: gi, isHeader: true})
		for ii := range g.Items {
			m.entries = append(m.entries, listEntry{groupIdx: gi, itemIdx: ii, isHeader: false})
		}
	}
}

func (m Model) calcTotal() int64 {
	var total int64
	for gi, g := range m.groups {
		for ii := range g.Items {
			if m.selected[[2]int{gi, ii}] {
				total += g.Items[ii].Size
			}
		}
	}
	return total
}

func (m Model) calcAllSize() int64 {
	var total int64
	for _, g := range m.groups {
		for _, item := range g.Items {
			if !whitelist.IsWhitelisted(m.whitelist, item.Path) {
				total += item.Size
			}
		}
	}
	return total
}
