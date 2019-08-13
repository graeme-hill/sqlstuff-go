package lib

type joinType int

const (
	JoinTypeInner joinType = iota
	JoinTypeLeftOuter
	JoinTypeRightOuter
	JoinTypeFullOuter
	JoinTypeCross
)

type setOpType int

const (
	SetOpUnion setOpType = iota
	SetOpUnionAll
	SetOpIntersect
	SetOpExcept
)

type dataType int

const (
	DataTypeSmallInt dataType = iota
	DataTypeInteger
	DataTypeBigInt
	DataTypeDecimal
	DataTypeNumeric
	DataTypeReal
	DataTypeDoublePrecision
	DataTypeSmallSerial
	DataTypeSerial
	DataTypeBigSerial
	DataTypeMoney
	DataTypeChar
	DataTypeVarChar
	DataTypeText
	DataTypeBytea
	DataTypeTimestamp
	DataTypeTimestampWithTimeZone
	DataTypeDate
	DataTypeTime
	DataTypeTimeWithTimeZone
	DataTypeInterval
	DataTypeBoolean
	DataTypeEnum
	DataTypePoint
	DataTypeLine
	DataTypeLineSegment
	DataTypeBox
	DataTypePath
	DataTypePolygon
	DataTypeCircle
	DataTypeCidr
	DataTypeInet
	DataTypeMacAddr
	DataTypeVarBit
	DataTypeBit
	DataTypeTextSearchVector
	DataTypeTextSearchQuery
	DataTypeUUID
	DataTypeXML
	DataTypeJSON
	DataTypeBinaryJSON
)

type binaryCondOpType int

const (
	BinaryCondOpIs binaryCondOpType = iota
	BinaryCondOpEqual
	BinaryCondOpNotEqual
	BinaryCondOpGreatThan
	BinaryCondOpGreatThanOrEqual
	BinaryCondOpLessThan
	BinaryCondOpLessThanOrEqual
)

type binaryExprOpType int

const (
	BinaryExprOpAdd binaryExprOpType = iota
	BinaryExprOpSubtract
	BinaryExprOpMultiply
	BinaryExprOpDivide
)

type logicalOpType int

const (
	LogicalOpAnd logicalOpType = iota
	LogicalOpOr
)

type unaryExprOpType int

const (
	UnaryExprOpNegative unaryExprOpType = iota
)

type Program struct {
	Statements []Statement
	Parameters []Parameter
}

type Parameter struct {
	Name string
}

type Statement interface {
	isStatement()
}

func (s Select) isStatement()      {}
func (i Insert) isStatement()      {}
func (s CreateTable) isStatement() {}
func (s AddColumn) isStatement()   {}
func (s DropColumn) isStatement()  {}
func (s DropTable) isStatement()   {}

type Expression interface {
	isExpression()
}

func (p ParameterExpression) isExpression() {}
func (f FunctionExpression) isExpression()  {}
func (b BinaryExpression) isExpression()    {}
func (b UnaryExpression) isExpression()     {}
func (s StringLiteral) isExpression()       {}
func (n NumberLiteral) isExpression()       {}
func (c ColumnExpression) isExpression()    {}

type ParameterExpression struct {
	Name string
}

type StringLiteral struct {
	Value string
}

type NumberLiteral struct {
	Value string
}

type ColumnExpression struct {
	ColumnName string
	TableName  string
}

func (c ColumnExpression) String() string {
	if c.TableName == "" {
		return c.ColumnName
	}
	return c.TableName + "." + c.ColumnName
}

type FunctionExpression struct {
	FuncName   string
	Parameters []Expression
}

type BinaryExpression struct {
	Left  Expression
	Right Expression
	Op    binaryExprOpType
}

type UnaryExpression struct {
	Right Expression
	Op    unaryExprOpType
}

type Field struct {
	Alias string
	Expr  Expression
}

type TargetTable struct {
	Alias     string
	TableName string
	Subselect *Select
}

type Condition interface {
	isCondition()
}

func (b BinaryCondition) isCondition()  {}
func (l LogicalCondition) isCondition() {}
func (l NullCondition) isCondition()    {}

type BinaryCondition struct {
	Left  Expression
	Right Expression
	Op    binaryCondOpType
}

type LogicalCondition struct {
	Left  Condition
	Right Condition
	Op    logicalOpType
}

type NullCondition struct{}

type OrderExpr struct {
	desc bool
	expr Expression
}

type NextSelect struct {
	SetOp setOpType
	Query Select
}

type Join struct {
	Type   joinType
	Target TargetTable
	On     Condition
}

type Limit struct {
	Count    int
	HasLimit bool
}

type Insert struct {
	Target  TargetTable
	Columns []ColumnExpression
	Values  []Expression
}

type Select struct {
	Fields  []Field
	From    TargetTable
	Joins   []Join
	Where   Condition
	Having  Condition
	OrderBy []OrderExpr
	Limit   Limit
	Next    *NextSelect
}

type CreateTable struct {
	Name    string
	Columns []ColumnDefinition
}

type ColumnDefinition struct {
	Name     string
	Type     dataType
	Param1   int
	Param2   int
	Values   []string
	Nullable bool
	Default  string
}

type AddColumn struct {
	TableName string
	Column    ColumnDefinition
}

type DropColumn struct {
	TableName  string
	ColumnName string
}

type DropTable struct {
	TableName string
}

type RenameTable struct {
	From string
	To   string
}

type RenameColumn struct {
	TableName string
	From      string
	To        string
}
