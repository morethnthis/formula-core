package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	// reader := bufio.NewReader(os.Stdin)

	// fmt.Print("请输入 LaTeX 公式: ")
	// latexFormula, _ := reader.ReadString('\n')
	latexFormula := `d8be33e3-7403-4298-a756-12131ee167c1 \\times b8a39a7a-87eb-4a87-83f2-0d2eb9c32f02 \\times 48242291-145e-4871-87d8-8257724d0bcf \\div 793ca9dd-4681-4b82-87af-b6c8b056a0a1`

	// 将 LaTeX 公式转换为 Go 表达式
	expression, err := convertLatexToExpression(latexFormula)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(expression)
	// var params []float64
	// var paramNames []string
	// fmt.Print("请输入参数名称（用逗号分隔）: ")
	// paramNamesInput, _ := reader.ReadString('\n')
	// paramNamesInput = strings.TrimSpace(paramNamesInput)
	// paramNames = strings.Split(paramNamesInput, ",")

	// fmt.Print("请输入参数值（用逗号分隔）: ")
	// paramsInput, _ := reader.ReadString('\n')
	// paramsInput = strings.TrimSpace(paramsInput)
	// paramsStr := strings.Split(paramsInput, ",")
	// paramsValue := make(map[string]float64)
	// for index, paramStr := range paramsStr {
	// 	var param float64
	// 	fmt.Sscanf(paramStr, "%f", &param)
	// 	params = append(params, param)
	// 	paramsValue[paramNames[index]] = param
	// }
	// fmt.Println(paramsValue)
	// result, err := calculateFormula(expression, paramNames, params)

	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }
	// fmt.Printf("公式计算结果为: %.4f\n", result)
}

// convertLatexToExpression 将 LaTeX 公式转换为 Go 表达式
func convertLatexToExpression(latexFormula string) (string, error) {
	// 定义 LaTeX 到 Go 表达式的替换规则，有序替换
	replacements := []struct {
		latex string
		expr  string
	}{
		{`\\frac{`, "("},          // 分数开始
		{`}{`, ")/("},             // 分数结束
		{`}`, ")"},                // 分数结束
		{`\\sqrt{`, "math.Sqrt("}, // 开根号
		{`\\cdot`, "*"},           // 乘号
		{`\\times`, "*"},          // 乘号
		{`\\div`, "/"},            // 除号
		{`\\left(`, "("},          // 左括号
		{`\\right)`, ")"},         // 右括号
		{`^`, "**"},               // 指数
	}

	// 替换 LaTeX 符号为 Go 表达式
	for _, replacement := range replacements {
		latexFormula = strings.ReplaceAll(latexFormula, replacement.latex, replacement.expr)
	}

	// 处理 ** 为 math.Pow()
	re := regexp.MustCompile(`(\w+)\*\*(\w+)`)
	latexFormula = re.ReplaceAllString(latexFormula, "math.Pow($1, $2)")

	return latexFormula, nil
}

// calculateFormula 计算 Go 表达式的值
func calculateFormula(formula string, paramNames []string, params []float64) (float64, error) {
	if len(paramNames) != len(params) {
		return 0, fmt.Errorf("参数名称和参数值的数量不匹配")
	}

	for i, paramName := range paramNames {
		formula = strings.ReplaceAll(formula, paramName, fmt.Sprintf("%f", params[i]))
	}

	expr, err := parser.ParseExpr(formula)
	if err != nil {
		return 0, err
	}
	fmt.Println(expr)
	return eval(expr)
}

func eval(expr ast.Expr) (float64, error) {
	switch v := expr.(type) {
	case *ast.BasicLit:
		return strconv.ParseFloat(v.Value, 64)
	case *ast.BinaryExpr:
		left, err := eval(v.X)
		if err != nil {
			return 0, err
		}
		right, err := eval(v.Y)
		if err != nil {
			return 0, err
		}
		return applyOp(v.Op, left, right)
	case *ast.CallExpr:
		fn, ok := v.Fun.(*ast.SelectorExpr)
		if !ok {
			return 0, fmt.Errorf("不支持的函数调用")
		}
		switch fn.Sel.Name {
		case "Pow":
			if len(v.Args) != 2 {
				return 0, fmt.Errorf("math.Pow 需要两个参数")
			}
			arg1, err := eval(v.Args[0])
			if err != nil {
				return 0, err
			}
			arg2, err := eval(v.Args[1])
			if err != nil {
				return 0, err
			}
			return math.Pow(arg1, arg2), nil
		case "Sqrt":
			if len(v.Args) != 1 {
				return 0, fmt.Errorf("math.Sqrt 需要一个参数")
			}
			arg, err := eval(v.Args[0])
			if err != nil {
				return 0, err
			}
			return math.Sqrt(arg), nil
		default:
			return 0, fmt.Errorf("不支持的函数调用")
		}
	case *ast.ParenExpr:
		return eval(v.X)
	default:
		return 0, fmt.Errorf("不支持的表达式类型")
	}
}
func applyOp(op token.Token, left, right float64) (float64, error) {
	switch op {
	case token.ADD:
		return left + right, nil
	case token.SUB:
		return left - right, nil
	case token.MUL:
		return left * right, nil
	case token.QUO:
		if right == 0 {
			return 0, fmt.Errorf("除以零")
		}
		return left / right, nil
	default:
		return 0, fmt.Errorf("不支持的二元操作符")
	}
}
