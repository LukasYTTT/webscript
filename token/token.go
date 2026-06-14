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
	ASSIGN = "="
	PLUS   = "+"
	DOT    = "."
	COMMA  = ","

	// Delimiters
	LBRACE = "{"
	RBRACE = "}"
	LPAREN = "("
	RPAREN = ")"

	// Keywords
	IMPORT = "IMPORT"
	FUNC   = "FUNC"
)

var keywords = map[string]TokenType{
	"import": IMPORT,
	"func":   FUNC,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
