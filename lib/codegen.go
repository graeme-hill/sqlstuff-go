package lib

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"io"
	"text/template"
	"os"
)

var templateString = `// This code was generate by a tool.

package {{.Package}}

{{range .Batches}}
/******************************************************************************
 * {{.Name}}
 ****************************************************************************/

{{range .ResultTypes}}
type {{.Name}} struct {
  {{range .Columns}}
  {{.Name}} {{.Type}}
  {{end}}
}
{{end}}

func {{.FuncName}}() ({{range .ResultTypes}}{{.Name}}, {{end}}error) {
  // to do...
}
{{end}}
`

var numberSequence = regexp.MustCompile(`([a-zA-Z])(\d+)([a-zA-Z]?)`)
var numberReplacement = []byte(`$1 $2 $3`)

type codeGenViewModel struct {
	Package string
	Batches []batchViewModel
}

type batchViewModel struct {
	Name        string
	ResultTypes []resultTypeViewModel
	FuncName    string
}

type resultTypeViewModel struct {
	Name    string
	Columns []columnViewModel
}

type columnViewModel struct {
	Name string
	Type string
}

func Generate(migrationDir string, queryDir string, dest string, pkg string) error {
	migrations, err := ReadMigrationsDir(migrationDir)
	if err != nil {
		return err
	}

	model, err := ModelFromMigrations(migrations)
	if err != nil {
		return err
	}

	batches, err := ReadQueriesFromDir(queryDir, model)
	if err != nil {
		return err
	}

	fileWriter, err := os.Create(dest)
	if err != nil {
		return err
	}

	err = writeCode(fileWriter, pkg, batches)
	if err != nil {
		return err
	}

	return nil
}

func writeCode(writer io.Writer, pkg string, batches []QueryBatch) error {
	vm, err := newViewModel(pkg, batches)
	if err != nil {
		return err
	}

	tmpl, err := template.New("sql").Parse(templateString)
	if err != nil {
		return err
	}

	err = tmpl.Execute(writer, vm)
	return err
}

func newViewModel(pkg string, batches []QueryBatch) (codeGenViewModel, error) {
	vm := codeGenViewModel{
		Package: pkg,
		Batches: []batchViewModel{},
	}

	for _, qb := range batches {
		resultTypes := []resultTypeViewModel{}

		for i, shape := range qb.Shapes {
			rvm, err := newResultTypeViewModel(i, len(qb.Shapes), qb.Name, shape)
			if err != nil {
				return codeGenViewModel{}, err
			}

			resultTypes = append(resultTypes, rvm)
		}

		vm.Batches = append(vm.Batches, batchViewModel{
			Name:        qb.Name,
			FuncName:    pascalCase(qb.Name),
			ResultTypes: resultTypes,
		})
	}

	return vm, nil
}

func newResultTypeViewModel(index int, of int, batchName string, shape []ColumnDefinition) (resultTypeViewModel, error) {
	columns := []columnViewModel{}

	for _, c := range shape {
		typ, err := goType(c)
		if err != nil {
			return resultTypeViewModel{}, err
		}

		columns = append(columns, columnViewModel{
			Name: pascalCase(c.Name),
			Type: typ,
		})
	}

	numSuffix := ""
	if of > 1 {
		numSuffix = strconv.Itoa(index + 1)
	}

	return resultTypeViewModel{
		Name:    fmt.Sprintf("%sResult%s", pascalCase(batchName), numSuffix),
		Columns: columns,
	}, nil
}

func goType(def ColumnDefinition) (string, error) {
	switch def.Type {
	case DataTypeBigInt:
		return "int64", nil
	case DataTypeBoolean:
		return "bool", nil
	case DataTypeInteger:
		return "int32", nil
	case DataTypeVarChar:
		return "string", nil
	case DataTypeText:
		return "string", nil
	default:
		return "", fmt.Errorf("Unsupported type %v", def.Type)
	}
}

func addWordBoundariesToNumbers(s string) string {
	b := []byte(s)
	b = numberSequence.ReplaceAll(b, numberReplacement)
	return string(b)
}

func toCamelInitCase(s string, initCase bool) string {
	s = addWordBoundariesToNumbers(s)
	s = strings.Trim(s, " ")
	n := ""
	capNext := initCase
	for _, v := range s {
		if v >= 'A' && v <= 'Z' {
			n += string(v)
		}
		if v >= '0' && v <= '9' {
			n += string(v)
		}
		if v >= 'a' && v <= 'z' {
			if capNext {
				n += strings.ToUpper(string(v))
			} else {
				n += string(v)
			}
		}
		if v == '_' || v == ' ' || v == '-' {
			capNext = true
		} else {
			capNext = false
		}
	}
	return n
}

func pascalCase(s string) string {
	return toCamelInitCase(s, true)
}

func camelCase(s string) string {
	if s == "" {
		return s
	}
	if r := rune(s[0]); r >= 'A' && r <= 'Z' {
		s = strings.ToLower(string(r)) + s[1:]
	}
	return toCamelInitCase(s, false)
}
