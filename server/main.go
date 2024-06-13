/*
 * @Author: morethanthis
 * @Date: 2024-06-07 10:39:39
 * @LastEditTime: 2024-06-07 17:37:48
 * @LastEditors: Do not edit
 * @Description: In User Settings Edit
 * @FilePath: /formula_test/server/main.go
 */
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Param struct {
	Id        string  `json:"id"`
	Name      string  `json:"name"`
	Desc      string  `json:"desc"`
	Type      int8    `json:"type"`
	Value     float64 `json:"value"`
	FormulaId string  `json:"formula_id"`
	CreateBy  string  `json:"create_by"`
	Remark    string  `json:"remark"`
}

type Formula struct {
	Id         string   `json:"id"`
	Name       string   `json:"name"`
	Desc       string   `json:"desc"`
	Expression string   `json:"expression"`
	RawLatex   string   `json:"raw_latex"`
	ParamIds   []string `json:"param_ids"`
	CreateBy   string   `json:"create_by"`
	Remark     string   `json:"remark"`
}

// 全局对象列表
var params []Param
var formulas []Formula

func main() {
	// 创建一个默认的 Gin 路由器
	r := gin.Default()

	// 定义一个路由组
	v1 := r.Group("/api/v1")
	{
		v1.POST("/formula", createFormula)
		v1.POST("/param", createParam)
		v1.POST("/clac", calculate)
		v1.GET("/formulas", getFormulas)
		v1.GET("/params", getParams)
	}

	// 运行服务器
	r.Run(":8080") // 服务器将在 8080 端口运行
}

// createParam 处理第二个 POST 请求
func createParam(c *gin.Context) {
	var json map[string]interface{}
	if err := c.BindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uuid := uuid.New().String()
	p := Param{
		Id:       uuid,
		Name:     json["name"].(string),
		Desc:     json["desc"].(string),
		Type:     int8(json["type"].(float64)),
		Value:    json["value"].(float64),
		CreateBy: time.Now().Format("2006-01-02 15:04:05"),
		Remark:   json["remark"].(string),
	}
	// 处理请求逻辑
	params = append(params, p)
	// file, err := os.OpenFile("param.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// if err != nil {
	// 	log.Fatalf("无法打开文件: %v", err)
	// }
	// defer file.Close()

	// // 创建一个新的日志记录器
	// logger := log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	// // 将切片内容写入日志文件
	// for _, item := range params {
	// 	fmt.Println(item)
	// 	fmt.Println("-------------")
	// 	logger.Println(item)
	// }
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": p})
}
func getParams(c *gin.Context) {
	fmt.Println(params)
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": params})
}

// createFormula 处理第一个 POST 请求
func createFormula(c *gin.Context) {
	var json map[string]interface{}
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	expression, _ := convertLatexToExpression(json["raw_latex"].(string))
	paramIds := json["param_ids"].([]interface{})
	paramIdsStr := make([]string, 0, len(paramIds))
	for _, v := range paramIds {
		switch val := v.(type) {
		case string:
			paramIdsStr = append(paramIdsStr, val)
		default:
			// handle other types if needed
		}
	}
	f := Formula{
		Id:         uuid.New().String(),
		Name:       json["name"].(string),
		Desc:       json["desc"].(string),
		Expression: expression,
		RawLatex:   json["raw_latex"].(string),
		ParamIds:   paramIdsStr,
		CreateBy:   time.Now().Format("2006-01-02 15:04:05"),
		Remark:     json["remark"].(string),
	}
	formulas = append(formulas, f)
	file, err := os.OpenFile("formula.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("无法打开文件: %v", err)
	}
	defer file.Close()

	// 创建一个新的日志记录器
	logger := log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	// 将切片内容写入日志文件
	for _, item := range params {
		logger.Println(item)
	}
	// 处理请求逻辑
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": f})
}
func getFormulas(c *gin.Context) {
	fmt.Println(formulas)
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": formulas})
}

func calculate(c *gin.Context) {

	var json map[string]interface{}
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fid := json["formula_id"]
	for _, formula := range formulas {
		if formula.Id == fid.(string) {
			result, err := calculateFormula(formula.Expression, formula.ParamIds, json["param_value"].([]float64))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "success", "data": result})
			return
		}
	}
	// 处理请求逻辑
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": json})
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

// convertLatexToExpression 将 LaTeX 公式转换为 Go 表达式
func convertLatexToExpression(latexFormula string) (string, error) {
	// 定义 LaTeX 到 Go 表达式的替换规则，有序替换
	replacements := []struct {
		latex string
		expr  string
	}{
		{"\\frac{", "("},          // 分数开始
		{"}{", ")/("},             // 分数结束
		{"}", ")"},                // 分数结束
		{"\\sqrt{", "math.Sqrt("}, // 开根号
		{"\\cdot", "*"},           // 乘号
		{"\\times", "*"},          // 乘号
		{"\\div", "/"},            // 除号
		{"\\left(", "("},          // 左括号
		{"\\right)", ")"},         // 右括号
		{"^", "**"},               // 指数
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
