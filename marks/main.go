package marks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Marks map[string]string

func ReadMarks(path string) (Marks, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Marks{}, nil
		}

		return nil, err
	}

	var m Marks
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return m, nil
}

func (m Marks) Write(path string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	data = append(data, '\n')

	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0666)
}

type FileMarks map[string]FileMark

func ReadFileMarks(path string) (FileMarks, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return FileMarks{}, nil
		}

		return nil, err
	}

	var m FileMarks
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return m, nil
}

func (m FileMarks) Write(path string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	data = append(data, '\n')

	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0666)
}

type FileMark struct {
	Path   string `json:"path"`
	Line   uint64 `json:"line"`
	Column uint64 `json:"col"`
}

func File(path string, line, col uint64) FileMark {
	return FileMark{path, line, col}
}

func (fm FileMark) String() string {
	return fmt.Sprintf("%s:%d:%d", fm.Path, fm.Line, fm.Column)
}
