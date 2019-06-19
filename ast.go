package main

// type setOpType int

// const (
// 	SetOpUnion setOpType = iota
// 	SetOpUnionAll
// 	SetOpIntersect
// 	SetOpExcept
// )

// type binaryCondOpType int

// const (
// 	BinaryCondOpIs binaryCondOpType = iota
// 	BinaryCondOpEqual
// 	BinaryCondOpNotEqual
// 	BinaryCondOpGreatThan
// 	BinaryCondOpGreatThanOrEqual
// 	BinaryCondOpLessThan
// 	BinaryCondOpLessThanOrEqual
// )

// type binaryExprOpType int

// const (
// 	BinaryExprOpAdd binaryExprOpType = iota
// 	BinaryExprOpSubtract
// 	BinaryExprOpMultiply
// 	BinaryExprOpDivide
// )

// type logicalOpType int

// const (
// 	LogicalOpAnd logicalOpType = iota
// 	LogicalOpOr
// )

// type Statement interface {
// 	isStatement()
// }

// func (s *Select) isStatement() {}

// type Expression interface {
// 	isExpression()
// }

// func (f *FunctionExpression) isExpression() {}
// func (b *BinaryExpression) isExpression()   {}
// func (s *StringLiteral) isExpression()      {}
// func (n *NumberLiteral) isExpression()      {}

// type StringLiteral struct {
// 	Value string
// }

// type NumberLiteral struct {
// 	Value string
// }

// type ColumnExpression struct {
// 	ColumnName string
// 	TableName  string
// }

// type FunctionExpression struct {
// 	FuncName   string
// 	Parameters []Expression
// }

// type BinaryExpression struct {
// 	Left  Expression
// 	Right Expression
// 	Op    binaryExprOpType
// }

// type Field struct {
// 	Alias string
// 	Expr  Expression
// }

// type SelectTarget struct {
// 	Alias     string
// 	TableName string
// 	Subselect *Select
// }

// type Condition interface {
// 	isCondition()
// }

// func (b *BinaryCondition) isCondition()  {}
// func (l *LogicalCondition) isCondition() {}

// type BinaryCondition struct {
// 	Left  Expression
// 	Right Expression
// 	Op    binaryCondOpType
// }

// type LogicalCondition struct {
// 	Left  Condition
// 	Right Condition
// 	Op    logicalOpType
// }

// type OrderExpr struct {
// 	desc bool
// 	expr Expression
// }

// type NextSelect struct {
// 	SetOp setOpType
// 	Query Select
// }

// type Join struct {
// 	// to do
// }

// type Select struct {
// 	Fields  []Field
// 	From    []SelectTarget
// 	Joins   []Join
// 	Where   Condition
// 	Having  Condition
// 	OrderBy []OrderExpr
// 	Next    *NextSelect
// }
