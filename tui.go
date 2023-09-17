package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	headerText         string = "Nachrichten"
	regionalHeaderText string = "Regional"
	nationalHeaderText string = "National"
	germanDateFormat   string = "15:04 02.01.06"
)

var (
	activeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("62")).BorderForeground(lipgloss.Color("62"))

	inactiveStyle = lipgloss.NewStyle()

	listActiveStyle = activeStyle.Copy().Padding(1, 1, 1, 1).Margin(0, 1, 0, 1).BorderStyle(lipgloss.RoundedBorder())

	listInactiveStyle = inactiveStyle.Copy().Padding(1, 1, 1, 1).Margin(0, 1, 0, 1).BorderStyle(lipgloss.RoundedBorder())

	readerTitleActiveStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return activeStyle.Copy().BorderStyle(b).Padding(0, 1)
	}()

	readerTitleInactiveStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return inactiveStyle.Copy().BorderStyle(b).Padding(0, 1)
	}()

	readerInfoActiveStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return readerTitleActiveStyle.Copy().BorderStyle(b)
	}()

	readerInfoInactiveStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return readerTitleInactiveStyle.Copy().BorderStyle(b)
	}()

	titleActiveStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("62")).
				Foreground(lipgloss.Color("230"))

	titleInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("230"))

	screenCentered = func(w, h int) lipgloss.Style {
		return lipgloss.NewStyle().
			Width(w).
			Align(lipgloss.Center).
			Height(h).
			AlignVertical(lipgloss.Center)
	}
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type keymap = struct {
	quit, enter, leave, next, prev, open, video, shortNews key.Binding
}

type Model struct {
	news               News
	keymap             keymap
	ready              bool
	lists              []list.Model
	listsActiveIndeces []int
	activeListIndex    int
	reader             viewport.Model
	spinner            spinner.Model
	focus              int
	readerFocused      bool
	width              int
	height             int
}

func (m *Model) InitLists(news [][]NewsEntry) {
	for i, n := range news {
		var items []list.Item
		for _, ne := range n {
			items = append(items, item{title: ne.TopLine, desc: ne.Title})
		}

		m.lists[i].SetItems(items)
		m.listsActiveIndeces = append(m.listsActiveIndeces, 0)
	}
}

func EmptyLists(count int) []list.Model {
	var lists []list.Model
	for i := 0; i < count; i++ {
		newList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
		newList.SetFilteringEnabled(false)
		newList.SetShowTitle(true)
		newList.SetShowStatusBar(false)
		newList.SetShowHelp(false)
		lists = append(lists, newList)
	}
	return lists
}

func NewDotSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return s
}

func InitialModel() Model {
	m := Model{
		keymap: keymap{
			quit: key.NewBinding(
				key.WithKeys("q", "esc", "ctrl+c"),
			),
			enter: key.NewBinding(
				key.WithKeys("l"),
			),
			leave: key.NewBinding(
				key.WithKeys("h"),
			),
			next: key.NewBinding(
				key.WithKeys("tab"),
			),
			prev: key.NewBinding(
				key.WithKeys("shift+tab"),
			),
			open: key.NewBinding(
				key.WithKeys("o"),
			),
			video: key.NewBinding(
				key.WithKeys("v"),
			),
			shortNews: key.NewBinding(
				key.WithKeys("s"),
			),
		},
		ready:              false,
		reader:             viewport.New(0, 0),
		spinner:            NewDotSpinner(),
		focus:              0,
		lists:              EmptyLists(2),
		listsActiveIndeces: []int{},
		activeListIndex:    0,
		width:              0,
		height:             0,
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(getNews(), m.spinner.Tick)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case loadedNews:
		m.news = News(msg)
		m.InitLists([][]NewsEntry{m.news.NationalNews, m.news.RegionalNews})
		for i := range m.lists {
			width, _ := m.listInnerDims()
			m.lists[i].Title = lipgloss.PlaceHorizontal(width, lipgloss.Center, headerText)
		}
		m.ready = true
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keymap.enter):
			m.readerFocused = true
		case key.Matches(msg, m.keymap.leave):
			m.readerFocused = false
		case key.Matches(msg, m.keymap.next):
			m.readerFocused = false
			m.activeListIndex = (m.activeListIndex + 1) % len(m.lists)
		case key.Matches(msg, m.keymap.prev):
			m.readerFocused = false
			m.activeListIndex = (len(m.lists) + m.activeListIndex - 1) % len(m.lists)
		case key.Matches(msg, m.keymap.open):
			article := m.SelectedArticle()
			_ = open_url(article.URL)
		case key.Matches(msg, m.keymap.video):
			article := m.SelectedArticle()
			_ = open_url(article.Video.VideoURLs.Big)
		case key.Matches(msg, m.keymap.shortNews):
			url, _ := getShortNewsURL()
			_ = open_url(url)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.reader.YPosition = m.readerHeaderHeight()

		for i := range m.lists {
			m.lists[i].SetSize(m.listOuterDims())
			width, _ := m.listInnerDims()
			m.lists[i].Title = lipgloss.PlaceHorizontal(width, lipgloss.Center, headerText)
		}

		m.reader.Width, m.reader.Height = m.readerDims()
	default:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	if !m.ready {
		return m, tea.Batch(cmds...)
	}

	if m.readerFocused {
		m.reader, cmd = m.reader.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.lists[m.activeListIndex], cmd = m.lists[m.activeListIndex].Update(msg)
		cmds = append(cmds, cmd)
		m.listsActiveIndeces[m.activeListIndex] = m.lists[m.activeListIndex].Index()
		m.reader.SetContent(ContentToText(m.SelectedArticle().Content, m.reader.Width))
	}

	return m, tea.Batch(cmds...)
}

func (m Model) listOuterDims() (int, int) {
	return m.width / 3, m.height - 5
}

func (m Model) listInnerDims() (int, int) {
	w, h := m.listOuterDims()
	return w - 6, h
}

func (m Model) listSelectorDims() (int, int) {
	w, h := m.listOuterDims()
	return w - 4, h
}

func (m Model) readerDims() (int, int) {
	lw, _ := m.listOuterDims()
	return m.width - lw - 7, m.height - m.readerHeaderHeight() - m.readerFooterHeight()
}

func (m Model) readerHeaderHeight() int {
	return lipgloss.Height(m.headerView("", ""))
}

func (m Model) readerFooterHeight() int {
	return lipgloss.Height(m.footerView())
}

func (m Model) SelectedArticle() NewsEntry {
	var article NewsEntry
	if m.activeListIndex == 0 {
		article = m.news.NationalNews[m.listsActiveIndeces[m.activeListIndex]]
	} else {
		article = m.news.RegionalNews[m.listsActiveIndeces[m.activeListIndex]]
	}
	return article
}

func (m Model) View() string {
	if !m.ready {
		content := fmt.Sprintf("%s Lade Nachrichten... press q to quit", m.spinner.View())
		return screenCentered(m.width, m.height).Render(content)
	}

	listHeader := m.listSelectorView([]string{nationalHeaderText, regionalHeaderText}, m.activeListIndex)
	listStyle := listActiveStyle
	if m.readerFocused {
		listStyle = listInactiveStyle
	}
	list := listStyle.Render(lipgloss.JoinVertical(lipgloss.Left, listHeader, m.lists[m.activeListIndex].View()))
	article := m.SelectedArticle()
	reader := fmt.Sprintf("%s\n%s\n%s", m.headerView(article.TopLine, article.Date.Format(germanDateFormat)), m.reader.View(), m.footerView())

	return lipgloss.JoinHorizontal(lipgloss.Top, list, reader)
}

func (m Model) listSelectorView(names []string, activeIndex int) string {
	width, _ := m.listSelectorDims()
	cellWidth := width / len(names)
	var widths []int
	for i := 0; i < len(names)-1; i++ {
		widths = append(widths, cellWidth)
	}
	widths = append(widths, width-(len(names)-1)*cellWidth)
	result := ""
	for i, n := range names {
		style := titleInactiveStyle
		if i == activeIndex {
			style = titleActiveStyle
		}
		result += style.Render(lipgloss.PlaceHorizontal(widths[i], lipgloss.Center, n))
	}
	return lipgloss.NewStyle().PaddingLeft(2).Render(result)
}

func (m Model) headerView(name string, date string) string {
	titleStyle := readerTitleInactiveStyle
	lineStyle := inactiveStyle
	dateStyle := readerInfoInactiveStyle
	if m.readerFocused {
		titleStyle = readerTitleActiveStyle
		lineStyle = activeStyle
		dateStyle = readerInfoActiveStyle
	}

	title := titleStyle.Render(name)
	date = dateStyle.Render(date)
	line := lineStyle.Render(strings.Repeat("─", max(0, m.reader.Width-lipgloss.Width(title)-lipgloss.Width(date))))

	return lipgloss.JoinHorizontal(lipgloss.Center, title, line, date)
}

func (m Model) footerView() string {
	infoStyle := readerInfoInactiveStyle
	lineStyle := inactiveStyle
	if m.readerFocused {
		infoStyle = readerInfoActiveStyle
		lineStyle = activeStyle
	}

	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.reader.ScrollPercent()*100))
	line := lineStyle.Render(strings.Repeat("─", max(0, m.reader.Width-lipgloss.Width(info))))

	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}
