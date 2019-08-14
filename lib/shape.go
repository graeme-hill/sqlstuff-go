package lib

import (
	"errors"
	"fmt"
	"strings"
)

// Enum for the possible return values of a query. Examples:
//   QueryResultTypeManyRows: SELECT foo FROM bar
//   QueryResultTypeOneRow: SELECT foo FROM bar LIMIT 1
//   QueryResultTypeCommand: INSERT INTO foo (bar) VALUES ('hello')
type queryResultType int

const (
	QueryResultTypeManyRows queryResultType = iota
	QueryResultTypeOneRow
	QueryResultTypeCommand
)

type TableUniqueConstraint struct {
	TableName        string
	UniqueConstraint Constraint
}

type Shape struct {
	Columns []ColumnDefinition
	Type    queryResultType
}

func getShape(stmt Statement, model Model) (Shape, error) {
	switch typed := stmt.(type) {
	case Select:
		return getSelectShape(typed, model)
	case Insert:
		return getInsertShape(typed, model)
	default:
		return Shape{}, errors.New("getShape not implemented for this type of statement yet")
	}
}

// A normal insert that looks like INSERT INTO foo (x) VALUES (1) will not have
// any result columns, but it you use RETURNING then it will behave like a
// SELECT.
func getInsertShape(query Insert, model Model) (Shape, error) {
	// Since RETURNS is not implemented yet always assume no fields are selected
	return Shape{
		Columns: []ColumnDefinition{},
		Type:    QueryResultTypeCommand,
	}, nil
}

// Returns the data types and names of the columns that will come out of the
// given query/model pair.
func getSelectShape(query Select, model Model) (Shape, error) {
	available, err := getAvailableColumns(query, model)
	if err != nil {
		return Shape{}, err
	}

	resultColumns := []ColumnDefinition{}
	for _, field := range query.Fields {
		def, err := fieldAsColumnDefinition(field, model, available)
		if err != nil {
			return Shape{}, err
		}
		resultColumns = append(resultColumns, def)
	}

	resultType, err := getSelectCardinality(query, model)
	if err != nil {
		return Shape{}, err
	}

	return Shape{
		Columns: resultColumns,
		Type:    resultType,
	}, nil
}

func conditionsCoverConstraint(constraint TableUniqueConstraint, conditions ...Condition) bool {
	fixed := []string{}
	for _, cond := range conditions {
		fixed = union(fixed, conditionFixedColumns(constraint, cond))
	}
	
	for _, col := range constraint.UniqueConstraint.Columns {
		found := false
		for _, fixedCol := range fixed {
			if fixedCol == col
		}
	}
}

// insersect([["a","b"], ["a","c","b"]]) -> ["a","b"]
func intersect(sets ...[]string) []string {
	counts := map[string]int{}

	for _, set := range sets {
		for _, col := range set {
			count, ok := counts[col]
			if !ok {
				count = 0
			}
			counts[col] = count + 1
		}
	}

	result := []string{}
	for col, count := range counts {
		if count == len(sets) {
			result = append(result, col)
		}
	}
	return result
}

// union([["a","b"], ["a","c","b"]]) -> ["a","b","c"]
func union(sets ...[]string) []string {
	found := map[string]struct{}{}
	for _, set := range sets {
		for _, col := range set {
			found[col] = struct{}{}
		}
	}

	result := []string{}
	for col := range found {
		result = append(result, col)
	}
	return result
}

func conditionFixedColumns(constraint TableUniqueConstraint, condition Condition) []string {
	// Break down multi-part logical conditions (eg: "a AND b" or "a OR b")
	logical, ok := condition.(LogicalCondition)
	if ok {
		if logical.Op == LogicalOpAnd {
			return union(
				conditionFixedColumns(constraint, logical.Left),
				conditionFixedColumns(constraint, logical.Right))
		}
		return intersect(
			conditionFixedColumns(constraint, logical.Left),
			conditionFixedColumns(constraint, logical.Right))
	}

	// Break down binary conditions (eg: 1+foo > 6)
	binary, ok := condition.(BinaryCondition)
	if ok {
		if binary.Op == BinaryCondOpEqual {
			l := binary.Left
			r := binary.Right
			return findFixedColumns(constraint, binary)
		}
		return []string{}
	}

	return []string{}
}

func findFixedColumns(constraint TableUniqueConstraint, cond BinaryCondition) []string {
	if cond.Op != BinaryCondOpEqual {
		return false
	}

	leftColumn, leftIsColumn := cond.Left.(ColumnExpression)
	rightColumn, rightIsColumn := cond.Right.(ColumnExpression)

	leftLiteral, leftIsLiteral := cond.Left.(Literal)
	rightLiteral, rightIsLiteral := cond.Right.(Literal)

	if leftIsColumn && rightIsLiteral {
		return []string{leftColumn.String()}
	} else if leftIsColumn && rightIsColumn {
		return []string{leftColumn.String(), rightColumn.String()}
	} else if leftIsLiteral && rightIsColumn {
		return []string{rightColumn.String()}
	}

	return []string{}
}

func anyConditionCoversConstraint(constraint TableUniqueConstraint, conditions ...Condition) bool {
	for _, cond := range conditions {
		if conditionCoversConstraint(constraint, cond) {
			return true
		}
	}
	return false
}

func allConditionsCoverConstraint(constraint TableUniqueConstraint, conditions ...Condition) bool {
	for _, cond := range conditions {
		if !conditionCoversConstraint(constraint, cond) {
			return false
		}
	}
	return true
}

// Returns true iff the conditions cover any of the constraints.
func checkConstraints(constraints []TableUniqueConstraint, conditions ...Condition) bool {
	for _, constraint := range constraints {
		if conditionsCoverConstraint(constraint, conditions...) {
			return true
		}
	}
	return false
}

// Returns true iff the given conditions are enough to select no more than one
// row from the given table.
func unique(m Model, tableName string, alias string, conditions ...Condition) (bool, error) {
	table, ok := m.Tables[tableName]
	if !ok {
		return false, fmt.Errorf("Unknown table '%s'", tableName)
	}

	name := alias
	if name == "" {
		name = tableName
	}

	uniqueConstraints := []TableUniqueConstraint{}
	for _, constraint := range table.Constraints {
		if constraint.IsUnique() {
			uniqueConstraints = append(uniqueConstraints, TableUniqueConstraint{
				TableName:        name,
				UniqueConstraint: constraint,
			})
		}
	}

	return checkConstraints(uniqueConstraints, conditions...), nil
}

// Can be a table, view, subselect, etc. Records which alias relates to which
// unique constraints.
type cardinalitySource struct {
	// The alias or inferred name used to refer to columns in this datab source.
	name string

	uniqueConstraints []TableUniqueConstraint
}

type cardinalityCalculator struct {
	model Model

	// The root table or subselect (ie: the FROM clause) and each JOIN are all
	// considered sources.
	sources []cardinalitySource

	// For each fully qualified column, keeps track of the other columns that are
	// engtangled. The clause "ON a.b=c.d" would entangle a.b and c.d.
	entanglements map[string][]string

	// Keeps track of fully qualified columns that are fixed. The clause
	// "WHERE u.id=1" would fix u.id.
	fixedCols map[string]struct{}
}

func newCardinalitySource(t TargetTable, m Model) (cardinalitySource, error) {

	// If there is a subselect
	if t.Subselect != nil {
		if t.Alias == "" {
			return cardinalitySource{}, errors.New("Subselects must have an alias")
		}
		virtualConstraints, err := getVirtualUniqueConstraints(*t.Subselect, m)
		if err != nil {
			return cardinalitySource{}, err
		}

		return cardinalitySource{
			name:              t.Alias,
			uniqueConstraints: virtualConstraints,
		}, nil
	}

	// If there is not a subselect and just a table name
	name := t.Alias
	if name == "" {
		name = t.TableName
	}

	tbl, ok := m.Tables[t.TableName]
	if !ok {
		return cardinalitySource{}, fmt.Errorf("Unknown table '%s'", t.TableName)
	}

	uniqueConstraints := getTableUniqueConstraints(tbl.Constraints, name)
	return cardinalitySource{
		name:              name,
		uniqueConstraints: uniqueConstraints,
	}, nil
}

func (c *cardinalityCalculator) addSource(t TargetTable) error {
	src, err := newCardinalitySource(t, c.model)
	if err != nil {
		return err
	}
	c.sources = append(c.sources, src)
	return nil
}

func (c *cardinalityCalculator) entangled(a string, b string) {
	existing, ok := c.entanglements[a]
	if ok {
		c.entanglements[a] = append(existing, b)
	} else {
		c.entanglements[a] = []string{b}
	}
}

func (c *cardinalityCalculator) fixed(a string) {
	c.fixedCols[a] = struct{}{}
}

func (c *cardinalityCalculator) setEntanglement(left Expression, right Expression) {
	leftCol, leftIsCol := left.(ColumnExpression)
	rightCol, rightIsCol := right.(ColumnExpression)

	if leftIsCol {
		if rightIsCol {
			c.entangled(leftCol.String(), rightCol.String())
			c.entangled(rightCol.String(), leftCol.String())
		} else {
			c.fixed(leftCol.String())
		}
	} else if rightIsCol {
		c.fixed(rightCol.String())
	}
}

func (c *cardinalityCalculator) updateEntanglements(cond Condition) {
	// TO DO: make this more sophisticated. Right now it only supports ANDed
	// equalities like this:
	//   u.id = x.user_id AND u.tid = x.tid
	// but it does not support ORs like this:
	//   (u.id = x.user_id AND u.tid = x.tid) OR (u.email = x.email AND u.tid = x.tid)

	// It's just a single binary expr like "1 > 2" or "u.email = $email"
	binary, ok := cond.(BinaryCondition)
	if ok {
		if binary.Op == BinaryCondOpEqual {
			c.setEntanglement(binary.Left, binary.Right)
		}
		return
	}

	// There is at least one logical op like "u.email = $email AND u.tid = $tid"
	logical, ok := cond.(LogicalCondition)
	if ok {
		if logical.Op == LogicalOpAnd {
			c.updateEntanglements(logical.Left)
			c.updateEntanglements(logical.Right)
		}
	}
}

func getEntanglements(cs cardinalitySource) map[string][]string {
	res := map[string][]string{}
	// TO DO
	return res
}

func getTableUniqueConstraints(constraints []Constraint, name string) []TableUniqueConstraint {
	uniqueConstraints := []TableUniqueConstraint{}
	for _, constraint := range constraints {
		if constraint.IsUnique() {
			uniqueConstraints = append(uniqueConstraints, TableUniqueConstraint{
				TableName:        name,
				UniqueConstraint: constraint,
			})
		}
	}
	return uniqueConstraints
}

func newCardinalityCalculator(s Select, m Model) (cardinalityCalculator, error) {
	calc := cardinalityCalculator{
		model: m,
	}

	err := calc.addSource(s.From)
	if err != nil {
		return cardinalityCalculator{}, err
	}

	for _, join := range s.Joins {
		err = calc.addSource(join.Target)
		if err != nil {
			return cardinalityCalculator{}, nil
		}
		calc.updateEntanglements(join.On)
	}

	calc.updateEntanglements(s.Where)

	return calc, nil
}

// Returns the unique definitions for this query.
func getVirtualUniqueConstraints(s Select, m Model) ([]TableUniqueConstraint, error) {

	/*
	   select u.name, u.email, u.id as uid, g.id as gid, g.name as gname (u.tid=ug.tid=g.tid) (u.id=ug.uid) (g.id=ug.gid)
	   from users u -> (tid, id) (tid, email)
	   join user_groups ug on ug.uid = u.id and ug.tid = u.tid (tid, uid, gid)!!
	   join groups g on g.id = ug.gid and g.tid = u.tid (tid, id) (tid, name)
	   where u.tid=$tid and u.email=$email

	   where u.tid=$tid and u.email=$email ---> (gid) | (gname)
	   where u.tid=$tid ---> (uid, gid) (uid, gname) (email, gid) (email, gname)

	   step 1: find unique constraints on each collection
	   step 2: map equivalent fields from on clauses
	   step 3: find matched constraints
	   step 4: mark all constrains from matched tables as matched
	   step 5: find linked constraints that are matched (loop until no change)

	*/

	// name := alias
	// if name == "" {
	// 	name = s.TableName
	// }

	// table, ok := m.Tables[s.TableName]
	// if !ok {
	// 	return nil, fmt.Errorf("Unknown table '%s'", s.TableName)
	// }

	// res := []TableUniqueConstraint{}
	// for _, c := range table.Constraints {
	// 	res = append(res, TableUniqueConstraint{
	// 		TableName: name,
	// 		UniqueConstraint: c,
	// 	})
	// }

	// return res, nil
	return nil, nil
}

// Determines whether the given join is a one or many join within the context
// of the given select (ie: it takes the parent select's where clause into
// account). TODO: also consider GROUP BY... and HAVING??
func getTargetCardinality(target TargetTable, s Select, m Model, conditions ...Condition) (queryResultType, error) {
	if target.Subselect != nil {
		// An arbitrarily complex sub-select
		subselect := *target.Subselect
		subCardinality, err := getSelectCardinality(subselect, m)
		if err != nil {
			return 0, err
		}
		if subCardinality == QueryResultTypeOneRow {
			// Even without looking at the ON or WHERE clauses of the parent query,
			// this subquery is already fetching max one row so no more work to do.
			return QueryResultTypeOneRow, nil
		}

		constraints, err := getVirtualUniqueConstraints(subselect, m)
		if err != nil {
			return 0, err
		}

		if checkConstraints(constraints, append(conditions, s.Where)...) {
			return QueryResultTypeOneRow, nil
		}
		return QueryResultTypeManyRows, nil
	} else {
		// Just a table name
		isUnique, err := unique(m, target.TableName, target.Alias, append(conditions, s.Where)...)
		if err != nil {
			return 0, err
		}
		if isUnique {
			return QueryResultTypeOneRow, err
		}
		return QueryResultTypeManyRows, err
	}
}

func getSelectCardinality(s Select, model Model) (queryResultType, error) {
	// If the top level query explcitly includes "LIMIT 1" then we know is a
	// single row result and we're done here.
	if s.Limit.HasLimit && s.Limit.Count == 1 {
		return QueryResultTypeOneRow, nil
	}

	// Evaluate the effect of each join on cardinality. If there is a join that
	// could select many rows even when filters from the parent select are
	// considered then we know it's a many result and we can stop processing.
	for _, join := range s.Joins {
		joinCardinality, err := getTargetCardinality(join.Target, s, model, join.On)
		if err != nil {
			return 0, err
		}
		if joinCardinality == QueryResultTypeManyRows {
			return QueryResultTypeManyRows, nil
		}
	}

	// Haven't ruled out a single row result yet so check the root table that is
	// being selected on.
	cardinality, err := getTargetCardinality(s.From, s, model)
	if err != nil {
		return 0, err
	}
	if cardinality == QueryResultTypeManyRows {
		return QueryResultTypeManyRows, nil
	}

	return QueryResultTypeOneRow, nil
}

func fieldAsColumnDefinition(
	field Field,
	model Model,
	available map[string][]ColumnDefinition,
) (ColumnDefinition, error) {
	def, err := exprAsColumnDefinition(field.Expr, model, available)
	if err != nil {
		return ColumnDefinition{}, err
	}

	if len(field.Alias) > 0 {
		def.Name = field.Alias
	}

	return def, nil
}

func exprAsColumnDefinition(
	expr Expression,
	model Model,
	available map[string][]ColumnDefinition,
) (ColumnDefinition, error) {
	switch typed := expr.(type) {
	case ColumnExpression:
		return findColumn(typed.TableName, typed.ColumnName, available)
	case FunctionExpression:
		return getFuncReturnType(typed)
	default:
		return ColumnDefinition{}, errors.New("Expression type not implemented yet")
	}
}

func getFuncReturnType(fnExpr FunctionExpression) (ColumnDefinition, error) {
	switch strings.ToUpper(fnExpr.FuncName) {
	case "COUNT":
		return ColumnDefinition{
			Type: DataTypeBigInt,
		}, nil
	default:
		return ColumnDefinition{}, fmt.Errorf("Function '%s' not supported", fnExpr.FuncName)
	}
}

func findColumn(
	table string,
	column string,
	available map[string][]ColumnDefinition,
) (ColumnDefinition, error) {
	if len(table) > 0 {
		return findAliasedColumn(table, column, available)
	}
	return findUnaliasedColumn(column, available)
}

func findAliasedColumn(
	table string,
	column string,
	available map[string][]ColumnDefinition,
) (ColumnDefinition, error) {
	defs, ok := available[table]
	if !ok {
		return ColumnDefinition{}, fmt.Errorf("Invalid table/alias '%s'", table)
	}

	for _, def := range defs {
		if def.Name == column {
			return def, nil
		}
	}

	return ColumnDefinition{}, fmt.Errorf("Column '%s' not found on '%s'", column, table)
}

func findUnaliasedColumn(
	column string,
	available map[string][]ColumnDefinition,
) (ColumnDefinition, error) {
	found := false
	result := ColumnDefinition{}

	for _, defs := range available {
		for _, def := range defs {
			if def.Name == column {
				if found {
					return ColumnDefinition{}, fmt.Errorf("Ambiguous column '%s'", column)
				}
				result = def
				found = true
			}
		}
	}

	if found {
		return result, nil
	}
	return ColumnDefinition{}, fmt.Errorf("Column '%s' not found", column)
}

func getAvailableColumns(
	query Select,
	model Model,
) (map[string][]ColumnDefinition, error) {
	// Map to store all columns available for select list
	available := map[string][]ColumnDefinition{}

	// FROM clause
	err := addTargetTable(available, model, query.From)
	if err != nil {
		return nil, err
	}

	// JOINs
	for _, join := range query.Joins {
		err = addTargetTable(available, model, join.Target)
		if err != nil {
			return nil, err
		}
	}

	return available, nil
}

func addTargetTable(
	available map[string][]ColumnDefinition,
	model Model,
	target TargetTable,
) error {
	if len(target.TableName) > 0 {
		// Just a table name (possibly aliased)
		key := target.TableName
		if len(target.Alias) > 0 {
			key = target.Alias
		}
		tbl, ok := model.Tables[target.TableName]
		if !ok {
			return fmt.Errorf("Unknown table '%s'", target.TableName)
		}

		_, exists := available[key]
		if exists {
			return fmt.Errorf("Duplicate alias '%s'", key)
		}

		available[key] = tbl.Columns
	} else {
		// With a subselect (must be aliased)
		if len(target.Alias) <= 0 {
			return errors.New("Subselect requires alias")
		}
		_, exists := available[target.Alias]
		if exists {
			return fmt.Errorf("Duplicate alias '%s'", target.Alias)
		}
		shape, err := getShape(*target.Subselect, model)
		if err != nil {
			return err
		}
		available[target.Alias] = shape.Columns
	}

	return nil
}
