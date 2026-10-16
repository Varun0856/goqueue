package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	apiURL       = "http://localhost:8080/jobs"
	refreshEvery = 5 * time.Second
)

type Job struct {
	ID      string `json:"ID"`
	Type    string `json:"Type"`
	Payload string `json:"Payload"`
	Status  string `json:"Status"`
	Result  string `json:"Result"`
	Error   string `json:"Error"`
}

type createJobRequest struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type createJobResponse struct {
	ID     string `json:"ID"`
	Status string `json:"status"`
}

type jobsLoadedMsg struct {
	jobs []Job
	err  error
}

type jobSubmittedMsg struct {
	job createJobResponse
	err error
}

type tickMsg time.Time

type viewMode int

const (
	dashboardView viewMode = iota
	detailsView
	submitView
)

// Palette
var (
	accent = lipgloss.Color("#7C5CFC")
	cyan   = lipgloss.Color("#56D4DD")
	green  = lipgloss.Color("#3DDC97")
	yellow = lipgloss.Color("#F5C451")
	red    = lipgloss.Color("#FF6B81")
	muted  = lipgloss.Color("#8492A6")
	white  = lipgloss.Color("#F4F7FF")
	dark   = lipgloss.Color("#1B2030")
	line   = lipgloss.Color("#394158")

	bannerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accent)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(muted)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(cyan)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(line).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(cyan)

	helpStyle = lipgloss.NewStyle().
			Foreground(muted)

	keyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(white).
			Background(dark).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(red)

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(green)

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(muted)

	valueStyle = lipgloss.NewStyle().
			Foreground(white)
)

type model struct {
	jobs      []Job
	selected  int
	view      viewMode
	err       error
	loading   bool
	lastFetch time.Time

	width  int
	height int

	formField  int
	jobType    string
	payload    textinput.Model
	message    string
	submitting bool
}

func initialModel() model {
	payload := textinput.New()
	payload.Placeholder = "Type your payload here..."
	payload.CharLimit = 4096
	payload.Width = 60

	return model{
		loading: true,
		view:    dashboardView,
		jobType: "uppercase",
		payload: payload,
		width:   100,
		height:  30,
	}
}

func fetchJobs() tea.Cmd {
	return func() tea.Msg {
		client := &http.Client{Timeout: 3 * time.Second}

		resp, err := client.Get(apiURL)
		if err != nil {
			return jobsLoadedMsg{err: err}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return jobsLoadedMsg{
				err: fmt.Errorf("API returned status %s", resp.Status),
			}
		}

		var jobs []Job
		if err := json.NewDecoder(resp.Body).Decode(&jobs); err != nil {
			return jobsLoadedMsg{err: err}
		}

		return jobsLoadedMsg{jobs: jobs}
	}
}

func submitJob(jobType, payload string) tea.Cmd {
	return func() tea.Msg {
		body, err := json.Marshal(createJobRequest{
			Type:    jobType,
			Payload: payload,
		})
		if err != nil {
			return jobSubmittedMsg{err: err}
		}

		req, err := http.NewRequest(
			http.MethodPost,
			apiURL,
			bytes.NewReader(body),
		)
		if err != nil {
			return jobSubmittedMsg{err: err}
		}

		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return jobSubmittedMsg{err: err}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			return jobSubmittedMsg{
				err: fmt.Errorf("API returned status %s", resp.Status),
			}
		}

		var result createJobResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return jobSubmittedMsg{err: err}
		}

		return jobSubmittedMsg{job: result}
	}
}

func tick() tea.Cmd {
	return tea.Tick(refreshEvery, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tea.Batch(fetchJobs(), tick())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if m.width < 40 {
			m.width = 40
		}
		if m.height < 12 {
			m.height = 12
		}

		available := m.width - 12
		if available < 10 {
			available = 10
		}
		if available > 60 {
			available = 60
		}
		m.payload.Width = available

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		switch m.view {
		case dashboardView:
			switch msg.String() {
			case "q":
				return m, tea.Quit

			case "r":
				m.loading = true
				return m, fetchJobs()

			case "s":
				m.view = submitView
				m.formField = 0
				m.jobType = "uppercase"
				m.payload.SetValue("")
				m.message = ""
				m.submitting = false
				m.payload.Blur()
				return m, nil

			case "up", "k":
				if m.selected > 0 {
					m.selected--
				}

			case "down", "j":
				if m.selected < len(m.jobs)-1 {
					m.selected++
				}

			case "enter":
				if len(m.jobs) > 0 {
					m.view = detailsView
				}
			}

		case detailsView:
			switch msg.String() {
			case "esc", "backspace":
				m.view = dashboardView

			case "r":
				m.loading = true
				return m, fetchJobs()

			case "q":
				return m, tea.Quit
			}

		case submitView:
			if m.submitting {
				if msg.String() == "esc" {
					m.view = dashboardView
				}
				return m, nil
			}

			switch msg.String() {
			case "esc":
				m.payload.Blur()
				m.view = dashboardView
				return m, nil

			case "tab", "down", "shift+tab", "up":
				if m.formField == 0 {
					m.formField = 1
					return m, m.payload.Focus()
				}

				m.formField = 0
				m.payload.Blur()
				return m, nil

			case "left", "right":
				if m.formField == 0 {
					if m.jobType == "uppercase" {
						m.jobType = "reverse"
					} else {
						m.jobType = "uppercase"
					}
				}

			case "enter":
				if m.formField == 0 {
					m.formField = 1
					return m, m.payload.Focus()
				}

				if strings.TrimSpace(m.payload.Value()) == "" {
					m.message = "Payload cannot be empty."
					break
				}

				m.message = "Submitting job..."
				m.submitting = true
				return m, submitJob(m.jobType, m.payload.Value())

			default:
				if m.formField == 1 {
					var cmd tea.Cmd
					m.payload, cmd = m.payload.Update(msg)
					return m, cmd
				}
			}
		}

	case tickMsg:
		cmds := []tea.Cmd{tick()}
		if m.view == dashboardView && !m.submitting {
			m.loading = true
			cmds = append(cmds, fetchJobs())
		}
		return m, tea.Batch(cmds...)

	case jobsLoadedMsg:
		m.loading = false
		m.lastFetch = time.Now()
		m.err = msg.err

		// Keep the previous data if the refresh fails.
		if msg.err == nil {
			m.jobs = msg.jobs

			if len(m.jobs) == 0 {
				m.selected = 0
				if m.view == detailsView {
					m.view = dashboardView
				}
			} else if m.selected >= len(m.jobs) {
				m.selected = len(m.jobs) - 1
			}
		}

	case jobSubmittedMsg:
		m.submitting = false

		if msg.err != nil {
			m.message = "Submission failed: " + msg.err.Error()
			break
		}

		m.message = fmt.Sprintf(
			"Job submitted: %s (%s)",
			msg.job.ID,
			msg.job.Status,
		)

		m.payload.Blur()
		m.view = dashboardView
		m.loading = true
		return m, fetchJobs()
	}

	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	switch m.view {
	case submitView:
		b.WriteString(m.renderSubmit())
	case detailsView:
		b.WriteString(m.renderDetails())
	default:
		b.WriteString(m.renderDashboard())
	}

	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Left,
		lipgloss.Top,
		b.String(),
	)
}

func (m model) renderHeader() string {
	banner := `
 ██████╗  ██████╗  ██████╗ ██╗   ██╗███████╗██╗   ██╗███████╗
██╔════╝ ██╔═══██╗██╔═══██╗██║   ██║██╔════╝██║   ██║██╔════╝
██║  ███╗██║   ██║██║   ██║██║   ██║█████╗  ██║   ██║█████╗  
██║   ██║██║   ██║██║▄▄ ██║██║   ██║██╔══╝  ██║   ██║██╔══╝  
╚██████╔╝╚██████╔╝╚██████╔╝╚██████╔╝███████╗╚██████╔╝███████╗
 ╚═════╝  ╚═════╝  ╚══▀▀═╝  ╚═════╝ ╚══════╝ ╚═════╝ ╚══════╝` // Use a compact wordmark on narrow terminals.
	if m.width < 78 {
		banner = "GOQUEUE"
	}

	var b strings.Builder
	b.WriteString(bannerStyle.Render(banner))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render(
		"CONCURRENT JOB PROCESSING  /  WORKER POOL  /  HTTP API",
	))
	b.WriteString("\n")

	status := successStyle.Render("● ONLINE")
	if m.err != nil {
		status = errorStyle.Render("● API ERROR")
	} else if m.loading {
		status = lipgloss.NewStyle().Foreground(yellow).Render("● REFRESHING")
	}

	b.WriteString(
		lipgloss.JoinHorizontal(
			lipgloss.Center,
			titleStyle.Render("JOB CONTROL CENTER"),
			"   ",
			status,
		),
	)
	b.WriteString("\n")

	return b.String()
}

func (m model) renderDashboard() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  JOB QUEUE"))
	b.WriteString("\n")

	summary := fmt.Sprintf(
		"  Total: %d    Queued: %d    Running: %d    Completed: %d    Failed: %d",
		len(m.jobs),
		countStatus(m.jobs, "queued"),
		countStatus(m.jobs, "running"),
		countStatus(m.jobs, "completed"),
		countStatus(m.jobs, "failed"),
	)
	b.WriteString(subtitleStyle.Render(summary))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(errorStyle.Render("  API CONNECTION FAILED"))
		b.WriteString("\n  " + m.err.Error())
		b.WriteString("\n  Start the GoQueue server and press R to retry.\n")
	} else if len(m.jobs) == 0 {
		b.WriteString("\n")
		b.WriteString(titleStyle.Render("  No jobs in the queue"))
		b.WriteString("\n")
		b.WriteString(subtitleStyle.Render("  Press S to submit your first job."))
		b.WriteString("\n")
	} else {
		idWidth := m.width - 47
		if idWidth > 36 {
			idWidth = 36
		}
		if idWidth < 8 {
			idWidth = 8
		}

		header := fmt.Sprintf(
			"  %-3s %-*s %-12s %-12s",
			"",
			idWidth,
			"JOB ID",
			"TYPE",
			"STATUS",
		)
		b.WriteString(titleStyle.Render(header))
		b.WriteString("\n")

		separatorWidth := m.width - 8
		if separatorWidth < 10 {
			separatorWidth = 10
		}
		b.WriteString(subtitleStyle.Render("  " + strings.Repeat("─", separatorWidth)))
		b.WriteString("\n")

		maxRows := m.height - 17
		if maxRows < 1 {
			maxRows = 1
		}

		for i, j := range m.jobs {
			if i >= maxRows {
				remaining := len(m.jobs) - maxRows
				b.WriteString(subtitleStyle.Render(
					fmt.Sprintf("\n  ... and %d more jobs", remaining),
				))
				break
			}

			cursor := "  "
			rowStyle := lipgloss.NewStyle()

			if i == m.selected {
				cursor = "› "
				rowStyle = selectedStyle
			}

			id := truncate(j.ID, idWidth)
			jobType := truncate(j.Type, 12)
			status := renderStatus(j.Status)

			row := fmt.Sprintf(
				"  %-2s %-*s %-12s %s",
				cursor,
				idWidth,
				id,
				jobType,
				status,
			)

			b.WriteString(rowStyle.Render(row))
			b.WriteString("\n")
		}
	}

	if !m.lastFetch.IsZero() {
		b.WriteString("\n")
		b.WriteString(subtitleStyle.Render(
			"  LAST SYNC  " + m.lastFetch.Format("15:04:05"),
		))
	}

	if m.message != "" {
		b.WriteString("\n\n  ")
		if strings.HasPrefix(m.message, "Submission failed") {
			b.WriteString(errorStyle.Render(m.message))
		} else {
			b.WriteString(successStyle.Render(m.message))
		}
	}

	return panelStyle.Width(maxInt(20, m.width-4)).Render(b.String())
}

func (m model) renderDetails() string {
	var b strings.Builder

	if len(m.jobs) == 0 || m.selected >= len(m.jobs) {
		b.WriteString(titleStyle.Render("No job selected"))
		return panelStyle.Width(maxInt(20, m.width-4)).Render(b.String())
	}

	j := m.jobs[m.selected]

	b.WriteString(titleStyle.Render("  JOB DETAILS"))
	b.WriteString("\n\n")

	writeDetail := func(label, value string) {
		b.WriteString("  ")
		b.WriteString(labelStyle.Render(label))
		b.WriteString("  ")
		b.WriteString(valueStyle.Render(value))
		b.WriteString("\n")
	}

	writeDetail("ID", j.ID)
	writeDetail("TYPE", j.Type)
	writeDetail("STATUS", j.Status)
	writeDetail("PAYLOAD", j.Payload)
	writeDetail("RESULT", j.Result)

	b.WriteString("\n")
	b.WriteString(labelStyle.Render("  ERROR"))
	b.WriteString("\n")
	if j.Error == "" {
		b.WriteString(subtitleStyle.Render("  None"))
	} else {
		b.WriteString(errorStyle.Render("  " + j.Error))
	}

	return panelStyle.Width(maxInt(20, m.width-4)).Render(b.String())
}

func (m model) renderSubmit() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  SUBMIT A JOB"))
	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render("  Choose an operation and provide its payload."))
	b.WriteString("\n\n")

	typeLabel := "  "
	if m.formField == 0 {
		typeLabel = "› "
	}

	b.WriteString(typeLabel)
	b.WriteString(labelStyle.Render("JOB TYPE"))
	b.WriteString("\n")

	typeOptions := []string{"uppercase", "reverse"}
	for _, option := range typeOptions {
		if m.jobType == option {
			b.WriteString("    ")
			b.WriteString(selectedStyle.Render("● " + strings.ToUpper(option)))
			b.WriteString("\n")
		} else {
			b.WriteString("    ")
			b.WriteString(subtitleStyle.Render("○ " + strings.ToUpper(option)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	payloadLabel := "  "
	if m.formField == 1 {
		payloadLabel = "› "
	}

	b.WriteString(payloadLabel)
	b.WriteString(labelStyle.Render("PAYLOAD"))
	b.WriteString("\n  ")
	b.WriteString(m.payload.View())
	b.WriteString("\n")

	if m.submitting {
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(yellow).Render("  Submitting job..."))
	} else if m.message != "" {
		b.WriteString("\n")
		if strings.HasPrefix(m.message, "Submission failed") ||
			m.message == "Payload cannot be empty." {
			b.WriteString(errorStyle.Render("  " + m.message))
		} else {
			b.WriteString(successStyle.Render("  " + m.message))
		}
	}

	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render("  ← / → changes the job type"))

	return panelStyle.Width(maxInt(20, m.width-4)).Render(b.String())
}

func (m model) renderFooter() string {
	var b strings.Builder

	b.WriteString(subtitleStyle.Render(strings.Repeat("─", maxInt(10, m.width-2))))
	b.WriteString("\n")

	var controls string

	switch m.view {
	case submitView:
		controls = renderKey("TAB", "Switch field") + "  " +
			renderKey("ENTER", "Submit") + "  " +
			renderKey("ESC", "Cancel") + "  " +
			renderKey("CTRL+C", "Quit")
	case detailsView:
		controls = renderKey("ESC", "Back") + "  " +
			renderKey("R", "Refresh") + "  " +
			renderKey("Q", "Quit")
	default:
		controls = renderKey("↑/↓", "Navigate") + "  " +
			renderKey("ENTER", "Details") + "  " +
			renderKey("S", "Submit") + "  " +
			renderKey("R", "Refresh") + "  " +
			renderKey("Q", "Quit")
	}

	b.WriteString(controls)
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("  GOQUEUE  •  TERMINAL JOB MANAGER"))

	return b.String()
}

func renderKey(key, description string) string {
	return keyStyle.Render(key) + " " + helpStyle.Render(description)
}

func renderStatus(status string) string {
	switch strings.ToLower(status) {
	case "queued":
		return lipgloss.NewStyle().Foreground(yellow).Bold(true).Render("● QUEUED")
	case "running":
		return lipgloss.NewStyle().Foreground(cyan).Bold(true).Render("● RUNNING")
	case "completed":
		return lipgloss.NewStyle().Foreground(green).Bold(true).Render("● COMPLETED")
	case "failed":
		return lipgloss.NewStyle().Foreground(red).Bold(true).Render("● FAILED")
	default:
		return lipgloss.NewStyle().Foreground(muted).Render("● " + strings.ToUpper(status))
	}
}

func countStatus(jobs []Job, status string) int {
	count := 0
	for _, j := range jobs {
		if strings.EqualFold(j.Status, status) {
			count++
		}
	}
	return count
}

func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if len([]rune(s)) <= width {
		return s
	}
	if width <= 3 {
		return string([]rune(s)[:width])
	}
	return string([]rune(s)[:width-3]) + "..."
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "GoQueue TUI error:", err)
		os.Exit(1)
	}
}
