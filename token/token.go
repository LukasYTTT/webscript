package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT  = "IDENT"
	STRING = "STRING"

	// Operators
	DOT   = "."
	COMMA = ","
	ARROW = "->"

	// Delimiters
	LBRACE = "{"
	RBRACE = "}"
	LPAREN = "("
	RPAREN = ")"

	// Keywords
	IMPORT = "IMPORT"
	
	// Legacy Keywords (can be phased out or kept for backwards compatibility)
	SERVER = "SERVER"
	ROUTE  = "ROUTE"
	PROXY  = "PROXY"
	STATIC = "STATIC"
)

var keywords = map[string]TokenType{
	"import": IMPORT,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
