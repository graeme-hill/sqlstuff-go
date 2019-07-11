package lib

import "fmt"

type Model struct {
	Tables map[string]*Table
}

type Table struct {
	Name    string
	Columns []ColumnDefinition
}

type ModelBuilder struct {
	model Model
}

func NewModelBuilder() *ModelBuilder {
	return &ModelBuilder{
		model: Model{
			Tables: map[string]*Table{},
		},
	}
}

func (m *ModelBuilder) handleStmt(stmt Statement) error {
	switch s := stmt.(type) {
	case CreateTable:
		return m.handleCreateTableStmt(s)
	case AddColumn:
		return m.handleAddColumnStmt(s)
	case DropColumn:
		return m.handleDropColumnStmt(s)
	default:
		// Ignore because we don't care about SELECT, INSERT, etc
		return nil
	}
}

func (m *ModelBuilder) handleCreateTableStmt(ct CreateTable) error {
	_, exists := m.model.Tables[ct.Name]
	if exists {
		return fmt.Errorf("Table named '%s' already exists", ct.Name)
	}

	tbl := &Table{
		Name:    ct.Name,
		Columns: ct.Columns,
	}

	m.model.Tables[tbl.Name] = tbl
	return nil
}

func (m *ModelBuilder) handleAddColumnStmt(ac AddColumn) error {
	tbl, tblExists := m.model.Tables[ac.TableName]
	if !tblExists {
		return fmt.Errorf(
			"Cannot add column '%s' because the table '%s' does not exist",
			ac.Column.Name,
			ac.TableName,
		)
	}

	for _, col := range tbl.Columns {
		if col.Name == ac.Column.Name {
			return fmt.Errorf(
				"Cannot add column '%s' to the table '%s' because the column already exists",
				ac.Column.Name,
				ac.TableName,
			)
		}
	}

	tbl.Columns = append(tbl.Columns, ac.Column)
	return nil
}

func (m *ModelBuilder) handleDropColumnStmt(dc DropColumn) error {
	tbl, tblExists := m.model.Tables[dc.TableName]
	if !tblExists {
		return fmt.Errorf(
			"Cannot drop column '%s' because the table '%s' does not exist",
			dc.ColumnName,
			dc.TableName,
		)
	}

	toRemove := -1
	for i, col := range tbl.Columns {
		if col.Name == dc.ColumnName {
			toRemove = i
			break
		}
	}

	if toRemove < 0 {
		return fmt.Errorf(
			"Cannot drop column '%s' from the table '%s' because the column does not exist",
			dc.ColumnName,
			dc.TableName,
		)
	}

	tbl.Columns = append(tbl.Columns[:toRemove], tbl.Columns[toRemove+1:]...)
	return nil
}

func ModelFromMigrations(migrations []*Migration) (Model, error) {
	builder := NewModelBuilder()
	for _, migration := range migrations {
		prog, err := Parse(migration.UpSQL)
		if err != nil {
			return Model{}, err
		}
		for _, stmt := range prog.Statements {
			err = builder.handleStmt(stmt)
			if err != nil {
				return Model{}, err
			}
		}
	}
	return builder.model, nil
}
