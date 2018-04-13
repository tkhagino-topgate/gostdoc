package gostdoc

import (
	"errors"
	"go/ast"
)

// ExprToTypeName convert ast.Expr to type name.
func ExprToTypeName(expr ast.Expr) (string, error) {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name, nil
	}
	if star, ok := expr.(*ast.StarExpr); ok {
		x, err := ExprToTypeName(star.X)
		if err != nil {
			return "", nil
		}
		return "*" + x, nil
	}
	if selector, ok := expr.(*ast.SelectorExpr); ok {
		x, err := ExprToTypeName(selector.X)
		if err != nil {
			return "", nil
		}
		sel, err := ExprToTypeName(selector.Sel)
		if err != nil {
			return "", nil
		}
		return x + "." + sel, nil
	}
	if array, ok := expr.(*ast.ArrayType); ok {
		x, err := ExprToTypeName(array.Elt)
		if err != nil {
			return "", nil
		}
		return "[]" + x, nil
	}

	// add
	if m, ok := expr.(*ast.MapType); ok {
		k, err := ExprToTypeName(m.Key)
		if err != nil {
			return "", nil
		}
		v, err := ExprToTypeName(m.Value)
		if err != nil {
			return "", nil
		}

		return "map[" + k + "]" + v, nil
	}

	if _, ok := expr.(*ast.StructType); ok {
		return "struct{}", nil
	}

	return "", errors.New("can't detect type name")
}

// ExprToBaseTypeName convert ast.Expr to type name without "*" and "[]".
func ExprToBaseTypeName(expr ast.Expr) (string, error) {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name, nil
	}
	if star, ok := expr.(*ast.StarExpr); ok {
		x, err := ExprToBaseTypeName(star.X)
		if err != nil {
			return "", nil
		}
		return x, nil
	}
	if selector, ok := expr.(*ast.SelectorExpr); ok {
		x, err := ExprToBaseTypeName(selector.X)
		if err != nil {
			return "", nil
		}
		sel, err := ExprToBaseTypeName(selector.Sel)
		if err != nil {
			return "", nil
		}
		return x + "." + sel, nil
	}
	if array, ok := expr.(*ast.ArrayType); ok {
		x, err := ExprToBaseTypeName(array.Elt)
		if err != nil {
			return "", nil
		}
		return x, nil
	}
	// add
	if m, ok := expr.(*ast.MapType); ok {
		k, err := ExprToTypeName(m.Key)
		if err != nil {
			return "", nil
		}
		v, err := ExprToTypeName(m.Value)
		if err != nil {
			return "", nil
		}

		return "map[" + k + "]" + v, nil
	}

	if _, ok := expr.(*ast.StructType); ok {
		return "struct{}", nil
	}
	return "", errors.New("can't detect type name")
}
