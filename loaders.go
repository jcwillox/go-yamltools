package yamltools

import (
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"path/filepath"
)

type TagProcessor = func(n *yaml.Node) error

// HandleCustomTag is used to define a custom YAML tags, it will recursively search YAML nodes
// for the tag and call the tag processor function.
func HandleCustomTag(n *yaml.Node, tag string, fn TagProcessor) error {
	if n.Tag == tag {
		err := fn(n)
		if err != nil {
			return err
		}
	} else {
		if n.Kind == yaml.SequenceNode {
			for _, n := range n.Content {
				err := HandleCustomTag(n, tag, fn)
				if err != nil {
					return err
				}
			}
		} else if n.Kind == yaml.MappingNode {
			// only need to check every second node (the values)
			for i := 1; i < len(n.Content); i += 2 {
				err := HandleCustomTag(n.Content[i], tag, fn)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Fragment is used to parse YAML into a node instead of an interface.
type Fragment struct {
	Content *yaml.Node
}

func (f *Fragment) UnmarshalYAML(n *yaml.Node) error {
	f.Content = n
	return nil
}

// LoadFileFragment reads in and parses a given file returning a YAML node.
func LoadFileFragment(path string) (*yaml.Node, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f Fragment
	err = yaml.Unmarshal(data, &f)
	if err != nil {
		return nil, err
	}
	return f.Content, nil
}

// LoadIncludeTag recursively searches for the !include tag from the given node
// and will replace the tag node with content of the included file.
func LoadIncludeTag(n *yaml.Node) error {
	return HandleCustomTag(n, "!include", func(n *yaml.Node) error {
		fragment, err := LoadFileFragment(n.Value)
		if err != nil {
			return err
		}
		*n = *fragment
		return nil
	})
}

// LoadIncludeDirNamedTag recursively searches for the !include_dir_named tag from the given node
// and will replace the tag node with map of filename to content for each file in the directory.
func LoadIncludeDirNamedTag(n *yaml.Node) error {
	return HandleCustomTag(n, "!include_dir_named", func(n *yaml.Node) error {
		content := make([]*yaml.Node, 0, 10)
		err := filepath.WalkDir(n.Value, func(path string, entry fs.DirEntry, err error) error {
			if path == n.Value {
				return nil
			}
			if entry.IsDir() {
				return nil
			}
			fragment, err := LoadFileFragment(path)
			if err != nil {
				return err
			}
			content = append(content, &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: fileNameWithoutExt(filepath.Base(path)),
			}, fragment)
			return nil
		})
		if err != nil {
			return err
		}
		*n = *&yaml.Node{
			Kind:    yaml.MappingNode,
			Tag:     "!!map",
			Content: content,
		}
		return nil
	})
}

func fileNameWithoutExt(path string) string {
	for i := len(path) - 1; i >= 0 && !os.IsPathSeparator(path[i]); i-- {
		if path[i] == '.' {
			return path[:i]
		}
	}
	return ""
}
