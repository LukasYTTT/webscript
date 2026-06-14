package parser

import (
	"fmt"
	"webscript/ast"
	"webscript/lexer"
	"webscript/token"
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.nextToken()
	p.nextToken()

	return p
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
	program.Imports = []*ast.ImportStatement{}
	program.Servers = []*ast.Server{}

	for p.curToken.Type != token.EOF {
		if p.curToken.Type == token.IMPORT {
			imp := p.parseImport()
			if imp != nil {
				program.Imports = append(program.Imports, imp)
			}
		} else if p.curToken.Type == token.IDENT {
			// Wir erwarten "http.server"
			if p.curToken.Literal == "http" && p.peekToken.Type == token.DOT {
				p.nextToken() // skip 'http'
				p.nextToken() // skip '.'
				if p.curToken.Type == token.IDENT && p.curToken.Literal == "server" {
					server := p.parseServer()
					if server != nil {
						program.Servers = append(program.Servers, server)
					}
				} else {
					p.errors = append(p.errors, fmt.Sprintf("erwartete 'server', bekam %q", p.curToken.Literal))
					p.nextToken()
				}
			} else {
				p.errors = append(p.errors, fmt.Sprintf("erwartete 'http.server', bekam %q", p.curToken.Literal))
				p.nextToken()
			}
		} else {
			p.errors = append(p.errors, fmt.Sprintf("Unerwartetes Token: %q", p.curToken.Literal))
			p.nextToken()
		}
	}

	return program
}

func (p *Parser) parseImport() *ast.ImportStatement {
	stmt := &ast.ImportStatement{}

	if !p.expectPeek(token.STRING) {
		return nil
	}
	stmt.Path = p.curToken.Literal

	p.nextToken() // consume string
	return stmt
}

func (p *Parser) parseServer() *ast.Server {
	server := &ast.Server{}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	
	if !p.expectPeek(token.STRING) {
		return nil
	}
	server.Domain = p.curToken.Literal

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	p.nextToken() // skip {

	server.Routes = []*ast.Route{}

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		// Wir erwarten http.route
		if p.curToken.Type == token.IDENT && p.curToken.Literal == "http" {
			p.nextToken() // skip http
			if p.curToken.Type == token.DOT {
				p.nextToken() // skip .
				if p.curToken.Literal == "route" {
					route := p.parseRoute()
					if route != nil {
						server.Routes = append(server.Routes, route)
					}
				}
			}
		} else {
			p.errors = append(p.errors, fmt.Sprintf("erwartete 'http.route' oder '}', bekam %q", p.curToken.Literal))
			p.nextToken()
		}
	}

	p.nextToken() // consume }
	return server
}

func (p *Parser) parseRoute() *ast.Route {
	route := &ast.Route{}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	if !p.expectPeek(token.STRING) {
		return nil
	}
	route.Path = p.curToken.Literal

	if !p.expectPeek(token.COMMA) {
		return nil
	}

	p.nextToken() // skip ,

	// now expecting target like http.proxy("localhost:8080") or http.static("./folder")
	if p.curToken.Type == token.IDENT && p.curToken.Literal == "http" {
		p.nextToken() // skip http
		if p.curToken.Type == token.DOT {
			p.nextToken() // skip .
			route.Target = p.parseTarget()
		}
	}

	// ParseTarget will consume until RPAREN, so next token should be the RPAREN for route
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	
	p.nextToken() // consume RPAREN
	
	return route
}

func (p *Parser) parseTarget() *ast.Target {
	target := &ast.Target{}

	if p.curToken.Type == token.IDENT && p.curToken.Literal == "proxy" {
		target.Type = ast.TargetProxy
	} else if p.curToken.Type == token.IDENT && p.curToken.Literal == "static" {
		target.Type = ast.TargetStatic
	} else {
		p.errors = append(p.errors, fmt.Sprintf("erwartete 'proxy' oder 'static', bekam %q", p.curToken.Literal))
		return nil
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	if !p.expectPeek(token.STRING) {
		return nil
	}
	target.Value = p.curToken.Literal

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return target
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("erwartete nächstes Token %s, bekam stattdessen %s (Literal: %s)", t, p.peekToken.Type, p.peekToken.Literal)
	p.errors = append(p.errors, msg)
}
