package gostdoc

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/favclip/genbase"
)

// ErrNotTargetStruct shows argument is target struct.
var ErrNotTargetStruct = errors.New("struct is not target")

// ErrUnSupportedFormat shows argument is target struct.
var ErrUnSupportedFormat = errors.New("format is not supprted")

// ParseOptions means options to Parse function.
type ParseOptions struct {
	IgnoreStructSuffix []string
}

// OutputOptions represents output options
type OutputOptions struct {
	Format OutputFormatType
}

// OutputFormatType represents output format type
type OutputFormatType string

//
const (
	OutputFormatTypeTsv      OutputFormatType = "tsv"
	OutputFormatTypeTsvShort OutputFormatType = "tsvshort"
	//OutputFormatTypeJSON OutputFormatType = "json"
)

// BuildSource represents source code of assembling..
type BuildSource struct {
	g         *genbase.Generator
	pkg       *genbase.PackageInfo
	typeInfos genbase.TypeInfos
	Structs   []*BuildStruct
}

// BuildStruct represents struct of assembling..
type BuildStruct struct {
	parent   *BuildSource
	typeInfo *genbase.TypeInfo

	Fields []*BuildField
}

// BuildField represents field of BuildStruct.
type BuildField struct {
	parent    *BuildStruct
	fieldInfo *genbase.FieldInfo

	Name  string
	Embed bool
	Tag   []*BuildTag
}

// BuildTag represents tag of BuildField.
type BuildTag struct {
	field *BuildField
	Name  string
	Value string
}

// Parse construct *BuildSource from package & type information.
func Parse(pkg *genbase.PackageInfo, typeInfos genbase.TypeInfos, opts *ParseOptions) (*BuildSource, error) {
	bu := &BuildSource{
		g:         genbase.NewGenerator(pkg),
		pkg:       pkg,
		typeInfos: typeInfos,
	}

	for _, typeInfo := range typeInfos {
		err := bu.parseStruct(typeInfo, opts)
		if err == ErrNotTargetStruct || err == genbase.ErrNotStructType {
			continue
		} else if err != nil {
			return nil, err
		}
	}

	return bu, nil
}

func (b *BuildSource) parseStruct(typeInfo *genbase.TypeInfo, opts *ParseOptions) error {
	structType, err := typeInfo.StructType()
	if err != nil {
		return err
	}

	structName := typeInfo.Name()

	ignore := false
	for _, suf := range opts.IgnoreStructSuffix {
		if strings.HasSuffix(structName, suf) {
			ignore = true
			break
		}
	}
	if ignore {
		return ErrNotTargetStruct
	}

	st := &BuildStruct{
		parent:   b,
		typeInfo: typeInfo,
	}

	for _, fieldInfo := range structType.FieldInfos() {
		if len(fieldInfo.Names) == 0 {
			// embed
			typeName, err := ExprToTypeName(fieldInfo.Type)
			if err != nil {
				return err
			}

			err = b.parseField(st, typeInfo, fieldInfo, typeName, true)
			if err != nil {
				return err
			}
		} else {
			for _, nameIdent := range fieldInfo.Names {
				err := b.parseField(st, typeInfo, fieldInfo, nameIdent.Name, false)
				if err != nil {
					return err
				}
			}
		}
	}

	b.Structs = append(b.Structs, st)

	return nil
}

func (b *BuildSource) parseField(st *BuildStruct, typeInfo *genbase.TypeInfo, fieldInfo *genbase.FieldInfo, name string, embed bool) error {
	field := &BuildField{
		parent:    st,
		fieldInfo: fieldInfo,
		Name:      name,
		Embed:     embed,
	}
	st.Fields = append(st.Fields, field)

	field.Tag = make([]*BuildTag, 0, 10)

	// tags
	if fieldInfo.Tag == nil {
		return nil
	}

	tagText := fieldInfo.Tag.Value[1 : len(fieldInfo.Tag.Value)-1]
	tagKeys := genbase.GetKeys(tagText)
	structTag := reflect.StructTag(tagText)

	for _, key := range tagKeys {
		tag := &BuildTag{
			field: field,
			Name:  key,
			Value: structTag.Get(key),
		}

		field.Tag = append(field.Tag, tag)
	}

	return nil
}

// Emit generate wrapper code.
func (b *BuildSource) Emit(opts *OutputOptions) ([]byte, error) {
	for _, st := range b.Structs {
		err := st.emit(b.g, opts)
		if err != nil {
			return nil, err
		}
	}

	return b.g.Buf.Bytes(), nil
}

func (st *BuildStruct) emit(g *genbase.Generator, opts *OutputOptions) error {
	for _, field := range st.Fields {
		typeName, err := ExprToBaseTypeName(field.fieldInfo.Type)
		if err != nil {
			return err
		}

		tags := make([]string, len(field.Tag))
		for i, tag := range field.Tag {
			tags[i] = tag.TagString()
		}

		switch opts.Format {
		case OutputFormatTypeTsv:
			g.Printf("%s\t%s\t%s\t%s\t%s\n",
				st.Name(), typeName, field.Name, strings.Join(tags, " "), strings.TrimRight(field.fieldInfo.Comment.Text(), "\n"))
		case OutputFormatTypeTsvShort:
			g.Printf("%s\t%s\t%s\n", st.Name(), typeName, field.Name)
		//case OutputFormatTypeJSON:
		default:
			return ErrUnSupportedFormat
		}

	}

	g.Printf("\n")

	return nil
}

// Name returns struct type name.
func (st *BuildStruct) Name() string {
	return st.typeInfo.Name()
}

// IsPtr returns field type is pointer.
func (f *BuildField) IsPtr() bool {
	return f.fieldInfo.IsPtr()
}

// IsArray returns field type is array.
func (f *BuildField) IsArray() bool {
	return f.fieldInfo.IsArray()
}

// IsPtrArray returns field type is pointer array.
func (f *BuildField) IsPtrArray() bool {
	return f.fieldInfo.IsPtrArray()
}

// IsArrayPtr returns field type is array of pointer.
func (f *BuildField) IsArrayPtr() bool {
	return f.fieldInfo.IsArrayPtr()
}

// IsPtrArrayPtr returns field type is pointer of pointer array.
func (f *BuildField) IsPtrArrayPtr() bool {
	return f.fieldInfo.IsPtrArrayPtr()
}

// TagString build tag string.
func (tag *BuildTag) TagString() string {
	return fmt.Sprintf("%s:\"%s\"", tag.Name, tag.Value)
}
