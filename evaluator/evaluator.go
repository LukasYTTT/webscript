package evaluator

import (
	"fmt"
	"webscript/ast"
	"webscript/object"
)

var NULL = &object.Null{}

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.AssignmentStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
		return val
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args)
	case *ast.MemberExpression:
		// Convert to a string identifier representation for lookup, e.g. "http.server"
		objStr := ""
		if ident, ok := node.Object.(*ast.Identifier); ok {
			objStr = ident.Value
		} else {
			return newError("object not an identifier")
		}
		fullName := objStr + "." + node.Property.Value
		return evalIdentifier(&ast.Identifier{Value: fullName}, env)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	}

	return nil
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range program.Statements {
		result = Eval(statement, env)
		if err, ok := result.(*object.Error); ok {
			return err
		}
	}
	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range block.Statements {
		result = Eval(statement, env)
		if result != nil && result.Type() == object.ERROR_OBJ {
			return result
		}
	}
	return result
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	return newError("identifier not found: " + node.Value)
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	if left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ {
		return evalStringInfixExpression(operator, left, right)
	}
	return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	if operator != "+" {
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object
	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return evaluated
	case *object.Builtin:
		return fn.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)
	for paramIdx, param := range fn.Parameters {
		if paramIdx < len(args) {
			env.Set(param.Value, args[paramIdx])
		}
	}
	return env
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}
