package presets

import (
	"errors"
	"os"
	"sort"
)

// DefaultParser holds presets parsing data
type DefaultParser struct {
	Presets map[string]map[string]string
}

// Parser holds presets parsing logic
type Parser interface {
	Exists(string) bool
	GetPresets() []string
	LookUpFiles(string) []string
	WriteFiles(string) (string, error)
}

// Exists check if preset exists
func (p *DefaultParser) Exists(preset string) (exists bool) {
	_, exists = p.Presets[preset]
	return
}

// GetPresets get all presets names
func (p *DefaultParser) GetPresets() (presets []string) {
	if len(p.Presets) == 0 {
		return
	}

	for key := range p.Presets {
		presets = append(presets, key)
	}
	sort.Strings(presets)
	return
}

// ErrPresetWriteAllBytes error throwed when did not write all preset file bytes
var ErrPresetWriteAllBytes = errors.New("failed to write all bytes")

// LookUpFiles check if preset files exist
func (p *DefaultParser) LookUpFiles(preset string) (foundFiles []string) {
	presetFiles := p.Presets[preset]

	for fileName := range presetFiles {
		if _, err := os.Stat(fileName); !os.IsNotExist(err) {
			foundFiles = append(foundFiles, fileName)
		}
	}
	return
}

// WriteFiles write preset files
func (p *DefaultParser) WriteFiles(preset string) (fileError string, err error) {
	presetFiles := p.Presets[preset]

	for fileName, fileContent := range presetFiles {
		var (
			file  *os.File
			lines int
		)
		file, err = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)

		if err != nil {
			fileError = fileName
			return
		}

		if lines, err = file.Write([]byte(fileContent)); err != nil {
			fileError = fileName
			return
		}

		if len([]byte(fileContent)) != lines {
			fileError = fileName
			err = ErrPresetWriteAllBytes
			return
		}

		if err = file.Sync(); err != nil {
			fileError = fileName
			return
		}

		file.Close()
	}

	return
}
