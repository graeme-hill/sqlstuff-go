package lib

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func Parse(sql string) (Program, error) {
	buffer := newTokenBuffer()
	p := parser{reader: buffer, parameters: []Parameter{}}
	var lexErr error = nil

	go (func() {
		lexErr = lex(sql, buffer.Write)
		buffer.Done()
	})()

	prog, err := p.scan()
	if err == nil {
		err = lexErr
	}
	return prog, err
}

type parser struct {
	reader     tokenReader
	parameters []Parameter
}

func (p *parser) scan() (Program, error) {
	statements := []Statement{}
	requireSemicolon := false

	for {
		tok, done, err := p.reader.Next()
		if err != nil {
			return Program{}, err
		}

		if done {
			break
		}

		if requireSemicolon {
			if tok.tokType == tokenTypeSemicolon {
				requireSemicolon = false
				continue
			}
		}

		// SELECT...
		if isKeyword(tok, "SELECT") {
			selectStatement, err := p.scanSelect()
			if err != nil {
				return Program{}, err
			}
			statements = append(statements, selectStatement)
			requireSemicolon = true
			continue
		}

		if isKeyword(tok, "INSERT") {
			insertStatement, err := p.scanInsert()
			if err != nil {
				return Program{}, err
			}
			statements = append(statements, insertStatement)
			requireSemicolon = true
			continue
		}

		// CREATE ...
		if isKeyword(tok, "CREATE") {
			tok, err = p.requireToken(tokenTypeWord)
			if err != nil {
				return Program{}, err
			}

			// CREATE TABLE ...
			if isKeyword(tok, "TABLE") {
				createTableStatement, err := p.scanCreateTable()
				if err != nil {
					return Program{}, err
				}
				statements = append(statements, createTableStatement)
				requireSemicolon = true
				continue
			}

			return Program{}, fmt.Errorf("Invalid CREATE statement at %s", tokenString(tok))
		}

		// ALTER ...
		if isKeyword(tok, "ALTER") {
			tok, err = p.requireToken(tokenTypeWord)
			if err != nil {
				return Program{}, err
			}

			// ALTER TABLE ...
			if isKeyword(tok, "TABLE") {
				alterTableStatement, err := p.scanAlterTable()
				if err != nil {
					return Program{}, err
				}
				statements = append(statements, alterTableStatement)
				requireSemicolon = true
				continue
			}

			return Program{}, fmt.Errorf("Invalid ALTER statement at %s", tokenString(tok))
		}

		return Program{}, fmt.Errorf("Expecting start of statement but got <%s>", tokenString(tok))
	}

	return Program{
		Statements: statements,
		Parameters: p.parameters,
	}, nil
}

func isKeyword(tok token, keyword string) bool {
	return tok.tokType == tokenTypeWord && strings.EqualFold(string(tok.value), keyword)
}

// Reads after "INSERT"
func (p *parser) scanInsert() (Statement, error) {
	result := Insert{
		Columns: []ColumnExpression{},
		Values:  []Expression{},
	}

	// Make sure it starts with INSERT INTO...
	if !p.checkWord("INTO") {
		return Insert{}, errors.New("Expecting INTO after INSERT")
	}

	// INSERT INTO table_name...
	target, err := p.scanTargetTable()
	if err != nil {
		return Insert{}, err
	}
	if target.Subselect != nil {
		return Insert{}, errors.New("Cannot insert into a subselect")
	}
	result.Target = target

	// INSERT INTO table_name (...
	_, err = p.requireToken(tokenTypeLParen)
	if err != nil {
		return Insert{}, err
	}

	// INSERT INTO table_name (col1, col2...
	exprs, err := p.scanExprList()
	if err != nil {
		return Insert{}, err
	}
	for _, expr := range exprs {
		colExpr, ok := expr.(ColumnExpression)
		if !ok {
			return Insert{}, errors.New("Invalid target column in INSERT")
		}
		if colExpr.TableName == "" {
			if target.Alias == "" {
				colExpr.TableName = target.TableName
			} else {
				colExpr.TableName = target.Alias
			}
		}
		result.Columns = append(result.Columns, colExpr)
	}

	// INSERT INTO table_name (col1, col2)...
	_, err = p.requireToken(tokenTypeRParen)
	if err != nil {
		return Insert{}, nil
	}

	// INSERT INTO table_name (col1, col2) VALUES...
	if !p.checkWord("VALUES") {
		return Insert{}, errors.New("Expecting VALUES")
	}

	// INSERT INTO table_name (col1, col2) VALUES (...
	_, err = p.requireToken(tokenTypeLParen)
	if err != nil {
		return Insert{}, err
	}

	// INSERT INTO table_name (col1, col2) VALUES ('one', 'two'...
	valExprs, err := p.scanExprList()
	if err != nil {
		return Insert{}, err
	}
	result.Values = valExprs

	// INSERT INTO table_name (col1, col2) VALUES ('one', 'two')
	_, err = p.requireToken(tokenTypeRParen)
	if err != nil {
		return Insert{}, nil
	}

	return result, nil
}

// Reads after "ALTER TABLE"
func (p *parser) scanAlterTable() (Statement, error) {

	// get table name
	nameTok, err := p.requireToken(tokenTypeWord)
	if err != nil {
		return AddColumn{}, nil
	}

	next, err := p.requireToken(tokenTypeWord)
	if err != nil {
		return AddColumn{}, nil
	}

	// ALTER TABLE ___ ADD ...
	if isKeyword(next, "ADD") {
		next, err = p.requireToken(tokenTypeWord)
		if err != nil {
			return AddColumn{}, nil
		}

		// ALTER TABLE ___ ADD COLUMN ...
		if isKeyword(next, "COLUMN") {
			def, _, err := p.scanColumnDef()
			if err != nil {
				return AddColumn{}, err
			}
			return AddColumn{
				TableName: string(nameTok.value),
				Column:    def,
			}, nil
		}
	}

	return AddColumn{}, fmt.Errorf("Unsupported ALTER TABLE statement at <%s>", tokenString(next))
}

// Reads after "CREATE TABLE"
func (p *parser) scanCreateTable() (CreateTable, error) {

	var more bool

	// get table name
	nameTok, err := p.requireToken(tokenTypeWord)
	if err != nil {
		return CreateTable{}, err
	}
	createTable := CreateTable{
		Name:    string(nameTok.value),
		Columns: []ColumnDefinition{},
	}

	// look for '(' to start column list
	_, err = p.requireToken(tokenTypeLParen)
	if err != nil {
		return CreateTable{}, err
	}

	// columns
	for {
		// Stop looping if hit ')'. There may be zero columns!
		_, foundRParen := p.peekToken(tokenTypeRParen)
		if foundRParen {
			_, _, err := p.reader.Next()
			if err != nil {
				return CreateTable{}, err
			}
			break
		}

		// Skip PRIMARY KEY because we don't need that info
		if p.checkWord("PRIMARY") {
			more, err = p.skipColumnDef()
			if err != nil {
				return CreateTable{}, nil
			}
			if !more {
				break
			}
		}

		// Skip CONSTRAINT because we don't need that info
		if p.checkWord("CONSTRAINT") {
			more, err = p.skipColumnDef()
			if err != nil {
				return CreateTable{}, nil
			}
			if !more {
				break
			}
		}

		col, more, err := p.scanColumnDef()
		if err != nil {
			return CreateTable{}, err
		}
		createTable.Columns = append(createTable.Columns, col)

		if !more {
			break
		}
	}

	return createTable, nil
}

func (p *parser) skipColumnDef() (more bool, err error) {
	parenCount := 0
	for {
		next, done, err := p.reader.Next()
		if done {
			return false, errors.New("Unexpected EOF in CREATE TABLE")
		}
		if err != nil {
			return false, err
		}

		if next.tokType == tokenTypeLParen {
			parenCount++
		} else if next.tokType == tokenTypeRParen {
			if parenCount <= 0 {
				return false, nil
			} else {
				parenCount--
			}
		} else if next.tokType == tokenTypeComma && parenCount <= 0 {
			return true, nil
		}
	}
}

// Reads an expression like "first_name VARCHAR(200) NOT NULL"
func (p *parser) scanColumnDef() (ColumnDefinition, bool, error) {
	def := ColumnDefinition{}
	colNameTok, err := p.requireToken(tokenTypeWord)
	if err != nil {
		return ColumnDefinition{}, false, err
	}
	def.Name = string(colNameTok.value)

	typ, err := p.scanDataType()
	if err != nil {
		return ColumnDefinition{}, false, err
	}
	def.Type = typ

	err = p.applyTypeParams(&def)
	if err != nil {
		return ColumnDefinition{}, false, err
	}

	err = p.applyConstraints(&def)
	if err != nil {
		return ColumnDefinition{}, false, err
	}

	next, done, err := p.reader.Next()
	if err != nil {
		return ColumnDefinition{}, false, err
	}
	more := !done && next.tokType == tokenTypeComma

	return def, more, nil
}

func (p *parser) applyTypeParams(def *ColumnDefinition) error {
	_, hasParams := p.checkToken(tokenTypeLParen)
	if hasParams {
		tok, err := p.requireToken(tokenTypeNumber)
		if err != nil {
			return err
		}
		num, err := strconv.Atoi(string(tok.value))
		if err != nil {
			return err
		}
		def.Param1 = num

		_, comma := p.checkToken(tokenTypeComma)
		if comma {
			tok, err = p.requireToken(tokenTypeNumber)
			if err != nil {
				return err
			}
			num, err = strconv.Atoi(string(tok.value))
			if err != nil {
				return err
			}
			def.Param2 = num
		}

		_, err = p.requireToken(tokenTypeRParen)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *parser) applyConstraints(def *ColumnDefinition) error {
	// Defaults
	def.Nullable = true

	// Flags to avoid duplicate constraints
	alreadyNullable := false
	alreadyPrimaryKey := false

	didAnything := false

	for {
		if p.checkWord("NULL") {
			didAnything = true
			if alreadyNullable {
				return errors.New("Cannot specify null constraint more than once")
			}
			alreadyNullable = true
			def.Nullable = true
		}

		if p.checkWord("NOT") {
			didAnything = true
			if p.checkWord("NULL") {
				if alreadyNullable {
					return errors.New("Cannot specify null constraint more than once")
				}
				alreadyNullable = true
				def.Nullable = false
			} else {
				return errors.New("Expecting 'NOT' to be followed by 'NULL' but was not.")
			}
		}

		if p.checkWord("PRIMARY") {
			didAnything = true
			if p.checkWord("KEY") {
				if alreadyPrimaryKey {
					return errors.New("Cannot specify PRIMARY KEY more than once")
				}
				alreadyPrimaryKey = true
				// don't actually record this, just validate
			}
		}

		if !didAnything {
			break
		}
		didAnything = false
	}

	return nil
}

func (p *parser) scanDataType() (dataType, error) {
	if p.checkWord("INT") {
		return DataTypeInteger, nil
	}
	if p.checkWord("VARCHAR") {
		return DataTypeVarChar, nil
	}
	if p.checkWord("UUID") {
		return DataTypeUUID, nil
	}
	if p.checkWord("TIMESTAMPTZ") {
		return DataTypeTimestampWithTimeZone, nil
	}
	if p.checkWord("JSONB") {
		return DataTypeBinaryJSON, nil
	}
	next, _, _ := p.reader.Peek()
	return 0, fmt.Errorf("Unknown data type <%s>", tokenString(next))
}

func (p *parser) requireToken(tokType tokenType) (token, error) {
	next, done, err := p.reader.Next()
	if err != nil {
		return token{}, err
	}
	if done {
		return token{}, fmt.Errorf(
			"Expected '%s' but got EOF",
			tokenValueString(token{tokType: tokType}))
	}
	if next.tokType != tokType {
		return token{}, fmt.Errorf(
			"Expected '%s' but got <%s>",
			tokenValueString(token{tokType: tokType}),
			tokenString(next))
	}
	return next, nil
}

func (p *parser) advance() error {
	_, _, err := p.reader.Next()
	return err
}

func (p *parser) checkWord(word string) bool {
	next, done, err := p.reader.Peek()
	if err != nil || done || next.tokType != tokenTypeWord || !strings.EqualFold(string(next.value), word) {
		return false
	}
	_, _, _ = p.reader.Next()
	return true
}

func (p *parser) peekToken(tokType tokenType) (token, bool) {
	next, done, err := p.reader.Peek()
	if err != nil {
		return token{}, false
	}
	if done {
		return token{}, false
	}
	if next.tokType != tokType {
		return token{}, false
	}
	return next, true
}

func (p *parser) checkToken(tokType tokenType) (token, bool) {
	tok, found := p.peekToken(tokType)
	if found {
		_ = p.advance()
	}
	return tok, found
}

func (p *parser) scanSelect() (Select, error) {
	fields, err := p.scanFieldList()
	if err != nil {
		return Select{}, err
	}

	target, err := p.scanTargetTable()
	if err != nil {
		return Select{}, err
	}

	joins, err := p.scanJoins()
	if err != nil {
		return Select{}, err
	}

	where, err := p.scanWhere()
	if err != nil {
		return Select{}, err
	}

	having, err := p.scanHaving()
	if err != nil {
		return Select{}, err
	}

	limit, err := p.scanLimit()
	if err != nil {
		return Select{}, err
	}

	return Select{
		Fields: fields,
		From:   target,
		Joins:  joins,
		Where:  where,
		Having: having,
		Limit:  limit,
	}, nil
}

func (p *parser) scanLimit() (Limit, error) {
	if !p.checkWord("LIMIT") {
		return Limit{}, nil
	}

	// We are in a LIMIT clause so expect a number or 'ALL'
	next, done, err := p.reader.Next()
	if err != nil {
		return Limit{}, err
	}
	if done {
		return Limit{}, errors.New("Expecting LIMIT value but got EOF")
	}
	if next.tokType == tokenTypeNumber {
		i, err := strconv.Atoi(string(next.value))
		if err != nil {
			return Limit{}, fmt.Errorf("Cannot convert LIMIT %s to int", string(next.value))
		}
		return Limit{
			HasLimit: true,
			Count:    i,
		}, nil
	} else if isKeyword(next, "ALL") {
		return Limit{}, nil
	}
	return Limit{}, fmt.Errorf("Invalid LIMIT value <%s>", tokenString(next))
}

// Reads a comma-separated list of expressions. eg:
//   1+1, 'hello', foo.bar
// The expressions can NOT have aliases like ""'hello' AS name"
func (p *parser) scanExprList() ([]Expression, error) {
	exprs := []Expression{}
	for {
		expr, err := p.scanExpr()
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)

		next, done, err := p.reader.Peek()
		if err != nil {
			return nil, err
		} else if done {
			break
		}

		if next.tokType != tokenTypeComma {
			break
		}
		_ = p.advance()
	}
	return exprs, nil
}

func (p *parser) scanFieldList() ([]Field, error) {
	fields := []Field{}
	for {
		expr, err := p.scanExpr()
		if err != nil {
			return nil, err
		}

		next, done, err := p.reader.Next()
		if done {
			return nil, errors.New("SELECT statement missing FROM")
		} else if err != nil {
			return nil, err
		}

		if next.tokType == tokenTypeComma {
			// it's a field with no alias, and there are more fields to come
			fields = append(fields, Field{Expr: expr})
			continue
		} else if isKeyword(next, "AS") {
			alias, err := p.scanAlias()
			if err != nil {
				return nil, err
			}
			fields = append(fields, Field{Expr: expr, Alias: alias})

			next, done, err = p.reader.Next()
			if done {
				return nil, errors.New("SELECT statement missing FROM")
			} else if err != nil {
				return nil, err
			}

			if next.tokType == tokenTypeComma {
				continue
			} else if isKeyword(next, "FROM") {
				break
			} else {
				return nil, errors.New("SELECT statement missing FROM")
			}
		} else if isKeyword(next, "FROM") {
			fields = append(fields, Field{Expr: expr})
			break
		} else {
			return nil, errors.New("SELECT statement missing FROM")
		}
	}
	return fields, nil
}

func (p *parser) scanWhere() (Condition, error) {
	if p.checkWord("WHERE") {
		cond, err := p.scanCondition()
		if err != nil {
			return NullCondition{}, err
		}
		return cond, nil
	}

	return NullCondition{}, nil
}

func (p *parser) scanHaving() (Condition, error) {
	if p.checkWord("HAVING") {
		cond, err := p.scanCondition()
		if err != nil {
			return NullCondition{}, err
		}
		return cond, nil
	}

	return NullCondition{}, nil
}

func (p *parser) scanJoins() ([]Join, error) {
	joins := []Join{}

	for {
		join := Join{}

		// Find join type (or detect that there are no more joins)
		if p.checkWord("LEFT") {
			p.checkWord("OUTER")
			if p.checkWord("JOIN") {
				join.Type = JoinTypeLeftOuter
			} else {
				return nil, errors.New("Expected 'JOIN' after 'LEFT [OUTER]'")
			}
		} else if p.checkWord("RIGHT") {
			p.checkWord("OUTER")
			if p.checkWord("JOIN") {
				join.Type = JoinTypeRightOuter
			} else {
				return nil, errors.New("Expected 'JOIN' after 'RIGHT [OUTER]'")
			}
		} else if p.checkWord("JOIN") {
			join.Type = JoinTypeInner
		} else if p.checkWord("INNER") {
			if p.checkWord("JOIN") {
				join.Type = JoinTypeInner
			} else {
				return nil, errors.New("Expected 'JOIN' after 'INNER'")
			}
		} else {
			break
		}

		// If got here then it is a join and we know the type, so now parse the
		// thing it is joining with.
		joinTarget, err := p.scanTargetTable()
		if err != nil {
			return nil, err
		}
		join.Target = joinTarget

		// Now parse the join condition (the ON clause).
		if !p.checkWord("ON") {
			return nil, errors.New("Missing ON clause in JOIN")
		}
		cond, err := p.scanCondition()
		if err != nil {
			return nil, err
		}
		join.On = cond

		joins = append(joins, join)
	}

	return joins, nil
}

func (p *parser) scanCondition() (Condition, error) {
	var left Condition
	left, err := p.scanBinaryCondition()
	if err != nil {
		return BinaryCondition{}, err
	}

	for {
		if p.checkWord("AND") {
			right, err := p.scanBinaryCondition()
			if err != nil {
				return BinaryCondition{}, err
			}
			left = LogicalCondition{
				Left:  left,
				Right: right,
				Op:    LogicalOpAnd,
			}
		} else if p.checkWord("OR") {
			right, err := p.scanBinaryCondition()
			if err != nil {
				return BinaryCondition{}, err
			}
			left = LogicalCondition{
				Left:  left,
				Right: right,
				Op:    LogicalOpOr,
			}
		} else {
			break
		}
	}

	return left, nil
}

func (p *parser) scanBinaryCondition() (BinaryCondition, error) {
	// Left
	left, err := p.scanExpr()
	if err != nil {
		return BinaryCondition{}, nil
	}

	// Operator
	next, done, err := p.reader.Next()
	if err != nil {
		return BinaryCondition{}, nil
	}
	if done {
		return BinaryCondition{}, errors.New("Expecting operator but got EOF")
	}

	op, err := getBinaryConditionOperator(next)
	if err != nil {
		return BinaryCondition{}, nil
	}

	// Right
	right, err := p.scanExpr()
	if err != nil {
		return BinaryCondition{}, nil
	}

	return BinaryCondition{
		Left:  left,
		Right: right,
		Op:    op,
	}, nil
}

func getBinaryConditionOperator(tok token) (binaryCondOpType, error) {
	switch tok.tokType {
	case tokenTypeLess:
		return BinaryCondOpLessThan, nil
	case tokenTypeLessOrEqual:
		return BinaryCondOpLessThanOrEqual, nil
	case tokenTypeGreater:
		return BinaryCondOpGreatThan, nil
	case tokenTypeGreaterOrEqual:
		return BinaryCondOpGreatThanOrEqual, nil
	case tokenTypeEqual:
		return BinaryCondOpEqual, nil
	case tokenTypeNotEqual:
		return BinaryCondOpNotEqual, nil
	case tokenTypeWord:
		if strings.EqualFold("IS", string(tok.value)) {
			return BinaryCondOpIs, nil
		}
	}

	return 0, fmt.Errorf("Unknown binary condition operator at <%s>", tokenString(tok))
}

func (p *parser) scanTargetTable() (TargetTable, error) {
	next, done, err := p.reader.Next()
	if err != nil {
		return TargetTable{}, err
	}
	if done {
		return TargetTable{}, fmt.Errorf("Expecting select target but got <%s>", tokenString(next))
	}

	target := TargetTable{}

	if next.tokType == tokenTypeWord {
		target.TableName = string(next.value)
	} else if next.tokType == tokenTypeLParen {
		next, done, err = p.reader.Next()
		if err != nil {
			return TargetTable{}, err
		}
		if done {
			return TargetTable{}, errors.New("Expecting SELECT but got EOF")
		}

		if isKeyword(next, "SELECT") {
			subSelect, err := p.scanSelect()
			if err != nil {
				return TargetTable{}, err
			}
			target.Subselect = &subSelect

			next, done, err := p.reader.Next()
			if err != nil {
				return TargetTable{}, err
			}
			if done {
				return TargetTable{}, errors.New("Expecting ')' but got EOF")
			}
			if next.tokType != tokenTypeRParen {
				return TargetTable{}, fmt.Errorf("Expecting ')' but got <%s>", tokenString(next))
			}
		} else {
			return TargetTable{}, errors.New("Parenthesis in FROM clause must contain sub-select")
		}
	}

	next, done, err = p.reader.Peek()
	if err != nil {
		return TargetTable{}, err
	}
	if done {
		return target, nil
	}

	if isTargetTableAlias(next) {
		target.Alias = string(next.value)
		err = p.advance()
		if err != nil {
			return TargetTable{}, nil
		}
	}

	return target, nil
}

func isTargetTableAlias(tok token) bool {
	if tok.tokType != tokenTypeWord {
		return false
	}

	switch strings.ToUpper(string(tok.value)) {
	case "INNER":
		return false
	case "RIGHT":
		return false
	case "LEFT":
		return false
	case "FULL":
		return false
	case "CROSS":
		return false
	case "WHERE":
		return false
	case "HAVING":
		return false
	case "GROUP":
		return false
	case "ORDER":
		return false
	case "LIMIT":
		return false
	default:
		return true
	}
}

func (p *parser) scanAlias() (string, error) {
	next, done, err := p.reader.Next()
	if err != nil {
		return "", err
	}

	if done {
		return "", errors.New("Missing alias")
	}

	if next.tokType != tokenTypeWord {
		return "", errors.New("Missing alias")
	}

	if !isValidAlias(next.value) {
		return "", fmt.Errorf("Invalid alias '%s'", string(next.value))
	}

	return string(next.value), nil
}

func isValidAlias(alias []rune) bool {
	return len(alias) > 0
}

/*
 ______                              _                _____ _          __  __
|  ____|                            (_)              / ____| |        / _|/ _|
| |__  __  ___ __  _ __ ___  ___ ___ _  ___  _ __   | (___ | |_ _   _| |_| |_
|  __| \ \/ / '_ \| '__/ _ \/ __/ __| |/ _ \| '_ \   \___ \| __| | | |  _|  _|
| |____ >  <| |_) | | |  __/\__ \__ \ | (_) | | | |  ____) | |_| |_| | | | |
|______/_/\_\ .__/|_|  \___||___/___/_|\___/|_| |_| |_____/ \__|\__,_|_| |_|
            | |
            |_|
*/
func (p *parser) scanExpr() (Expression, error) {
	left, err := p.scanSubExpr()
	if err != nil {
		return ColumnExpression{}, err
	}

	for {
		opToken, done, err := p.reader.Peek()
		if err != nil {
			return ColumnExpression{}, err
		}
		if done {
			break
		}

		opType, isOp := getExprBinaryOpType(opToken)
		if !isOp {
			break
		}

		_, _, err = p.reader.Next()
		if err != nil {
			return ColumnExpression{}, nil
		}

		right, err := p.scanSubExpr()
		if err != nil {
			return ColumnExpression{}, err
		}

		left = binaryExprTreeAppend(left, right, opType)
	}

	return left, nil
}

func (p *parser) foundParameter(param Parameter) {
	for _, existing := range p.parameters {
		if existing.Name == param.Name {
			// Parameter already exists so do nothing
			return
		}
	}
	p.parameters = append(p.parameters, param)
}

func (p *parser) scanSubExpr() (Expression, error) {
	tok, done, err := p.reader.Next()
	if err != nil {
		return ColumnExpression{}, err
	}
	if done {
		return ColumnExpression{}, errors.New("Expecting expression but found EOF")
	}

	// Parameters (eg: $foo)
	if tok.tokType == tokenTypeParameter {
		param := ParameterExpression{
			Name: string(tok.value),
		}
		p.foundParameter(Parameter{Name: param.Name})
		return param, nil
	}

	// Number literals
	if tok.tokType == tokenTypeNumber {
		return NumberLiteral{
			Value: string(tok.value),
		}, nil
	}

	// String literals
	if tok.tokType == tokenTypeString {
		return StringLiteral{
			Value: string(tok.value),
		}, nil
	}

	// Unary expressions
	if tok.tokType == tokenTypeMinus {
		right, err := p.scanSubExpr()
		if err != nil {
			return ColumnExpression{}, err
		}

		return UnaryExpression{
			Right: right,
			Op:    UnaryExprOpNegative,
		}, nil
	}

	// Parentheticals
	if tok.tokType == tokenTypeLParen {
		return p.scanParenthetical()
	}

	// Columns and functions calls
	if tok.tokType == tokenTypeWord {
		return p.scanColumnOrCall(tok)
	}

	// Not recognized so it must be a syntax error
	return ColumnExpression{}, fmt.Errorf("Unexpected <%s>", tokenString(tok))
}

func (p *parser) scanColumnOrCall(firstToken token) (Expression, error) {
	secondToken, done, err := p.reader.Peek()
	if err != nil {
		return ColumnExpression{}, err
	}
	if done {
		return ColumnExpression{
			ColumnName: string(firstToken.value),
		}, nil
	}

	// Table qualified column name
	if secondToken.tokType == tokenTypeDot {
		_, _, err = p.reader.Next()
		if err != nil {
			return ColumnExpression{}, err
		}

		colToken, _, err := p.reader.Next()
		if err != nil {
			return ColumnExpression{}, err
		}
		if colToken.tokType != tokenTypeWord {
			return ColumnExpression{}, fmt.Errorf(
				"Expected column name but got <%s>", tokenString(colToken))
		}
		return ColumnExpression{
			ColumnName: string(colToken.value),
			TableName:  string(firstToken.value),
		}, nil
	}

	// Function call
	if secondToken.tokType == tokenTypeLParen {
		_, _, err = p.reader.Next()
		if err != nil {
			return ColumnExpression{}, err
		}

		params, err := p.scanFunctionParams()
		if err != nil {
			return ColumnExpression{}, nil
		}
		return FunctionExpression{
			FuncName:   string(firstToken.value),
			Parameters: params,
		}, nil
	}

	return ColumnExpression{
		ColumnName: string(firstToken.value),
	}, nil
}

func (p *parser) scanFunctionParams() ([]Expression, error) {
	params := []Expression{}
	for {
		expr, err := p.scanExpr()
		if err != nil {
			return nil, err
		}

		params = append(params, expr)
		next, done, err := p.reader.Next()
		if err != nil {
			return nil, err
		}
		if done {
			return nil, errors.New("Expecting next parameter")
		}

		// Comma so go to next param
		if next.tokType == tokenTypeComma {
			continue
		}

		// Close paren so finished
		if next.tokType == tokenTypeRParen {
			break
		}

		// Unknown so error
		return nil, fmt.Errorf("Expected ',' or ')' but got <%s>", tokenString(next))
	}

	return params, nil
}

func (p *parser) scanParenthetical() (Expression, error) {
	expr, err := p.scanExpr()
	if err != nil {
		return ColumnExpression{}, err
	}

	next, done, err := p.reader.Next()
	if err != nil {
		return ColumnExpression{}, err
	}
	if done {
		return ColumnExpression{}, errors.New("Expecting ')' but got EOF")
	}
	if next.tokType != tokenTypeRParen {
		return ColumnExpression{}, fmt.Errorf("Expecting ')' but got <%s>", tokenString(next))
	}
	return expr, nil
}

func tokenString(tok token) string {
	return fmt.Sprintf(
		"%d:%d -> %s",
		tok.location.line,
		tok.location.col,
		tokenValueString(tok))
}

func tokenValueString(tok token) string {
	switch tok.tokType {
	case tokenTypeWord:
		return fmt.Sprintf("word: %s", string(tok.value))
	case tokenTypeLParen:
		return "("
	case tokenTypeRParen:
		return ")"
	case tokenTypeString:
		return fmt.Sprintf("string: '%s'", string(tok.value))
	case tokenTypeDot:
		return "."
	case tokenTypeComma:
		return ","
	case tokenTypeSemicolon:
		return ";"
	case tokenTypePlus:
		return "+"
	case tokenTypeMinus:
		return "-"
	case tokenTypeSlash:
		return "/"
	case tokenTypeAsterisk:
		return "*"
	default:
		return "?"
	}
}

func getExprBinaryOpType(tok token) (binaryExprOpType, bool) {
	switch tok.tokType {
	case tokenTypePlus:
		return BinaryExprOpAdd, true
	case tokenTypeMinus:
		return BinaryExprOpSubtract, true
	case tokenTypeAsterisk:
		return BinaryExprOpMultiply, true
	case tokenTypeSlash:
		return BinaryExprOpDivide, true
	}

	return 0, false
}

func binaryExprTreeAppend(left Expression, right Expression, rightOp binaryExprOpType) Expression {
	leftBinary, leftIsBinary := left.(BinaryExpression)
	if !leftIsBinary {
		return BinaryExpression{
			Left:  left,
			Right: right,
			Op:    rightOp,
		}
	}

	leftOp := leftBinary.Op
	if rightIsGreaterPrecendence(leftOp, rightOp) {
		return BinaryExpression{
			Left: leftBinary.Left,
			Right: BinaryExpression{
				Left:  leftBinary.Right,
				Right: right,
				Op:    rightOp,
			},
			Op: leftOp,
		}
	} else {
		return BinaryExpression{
			Left:  leftBinary,
			Right: right,
			Op:    rightOp,
		}
	}
}

func rightIsGreaterPrecendence(left binaryExprOpType, right binaryExprOpType) bool {
	return getPrecendence(left) > getPrecendence(right)
}

func getPrecendence(op binaryExprOpType) int {
	switch op {
	case BinaryExprOpAdd:
		return 1
	case BinaryExprOpSubtract:
		return 1
	case BinaryExprOpMultiply:
		return 2
	case BinaryExprOpDivide:
		return 2
	default:
		return 100
	}
}
