package eval

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
)

func Expr(expr string) (float64, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return 0, nil
	}
	if v, err := strconv.ParseFloat(expr, 64); err == nil {
		return v, nil
	}
	n, err := parser.ParseExpr(expr)
	if err != nil {
		return 0, fmt.Errorf("invalid expression: %s", expr)
	}
	return node(n)
}

func node(nd ast.Expr) (float64, error) {
	switch n := nd.(type) {
	case *ast.BasicLit:
		if n.Kind == token.INT || n.Kind == token.FLOAT {
			return strconv.ParseFloat(n.Value, 64)
		}
	case *ast.UnaryExpr:
		val, err := node(n.X)
		if err != nil {
			return 0, err
		}
		if n.Op == token.SUB {
			return -val, nil
		}
		return val, nil
	case *ast.BinaryExpr:
		lhs, err := node(n.X)
		if err != nil {
			return 0, err
		}
		rhs, err := node(n.Y)
		if err != nil {
			return 0, err
		}
		switch n.Op {
		case token.ADD:
			return lhs + rhs, nil
		case token.SUB:
			return lhs - rhs, nil
		case token.MUL:
			return lhs * rhs, nil
		case token.QUO:
			if rhs == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return lhs / rhs, nil
		}
	case *ast.ParenExpr:
		return node(n.X)
	}
	return 0, fmt.Errorf("unsupported expression")
}
