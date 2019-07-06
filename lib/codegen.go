package lib

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"io"
	"text/template"
	"bytes"
	"go/format"
	"io/ioutil"
)

var templateString = `// This code was generate by a tool.

package {{.Package}}

import (
  "database/sql"
  _ "github.com/lib/pq"
)

type DBClient interface {
	{{- range .Batches}}
	{{.FuncName}}() ({{range .Queries}}[]{{.Result.Name}}, {{end}}error)
	{{end -}}
	Close()
}

type SQLDBClient struct {
	db *sql.DB
}

func NewDBClient(connectionString string) (DBClient, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return SQLDBClient{}, err
	}

	return SQLDBClient{
		db: db,
	}, nil
}

func (client SQLDBClient) Close() {
	client.db.Close()
}

{{range .Batches}}
/******************************************************************************
 * {{.Name}}
 ****************************************************************************/

{{range .Queries}}
type {{.Result.Name}} struct {
  {{range .Result.Columns -}}
	{{.Name}} {{.Type}}
	{{end}}
}
{{end}}

func (client SQLDBClient) {{.FuncName}}() ({{range .Queries}}r{{.Index}} []{{.Result.Name}}, {{end}}err error) {
	sql := "{{.SQL}}"
	rows, err := client.db.Query(sql)
	if err != nil {
		return 
	}
	defer rows.Close()

	{{range .Queries}}
	{{if (gt .Index 1)}}
	if !rows.NextResultSet() {
		err = fmt.Errorf("Expecting more result sets: %v", rows.Err())
		return
	}
	{{end}}

	for rows.Next() {
		var (
			{{range .Result.Columns -}}
			{{.NameLower}} {{.Type}}
			{{end}}
		)
		err = rows.Scan({{range .Result.Columns}}{{if (gt .Index 1)}}, {{end}}&{{.NameLower}}{{end}})
		if err != nil {
			return
		}

		r{{.Index}} = append(r{{.Index}}, {{.Result.Name}}{
			{{range .Result.Columns -}}
			{{.Name}}: {{.NameLower}},
			{{end}}
		})
	}
	{{- end}}
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
	Queries []queryViewModel
	FuncName    string
	SQL string
}

type resultTypeViewModel struct {
	Name    string
	Columns []columnViewModel
}

type columnViewModel struct {
	Name string
	NameLower string
	Type string
	Index int
}

type queryViewModel struct {
	Result resultTypeViewModel
	Index int
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

	buf := bytes.Buffer{}
	err = writeCode(&buf, pkg, batches)
	if err != nil {
		return err
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		formatted = buf.Bytes()
		fmt.Printf("gofmt failed: %v", err)
	}

	err = ioutil.WriteFile(dest, formatted, 0644)
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
		queries := []queryViewModel{}

		for i, shape := range qb.Shapes {
			rvm, err := newResultTypeViewModel(i, len(qb.Shapes), qb.Name, shape)
			if err != nil {
				return codeGenViewModel{}, err
			}

			queries = append(queries, queryViewModel{
				Result: rvm,
				Index: i+1,
			})
		}

		vm.Batches = append(vm.Batches, batchViewModel{
			Name:        qb.Name,
			FuncName:    pascalCase(qb.Name),
			Queries: queries,
			SQL: formatSQLForGo(qb.SQL),
		})
	}

	return vm, nil
}

func newResultTypeViewModel(index int, of int, batchName string, shape []ColumnDefinition) (resultTypeViewModel, error) {
	columns := []columnViewModel{}

	for i, c := range shape {
		typ, err := goType(c)
		if err != nil {
			return resultTypeViewModel{}, err
		}

		columns = append(columns, columnViewModel{
			Name: pascalCase(c.Name),
			NameLower: camelCase(c.Name),
			Type: typ,
			Index: i+1,
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

func formatSQLForGo(sql string) string {
	return strings.ReplaceAll(strings.ReplaceAll(sql, "\"", "\"\""), "\n", "\\n")
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
