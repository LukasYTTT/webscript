package ast

import "bytes"

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

// Program is the root node
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}
func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// Identifier (e.g. port)
type Identifier struct {
	TokenLiteralValue string
	Value             string
}
func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.TokenLiteralValue }
func (i *Identifier) String() string       { return i.Value }

// StringLiteral (e.g. "localhost")
type StringLiteral struct {
	TokenLiteralValue string
	Value             string
}
func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.TokenLiteralValue }
func (sl *StringLiteral) String() string       { return `"` + sl.Value + `"` }

// AssignmentStatement (e.g. port = 3000)
type AssignmentStatement struct {
	TokenLiteralValue string // '='
	Name              *Identifier
	Value             Expression
}
func (as *AssignmentStatement) statementNode()       {}
func (as *AssignmentStatement) TokenLiteral() string { return as.TokenLiteralValue }
func (as *AssignmentStatement) String() string {
	return as.Name.String() + " = " + as.Value.String()
}

// ExpressionStatement (e.g. http.server(...))
type ExpressionStatement struct {
	TokenLiteralValue string
	Expression        Expression
}
func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.TokenLiteralValue }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// CallExpression (e.g. http.route(...))
type CallExpression struct {
	TokenLiteralValue string // '('
	Function          Expression // Identifier or MemberExpression
	Arguments         []Expression
}
func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.TokenLiteralValue }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	// Simple join
	for i, a := range args {
		out.WriteString(a)
		if i < len(args)-1 {
			out.WriteString(", ")
		}
	}
	out.WriteString(")")
	return out.String()
}

// MemberExpression (e.g. http.server)
type MemberExpression struct {
	TokenLiteralValue string // '.'
	Object            Expression
	Property          *Identifier
}
func (me *MemberExpression) expressionNode()      {}
func (me *MemberExpression) TokenLiteral() string { return me.TokenLiteralValue }
func (me *MemberExpression) String() string {
	return me.Object.String() + "." + me.Property.String()
}

// InfixExpression (e.g. "localhost:" + port)
type InfixExpression struct {
	TokenLiteralValue string // '+'
	Left              Expression
	Operator          string
	Right             Expression
}
func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.TokenLiteralValue }
func (ie *InfixExpression) String() string {
	return "(" + ie.Left.String() + " " + ie.Operator + " " + ie.Right.String() + ")"
}

// BlockStatement (e.g. { ... })
type BlockStatement struct {
	TokenLiteralValue string // '{'
	Statements        []Statement
}
func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.TokenLiteralValue }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// FunctionLiteral (e.g. func(req, res) { ... })
type FunctionLiteral struct {
	TokenLiteralValue string // 'func'
	Parameters        []*Identifier
	Body              *BlockStatement
}
func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.TokenLiteralValue }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer
	out.WriteString("func(")
	for i, p := range fl.Parameters {
		out.WriteString(p.String())
		if i < len(fl.Parameters)-1 {
			out.WriteString(", ")
		}
	}
	out.WriteString(") { ")
	out.WriteString(fl.Body.String())
	out.WriteString(" }")
	return out.String()
}
