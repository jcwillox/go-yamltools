# YAML Tools

A set of utility methods to help with parsing and transforming YAML using [go-yaml/yaml](https://github.com/go-yaml/yaml).

## Install

```sh
go get -u github.com/jcwillox/go-yamltools
```

## Examples

### Advanced YAML Parsing

```go
import (
  "gopkg.in/yaml.v3"
  "github.com/jcwillox/go-yamltools"
)

type ConfigList []*Config
type Config struct {
	Path      string `yaml:",omitempty"`
	Force     bool
	Recursive bool
}

func (c *ConfigList) UnmarshalYAML(n *yaml.Node) error {
	// will ensure the order of the keys is preserved
	// by converting a map to a list of single key maps
	// has no effect if the root node is not a map
	//   key1: val1
	//   key2: val2
	// becomes
	//   - key1: val1
	//   - key2: val2
	n = yamltools.MapToSliceMap(n)
	// ensure the node is a list, i.e. if it was a string
	n = yamltools.EnsureList(n)
	// this type alias is unfortunately necessary otherwise when
	// we call `n.Decode` it will call `UnmarshalYAML` again
	type ConfigListT ConfigList
	return n.Decode((*ConfigListT)(c))
}

func (c *Config) UnmarshalYAML(n *yaml.Node) error {
	// will ensure the node is a map and that the value is not null
	//   key1: null
	//   > key1: {}
	//   string2
	//   > string2: {}
	n = yamltools.EnsureMapMap(n)
	// this a rather complex transformation used to flatten a map to
	// make it easier to use with structs
	//   ~/projects:
	//     key2: val2
	// becomes
	//   key2: val2
	//   path: ~/projects
	n = yamltools.MapKeyIntoValueMap(n, "path")
	type ConfigT Config
	return n.Decode((*ConfigT)(c))
}
```

### YAML Tag Handlers

This project adds support for YAML tags such as the `!include` tag to load in a YAML fragment. It also support the definition of custom tags.

**Include Tag**

```go
import (
  "gopkg.in/yaml.v3"
  "github.com/jcwillox/go-yamltools"
)

type Config struct {
	...
}

func (c *Config) UnmarshalYAML(n *yaml.Node) error {
	// will recursively load in any `!include` tags replacing the
	// node with the loaded fragment
	err := yamltools.LoadIncludeTag(n)
	if err != nil {
		return err
	}
	// support the `!include_dir_named` tag, which will load in all
	// files in a directory into a map of each filename to the fragment
	err = yamltools.LoadIncludeDirNamedTag(n)
	if err != nil {
		return err
	}
	// this type alias is unfortunately necessary otherwise when
	// we call `n.Decode` it will call `UnmarshalYAML` again
	type ConfigT Config
	return n.Decode((*ConfigT)(c))
}

```

**Custom Tags**

Custom tags can be defined using the [`HandleCustomTag`](<#func-handlecustomtag>) function, and use the [`LoadFileFragment`](<#func-loadfilefragment>) function which reads in, parses the file and returns a YAML node.

See the [`LoadIncludeDirNamedTag`](<#func-loadincludedirnamedtag>) function definition for a more advanced example.

```go
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
```

<!-- gomarkdoc:embed:start -->

## Index

- [func EnsureFlatList(n *yaml.Node) *yaml.Node](<#func-ensureflatlist>)
- [func EnsureList(n *yaml.Node) *yaml.Node](<#func-ensurelist>)
- [func EnsureMapMap(n *yaml.Node) *yaml.Node](<#func-ensuremapmap>)
- [func HandleCustomTag(n *yaml.Node, tag string, fn TagProcessor) error](<#func-handlecustomtag>)
- [func IsScalarMap(n *yaml.Node) bool](<#func-isscalarmap>)
- [func ListToMapVal(n *yaml.Node, key string) *yaml.Node](<#func-listtomapval>)
- [func LoadFileFragment(path string) (*yaml.Node, error)](<#func-loadfilefragment>)
- [func LoadIncludeDirNamedTag(n *yaml.Node) error](<#func-loadincludedirnamedtag>)
- [func LoadIncludeTag(n *yaml.Node) error](<#func-loadincludetag>)
- [func MapKeyIntoValueMap(n *yaml.Node, keyKey string) *yaml.Node](<#func-mapkeyintovaluemap>)
- [func MapKeys(n *yaml.Node) []string](<#func-mapkeys>)
- [func MapSplitKeyVal(n *yaml.Node, keyKey, valKey string) *yaml.Node](<#func-mapsplitkeyval>)
- [func MapToSliceMap(n *yaml.Node) *yaml.Node](<#func-maptoslicemap>)
- [func ParseBoolNode(n *yaml.Node) (value bool, ok bool)](<#func-parseboolnode>)
- [func ScalarToList(n *yaml.Node) *yaml.Node](<#func-scalartolist>)
- [func ScalarToMap(n *yaml.Node) *yaml.Node](<#func-scalartomap>)
- [func ScalarToMapVal(n *yaml.Node, key string) *yaml.Node](<#func-scalartomapval>)
- [type Fragment](<#type-fragment>)
  - [func (f *Fragment) UnmarshalYAML(n *yaml.Node) error](<#func-fragment-unmarshalyaml>)
- [type TagProcessor](<#type-tagprocessor>)


## func EnsureFlatList

```go
func EnsureFlatList(n *yaml.Node) *yaml.Node
```

EnsureFlatList will flatten nested lists of yaml nodes expects to be passed a yaml.SequenceNode

## func EnsureList

```go
func EnsureList(n *yaml.Node) *yaml.Node
```

EnsureList will ensure that the base node is a SequenceNode

```
key: val
=========
- key: val
```

## func EnsureMapMap

```go
func EnsureMapMap(n *yaml.Node) *yaml.Node
```

EnsureMapMap will ensure that the node is a map and that the value is not null

```
key: null
> key: {}
string
> string: {}
```

## func HandleCustomTag

```go
func HandleCustomTag(n *yaml.Node, tag string, fn TagProcessor) error
```

HandleCustomTag is used to define a custom YAML tags, it will recursively search YAML nodes for the tag and call the tag processor function.

## func IsScalarMap

```go
func IsScalarMap(n *yaml.Node) bool
```

IsScalarMap tests if n is a map that contains only scalar keys and values

## func ListToMapVal

```go
func ListToMapVal(n *yaml.Node, key string) *yaml.Node
```

ListToMapVal converts a sequence node to a mapping of \{key: node\} does nothing if n is not a sequence node.

```
[i1, i2]
key: [i1, i2]
```

## func LoadFileFragment

```go
func LoadFileFragment(path string) (*yaml.Node, error)
```

LoadFileFragment reads in and parses a given file returning a YAML node.

## func LoadIncludeDirNamedTag

```go
func LoadIncludeDirNamedTag(n *yaml.Node) error
```

LoadIncludeDirNamedTag recursively searches for the \!include\_dir\_named tag from the given node and will replace the tag node with map of filename to content for each file in the directory.

## func LoadIncludeTag

```go
func LoadIncludeTag(n *yaml.Node) error
```

LoadIncludeTag recursively searches for the \!include tag from the given node and will replace the tag node with content of the included file.

## func MapKeyIntoValueMap

```go
func MapKeyIntoValueMap(n *yaml.Node, keyKey string) *yaml.Node
```

MapKeyIntoValueMap if the value is a map moves the key into the map with the specified name

```
key1:
  key2: val2
============
key2: val2
keyKey: key1
```

## func MapKeys

```go
func MapKeys(n *yaml.Node) []string
```

MapKeys will return a list of the keys in a map, or an empty list if the node is not a map

## func MapSplitKeyVal

```go
func MapSplitKeyVal(n *yaml.Node, keyKey, valKey string) *yaml.Node
```

MapSplitKeyVal splits a maps key and val into their own maps using the specified keys

```
key: val
===========
keyKey: key
valKey: val
```

## func MapToSliceMap

```go
func MapToSliceMap(n *yaml.Node) *yaml.Node
```

MapToSliceMap converts a map to a slice of maps with one key each

```
key1:
  key2: val1
key3: val2
==============
- key1:
    key2: val1
- key3: val2
```

## func ParseBoolNode

```go
func ParseBoolNode(n *yaml.Node) (value bool, ok bool)
```

ParseBoolNode will parse a boolean node and return its value if the node is not a boolean node \`ok\` will be false. This is useful if you want to apply a different transform based on a boolean flag in the YAML.

## func ScalarToList

```go
func ScalarToList(n *yaml.Node) *yaml.Node
```

ScalarToList wraps a scalar node in a sequence node

```
string
========
- string
```

## func ScalarToMap

```go
func ScalarToMap(n *yaml.Node) *yaml.Node
```

ScalarToMap will convert a scalar node to a mapping of \{scalar: nil\}

```
string
========
string: null
```

## func ScalarToMapVal

```go
func ScalarToMapVal(n *yaml.Node, key string) *yaml.Node
```

ScalarToMapVal converts a scalar node to a mapping of \{key: node\} does nothing if n is not a scalar node.

```
string
> key: string
```

## type Fragment

Fragment is used to parse YAML into a node instead of an interface.

```go
type Fragment struct {
    Content *yaml.Node
}
```

### func \(\*Fragment\) UnmarshalYAML

```go
func (f *Fragment) UnmarshalYAML(n *yaml.Node) error
```

## type TagProcessor

```go
type TagProcessor = func(n *yaml.Node) error
```

<!-- gomarkdoc:embed:end -->