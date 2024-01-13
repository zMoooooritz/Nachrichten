package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Settings     Settings     `yaml:"Settings,omitempty"`
	Keys         Keys         `yaml:"Keys,omitempty"`
	Applications Applications `yaml:"Application,omitempty"`
	Theme        Theme        `yaml:"Theme,omitempty"`
}

type Settings struct {
	HideHelpOnStartup bool `yaml:"HideHelpOnStartup"`
	PreloadThumbnails bool `yaml:"PreloadThumbnails"`
}

type Keys struct {
	Up              []string `yaml:"Up"`
	Down            []string `yaml:"Down"`
	Left            []string `yaml:"Left"`
	Right           []string `yaml:"Right"`
	Prev            []string `yaml:"Prev"`
	Next            []string `yaml:"Next"`
	Full            []string `yaml:"Full"`
	Start           []string `yaml:"Start"`
	End             []string `yaml:"End"`
	Quit            []string `yaml:"Quit"`
	ToggleThumbnail []string `yaml:"ToggleThumbnail"`
	OpenArticle     []string `yaml:"OpenArticle"`
	OpenVideo       []string `yaml:"OpenVideo"`
	OpenShortNews   []string `yaml:"OpenShortNews"`
	Help            []string `yaml:"Help"`
}

type Theme struct {
	PrimaryColor         string `yaml:"PrimaryColor"`
	ShadedColor          string `yaml:"ShadedColor"`
	HighlightColor       string `yaml:"HighlightColor"`
	HighlightShadedColor string `yaml:"HighlightShadedColor"`
	WarningColor         string `yaml:"WarningColor"`
	WarningShadedColor   string `yaml:"WarningShadedColor"`
	ReaderHighlightColor string `yaml:"ReaderHighlightColor"`
	ReaderHeadingColor   string `yaml:"ReaderHeadingColor"`
}

type Applications struct {
	Image Application `yaml:"Image,omitempty"`
	Audio Application `yaml:"Audio,omitempty"`
	Video Application `yaml:"Video,omitempty"`
	HTML  Application `yaml:"HTML,omitempty"`
}

type Application struct {
	Path string   `yaml:"Path"`
	Args []string `yaml:"Args"`
}

func Load(configFile string) (Configuration, error) {
	config := defaultConfiguration()
	// no config file supplied, use default values
	if configFile == "" {
		return config, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return Configuration{}, fmt.Errorf("Configuration error: %s", err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return Configuration{}, fmt.Errorf("Configuration error: %s", err)
	}

	return config, nil
}

func defaultConfiguration() Configuration {
	return Configuration{
		Settings: Settings{
			HideHelpOnStartup: false,
			PreloadThumbnails: false,
		},
		Keys:         defaultKeys(),
		Applications: Applications{},
		Theme:        gruvboxTheme(),
	}
}

func defaultKeys() Keys {
	return Keys{
		Up:              []string{"k", "up"},
		Down:            []string{"j", "down"},
		Left:            []string{"h", "left"},
		Right:           []string{"l", "right"},
		Prev:            []string{"shift+tab"},
		Next:            []string{"tab"},
		Full:            []string{"f"},
		Start:           []string{"g", "home"},
		End:             []string{"G", "end"},
		Quit:            []string{"q", "esc", "ctrl+c"},
		ToggleThumbnail: []string{"i"},
		OpenArticle:     []string{"o"},
		OpenVideo:       []string{"v"},
		OpenShortNews:   []string{"s"},
		Help:            []string{"?"},
	}
}

func gruvboxTheme() Theme {
	return Theme{
		PrimaryColor:         "#EBDBB2",
		ShadedColor:          "#928374",
		HighlightColor:       "#458588",
		HighlightShadedColor: "#83A598",
		WarningColor:         "#FB4934",
		WarningShadedColor:   "#CC241D",
		ReaderHighlightColor: "#FABD2F",
		ReaderHeadingColor:   "#8EC07C",
	}
}
