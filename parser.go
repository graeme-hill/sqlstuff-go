package main

import (
	"errors"
	"fmt"
	"strings"
)

func Parse(sql string) ([]Statement, error) {
	buffer := newTokenBuffer()
	p := parser{reader: buffer}
	var lexErr error = nil

	go (func() {
		lexErr = lex(sql, buffer.Write)
		buffer.Done()
	})()

	statements, err := p.scan()
	if err == nil {
		err = lexErr
	}
	return statements, err
}

type parser struct {
	reader tokenReader
}

func (p *parser) scan() ([]Statement, error) {
	statements := []Statement{}
	requireSemicolon := false

	for {
		tok, done, err := p.reader.Next()
		if err != nil {
			return nil, err
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

		if tok.tokType == tokenTypeWord && strings.ToUpper(string(tok.value)) == "SELECT" {
			selectStatement, err := p.scanSelect()
			if err != nil {
				return nil, err
			}
			statements = append(statements, selectStatement)
			requireSemicolon = true
			continue
		}

		return nil, fmt.Errorf("Expecting start of statement but got <%s>", tokenString(tok))
	}

	return statements, nil
}

func (p *parser) scanSelect() (Select, error) {
	fields, err := p.scanFieldList()
	if err != nil {
		return Select{}, err
	}

	target, err := p.scanSelectTarget()
	if err != nil {
		return Select{}, err
	}

	return Select{
		Fields: fields,
		From:   target,
	}, nil
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
		} else if next.tokType == tokenTypeWord && strings.ToUpper(string(next.value)) == "AS" {
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
			} else if next.tokType == tokenTypeWord && strings.ToUpper(string(next.value)) == "FROM" {
				break
			} else {
				return nil, errors.New("SELECT statement missing FROM")
			}
		} else if next.tokType == tokenTypeWord && strings.ToUpper(string(next.value)) == "FROM" {
			fields = append(fields, Field{Expr: expr})
			break
		} else {
			return nil, errors.New("SELECT statement missing FROM")
		}
	}
	return fields, nil
}

func (p *parser) scanSelectTarget() (SelectTarget, error) {
	next, done, err := p.reader.Next()
	if err != nil {
		return SelectTarget{}, err
	}
	if done {
		return SelectTarget{}, fmt.Errorf("Expecting select target but got <%s>", tokenString(next))
	}

	target := SelectTarget{}

	if next.tokType == tokenTypeWord {
		target.TableName = string(next.value)
	} else if next.tokType == tokenTypeLParen {
		next, done, err = p.reader.Next()
		if err != nil {
			return SelectTarget{}, err
		}
		if done {
			return SelectTarget{}, errors.New("Expecting SELECT but got EOF")
		}

		if next.tokType == tokenTypeWord && strings.ToUpper(string(next.value)) == "SELECT" {
			subSelect, err := p.scanSelect()
			if err != nil {
				return SelectTarget{}, err
			}
			target.Subselect = &subSelect

			next, done, err := p.reader.Next()
			if err != nil {
				return SelectTarget{}, err
			}
			if done {
				return SelectTarget{}, errors.New("Expecting ')' but got EOF")
			}
			if next.tokType != tokenTypeRParen {
				return SelectTarget{}, fmt.Errorf("Expecting ')' but got <%s>", tokenString(next))
			}
		} else {
			return SelectTarget{}, errors.New("Parenthesis in FROM clause must contain sub-select")
		}
	}

	next, done, err = p.reader.Peek()
	if err != nil {
		return SelectTarget{}, err
	}
	if done {
		return target, nil
	}

	if next.tokType == tokenTypeWord && strings.ToUpper(string(next.value)) == "AS" {
		_, _, err := p.reader.Next()
		if err != nil {
			return SelectTarget{}, err
		}

		next, done, err = p.reader.Next()
		if err != nil {
			return SelectTarget{}, err
		}
		if done {
			return SelectTarget{}, errors.New("Expected alias but got EOF")
		}
		if next.tokType != tokenTypeWord {
			return SelectTarget{}, fmt.Errorf("Expected alias but got <%s>", tokenString(next))
		}

		target.Alias = string(next.value)
	}

	return target, nil
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

	// not compelte!
	return left, nil
}

func (p *parser) scanSubExpr() (Expression, error) {
	tok, done, err := p.reader.Next()
	if err != nil {
		return ColumnExpression{}, err
	}
	if done {
		return ColumnExpression{}, errors.New("Expecting expression but found EOF")
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
