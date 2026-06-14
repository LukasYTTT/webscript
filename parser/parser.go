package parser

import (
	"fmt"
	"webscript/ast"
	"webscript/lexer"
	"webscript/token"
)

const (
	_ int = iota
	LOWEST
	ASSIGN
	SUM
	CALL
	INDEX
)

var precedences = map[token.TokenType]int{
	token.ASSIGN: ASSIGN,
	token.PLUS:   SUM,
	token.LPAREN: CALL,
	token.DOT:    INDEX,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.FUNC, p.parseFunctionLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.DOT, p.parseMemberExpression)

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.IMPORT:
		// Convert old Import to a CallExpression: std_import("lib") so we don't need a special AST node
		// Actually, let's keep it simple and just skip imports for the evaluator right now, or parse it as a function call.
		// For backward compatibility, let's parse it as a special statement or expression.
		// Let's parse it as an ExpressionStatement -> CallExpression(Identifier("import"), StringLiteral("..."))
		tok := p.curToken
		p.nextToken()
		if p.curToken.Type == token.STRING {
			stmt := &ast.ExpressionStatement{
				TokenLiteralValue: tok.Literal,
				Expression: &ast.CallExpression{
					TokenLiteralValue: "(",
					Function: &ast.Identifier{TokenLiteralValue: "import", Value: "import"},
					Arguments: []ast.Expression{&ast.StringLiteral{TokenLiteralValue: p.curToken.Literal, Value: p.curToken.Literal}},
				},
			}
			return stmt
		}
		return nil
	case token.IDENT:
		// Check if assignment
		if p.peekToken.Type == token.ASSIGN {
			return p.parseAssignmentStatement()
		}
		return p.parseExpressionStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseAssignmentStatement() *ast.AssignmentStatement {
	stmt := &ast.AssignmentStatement{TokenLiteralValue: "="}
	stmt.Name = &ast.Identifier{TokenLiteralValue: p.curToken.Literal, Value: p.curToken.Literal}

	p.nextToken() // move to '='
	p.nextToken() // move to expression

	stmt.Value = p.parseExpression(LOWEST)

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{TokenLiteralValue: p.curToken.Literal}

	stmt.Expression = p.parseExpression(LOWEST)

	// If the expression is a CallExpression, it might be followed by a block (like http.server(...) { ... })
	if p.peekToken.Type == token.LBRACE {
		p.nextToken()
		// We can add this block as an argument to the function call
		block := p.parseBlockStatement()
		
		if call, ok := stmt.Expression.(*ast.CallExpression); ok {
			// Convert Block to a FunctionLiteral with no params so it can be passed as an argument
			funcLit := &ast.FunctionLiteral{
				TokenLiteralValue: "func",
				Parameters:        []*ast.Identifier{},
				Body:              block,
			}
			call.Arguments = append(call.Arguments, funcLit)
		}
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.errors = append(p.errors, fmt.Sprintf("no prefix parse function for %s found", p.curToken.Type))
		return nil
	}
	leftExp := prefix()

	for p.peekToken.Type != token.EOF && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{TokenLiteralValue: p.curToken.Literal, Value: p.curToken.Literal}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{TokenLiteralValue: p.curToken.Literal, Value: p.curToken.Literal}
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		TokenLiteralValue: p.curToken.Literal,
		Operator:          p.curToken.Literal,
		Left:              left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseMemberExpression(left ast.Expression) ast.Expression {
	expression := &ast.MemberExpression{
		TokenLiteralValue: p.curToken.Literal,
		Object:            left,
	}

	p.nextToken()
	expression.Property = p.parseIdentifier().(*ast.Identifier)

	return expression
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{TokenLiteralValue: p.curToken.Literal, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.peekToken.Type == token.RPAREN {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekToken.Type == token.COMMA {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if p.peekToken.Type == token.RPAREN {
		p.nextToken()
	} else {
		p.errors = append(p.errors, "expected RPAREN after arguments")
	}

	return args
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{TokenLiteralValue: p.curToken.Literal}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{TokenLiteralValue: p.curToken.Literal}

	if p.peekToken.Type != token.LPAREN {
		return nil
	}
	p.nextToken()

	lit.Parameters = p.parseFunctionParameters()

	if p.peekToken.Type != token.LBRACE {
		return nil
	}
	p.nextToken()

	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekToken.Type == token.RPAREN {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{TokenLiteralValue: p.curToken.Literal, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekToken.Type == token.COMMA {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{TokenLiteralValue: p.curToken.Literal, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if p.peekToken.Type != token.RPAREN {
		return nil
	}
	p.nextToken()

	return identifiers
}
