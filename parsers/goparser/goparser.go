package goparser

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	uni "unicode"
	"unicode/utf8"
)

type Position struct {
	Offset  int
	LineIdx int
	CharIdx int
}

func NewToken(kind string, start, end Position, val []rune) *Token {
	return &Token{
		Kind:  kind,
		Start: start,
		End:   end,
		Value: val,
	}
}

type Token struct {
	Kind  string
	Start Position
	End   Position
	Value []rune
}

func (t *Token) Fork() *Token {
	return &Token{
		Kind:  t.Kind,
		Start: t.Start,
		End:   t.End,
		Value: t.Value,
	}
}

type Node interface {
	Kind() string
	Range() (Position, Position)
	SetRange(Position, Position)
	RangeStart() Position
	RangeEnd() Position
	BuildLink()
	Parent() Node
	SetParent(Node)
	SelfField() string
	SetSelfField(string)
	Fields() []string
	ReplaceSelf(Node)
	SetReplaceSelf(func(Node))
	Child(field string) Node
	SetChild(nodes []Node)
	Fork() Node
	Visit(func(Node) (visitChildren, exit bool), func(Node) (exit bool)) (exit bool)
	FilePath() string
	FileContent() []rune
	Code() []rune
	Dump(hook func(Node, map[string]string) string) map[string]string
	IsDummy() bool
	UnpackNodes() []Node
	Any() any
	SetAny(any)
}

const TokenTypeDummy = "dummy"
const TokenTypeEndOfFile = "end_of_file"
const TokenTypePseudo = "pseudo"
const TokenTypeWhitespace = "whitespace"
const TokenTypeNewline = "newline"
const TokenTypeComment = "comment"
const TokenTypeIdent = "ident"
const TokenTypeNumber = "number"
const TokenTypeString = "string"
const TokenTypeOpAnd = "&"
const TokenTypeOpAndAnd = "&&"
const TokenTypeOpAndCaret = "&^"
const TokenTypeOpAndCaretEqual = "&^="
const TokenTypeOpAndEqual = "&="
const TokenTypeOpBar = "|"
const TokenTypeOpBarBar = "||"
const TokenTypeOpBarEqual = "|="
const TokenTypeOpCaret = "^"
const TokenTypeOpCaretEqual = "^="
const TokenTypeOpColon = ":"
const TokenTypeOpColonEqual = ":="
const TokenTypeOpComma = ","
const TokenTypeOpDot = "."
const TokenTypeOpDotDotDot = "..."
const TokenTypeOpEqual = "="
const TokenTypeOpEqualEqual = "=="
const TokenTypeOpGreater = ">"
const TokenTypeOpGreaterEqual = ">="
const TokenTypeOpGreaterGreater = ">>"
const TokenTypeOpGreaterGreaterEqual = ">>="
const TokenTypeOpLeftBrace = "{"
const TokenTypeOpLeftBracket = "["
const TokenTypeOpLeftParen = "("
const TokenTypeOpLess = "<"
const TokenTypeOpLessEqual = "<="
const TokenTypeOpLessLess = "<<"
const TokenTypeOpLessLessEqual = "<<="
const TokenTypeOpLessMinus = "<-"
const TokenTypeOpMinus = "-"
const TokenTypeOpMinusEqual = "-="
const TokenTypeOpMinusMinus = "--"
const TokenTypeOpNot = "!"
const TokenTypeOpNotEqual = "!="
const TokenTypeOpPercent = "%"
const TokenTypeOpPercentEqual = "%="
const TokenTypeOpPlus = "+"
const TokenTypeOpPlusEqual = "+="
const TokenTypeOpPlusPlus = "++"
const TokenTypeOpRightBrace = "}"
const TokenTypeOpRightBracket = "]"
const TokenTypeOpRightParen = ")"
const TokenTypeOpSemi = ";"
const TokenTypeOpSlash = "/"
const TokenTypeOpSlashEqual = "/="
const TokenTypeOpStar = "*"
const TokenTypeOpStarEqual = "*="
const TokenTypeOpTilde = "~"
const TokenTypeKwBreak = "kw_break"
const TokenTypeKwCase = "kw_case"
const TokenTypeKwChan = "kw_chan"
const TokenTypeKwConst = "kw_const"
const TokenTypeKwContinue = "kw_continue"
const TokenTypeKwDefault = "kw_default"
const TokenTypeKwDefer = "kw_defer"
const TokenTypeKwElse = "kw_else"
const TokenTypeKwFallthrough = "kw_fallthrough"
const TokenTypeKwFor = "kw_for"
const TokenTypeKwFunc = "kw_func"
const TokenTypeKwGo = "kw_go"
const TokenTypeKwGoto = "kw_goto"
const TokenTypeKwIf = "kw_if"
const TokenTypeKwImport = "kw_import"
const TokenTypeKwInterface = "kw_interface"
const TokenTypeKwMap = "kw_map"
const TokenTypeKwPackage = "kw_package"
const TokenTypeKwRange = "kw_range"
const TokenTypeKwReturn = "kw_return"
const TokenTypeKwSelect = "kw_select"
const TokenTypeKwStruct = "kw_struct"
const TokenTypeKwSwitch = "kw_switch"
const TokenTypeKwType = "kw_type"
const TokenTypeKwVar = "kw_var"

const NodeTypeDummy = "dummy"
const NodeTypeToken = "token"
const NodeTypeNodes = "nodes"
const NodeTypeFile = "file"
const NodeTypeAssignStmt = "assign_stmt"
const NodeTypeBlockStmt = "block_stmt"
const NodeTypeBranchStmt = "branch_stmt"
const NodeTypeDeferStmt = "defer_stmt"
const NodeTypeGoStmt = "go_stmt"
const NodeTypeSendStmt = "send_stmt"
const NodeTypeExprStmt = "expr_stmt"
const NodeTypeIncDecStmt = "inc_dec_stmt"
const NodeTypeIfStmt = "if_stmt"
const NodeTypeForStmt = "for_stmt"
const NodeTypeRangeStmt = "range_stmt"
const NodeTypeSelectStmt = "select_stmt"
const NodeTypeSwitchStmt = "switch_stmt"
const NodeTypeTypeSwitchStmt = "type_switch_stmt"
const NodeTypeReturnStmt = "return_stmt"
const NodeTypeBinaryExpr = "binary_expr"
const NodeTypeCallExpr = "call_expr"
const NodeTypeIndexExpr = "index_expr"
const NodeTypeKeyValueExpr = "key_value_expr"
const NodeTypeParenExpr = "paren_expr"
const NodeTypeSelectorExpr = "selector_expr"
const NodeTypeStarExpr = "star_expr"
const NodeTypeTypeAssertExpr = "type_assert_expr"
const NodeTypeSliceExpr = "slice_expr"
const NodeTypeUnaryExpr = "unary_expr"
const NodeTypeArrayType = "array_type"
const NodeTypeChanType = "chan_type"
const NodeTypeFunctionType = "function_type"
const NodeTypeInterfaceType = "interface_type"
const NodeTypeMapType = "map_type"
const NodeTypeStructType = "struct_type"
const NodeTypeBasicLit = "basic_lit"
const NodeTypeCompositeLit = "composite_lit"
const NodeTypeFunctionLit = "function_lit"
const NodeTypeCaseClause = "case_clause"
const NodeTypeCommonClause = "common_clause"
const NodeTypeFieldList = "field_list"
const NodeTypeField = "field"
const NodeTypeImportDecl = "import_decl"
const NodeTypeImportSpec = "import_spec"
const NodeTypeConstSpec = "const_spec"
const NodeTypeVarSpec = "var_spec"
const NodeTypeTypeEqSpec = "type_eq_spec"
const NodeTypeTypeSpec = "type_spec"
const NodeTypeConstDecl = "const_decl"
const NodeTypeVarDecl = "var_decl"
const NodeTypeTypeDecl = "type_decl"
const NodeTypeEllipsis = "ellipsis"
const NodeTypeLabeledStmt = "labeled_stmt"
const NodeTypeGenericTypeInstantiation = "generic_type_instantiation"
const NodeTypeIdent = "ident"
const NodeTypeMakeExpr = "make_expr"
const NodeTypeNewExpr = "new_expr"
const NodeTypePackageIdent = "package_ident"
const NodeTypeImportDot = "import_dot"
const NodeTypeImportIdent = "import_ident"
const NodeTypeImportPath = "import_path"
const NodeTypeConstIdent = "const_ident"
const NodeTypeVarIdent = "var_ident"
const NodeTypeTypeIdent = "type_ident"
const NodeTypeFunctionIdent = "function_ident"
const NodeTypeMethodIdent = "method_ident"
const NodeTypeGenericParameter = "generic_parameter"
const NodeTypeGenericParameterIdent = "generic_parameter_ident"
const NodeTypeGenericUnionConstraint = "generic_union_constraint"
const NodeTypeGenericUnderlyingTypeConstraint = "generic_underlying_type_constraint"
const NodeTypeGenericTypeConstraint = "generic_type_constraint"
const NodeTypeEllipsisParameter = "ellipsis_parameter"
const NodeTypeParameter = "parameter"
const NodeTypeParameterIdent = "parameter_ident"
const NodeTypeFunctionResult = "function_result"
const NodeTypeFunctionResultIdent = "function_result_ident"
const NodeTypeFunctionDecl = "function_decl"
const NodeTypeMethodDecl = "method_decl"
const NodeTypeReceiverIdent = "receiver_ident"
const NodeTypeReceiverTypeIdent = "receiver_type_ident"
const NodeTypeReceiverGenericTypeIdent = "receiver_generic_type_ident"
const NodeTypeReceiver = "receiver"

func errorContext(filePath string, fileContent []rune, offset, lineIdx, charIdx int) string {
	var lineStartOffset int
	for i := offset; i >= 0; i-- {
		if i < len(fileContent) && fileContent[i] == '\n' {
			lineStartOffset = i + 1
			break
		}
	}
	lineText := regexp.MustCompile("[^\\t]").ReplaceAllString(string(fileContent[lineStartOffset:offset]), " ")

	lines := strings.Split(string(fileContent), "\n")
	contextLines := 3
	startLine := lineIdx - contextLines
	if startLine < 0 {
		startLine = 0
	}
	endLine := lineIdx + contextLines
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== error context (%s:%d:%d) ===\n", filePath, lineIdx+1, charIdx+1))
	for i := startLine; i <= endLine; i++ {
		prefix := "   "
		var t string
		if i == lineIdx {
			prefix = ">>>"
			t = fmt.Sprintf("          %s^\n", lineText)
		}
		sb.WriteString(fmt.Sprintf("%s %4d: %s\n", prefix, i+1, lines[i]))
		if t != "" {
			sb.WriteString(t)
		}
	}
	sb.WriteString("=== end of error context ===")
	return sb.String()
}

func toSnakeCase(camelCaseString string) string {
	var sb strings.Builder
	for i, char := range camelCaseString {
		if uni.IsUpper(char) && i != 0 {
			sb.WriteRune('_')
		}
		sb.WriteRune(uni.ToLower(char))
	}
	return sb.String()
}

func toCamelCase(s string) string {
	var sb strings.Builder
	shouldUpper := true
	for _, r := range s {
		if r == '_' { // 下划线分隔符，将下一个字符转换为大写字母
			shouldUpper = true
		} else {
			if shouldUpper && uni.IsLetter(r) { // 需要转换为大写字母
				sb.WriteRune(uni.ToUpper(r))
				shouldUpper = false
			} else {
				sb.WriteRune(uni.ToLower(r))
			}
		}
	}
	return sb.String()
}

func DecodeBytes(bs []byte) ([]rune, [][3]int) {
	var encoding string
	var r *bufio.Reader

	file := bytes.NewBuffer(bs)

	skipBytes := 0
	// check BOM
	if len(bs) > 2 && bs[0] == 0xef && bs[1] == 0xbb && bs[2] == 0xbf {
		encoding = "utf-8-bom"
		r = bufio.NewReader(file)
		skipBytes = 3
	} else if len(bs) > 1 && bs[0] == 0xff && bs[1] == 0xfe {
		encoding = "utf-16le-bom"
		r = bufio.NewReader(transform.NewReader(file, unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()))
		skipBytes = 2
	} else if len(bs) > 1 && bs[0] == 0xfe && bs[1] == 0xff {
		encoding = "utf-16be-bom"
		r = bufio.NewReader(transform.NewReader(file, unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewDecoder()))
		skipBytes = 2
	} else if utf8.Valid(bs) {
		encoding = "utf-8"
		r = bufio.NewReader(file)
	} else {
		encoding = "gbk"
		r = bufio.NewReader(transform.NewReader(file, simplifiedchinese.GBK.NewDecoder()))
	}

	// r: rune-offset, b: byte-offset, s: size
	offsets := make([][3]int, 0)
	offsets = append(offsets, [3]int{0, 0, skipBytes})
	byteOffset := skipBytes
	result := make([]rune, 0)
	// read and decode text
	for {
		c, s, err := r.ReadRune()
		if err != nil {
			break
		}
		if c == 0xfeff {
			continue
		}
		if s > 1 {
			offsets = append(offsets, [3]int{len(result), byteOffset, s})
		}
		byteOffset += s
		result = append(result, c)
	}

	_ = encoding

	return result, offsets
}

func TypeNameOf(node Node) string {
	structName := reflect.ValueOf(node).Elem().Type().Name()
	name := structName[:len(structName)-4]
	return toSnakeCase(name)
}

func equalRune(a, b rune) bool {
	return a == b
}

func inRange(r, s, e rune) bool {
	return s <= r && e >= r
}

func nodesSetParent(targets []Node, parent Node, field string) {
	for i, target := range targets {
		target.SetParent(parent)
		target.SetSelfField(strconv.Itoa(i))
		if field != "" {
			target.SetSelfField(field)
		}
	}
}

func nodesVisit(nodes []Node, before func(Node) (visitChild, exit bool), after func(Node) (exit bool)) (exit bool) {
	for _, node := range nodes {
		if node.Visit(before, after) {
			return true
		}
	}
	return false
}

var creationHook = func(Node) {}

func SetCreationHook(h func(Node)) {
	creationHook = h
}

var DummyNode Node

func init() {
	DummyNode = &BaseNode{kind: NodeTypeDummy}
}

func NewBaseNode(filePath string, fileContent []rune, kind string, start, end Position) *BaseNode {
	return &BaseNode{filePath: filePath, fileContent: fileContent, kind: kind, start: start, end: end}
}

type BaseNode struct {
	filePath    string
	fileContent []rune
	kind        string
	start       Position
	end         Position
	parent      Node
	selfField   string
	replaceFun  func(Node)
	any_        any
}

func (n *BaseNode) FilePath() string {
	return n.filePath
}

func (n *BaseNode) FileContent() []rune {
	return n.fileContent
}

func (n *BaseNode) Kind() string {
	return n.kind
}

func (n *BaseNode) Range() (Position, Position) {
	return n.start, n.end
}

func (n *BaseNode) SetRange(start, end Position) {
	n.start = start
	n.end = end
}

func (n *BaseNode) RangeStart() Position {
	return n.start
}

func (n *BaseNode) RangeEnd() Position {
	return n.end
}

func (n *BaseNode) BuildLink() {
}

func (n *BaseNode) Parent() Node {
	return n.parent
}

func (n *BaseNode) SetParent(v Node) {
	n.parent = v
}

func (n *BaseNode) SelfField() string {
	return n.selfField
}

func (n *BaseNode) SetSelfField(v string) {
	n.selfField = v
}

func (n *BaseNode) ReplaceSelf(node Node) {
	node.SetReplaceSelf(n.replaceFun)
	node.SetParent(n.Parent())
	node.SetSelfField(n.SelfField())
	n.replaceFun(node)
}

func (n *BaseNode) SetReplaceSelf(fun func(Node)) {
	n.replaceFun = fun
}

func (n *BaseNode) Fields() []string {
	return nil
}

func (n *BaseNode) Child(_ string) Node {
	return DummyNode
}

func (n *BaseNode) SetChild(_ []Node) {
}

func (n *BaseNode) fork() *BaseNode {
	return &BaseNode{
		filePath:    n.filePath,
		fileContent: n.fileContent,
		kind:        n.kind,
		start:       n.start,
		end:         n.end,
		parent:      n.parent,
		selfField:   n.selfField,
		replaceFun:  n.replaceFun,
	}
}

func (n *BaseNode) Fork() Node {
	return n.fork()
}

func (n *BaseNode) Visit(func(Node) (bool, bool), func(Node) bool) bool {
	return false
}

func (n *BaseNode) Code() []rune {
	if n.fileContent == nil {
		return nil
	}
	code := n.fileContent
	start := 0
	end := len(code)
	if n.end.Offset <= len(code) && n.end.Offset >= 0 {
		end = n.end.Offset
	}
	if n.start.Offset >= 0 && n.start.Offset <= end {
		start = n.start.Offset
	}
	return code[start:end]
}

func (n *BaseNode) Dump(func(Node, map[string]string) string) map[string]string {
	return map[string]string{
		"kind": "?",
	}
}

func (n *BaseNode) IsDummy() bool {
	return n.kind == NodeTypeDummy
}

func (n *BaseNode) UnpackNodes() []Node {
	return nil
}

func (n *BaseNode) Any() any {
	return n.any_
}

func (n *BaseNode) SetAny(any_ any) {
	n.any_ = any_
}

func NewNodesNode(nodes []Node) Node {
	if len(nodes) == 0 {
		return DummyNode
	}
	filePath := nodes[0].FilePath()
	fileContent := nodes[0].FileContent()
	start := nodes[0].RangeStart()
	end := nodes[len(nodes)-1].RangeEnd()
	ret := &NodesNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeNodes, start, end),
		nodes:    nodes,
	}
	creationHook(ret)
	return ret
}

type NodesNode struct {
	*BaseNode
	nodes []Node
}

func (n *NodesNode) Nodes() []Node {
	return n.nodes
}

func (n *NodesNode) SetNodes(v []Node) {
	n.nodes = v
}

func (n *NodesNode) Fields() []string {
	ret := make([]string, 0)
	for i := 0; i < len(n.nodes); i++ {
		ret = append(ret, strconv.Itoa(i))
	}
	return ret
}

func (n *NodesNode) BuildLink() {
	nodesSetParent(n.nodes, n, "")
	for _, target := range n.nodes {
		target.BuildLink()
		target.SetReplaceSelf(func(n Node) {
			i, _ := strconv.Atoi(n.SelfField())
			n.Parent().(*NodesNode).Nodes()[i] = n
		})
	}
}

func (n *NodesNode) Child(field string) Node {
	index, err := strconv.Atoi(field)
	if err != nil {
		return DummyNode
	}
	if index >= 0 && index < len(n.nodes) {
		return n.nodes[index]
	}
	return DummyNode
}

func (n *NodesNode) SetChild(nodes []Node) {
	n.nodes = nodes
}

func (n *NodesNode) Fork() Node {
	nodes := make([]Node, 0)
	for _, n := range n.nodes {
		nodes = append(nodes, n.Fork())
	}
	_ret := &NodesNode{
		BaseNode: n.BaseNode.fork(),
		nodes:    nodes,
	}
	nodesSetParent(_ret.nodes, _ret, "")
	return _ret
}

func (n *NodesNode) Visit(beforeChildren func(Node) (visitChildren, exit bool), afterChildren func(Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if nodesVisit(n.nodes, beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *NodesNode) dumpNodes(hook func(Node, map[string]string) string) string {
	items := make([]string, 0)
	for _, t := range n.nodes {
		items = append(items, DumpNode(t, hook))
	}
	return fmt.Sprintf("[%s]", strings.Join(items, ", "))
}

func (n *NodesNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	return map[string]string{
		"kind":  "\"nodes\"",
		"nodes": n.dumpNodes(hook),
	}
}

func (n *NodesNode) UnpackNodes() []Node {
	return n.Nodes()
}

func NewTokenNode(filePath string, fileContent []rune, token *Token) Node {
	ret := &TokenNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeToken, token.Start, token.End),
		token:    token,
	}
	creationHook(ret)
	return ret
}

type TokenNode struct {
	*BaseNode
	token *Token
}

func (n *TokenNode) Token() *Token {
	return n.token
}

func (n *TokenNode) Visit(beforeChildren func(Node) (visitChildren, exit bool), afterChildren func(Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *TokenNode) Fork() Node {
	return &TokenNode{
		BaseNode: n.BaseNode.fork(),
		token:    n.token,
	}
}

func (n *TokenNode) Dump(func(Node, map[string]string) string) map[string]string {
	val := string(n.Code())
	val = strings.ReplaceAll(val, "\\", "\\\\")
	val = strings.ReplaceAll(val, "\"", "\\\"")
	val = strings.ReplaceAll(val, "\n", "\\n")
	val = strings.ReplaceAll(val, "\r", "\\r")
	val = strings.ReplaceAll(val, "\t", "\\t")
	val = fmt.Sprintf("\"%s\"", val)
	return map[string]string{
		"kind": "\"token\"",
		"code": val,
	}
}

func NewFileNode(filePath string, fileContent []rune, package_ Node, imports Node, declarations Node, start, end Position) Node {
	if package_ == nil {
		package_ = DummyNode
	}
	if imports == nil {
		imports = DummyNode
	}
	if declarations == nil {
		declarations = DummyNode
	}
	_1 := &FileNode{
		BaseNode:     NewBaseNode(filePath, fileContent, NodeTypeFile, start, end),
		package_:     package_,
		imports:      imports,
		declarations: declarations,
	}
	creationHook(_1)
	return _1
}

type FileNode struct {
	*BaseNode
	package_     Node
	imports      Node
	declarations Node
}

func (n *FileNode) Package() Node {
	return n.package_
}

func (n *FileNode) SetPackage(v Node) {
	n.package_ = v
}

func (n *FileNode) Imports() Node {
	return n.imports
}

func (n *FileNode) SetImports(v Node) {
	n.imports = v
}

func (n *FileNode) Declarations() Node {
	return n.declarations
}

func (n *FileNode) SetDeclarations(v Node) {
	n.declarations = v
}

func (n *FileNode) BuildLink() {
	if !n.Package().IsDummy() {
		package_ := n.Package()
		package_.BuildLink()
		package_.SetParent(n)
		package_.SetSelfField("package_")
		package_.SetReplaceSelf(func(n Node) {
			n.Parent().(*FileNode).SetPackage(n)
		})
	}
	if !n.Imports().IsDummy() {
		imports := n.Imports()
		imports.BuildLink()
		imports.SetParent(n)
		imports.SetSelfField("imports")
		imports.SetReplaceSelf(func(n Node) {
			n.Parent().(*FileNode).SetImports(n)
		})
	}
	if !n.Declarations().IsDummy() {
		declarations := n.Declarations()
		declarations.BuildLink()
		declarations.SetParent(n)
		declarations.SetSelfField("declarations")
		declarations.SetReplaceSelf(func(n Node) {
			n.Parent().(*FileNode).SetDeclarations(n)
		})
	}
}

func (n *FileNode) Fields() []string {
	return []string{
		"package_",
		"imports",
		"declarations",
	}
}

func (n *FileNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "package_" {
		return n.Package()
	}
	if field == "imports" {
		return n.Imports()
	}
	if field == "declarations" {
		return n.Declarations()
	}
	return nil
}

func (n *FileNode) SetChild(nodes []Node) {
	if len(nodes) != 3 {
		return
	}
	n.SetPackage(nodes[0])
	n.SetImports(nodes[1])
	n.SetDeclarations(nodes[2])
}

func (n *FileNode) Fork() Node {
	_ret := &FileNode{
		BaseNode:     n.BaseNode.fork(),
		package_:     n.package_.Fork(),
		imports:      n.imports.Fork(),
		declarations: n.declarations.Fork(),
	}
	_ret.package_.SetParent(_ret)
	_ret.imports.SetParent(_ret)
	_ret.declarations.SetParent(_ret)
	return _ret
}

func (n *FileNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.package_.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.imports.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.declarations.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *FileNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"file\""
	ret["package"] = DumpNode(n.Package(), hook)
	ret["imports"] = DumpNode(n.Imports(), hook)
	ret["declarations"] = DumpNode(n.Declarations(), hook)
	return ret
}

func NewAssignStmtNode(filePath string, fileContent []rune, lhs Node, op Node, rhs Node, start, end Position) Node {
	if lhs == nil {
		lhs = DummyNode
	}
	if op == nil {
		op = DummyNode
	}
	if rhs == nil {
		rhs = DummyNode
	}
	_1 := &AssignStmtNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeAssignStmt, start, end),
		lhs:      lhs,
		op:       op,
		rhs:      rhs,
	}
	creationHook(_1)
	return _1
}

type AssignStmtNode struct {
	*BaseNode
	lhs Node
	op  Node
	rhs Node
}

func (n *AssignStmtNode) Lhs() Node {
	return n.lhs
}

func (n *AssignStmtNode) SetLhs(v Node) {
	n.lhs = v
}

func (n *AssignStmtNode) Op() Node {
	return n.op
}

func (n *AssignStmtNode) SetOp(v Node) {
	n.op = v
}

func (n *AssignStmtNode) Rhs() Node {
	return n.rhs
}

func (n *AssignStmtNode) SetRhs(v Node) {
	n.rhs = v
}

func (n *AssignStmtNode) BuildLink() {
	if !n.Lhs().IsDummy() {
		lhs := n.Lhs()
		lhs.BuildLink()
		lhs.SetParent(n)
		lhs.SetSelfField("lhs")
		lhs.SetReplaceSelf(func(n Node) {
			n.Parent().(*AssignStmtNode).SetLhs(n)
		})
	}
	if !n.Op().IsDummy() {
		op := n.Op()
		op.BuildLink()
		op.SetParent(n)
		op.SetSelfField("op")
		op.SetReplaceSelf(func(n Node) {
			n.Parent().(*AssignStmtNode).SetOp(n)
		})
	}
	if !n.Rhs().IsDummy() {
		rhs := n.Rhs()
		rhs.BuildLink()
		rhs.SetParent(n)
		rhs.SetSelfField("rhs")
		rhs.SetReplaceSelf(func(n Node) {
			n.Parent().(*AssignStmtNode).SetRhs(n)
		})
	}
}

func (n *AssignStmtNode) Fields() []string {
	return []string{
		"lhs",
		"op",
		"rhs",
	}
}

func (n *AssignStmtNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "lhs" {
		return n.Lhs()
	}
	if field == "op" {
		return n.Op()
	}
	if field == "rhs" {
		return n.Rhs()
	}
	return nil
}

func (n *AssignStmtNode) SetChild(nodes []Node) {
	if len(nodes) != 3 {
		return
	}
	n.SetLhs(nodes[0])
	n.SetOp(nodes[1])
	n.SetRhs(nodes[2])
}

func (n *AssignStmtNode) Fork() Node {
	_ret := &AssignStmtNode{
		BaseNode: n.BaseNode.fork(),
		lhs:      n.lhs.Fork(),
		op:       n.op.Fork(),
		rhs:      n.rhs.Fork(),
	}
	_ret.lhs.SetParent(_ret)
	_ret.op.SetParent(_ret)
	_ret.rhs.SetParent(_ret)
	return _ret
}

func (n *AssignStmtNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.lhs.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.op.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.rhs.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *AssignStmtNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"assign_stmt\""
	ret["lhs"] = DumpNode(n.Lhs(), hook)
	ret["op"] = DumpNode(n.Op(), hook)
	ret["rhs"] = DumpNode(n.Rhs(), hook)
	return ret
}

func NewBlockStmtNode(filePath string, fileContent []rune, list Node, start, end Position) Node {
	if list == nil {
		list = DummyNode
	}
	_1 := &BlockStmtNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeBlockStmt, start, end),
		list:     list,
	}
	creationHook(_1)
	return _1
}

type BlockStmtNode struct {
	*BaseNode
	list Node
}

func (n *BlockStmtNode) List() Node {
	return n.list
}

func (n *BlockStmtNode) SetList(v Node) {
	n.list = v
}

func (n *BlockStmtNode) BuildLink() {
	if !n.List().IsDummy() {
		list := n.List()
		list.BuildLink()
		list.SetParent(n)
		list.SetSelfField("list")
		list.SetReplaceSelf(func(n Node) {
			n.Parent().(*BlockStmtNode).SetList(n)
		})
	}
}

func (n *BlockStmtNode) Fields() []string {
	return []string{
		"list",
	}
}

func (n *BlockStmtNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "list" {
		return n.List()
	}
	return nil
}

func (n *BlockStmtNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetList(nodes[0])
}

func (n *BlockStmtNode) Fork() Node {
	_ret := &BlockStmtNode{
		BaseNode: n.BaseNode.fork(),
		list:     n.list.Fork(),
	}
	_ret.list.SetParent(_ret)
	return _ret
}

func (n *BlockStmtNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.list.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *BlockStmtNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"block_stmt\""
	ret["list"] = DumpNode(n.List(), hook)
	return ret
}

func NewBranchStmtNode(filePath string, fileContent []rune, tok Node, label Node, start, end Position) Node {
	if tok == nil {
		tok = DummyNode
	}
	if label == nil {
		label = DummyNode
	}
	_1 := &BranchStmtNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeBranchStmt, start, end),
		tok:      tok,
		label:    label,
	}
	creationHook(_1)
	return _1
}

type BranchStmtNode struct {
	*BaseNode
	tok   Node
	label Node
}

func (n *BranchStmtNode) Tok() Node {
	return n.tok
}

func (n *BranchStmtNode) SetTok(v Node) {
	n.tok = v
}

func (n *BranchStmtNode) Label() Node {
	return n.label
}

func (n *BranchStmtNode) SetLabel(v Node) {
	n.label = v
}

func (n *BranchStmtNode) BuildLink() {
	if !n.Tok().IsDummy() {
		tok := n.Tok()
		tok.BuildLink()
		tok.SetParent(n)
		tok.SetSelfField("tok")
		tok.SetReplaceSelf(func(n Node) {
			n.Parent().(*BranchStmtNode).SetTok(n)
		})
	}
	if !n.Label().IsDummy() {
		label := n.Label()
		label.BuildLink()
		label.SetParent(n)
		label.SetSelfField("label")
		label.SetReplaceSelf(func(n Node) {
			n.Parent().(*BranchStmtNode).SetLabel(n)
		})
	}
}

func (n *BranchStmtNode) Fields() []string {
	return []string{
		"tok",
		"label",
	}
}

func (n *BranchStmtNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "tok" {
		return n.Tok()
	}
	if field == "label" {
		return n.Label()
	}
	return nil
}

func (n *BranchStmtNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetTok(nodes[0])
	n.SetLabel(nodes[1])
}

func (n *BranchStmtNode) Fork() Node {
	_ret := &BranchStmtNode{
		BaseNode: n.BaseNode.fork(),
		tok:      n.tok.Fork(),
		label:    n.label.Fork(),
	}
	_ret.tok.SetParent(_ret)
	_ret.label.SetParent(_ret)
	return _ret
}

func (n *BranchStmtNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.tok.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.label.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *BranchStmtNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"branch_stmt\""
	ret["tok"] = DumpNode(n.Tok(), hook)
	ret["label"] = DumpNode(n.Label(), hook)
	return ret
}

func NewDeferStmtNode(filePath string, fileContent []rune, call Node, start, end Position) Node {
	if call == nil {
		call = DummyNode
	}
	_1 := &DeferStmtNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeDeferStmt, start, end),
		call:     call,
	}
	creationHook(_1)
	return _1
}

type DeferStmtNode struct {
	*BaseNode
	call Node
}

func (n *DeferStmtNode) Call() Node {
	return n.call
}

func (n *DeferStmtNode) SetCall(v Node) {
	n.call = v
}

func (n *DeferStmtNode) BuildLink() {
	if !n.Call().IsDummy() {
		call := n.Call()
		call.BuildLink()
		call.SetParent(n)
		call.SetSelfField("call")
		call.SetReplaceSelf(func(n Node) {
			n.Parent().(*DeferStmtNode).SetCall(n)
		})
	}
}

func (n *DeferStmtNode) Fields() []string {
	return []string{
		"call",
	}
}

func (n *DeferStmtNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "call" {
		return n.Call()
	}
	return nil
}

func (n *DeferStmtNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetCall(nodes[0])
}

func (n *DeferStmtNode) Fork() Node {
	_ret := &DeferStmtNode{
		BaseNode: n.BaseNode.fork(),
		call:     n.call.Fork(),
	}
	_ret.call.SetParent(_ret)
	return _ret
}

func (n *DeferStmtNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.call.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *DeferStmtNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"defer_stmt\""
	ret["call"] = DumpNode(n.Call(), hook)
	return ret
}

func NewGoStmtNode(filePath string, fileContent []rune, call Node, start, end Position) Node {
	if call == nil {
		call = DummyNode
	}
	_1 := &GoStmtNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeGoStmt, start, end),
		call:     call,
	}
	creationHook(_1)
	return _1
}

type GoStmtNode struct {
	*BaseNode
	call Node
}

func (n *GoStmtNode) Call() Node {
	return n.call
}

func (n *GoStmtNode) SetCall(v Node) {
	n.call = v
}

func (n *GoStmtNode) BuildLink() {
	if !n.Call().IsDummy() {
		call := n.Call()
		call.BuildLink()
		call.SetParent(n)
		call.SetSelfField("call")
		call.SetReplaceSelf(func(n Node) {
			n.Parent().(*GoStmtNode).SetCall(n)
		})
	}
}

func (n *GoStmtNode) Fields() []string {
	return []string{
		"call",
	}
}

func (n *GoStmtNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "call" {
		return n.Call()
	}
	return nil
}

func (n *GoStmtNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetCall(nodes[0])
}

func (n *GoStmtNode) Fork() Node {
	_ret := &GoStmtNode{
		BaseNode: n.BaseNode.fork(),
		call:     n.call.Fork(),
	}
	_ret.call.SetParent(_ret)
	return _ret
}

func (n *GoStmtNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.call.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *GoStmtNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"go_stmt\""
	ret["call"] = DumpNode(n.Call(), hook)
	return ret
}

func NewSendStmtNode(filePath string, fileContent []rune, chan_ Node, value Node, start, end Position) Node {
	if chan_ == nil {
		chan_ = DummyNode
	}
	if value == nil {
		value = DummyNode
	}
	_1 := &SendStmtNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeSendStmt, start, end),
		chan_:    chan_,
		value:    value,
	}
	creationHook(_1)
	return _1
}

type SendStmtNode struct {
	*BaseNode
	chan_ Node
	value Node
}

func (n *SendStmtNode) Chan() Node {
	return n.chan_
}

func (n *SendStmtNode) SetChan(v Node) {
	n.chan_ = v
}

func (n *SendStmtNode) Value() Node {
	return n.value
}

func (n *SendStmtNode) SetValue(v Node) {
	n.value = v
}

func (n *SendStmtNode) BuildLink() {
	if !n.Chan().IsDummy() {
		chan_ := n.Chan()
		chan_.BuildLink()
		chan_.SetParent(n)
		chan_.SetSelfField("chan_")
		chan_.SetReplaceSelf(func(n Node) {
			n.Parent().(*SendStmtNode).SetChan(n)
		})
	}
	if !n.Value().IsDummy() {
		value := n.Value()
		value.BuildLink()
		value.SetParent(n)
		value.SetSelfField("value")
		value.SetReplaceSelf(func(n Node) {
			n.Parent().(*SendStmtNode).SetValue(n)
		})
	}
}

func (n *SendStmtNode) Fields() []string {
	return []string{
		"chan_",
		"value",
	}
}

func (n *SendStmtNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "chan_" {
		return n.Chan()
	}
	if field == "value" {
		return n.Value()
	}
	return nil
}

func (n *SendStmtNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetChan(nodes[0])
	n.SetValue(nodes[1])
}

func (n *SendStmtNode) Fork() Node {
	_ret := &SendStmtNode{
		BaseNode: n.BaseNode.fork(),
		chan_:    n.chan_.Fork(),
		value:    n.value.Fork(),
	}
	_ret.chan_.SetParent(_ret)
	_ret.value.SetParent(_ret)
	return _ret
}

func (n *SendStmtNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.chan_.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.value.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *SendStmtNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"send_stmt\""
	ret["chan"] = DumpNode(n.Chan(), hook)
	ret["value"] = DumpNode(n.Value(), hook)
	return ret
}

func NewExprStmtNode(filePath string, fileContent []rune, x Node, start, end Position) Node {
	if x == nil {
		x = DummyNode
	}
	_1 := &ExprStmtNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeExprStmt, start, end),
		x:        x,
	}
	creationHook(_1)
	return _1
}

type ExprStmtNode struct {
	*BaseNode
	x Node
}

func (n *ExprStmtNode) X() Node {
	return n.x
}

func (n *ExprStmtNode) SetX(v Node) {
	n.x = v
}

func (n *ExprStmtNode) BuildLink() {
	if !n.X().IsDummy() {
		x := n.X()
		x.BuildLink()
		x.SetParent(n)
		x.SetSelfField("x")
		x.SetReplaceSelf(func(n Node) {
			n.Parent().(*ExprStmtNode).SetX(n)
		})
	}
}

func (n *ExprStmtNode) Fields() []string {
	return []string{
		"x",
	}
}

func (n *ExprStmtNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "x" {
		return n.X()
	}
	return nil
}

func (n *ExprStmtNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetX(nodes[0])
}

func (n *ExprStmtNode) Fork() Node {
	_ret := &ExprStmtNode{
		BaseNode: n.BaseNode.fork(),
		x:        n.x.Fork(),
	}
	_ret.x.SetParent(_ret)
	return _ret
}

func (n *ExprStmtNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.x.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *ExprStmtNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"expr_stmt\""
	ret["x"] = DumpNode(n.X(), hook)
	return ret
}

func NewIncDecStmtNode(filePath string, fileContent []rune, x Node, tok Node, start, end Position) Node {
	if x == nil {
		x = DummyNode
	}
	if tok == nil {
		tok = DummyNode
	}
	_1 := &IncDecStmtNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeIncDecStmt, start, end),
		x:        x,
		tok:      tok,
	}
	creationHook(_1)
	return _1
}

type IncDecStmtNode struct {
	*BaseNode
	x   Node
	tok Node
}

func (n *IncDecStmtNode) X() Node {
	return n.x
}

func (n *IncDecStmtNode) SetX(v Node) {
	n.x = v
}

func (n *IncDecStmtNode) Tok() Node {
	return n.tok
}

func (n *IncDecStmtNode) SetTok(v Node) {
	n.tok = v
}

func (n *IncDecStmtNode) BuildLink() {
	if !n.X().IsDummy() {
		x := n.X()
		x.BuildLink()
		x.SetParent(n)
		x.SetSelfField("x")
		x.SetReplaceSelf(func(n Node) {
			n.Parent().(*IncDecStmtNode).SetX(n)
		})
	}
	if !n.Tok().IsDummy() {
		tok := n.Tok()
		tok.BuildLink()
		tok.SetParent(n)
		tok.SetSelfField("tok")
		tok.SetReplaceSelf(func(n Node) {
			n.Parent().(*IncDecStmtNode).SetTok(n)
		})
	}
}

func (n *IncDecStmtNode) Fields() []string {
	return []string{
		"x",
		"tok",
	}
}

func (n *IncDecStmtNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "x" {
		return n.X()
	}
	if field == "tok" {
		return n.Tok()
	}
	return nil
}

func (n *IncDecStmtNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetX(nodes[0])
	n.SetTok(nodes[1])
}

func (n *IncDecStmtNode) Fork() Node {
	_ret := &IncDecStmtNode{
		BaseNode: n.BaseNode.fork(),
		x:        n.x.Fork(),
		tok:      n.tok.Fork(),
	}
	_ret.x.SetParent(_ret)
	_ret.tok.SetParent(_ret)
	return _ret
}

func (n *IncDecStmtNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.x.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.tok.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *IncDecStmtNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"inc_dec_stmt\""
	ret["x"] = DumpNode(n.X(), hook)
	ret["tok"] = DumpNode(n.Tok(), hook)
	return ret
}

func NewIfStmtNode(filePath string, fileContent []rune, init Node, cond Node, body Node, else_ Node, start, end Position) Node {
	if init == nil {
		init = DummyNode
	}
	if cond == nil {
		cond = DummyNode
	}
	if body == nil {
		body = DummyNode
	}
	if else_ == nil {
		else_ = DummyNode
	}
	_1 := &IfStmtNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeIfStmt, start, end),
		init:     init,
		cond:     cond,
		body:     body,
		else_:    else_,
	}
	creationHook(_1)
	return _1
}

type IfStmtNode struct {
	*BaseNode
	init  Node
	cond  Node
	body  Node
	else_ Node
}

func (n *IfStmtNode) Init() Node {
	return n.init
}

func (n *IfStmtNode) SetInit(v Node) {
	n.init = v
}

func (n *IfStmtNode) Cond() Node {
	return n.cond
}

func (n *IfStmtNode) SetCond(v Node) {
	n.cond = v
}

func (n *IfStmtNode) Body() Node {
	return n.body
}

func (n *IfStmtNode) SetBody(v Node) {
	n.body = v
}

func (n *IfStmtNode) Else() Node {
	return n.else_
}

func (n *IfStmtNode) SetElse(v Node) {
	n.else_ = v
}

func (n *IfStmtNode) BuildLink() {
	if !n.Init().IsDummy() {
		init := n.Init()
		init.BuildLink()
		init.SetParent(n)
		init.SetSelfField("init")
		init.SetReplaceSelf(func(n Node) {
			n.Parent().(*IfStmtNode).SetInit(n)
		})
	}
	if !n.Cond().IsDummy() {
		cond := n.Cond()
		cond.BuildLink()
		cond.SetParent(n)
		cond.SetSelfField("cond")
		cond.SetReplaceSelf(func(n Node) {
			n.Parent().(*IfStmtNode).SetCond(n)
		})
	}
	if !n.Body().IsDummy() {
		body := n.Body()
		body.BuildLink()
		body.SetParent(n)
		body.SetSelfField("body")
		body.SetReplaceSelf(func(n Node) {
			n.Parent().(*IfStmtNode).SetBody(n)
		})
	}
	if !n.Else().IsDummy() {
		else_ := n.Else()
		else_.BuildLink()
		else_.SetParent(n)
		else_.SetSelfField("else_")
		else_.SetReplaceSelf(func(n Node) {
			n.Parent().(*IfStmtNode).SetElse(n)
		})
	}
}

func (n *IfStmtNode) Fields() []string {
	return []string{
		"init",
		"cond",
		"body",
		"else_",
	}
}

func (n *IfStmtNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "init" {
		return n.Init()
	}
	if field == "cond" {
		return n.Cond()
	}
	if field == "body" {
		return n.Body()
	}
	if field == "else_" {
		return n.Else()
	}
	return nil
}

func (n *IfStmtNode) SetChild(nodes []Node) {
	if len(nodes) != 4 {
		return
	}
	n.SetInit(nodes[0])
	n.SetCond(nodes[1])
	n.SetBody(nodes[2])
	n.SetElse(nodes[3])
}

func (n *IfStmtNode) Fork() Node {
	_ret := &IfStmtNode{
		BaseNode: n.BaseNode.fork(),
		init:     n.init.Fork(),
		cond:     n.cond.Fork(),
		body:     n.body.Fork(),
		else_:    n.else_.Fork(),
	}
	_ret.init.SetParent(_ret)
	_ret.cond.SetParent(_ret)
	_ret.body.SetParent(_ret)
	_ret.else_.SetParent(_ret)
	return _ret
}

func (n *IfStmtNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.init.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.cond.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.body.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.else_.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *IfStmtNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"if_stmt\""
	ret["init"] = DumpNode(n.Init(), hook)
	ret["cond"] = DumpNode(n.Cond(), hook)
	ret["body"] = DumpNode(n.Body(), hook)
	ret["else"] = DumpNode(n.Else(), hook)
	return ret
}

func NewForStmtNode(filePath string, fileContent []rune, init Node, cond Node, post Node, body Node, start, end Position) Node {
	if init == nil {
		init = DummyNode
	}
	if cond == nil {
		cond = DummyNode
	}
	if post == nil {
		post = DummyNode
	}
	if body == nil {
		body = DummyNode
	}
	_1 := &ForStmtNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeForStmt, start, end),
		init:     init,
		cond:     cond,
		post:     post,
		body:     body,
	}
	creationHook(_1)
	return _1
}

type ForStmtNode struct {
	*BaseNode
	init Node
	cond Node
	post Node
	body Node
}

func (n *ForStmtNode) Init() Node {
	return n.init
}

func (n *ForStmtNode) SetInit(v Node) {
	n.init = v
}

func (n *ForStmtNode) Cond() Node {
	return n.cond
}

func (n *ForStmtNode) SetCond(v Node) {
	n.cond = v
}

func (n *ForStmtNode) Post() Node {
	return n.post
}

func (n *ForStmtNode) SetPost(v Node) {
	n.post = v
}

func (n *ForStmtNode) Body() Node {
	return n.body
}

func (n *ForStmtNode) SetBody(v Node) {
	n.body = v
}

func (n *ForStmtNode) BuildLink() {
	if !n.Init().IsDummy() {
		init := n.Init()
		init.BuildLink()
		init.SetParent(n)
		init.SetSelfField("init")
		init.SetReplaceSelf(func(n Node) {
			n.Parent().(*ForStmtNode).SetInit(n)
		})
	}
	if !n.Cond().IsDummy() {
		cond := n.Cond()
		cond.BuildLink()
		cond.SetParent(n)
		cond.SetSelfField("cond")
		cond.SetReplaceSelf(func(n Node) {
			n.Parent().(*ForStmtNode).SetCond(n)
		})
	}
	if !n.Post().IsDummy() {
		post := n.Post()
		post.BuildLink()
		post.SetParent(n)
		post.SetSelfField("post")
		post.SetReplaceSelf(func(n Node) {
			n.Parent().(*ForStmtNode).SetPost(n)
		})
	}
	if !n.Body().IsDummy() {
		body := n.Body()
		body.BuildLink()
		body.SetParent(n)
		body.SetSelfField("body")
		body.SetReplaceSelf(func(n Node) {
			n.Parent().(*ForStmtNode).SetBody(n)
		})
	}
}

func (n *ForStmtNode) Fields() []string {
	return []string{
		"init",
		"cond",
		"post",
		"body",
	}
}

func (n *ForStmtNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "init" {
		return n.Init()
	}
	if field == "cond" {
		return n.Cond()
	}
	if field == "post" {
		return n.Post()
	}
	if field == "body" {
		return n.Body()
	}
	return nil
}

func (n *ForStmtNode) SetChild(nodes []Node) {
	if len(nodes) != 4 {
		return
	}
	n.SetInit(nodes[0])
	n.SetCond(nodes[1])
	n.SetPost(nodes[2])
	n.SetBody(nodes[3])
}

func (n *ForStmtNode) Fork() Node {
	_ret := &ForStmtNode{
		BaseNode: n.BaseNode.fork(),
		init:     n.init.Fork(),
		cond:     n.cond.Fork(),
		post:     n.post.Fork(),
		body:     n.body.Fork(),
	}
	_ret.init.SetParent(_ret)
	_ret.cond.SetParent(_ret)
	_ret.post.SetParent(_ret)
	_ret.body.SetParent(_ret)
	return _ret
}

func (n *ForStmtNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.init.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.cond.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.post.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.body.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *ForStmtNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"for_stmt\""
	ret["init"] = DumpNode(n.Init(), hook)
	ret["cond"] = DumpNode(n.Cond(), hook)
	ret["post"] = DumpNode(n.Post(), hook)
	ret["body"] = DumpNode(n.Body(), hook)
	return ret
}

func NewRangeStmtNode(filePath string, fileContent []rune, key Node, value Node, x Node, body Node, tok Node, start, end Position) Node {
	if key == nil {
		key = DummyNode
	}
	if value == nil {
		value = DummyNode
	}
	if x == nil {
		x = DummyNode
	}
	if body == nil {
		body = DummyNode
	}
	if tok == nil {
		tok = DummyNode
	}
	_1 := &RangeStmtNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeRangeStmt, start, end),
		key:      key,
		value:    value,
		x:        x,
		body:     body,
		tok:      tok,
	}
	creationHook(_1)
	return _1
}

type RangeStmtNode struct {
	*BaseNode
	key   Node
	value Node
	x     Node
	body  Node
	tok   Node
}

func (n *RangeStmtNode) Key() Node {
	return n.key
}

func (n *RangeStmtNode) SetKey(v Node) {
	n.key = v
}

func (n *RangeStmtNode) Value() Node {
	return n.value
}

func (n *RangeStmtNode) SetValue(v Node) {
	n.value = v
}

func (n *RangeStmtNode) X() Node {
	return n.x
}

func (n *RangeStmtNode) SetX(v Node) {
	n.x = v
}

func (n *RangeStmtNode) Body() Node {
	return n.body
}

func (n *RangeStmtNode) SetBody(v Node) {
	n.body = v
}

func (n *RangeStmtNode) Tok() Node {
	return n.tok
}

func (n *RangeStmtNode) SetTok(v Node) {
	n.tok = v
}

func (n *RangeStmtNode) BuildLink() {
	if !n.Key().IsDummy() {
		key := n.Key()
		key.BuildLink()
		key.SetParent(n)
		key.SetSelfField("key")
		key.SetReplaceSelf(func(n Node) {
			n.Parent().(*RangeStmtNode).SetKey(n)
		})
	}
	if !n.Value().IsDummy() {
		value := n.Value()
		value.BuildLink()
		value.SetParent(n)
		value.SetSelfField("value")
		value.SetReplaceSelf(func(n Node) {
			n.Parent().(*RangeStmtNode).SetValue(n)
		})
	}
	if !n.X().IsDummy() {
		x := n.X()
		x.BuildLink()
		x.SetParent(n)
		x.SetSelfField("x")
		x.SetReplaceSelf(func(n Node) {
			n.Parent().(*RangeStmtNode).SetX(n)
		})
	}
	if !n.Body().IsDummy() {
		body := n.Body()
		body.BuildLink()
		body.SetParent(n)
		body.SetSelfField("body")
		body.SetReplaceSelf(func(n Node) {
			n.Parent().(*RangeStmtNode).SetBody(n)
		})
	}
	if !n.Tok().IsDummy() {
		tok := n.Tok()
		tok.BuildLink()
		tok.SetParent(n)
		tok.SetSelfField("tok")
		tok.SetReplaceSelf(func(n Node) {
			n.Parent().(*RangeStmtNode).SetTok(n)
		})
	}
}

func (n *RangeStmtNode) Fields() []string {
	return []string{
		"key",
		"value",
		"x",
		"body",
		"tok",
	}
}

func (n *RangeStmtNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "key" {
		return n.Key()
	}
	if field == "value" {
		return n.Value()
	}
	if field == "x" {
		return n.X()
	}
	if field == "body" {
		return n.Body()
	}
	if field == "tok" {
		return n.Tok()
	}
	return nil
}

func (n *RangeStmtNode) SetChild(nodes []Node) {
	if len(nodes) != 5 {
		return
	}
	n.SetKey(nodes[0])
	n.SetValue(nodes[1])
	n.SetX(nodes[2])
	n.SetBody(nodes[3])
	n.SetTok(nodes[4])
}

func (n *RangeStmtNode) Fork() Node {
	_ret := &RangeStmtNode{
		BaseNode: n.BaseNode.fork(),
		key:      n.key.Fork(),
		value:    n.value.Fork(),
		x:        n.x.Fork(),
		body:     n.body.Fork(),
		tok:      n.tok.Fork(),
	}
	_ret.key.SetParent(_ret)
	_ret.value.SetParent(_ret)
	_ret.x.SetParent(_ret)
	_ret.body.SetParent(_ret)
	_ret.tok.SetParent(_ret)
	return _ret
}

func (n *RangeStmtNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.key.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.value.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.x.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.body.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.tok.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *RangeStmtNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"range_stmt\""
	ret["key"] = DumpNode(n.Key(), hook)
	ret["value"] = DumpNode(n.Value(), hook)
	ret["x"] = DumpNode(n.X(), hook)
	ret["body"] = DumpNode(n.Body(), hook)
	ret["tok"] = DumpNode(n.Tok(), hook)
	return ret
}

func NewSelectStmtNode(filePath string, fileContent []rune, body Node, start, end Position) Node {
	if body == nil {
		body = DummyNode
	}
	_1 := &SelectStmtNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeSelectStmt, start, end),
		body:     body,
	}
	creationHook(_1)
	return _1
}

type SelectStmtNode struct {
	*BaseNode
	body Node
}

func (n *SelectStmtNode) Body() Node {
	return n.body
}

func (n *SelectStmtNode) SetBody(v Node) {
	n.body = v
}

func (n *SelectStmtNode) BuildLink() {
	if !n.Body().IsDummy() {
		body := n.Body()
		body.BuildLink()
		body.SetParent(n)
		body.SetSelfField("body")
		body.SetReplaceSelf(func(n Node) {
			n.Parent().(*SelectStmtNode).SetBody(n)
		})
	}
}

func (n *SelectStmtNode) Fields() []string {
	return []string{
		"body",
	}
}

func (n *SelectStmtNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "body" {
		return n.Body()
	}
	return nil
}

func (n *SelectStmtNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetBody(nodes[0])
}

func (n *SelectStmtNode) Fork() Node {
	_ret := &SelectStmtNode{
		BaseNode: n.BaseNode.fork(),
		body:     n.body.Fork(),
	}
	_ret.body.SetParent(_ret)
	return _ret
}

func (n *SelectStmtNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.body.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *SelectStmtNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"select_stmt\""
	ret["body"] = DumpNode(n.Body(), hook)
	return ret
}

func NewSwitchStmtNode(filePath string, fileContent []rune, init Node, tag Node, body Node, start, end Position) Node {
	if init == nil {
		init = DummyNode
	}
	if tag == nil {
		tag = DummyNode
	}
	if body == nil {
		body = DummyNode
	}
	_1 := &SwitchStmtNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeSwitchStmt, start, end),
		init:     init,
		tag:      tag,
		body:     body,
	}
	creationHook(_1)
	return _1
}

type SwitchStmtNode struct {
	*BaseNode
	init Node
	tag  Node
	body Node
}

func (n *SwitchStmtNode) Init() Node {
	return n.init
}

func (n *SwitchStmtNode) SetInit(v Node) {
	n.init = v
}

func (n *SwitchStmtNode) Tag() Node {
	return n.tag
}

func (n *SwitchStmtNode) SetTag(v Node) {
	n.tag = v
}

func (n *SwitchStmtNode) Body() Node {
	return n.body
}

func (n *SwitchStmtNode) SetBody(v Node) {
	n.body = v
}

func (n *SwitchStmtNode) BuildLink() {
	if !n.Init().IsDummy() {
		init := n.Init()
		init.BuildLink()
		init.SetParent(n)
		init.SetSelfField("init")
		init.SetReplaceSelf(func(n Node) {
			n.Parent().(*SwitchStmtNode).SetInit(n)
		})
	}
	if !n.Tag().IsDummy() {
		tag := n.Tag()
		tag.BuildLink()
		tag.SetParent(n)
		tag.SetSelfField("tag")
		tag.SetReplaceSelf(func(n Node) {
			n.Parent().(*SwitchStmtNode).SetTag(n)
		})
	}
	if !n.Body().IsDummy() {
		body := n.Body()
		body.BuildLink()
		body.SetParent(n)
		body.SetSelfField("body")
		body.SetReplaceSelf(func(n Node) {
			n.Parent().(*SwitchStmtNode).SetBody(n)
		})
	}
}

func (n *SwitchStmtNode) Fields() []string {
	return []string{
		"init",
		"tag",
		"body",
	}
}

func (n *SwitchStmtNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "init" {
		return n.Init()
	}
	if field == "tag" {
		return n.Tag()
	}
	if field == "body" {
		return n.Body()
	}
	return nil
}

func (n *SwitchStmtNode) SetChild(nodes []Node) {
	if len(nodes) != 3 {
		return
	}
	n.SetInit(nodes[0])
	n.SetTag(nodes[1])
	n.SetBody(nodes[2])
}

func (n *SwitchStmtNode) Fork() Node {
	_ret := &SwitchStmtNode{
		BaseNode: n.BaseNode.fork(),
		init:     n.init.Fork(),
		tag:      n.tag.Fork(),
		body:     n.body.Fork(),
	}
	_ret.init.SetParent(_ret)
	_ret.tag.SetParent(_ret)
	_ret.body.SetParent(_ret)
	return _ret
}

func (n *SwitchStmtNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.init.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.tag.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.body.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *SwitchStmtNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"switch_stmt\""
	ret["init"] = DumpNode(n.Init(), hook)
	ret["tag"] = DumpNode(n.Tag(), hook)
	ret["body"] = DumpNode(n.Body(), hook)
	return ret
}

func NewTypeSwitchStmtNode(filePath string, fileContent []rune, init Node, assign Node, body Node, start, end Position) Node {
	if init == nil {
		init = DummyNode
	}
	if assign == nil {
		assign = DummyNode
	}
	if body == nil {
		body = DummyNode
	}
	_1 := &TypeSwitchStmtNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeTypeSwitchStmt, start, end),
		init:     init,
		assign:   assign,
		body:     body,
	}
	creationHook(_1)
	return _1
}

type TypeSwitchStmtNode struct {
	*BaseNode
	init   Node
	assign Node
	body   Node
}

func (n *TypeSwitchStmtNode) Init() Node {
	return n.init
}

func (n *TypeSwitchStmtNode) SetInit(v Node) {
	n.init = v
}

func (n *TypeSwitchStmtNode) Assign() Node {
	return n.assign
}

func (n *TypeSwitchStmtNode) SetAssign(v Node) {
	n.assign = v
}

func (n *TypeSwitchStmtNode) Body() Node {
	return n.body
}

func (n *TypeSwitchStmtNode) SetBody(v Node) {
	n.body = v
}

func (n *TypeSwitchStmtNode) BuildLink() {
	if !n.Init().IsDummy() {
		init := n.Init()
		init.BuildLink()
		init.SetParent(n)
		init.SetSelfField("init")
		init.SetReplaceSelf(func(n Node) {
			n.Parent().(*TypeSwitchStmtNode).SetInit(n)
		})
	}
	if !n.Assign().IsDummy() {
		assign := n.Assign()
		assign.BuildLink()
		assign.SetParent(n)
		assign.SetSelfField("assign")
		assign.SetReplaceSelf(func(n Node) {
			n.Parent().(*TypeSwitchStmtNode).SetAssign(n)
		})
	}
	if !n.Body().IsDummy() {
		body := n.Body()
		body.BuildLink()
		body.SetParent(n)
		body.SetSelfField("body")
		body.SetReplaceSelf(func(n Node) {
			n.Parent().(*TypeSwitchStmtNode).SetBody(n)
		})
	}
}

func (n *TypeSwitchStmtNode) Fields() []string {
	return []string{
		"init",
		"assign",
		"body",
	}
}

func (n *TypeSwitchStmtNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "init" {
		return n.Init()
	}
	if field == "assign" {
		return n.Assign()
	}
	if field == "body" {
		return n.Body()
	}
	return nil
}

func (n *TypeSwitchStmtNode) SetChild(nodes []Node) {
	if len(nodes) != 3 {
		return
	}
	n.SetInit(nodes[0])
	n.SetAssign(nodes[1])
	n.SetBody(nodes[2])
}

func (n *TypeSwitchStmtNode) Fork() Node {
	_ret := &TypeSwitchStmtNode{
		BaseNode: n.BaseNode.fork(),
		init:     n.init.Fork(),
		assign:   n.assign.Fork(),
		body:     n.body.Fork(),
	}
	_ret.init.SetParent(_ret)
	_ret.assign.SetParent(_ret)
	_ret.body.SetParent(_ret)
	return _ret
}

func (n *TypeSwitchStmtNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.init.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.assign.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.body.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *TypeSwitchStmtNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"type_switch_stmt\""
	ret["init"] = DumpNode(n.Init(), hook)
	ret["assign"] = DumpNode(n.Assign(), hook)
	ret["body"] = DumpNode(n.Body(), hook)
	return ret
}

func NewReturnStmtNode(filePath string, fileContent []rune, results Node, start, end Position) Node {
	if results == nil {
		results = DummyNode
	}
	_1 := &ReturnStmtNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeReturnStmt, start, end),
		results:  results,
	}
	creationHook(_1)
	return _1
}

type ReturnStmtNode struct {
	*BaseNode
	results Node
}

func (n *ReturnStmtNode) Results() Node {
	return n.results
}

func (n *ReturnStmtNode) SetResults(v Node) {
	n.results = v
}

func (n *ReturnStmtNode) BuildLink() {
	if !n.Results().IsDummy() {
		results := n.Results()
		results.BuildLink()
		results.SetParent(n)
		results.SetSelfField("results")
		results.SetReplaceSelf(func(n Node) {
			n.Parent().(*ReturnStmtNode).SetResults(n)
		})
	}
}

func (n *ReturnStmtNode) Fields() []string {
	return []string{
		"results",
	}
}

func (n *ReturnStmtNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "results" {
		return n.Results()
	}
	return nil
}

func (n *ReturnStmtNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetResults(nodes[0])
}

func (n *ReturnStmtNode) Fork() Node {
	_ret := &ReturnStmtNode{
		BaseNode: n.BaseNode.fork(),
		results:  n.results.Fork(),
	}
	_ret.results.SetParent(_ret)
	return _ret
}

func (n *ReturnStmtNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.results.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *ReturnStmtNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"return_stmt\""
	ret["results"] = DumpNode(n.Results(), hook)
	return ret
}

func NewBinaryExprNode(filePath string, fileContent []rune, x Node, y Node, op Node, start, end Position) Node {
	if x == nil {
		x = DummyNode
	}
	if y == nil {
		y = DummyNode
	}
	if op == nil {
		op = DummyNode
	}
	_1 := &BinaryExprNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeBinaryExpr, start, end),
		x:        x,
		y:        y,
		op:       op,
	}
	creationHook(_1)
	return _1
}

type BinaryExprNode struct {
	*BaseNode
	x  Node
	y  Node
	op Node
}

func (n *BinaryExprNode) X() Node {
	return n.x
}

func (n *BinaryExprNode) SetX(v Node) {
	n.x = v
}

func (n *BinaryExprNode) Y() Node {
	return n.y
}

func (n *BinaryExprNode) SetY(v Node) {
	n.y = v
}

func (n *BinaryExprNode) Op() Node {
	return n.op
}

func (n *BinaryExprNode) SetOp(v Node) {
	n.op = v
}

func (n *BinaryExprNode) BuildLink() {
	if !n.X().IsDummy() {
		x := n.X()
		x.BuildLink()
		x.SetParent(n)
		x.SetSelfField("x")
		x.SetReplaceSelf(func(n Node) {
			n.Parent().(*BinaryExprNode).SetX(n)
		})
	}
	if !n.Y().IsDummy() {
		y := n.Y()
		y.BuildLink()
		y.SetParent(n)
		y.SetSelfField("y")
		y.SetReplaceSelf(func(n Node) {
			n.Parent().(*BinaryExprNode).SetY(n)
		})
	}
	if !n.Op().IsDummy() {
		op := n.Op()
		op.BuildLink()
		op.SetParent(n)
		op.SetSelfField("op")
		op.SetReplaceSelf(func(n Node) {
			n.Parent().(*BinaryExprNode).SetOp(n)
		})
	}
}

func (n *BinaryExprNode) Fields() []string {
	return []string{
		"x",
		"y",
		"op",
	}
}

func (n *BinaryExprNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "x" {
		return n.X()
	}
	if field == "y" {
		return n.Y()
	}
	if field == "op" {
		return n.Op()
	}
	return nil
}

func (n *BinaryExprNode) SetChild(nodes []Node) {
	if len(nodes) != 3 {
		return
	}
	n.SetX(nodes[0])
	n.SetY(nodes[1])
	n.SetOp(nodes[2])
}

func (n *BinaryExprNode) Fork() Node {
	_ret := &BinaryExprNode{
		BaseNode: n.BaseNode.fork(),
		x:        n.x.Fork(),
		y:        n.y.Fork(),
		op:       n.op.Fork(),
	}
	_ret.x.SetParent(_ret)
	_ret.y.SetParent(_ret)
	_ret.op.SetParent(_ret)
	return _ret
}

func (n *BinaryExprNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.x.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.y.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.op.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *BinaryExprNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"binary_expr\""
	ret["x"] = DumpNode(n.X(), hook)
	ret["y"] = DumpNode(n.Y(), hook)
	ret["op"] = DumpNode(n.Op(), hook)
	return ret
}

func NewCallExprNode(filePath string, fileContent []rune, fun Node, typeArgs Node, args Node, start, end Position) Node {
	if fun == nil {
		fun = DummyNode
	}
	if typeArgs == nil {
		typeArgs = DummyNode
	}
	if args == nil {
		args = DummyNode
	}
	_1 := &CallExprNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeCallExpr, start, end),
		fun:      fun,
		typeArgs: typeArgs,
		args:     args,
	}
	creationHook(_1)
	return _1
}

type CallExprNode struct {
	*BaseNode
	fun      Node
	typeArgs Node
	args     Node
}

func (n *CallExprNode) Fun() Node {
	return n.fun
}

func (n *CallExprNode) SetFun(v Node) {
	n.fun = v
}

func (n *CallExprNode) TypeArgs() Node {
	return n.typeArgs
}

func (n *CallExprNode) SetTypeArgs(v Node) {
	n.typeArgs = v
}

func (n *CallExprNode) Args() Node {
	return n.args
}

func (n *CallExprNode) SetArgs(v Node) {
	n.args = v
}

func (n *CallExprNode) BuildLink() {
	if !n.Fun().IsDummy() {
		fun := n.Fun()
		fun.BuildLink()
		fun.SetParent(n)
		fun.SetSelfField("fun")
		fun.SetReplaceSelf(func(n Node) {
			n.Parent().(*CallExprNode).SetFun(n)
		})
	}
	if !n.TypeArgs().IsDummy() {
		typeArgs := n.TypeArgs()
		typeArgs.BuildLink()
		typeArgs.SetParent(n)
		typeArgs.SetSelfField("type_args")
		typeArgs.SetReplaceSelf(func(n Node) {
			n.Parent().(*CallExprNode).SetTypeArgs(n)
		})
	}
	if !n.Args().IsDummy() {
		args := n.Args()
		args.BuildLink()
		args.SetParent(n)
		args.SetSelfField("args")
		args.SetReplaceSelf(func(n Node) {
			n.Parent().(*CallExprNode).SetArgs(n)
		})
	}
}

func (n *CallExprNode) Fields() []string {
	return []string{
		"fun",
		"type_args",
		"args",
	}
}

func (n *CallExprNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "fun" {
		return n.Fun()
	}
	if field == "type_args" {
		return n.TypeArgs()
	}
	if field == "args" {
		return n.Args()
	}
	return nil
}

func (n *CallExprNode) SetChild(nodes []Node) {
	if len(nodes) != 3 {
		return
	}
	n.SetFun(nodes[0])
	n.SetTypeArgs(nodes[1])
	n.SetArgs(nodes[2])
}

func (n *CallExprNode) Fork() Node {
	_ret := &CallExprNode{
		BaseNode: n.BaseNode.fork(),
		fun:      n.fun.Fork(),
		typeArgs: n.typeArgs.Fork(),
		args:     n.args.Fork(),
	}
	_ret.fun.SetParent(_ret)
	_ret.typeArgs.SetParent(_ret)
	_ret.args.SetParent(_ret)
	return _ret
}

func (n *CallExprNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.fun.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.typeArgs.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.args.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *CallExprNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"call_expr\""
	ret["fun"] = DumpNode(n.Fun(), hook)
	ret["type_args"] = DumpNode(n.TypeArgs(), hook)
	ret["args"] = DumpNode(n.Args(), hook)
	return ret
}

func NewIndexExprNode(filePath string, fileContent []rune, x Node, index Node, start, end Position) Node {
	if x == nil {
		x = DummyNode
	}
	if index == nil {
		index = DummyNode
	}
	_1 := &IndexExprNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeIndexExpr, start, end),
		x:        x,
		index:    index,
	}
	creationHook(_1)
	return _1
}

type IndexExprNode struct {
	*BaseNode
	x     Node
	index Node
}

func (n *IndexExprNode) X() Node {
	return n.x
}

func (n *IndexExprNode) SetX(v Node) {
	n.x = v
}

func (n *IndexExprNode) Index() Node {
	return n.index
}

func (n *IndexExprNode) SetIndex(v Node) {
	n.index = v
}

func (n *IndexExprNode) BuildLink() {
	if !n.X().IsDummy() {
		x := n.X()
		x.BuildLink()
		x.SetParent(n)
		x.SetSelfField("x")
		x.SetReplaceSelf(func(n Node) {
			n.Parent().(*IndexExprNode).SetX(n)
		})
	}
	if !n.Index().IsDummy() {
		index := n.Index()
		index.BuildLink()
		index.SetParent(n)
		index.SetSelfField("index")
		index.SetReplaceSelf(func(n Node) {
			n.Parent().(*IndexExprNode).SetIndex(n)
		})
	}
}

func (n *IndexExprNode) Fields() []string {
	return []string{
		"x",
		"index",
	}
}

func (n *IndexExprNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "x" {
		return n.X()
	}
	if field == "index" {
		return n.Index()
	}
	return nil
}

func (n *IndexExprNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetX(nodes[0])
	n.SetIndex(nodes[1])
}

func (n *IndexExprNode) Fork() Node {
	_ret := &IndexExprNode{
		BaseNode: n.BaseNode.fork(),
		x:        n.x.Fork(),
		index:    n.index.Fork(),
	}
	_ret.x.SetParent(_ret)
	_ret.index.SetParent(_ret)
	return _ret
}

func (n *IndexExprNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.x.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.index.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *IndexExprNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"index_expr\""
	ret["x"] = DumpNode(n.X(), hook)
	ret["index"] = DumpNode(n.Index(), hook)
	return ret
}

func NewKeyValueExprNode(filePath string, fileContent []rune, key Node, value Node, start, end Position) Node {
	if key == nil {
		key = DummyNode
	}
	if value == nil {
		value = DummyNode
	}
	_1 := &KeyValueExprNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeKeyValueExpr, start, end),
		key:      key,
		value:    value,
	}
	creationHook(_1)
	return _1
}

type KeyValueExprNode struct {
	*BaseNode
	key   Node
	value Node
}

func (n *KeyValueExprNode) Key() Node {
	return n.key
}

func (n *KeyValueExprNode) SetKey(v Node) {
	n.key = v
}

func (n *KeyValueExprNode) Value() Node {
	return n.value
}

func (n *KeyValueExprNode) SetValue(v Node) {
	n.value = v
}

func (n *KeyValueExprNode) BuildLink() {
	if !n.Key().IsDummy() {
		key := n.Key()
		key.BuildLink()
		key.SetParent(n)
		key.SetSelfField("key")
		key.SetReplaceSelf(func(n Node) {
			n.Parent().(*KeyValueExprNode).SetKey(n)
		})
	}
	if !n.Value().IsDummy() {
		value := n.Value()
		value.BuildLink()
		value.SetParent(n)
		value.SetSelfField("value")
		value.SetReplaceSelf(func(n Node) {
			n.Parent().(*KeyValueExprNode).SetValue(n)
		})
	}
}

func (n *KeyValueExprNode) Fields() []string {
	return []string{
		"key",
		"value",
	}
}

func (n *KeyValueExprNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "key" {
		return n.Key()
	}
	if field == "value" {
		return n.Value()
	}
	return nil
}

func (n *KeyValueExprNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetKey(nodes[0])
	n.SetValue(nodes[1])
}

func (n *KeyValueExprNode) Fork() Node {
	_ret := &KeyValueExprNode{
		BaseNode: n.BaseNode.fork(),
		key:      n.key.Fork(),
		value:    n.value.Fork(),
	}
	_ret.key.SetParent(_ret)
	_ret.value.SetParent(_ret)
	return _ret
}

func (n *KeyValueExprNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.key.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.value.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *KeyValueExprNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"key_value_expr\""
	ret["key"] = DumpNode(n.Key(), hook)
	ret["value"] = DumpNode(n.Value(), hook)
	return ret
}

func NewParenExprNode(filePath string, fileContent []rune, x Node, start, end Position) Node {
	if x == nil {
		x = DummyNode
	}
	_1 := &ParenExprNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeParenExpr, start, end),
		x:        x,
	}
	creationHook(_1)
	return _1
}

type ParenExprNode struct {
	*BaseNode
	x Node
}

func (n *ParenExprNode) X() Node {
	return n.x
}

func (n *ParenExprNode) SetX(v Node) {
	n.x = v
}

func (n *ParenExprNode) BuildLink() {
	if !n.X().IsDummy() {
		x := n.X()
		x.BuildLink()
		x.SetParent(n)
		x.SetSelfField("x")
		x.SetReplaceSelf(func(n Node) {
			n.Parent().(*ParenExprNode).SetX(n)
		})
	}
}

func (n *ParenExprNode) Fields() []string {
	return []string{
		"x",
	}
}

func (n *ParenExprNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "x" {
		return n.X()
	}
	return nil
}

func (n *ParenExprNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetX(nodes[0])
}

func (n *ParenExprNode) Fork() Node {
	_ret := &ParenExprNode{
		BaseNode: n.BaseNode.fork(),
		x:        n.x.Fork(),
	}
	_ret.x.SetParent(_ret)
	return _ret
}

func (n *ParenExprNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.x.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *ParenExprNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"paren_expr\""
	ret["x"] = DumpNode(n.X(), hook)
	return ret
}

func NewSelectorExprNode(filePath string, fileContent []rune, x Node, sel Node, start, end Position) Node {
	if x == nil {
		x = DummyNode
	}
	if sel == nil {
		sel = DummyNode
	}
	_1 := &SelectorExprNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeSelectorExpr, start, end),
		x:        x,
		sel:      sel,
	}
	creationHook(_1)
	return _1
}

type SelectorExprNode struct {
	*BaseNode
	x   Node
	sel Node
}

func (n *SelectorExprNode) X() Node {
	return n.x
}

func (n *SelectorExprNode) SetX(v Node) {
	n.x = v
}

func (n *SelectorExprNode) Sel() Node {
	return n.sel
}

func (n *SelectorExprNode) SetSel(v Node) {
	n.sel = v
}

func (n *SelectorExprNode) BuildLink() {
	if !n.X().IsDummy() {
		x := n.X()
		x.BuildLink()
		x.SetParent(n)
		x.SetSelfField("x")
		x.SetReplaceSelf(func(n Node) {
			n.Parent().(*SelectorExprNode).SetX(n)
		})
	}
	if !n.Sel().IsDummy() {
		sel := n.Sel()
		sel.BuildLink()
		sel.SetParent(n)
		sel.SetSelfField("sel")
		sel.SetReplaceSelf(func(n Node) {
			n.Parent().(*SelectorExprNode).SetSel(n)
		})
	}
}

func (n *SelectorExprNode) Fields() []string {
	return []string{
		"x",
		"sel",
	}
}

func (n *SelectorExprNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "x" {
		return n.X()
	}
	if field == "sel" {
		return n.Sel()
	}
	return nil
}

func (n *SelectorExprNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetX(nodes[0])
	n.SetSel(nodes[1])
}

func (n *SelectorExprNode) Fork() Node {
	_ret := &SelectorExprNode{
		BaseNode: n.BaseNode.fork(),
		x:        n.x.Fork(),
		sel:      n.sel.Fork(),
	}
	_ret.x.SetParent(_ret)
	_ret.sel.SetParent(_ret)
	return _ret
}

func (n *SelectorExprNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.x.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.sel.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *SelectorExprNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"selector_expr\""
	ret["x"] = DumpNode(n.X(), hook)
	ret["sel"] = DumpNode(n.Sel(), hook)
	return ret
}

func NewStarExprNode(filePath string, fileContent []rune, x Node, start, end Position) Node {
	if x == nil {
		x = DummyNode
	}
	_1 := &StarExprNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeStarExpr, start, end),
		x:        x,
	}
	creationHook(_1)
	return _1
}

type StarExprNode struct {
	*BaseNode
	x Node
}

func (n *StarExprNode) X() Node {
	return n.x
}

func (n *StarExprNode) SetX(v Node) {
	n.x = v
}

func (n *StarExprNode) BuildLink() {
	if !n.X().IsDummy() {
		x := n.X()
		x.BuildLink()
		x.SetParent(n)
		x.SetSelfField("x")
		x.SetReplaceSelf(func(n Node) {
			n.Parent().(*StarExprNode).SetX(n)
		})
	}
}

func (n *StarExprNode) Fields() []string {
	return []string{
		"x",
	}
}

func (n *StarExprNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "x" {
		return n.X()
	}
	return nil
}

func (n *StarExprNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetX(nodes[0])
}

func (n *StarExprNode) Fork() Node {
	_ret := &StarExprNode{
		BaseNode: n.BaseNode.fork(),
		x:        n.x.Fork(),
	}
	_ret.x.SetParent(_ret)
	return _ret
}

func (n *StarExprNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.x.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *StarExprNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"star_expr\""
	ret["x"] = DumpNode(n.X(), hook)
	return ret
}

func NewTypeAssertExprNode(filePath string, fileContent []rune, x Node, type_ Node, start, end Position) Node {
	if x == nil {
		x = DummyNode
	}
	if type_ == nil {
		type_ = DummyNode
	}
	_1 := &TypeAssertExprNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeTypeAssertExpr, start, end),
		x:        x,
		type_:    type_,
	}
	creationHook(_1)
	return _1
}

type TypeAssertExprNode struct {
	*BaseNode
	x     Node
	type_ Node
}

func (n *TypeAssertExprNode) X() Node {
	return n.x
}

func (n *TypeAssertExprNode) SetX(v Node) {
	n.x = v
}

func (n *TypeAssertExprNode) Type() Node {
	return n.type_
}

func (n *TypeAssertExprNode) SetType(v Node) {
	n.type_ = v
}

func (n *TypeAssertExprNode) BuildLink() {
	if !n.X().IsDummy() {
		x := n.X()
		x.BuildLink()
		x.SetParent(n)
		x.SetSelfField("x")
		x.SetReplaceSelf(func(n Node) {
			n.Parent().(*TypeAssertExprNode).SetX(n)
		})
	}
	if !n.Type().IsDummy() {
		type_ := n.Type()
		type_.BuildLink()
		type_.SetParent(n)
		type_.SetSelfField("type_")
		type_.SetReplaceSelf(func(n Node) {
			n.Parent().(*TypeAssertExprNode).SetType(n)
		})
	}
}

func (n *TypeAssertExprNode) Fields() []string {
	return []string{
		"x",
		"type_",
	}
}

func (n *TypeAssertExprNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "x" {
		return n.X()
	}
	if field == "type_" {
		return n.Type()
	}
	return nil
}

func (n *TypeAssertExprNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetX(nodes[0])
	n.SetType(nodes[1])
}

func (n *TypeAssertExprNode) Fork() Node {
	_ret := &TypeAssertExprNode{
		BaseNode: n.BaseNode.fork(),
		x:        n.x.Fork(),
		type_:    n.type_.Fork(),
	}
	_ret.x.SetParent(_ret)
	_ret.type_.SetParent(_ret)
	return _ret
}

func (n *TypeAssertExprNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.x.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.type_.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *TypeAssertExprNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"type_assert_expr\""
	ret["x"] = DumpNode(n.X(), hook)
	ret["type"] = DumpNode(n.Type(), hook)
	return ret
}

func NewSliceExprNode(filePath string, fileContent []rune, x Node, low Node, high Node, max_ Node, start, end Position) Node {
	if x == nil {
		x = DummyNode
	}
	if low == nil {
		low = DummyNode
	}
	if high == nil {
		high = DummyNode
	}
	if max_ == nil {
		max_ = DummyNode
	}
	_1 := &SliceExprNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeSliceExpr, start, end),
		x:        x,
		low:      low,
		high:     high,
		max_:     max_,
	}
	creationHook(_1)
	return _1
}

type SliceExprNode struct {
	*BaseNode
	x    Node
	low  Node
	high Node
	max_ Node
}

func (n *SliceExprNode) X() Node {
	return n.x
}

func (n *SliceExprNode) SetX(v Node) {
	n.x = v
}

func (n *SliceExprNode) Low() Node {
	return n.low
}

func (n *SliceExprNode) SetLow(v Node) {
	n.low = v
}

func (n *SliceExprNode) High() Node {
	return n.high
}

func (n *SliceExprNode) SetHigh(v Node) {
	n.high = v
}

func (n *SliceExprNode) Max() Node {
	return n.max_
}

func (n *SliceExprNode) SetMax(v Node) {
	n.max_ = v
}

func (n *SliceExprNode) BuildLink() {
	if !n.X().IsDummy() {
		x := n.X()
		x.BuildLink()
		x.SetParent(n)
		x.SetSelfField("x")
		x.SetReplaceSelf(func(n Node) {
			n.Parent().(*SliceExprNode).SetX(n)
		})
	}
	if !n.Low().IsDummy() {
		low := n.Low()
		low.BuildLink()
		low.SetParent(n)
		low.SetSelfField("low")
		low.SetReplaceSelf(func(n Node) {
			n.Parent().(*SliceExprNode).SetLow(n)
		})
	}
	if !n.High().IsDummy() {
		high := n.High()
		high.BuildLink()
		high.SetParent(n)
		high.SetSelfField("high")
		high.SetReplaceSelf(func(n Node) {
			n.Parent().(*SliceExprNode).SetHigh(n)
		})
	}
	if !n.Max().IsDummy() {
		max_ := n.Max()
		max_.BuildLink()
		max_.SetParent(n)
		max_.SetSelfField("max_")
		max_.SetReplaceSelf(func(n Node) {
			n.Parent().(*SliceExprNode).SetMax(n)
		})
	}
}

func (n *SliceExprNode) Fields() []string {
	return []string{
		"x",
		"low",
		"high",
		"max_",
	}
}

func (n *SliceExprNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "x" {
		return n.X()
	}
	if field == "low" {
		return n.Low()
	}
	if field == "high" {
		return n.High()
	}
	if field == "max_" {
		return n.Max()
	}
	return nil
}

func (n *SliceExprNode) SetChild(nodes []Node) {
	if len(nodes) != 4 {
		return
	}
	n.SetX(nodes[0])
	n.SetLow(nodes[1])
	n.SetHigh(nodes[2])
	n.SetMax(nodes[3])
}

func (n *SliceExprNode) Fork() Node {
	_ret := &SliceExprNode{
		BaseNode: n.BaseNode.fork(),
		x:        n.x.Fork(),
		low:      n.low.Fork(),
		high:     n.high.Fork(),
		max_:     n.max_.Fork(),
	}
	_ret.x.SetParent(_ret)
	_ret.low.SetParent(_ret)
	_ret.high.SetParent(_ret)
	_ret.max_.SetParent(_ret)
	return _ret
}

func (n *SliceExprNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.x.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.low.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.high.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.max_.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *SliceExprNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"slice_expr\""
	ret["x"] = DumpNode(n.X(), hook)
	ret["low"] = DumpNode(n.Low(), hook)
	ret["high"] = DumpNode(n.High(), hook)
	ret["max"] = DumpNode(n.Max(), hook)
	return ret
}

func NewUnaryExprNode(filePath string, fileContent []rune, op Node, x Node, start, end Position) Node {
	if op == nil {
		op = DummyNode
	}
	if x == nil {
		x = DummyNode
	}
	_1 := &UnaryExprNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeUnaryExpr, start, end),
		op:       op,
		x:        x,
	}
	creationHook(_1)
	return _1
}

type UnaryExprNode struct {
	*BaseNode
	op Node
	x  Node
}

func (n *UnaryExprNode) Op() Node {
	return n.op
}

func (n *UnaryExprNode) SetOp(v Node) {
	n.op = v
}

func (n *UnaryExprNode) X() Node {
	return n.x
}

func (n *UnaryExprNode) SetX(v Node) {
	n.x = v
}

func (n *UnaryExprNode) BuildLink() {
	if !n.Op().IsDummy() {
		op := n.Op()
		op.BuildLink()
		op.SetParent(n)
		op.SetSelfField("op")
		op.SetReplaceSelf(func(n Node) {
			n.Parent().(*UnaryExprNode).SetOp(n)
		})
	}
	if !n.X().IsDummy() {
		x := n.X()
		x.BuildLink()
		x.SetParent(n)
		x.SetSelfField("x")
		x.SetReplaceSelf(func(n Node) {
			n.Parent().(*UnaryExprNode).SetX(n)
		})
	}
}

func (n *UnaryExprNode) Fields() []string {
	return []string{
		"op",
		"x",
	}
}

func (n *UnaryExprNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "op" {
		return n.Op()
	}
	if field == "x" {
		return n.X()
	}
	return nil
}

func (n *UnaryExprNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetOp(nodes[0])
	n.SetX(nodes[1])
}

func (n *UnaryExprNode) Fork() Node {
	_ret := &UnaryExprNode{
		BaseNode: n.BaseNode.fork(),
		op:       n.op.Fork(),
		x:        n.x.Fork(),
	}
	_ret.op.SetParent(_ret)
	_ret.x.SetParent(_ret)
	return _ret
}

func (n *UnaryExprNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.op.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.x.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *UnaryExprNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"unary_expr\""
	ret["op"] = DumpNode(n.Op(), hook)
	ret["x"] = DumpNode(n.X(), hook)
	return ret
}

func NewArrayTypeNode(filePath string, fileContent []rune, len_ Node, elt Node, start, end Position) Node {
	if len_ == nil {
		len_ = DummyNode
	}
	if elt == nil {
		elt = DummyNode
	}
	_1 := &ArrayTypeNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeArrayType, start, end),
		len_:     len_,
		elt:      elt,
	}
	creationHook(_1)
	return _1
}

type ArrayTypeNode struct {
	*BaseNode
	len_ Node
	elt  Node
}

func (n *ArrayTypeNode) Len() Node {
	return n.len_
}

func (n *ArrayTypeNode) SetLen(v Node) {
	n.len_ = v
}

func (n *ArrayTypeNode) Elt() Node {
	return n.elt
}

func (n *ArrayTypeNode) SetElt(v Node) {
	n.elt = v
}

func (n *ArrayTypeNode) BuildLink() {
	if !n.Len().IsDummy() {
		len_ := n.Len()
		len_.BuildLink()
		len_.SetParent(n)
		len_.SetSelfField("len_")
		len_.SetReplaceSelf(func(n Node) {
			n.Parent().(*ArrayTypeNode).SetLen(n)
		})
	}
	if !n.Elt().IsDummy() {
		elt := n.Elt()
		elt.BuildLink()
		elt.SetParent(n)
		elt.SetSelfField("elt")
		elt.SetReplaceSelf(func(n Node) {
			n.Parent().(*ArrayTypeNode).SetElt(n)
		})
	}
}

func (n *ArrayTypeNode) Fields() []string {
	return []string{
		"len_",
		"elt",
	}
}

func (n *ArrayTypeNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "len_" {
		return n.Len()
	}
	if field == "elt" {
		return n.Elt()
	}
	return nil
}

func (n *ArrayTypeNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetLen(nodes[0])
	n.SetElt(nodes[1])
}

func (n *ArrayTypeNode) Fork() Node {
	_ret := &ArrayTypeNode{
		BaseNode: n.BaseNode.fork(),
		len_:     n.len_.Fork(),
		elt:      n.elt.Fork(),
	}
	_ret.len_.SetParent(_ret)
	_ret.elt.SetParent(_ret)
	return _ret
}

func (n *ArrayTypeNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.len_.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.elt.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *ArrayTypeNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"array_type\""
	ret["len"] = DumpNode(n.Len(), hook)
	ret["elt"] = DumpNode(n.Elt(), hook)
	return ret
}

func NewChanTypeNode(filePath string, fileContent []rune, dir Node, value Node, start, end Position) Node {
	if dir == nil {
		dir = DummyNode
	}
	if value == nil {
		value = DummyNode
	}
	_1 := &ChanTypeNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeChanType, start, end),
		dir:      dir,
		value:    value,
	}
	creationHook(_1)
	return _1
}

type ChanTypeNode struct {
	*BaseNode
	dir   Node
	value Node
}

func (n *ChanTypeNode) Dir() Node {
	return n.dir
}

func (n *ChanTypeNode) SetDir(v Node) {
	n.dir = v
}

func (n *ChanTypeNode) Value() Node {
	return n.value
}

func (n *ChanTypeNode) SetValue(v Node) {
	n.value = v
}

func (n *ChanTypeNode) BuildLink() {
	if !n.Dir().IsDummy() {
		dir := n.Dir()
		dir.BuildLink()
		dir.SetParent(n)
		dir.SetSelfField("dir")
		dir.SetReplaceSelf(func(n Node) {
			n.Parent().(*ChanTypeNode).SetDir(n)
		})
	}
	if !n.Value().IsDummy() {
		value := n.Value()
		value.BuildLink()
		value.SetParent(n)
		value.SetSelfField("value")
		value.SetReplaceSelf(func(n Node) {
			n.Parent().(*ChanTypeNode).SetValue(n)
		})
	}
}

func (n *ChanTypeNode) Fields() []string {
	return []string{
		"dir",
		"value",
	}
}

func (n *ChanTypeNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "dir" {
		return n.Dir()
	}
	if field == "value" {
		return n.Value()
	}
	return nil
}

func (n *ChanTypeNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetDir(nodes[0])
	n.SetValue(nodes[1])
}

func (n *ChanTypeNode) Fork() Node {
	_ret := &ChanTypeNode{
		BaseNode: n.BaseNode.fork(),
		dir:      n.dir.Fork(),
		value:    n.value.Fork(),
	}
	_ret.dir.SetParent(_ret)
	_ret.value.SetParent(_ret)
	return _ret
}

func (n *ChanTypeNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.dir.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.value.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *ChanTypeNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"chan_type\""
	ret["dir"] = DumpNode(n.Dir(), hook)
	ret["value"] = DumpNode(n.Value(), hook)
	return ret
}

func NewFunctionTypeNode(filePath string, fileContent []rune, params Node, results Node, start, end Position) Node {
	if params == nil {
		params = DummyNode
	}
	if results == nil {
		results = DummyNode
	}
	_1 := &FunctionTypeNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeFunctionType, start, end),
		params:   params,
		results:  results,
	}
	creationHook(_1)
	return _1
}

type FunctionTypeNode struct {
	*BaseNode
	params  Node
	results Node
}

func (n *FunctionTypeNode) Params() Node {
	return n.params
}

func (n *FunctionTypeNode) SetParams(v Node) {
	n.params = v
}

func (n *FunctionTypeNode) Results() Node {
	return n.results
}

func (n *FunctionTypeNode) SetResults(v Node) {
	n.results = v
}

func (n *FunctionTypeNode) BuildLink() {
	if !n.Params().IsDummy() {
		params := n.Params()
		params.BuildLink()
		params.SetParent(n)
		params.SetSelfField("params")
		params.SetReplaceSelf(func(n Node) {
			n.Parent().(*FunctionTypeNode).SetParams(n)
		})
	}
	if !n.Results().IsDummy() {
		results := n.Results()
		results.BuildLink()
		results.SetParent(n)
		results.SetSelfField("results")
		results.SetReplaceSelf(func(n Node) {
			n.Parent().(*FunctionTypeNode).SetResults(n)
		})
	}
}

func (n *FunctionTypeNode) Fields() []string {
	return []string{
		"params",
		"results",
	}
}

func (n *FunctionTypeNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "params" {
		return n.Params()
	}
	if field == "results" {
		return n.Results()
	}
	return nil
}

func (n *FunctionTypeNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetParams(nodes[0])
	n.SetResults(nodes[1])
}

func (n *FunctionTypeNode) Fork() Node {
	_ret := &FunctionTypeNode{
		BaseNode: n.BaseNode.fork(),
		params:   n.params.Fork(),
		results:  n.results.Fork(),
	}
	_ret.params.SetParent(_ret)
	_ret.results.SetParent(_ret)
	return _ret
}

func (n *FunctionTypeNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.params.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.results.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *FunctionTypeNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"function_type\""
	ret["params"] = DumpNode(n.Params(), hook)
	ret["results"] = DumpNode(n.Results(), hook)
	return ret
}

func NewInterfaceTypeNode(filePath string, fileContent []rune, methods Node, start, end Position) Node {
	if methods == nil {
		methods = DummyNode
	}
	_1 := &InterfaceTypeNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeInterfaceType, start, end),
		methods:  methods,
	}
	creationHook(_1)
	return _1
}

type InterfaceTypeNode struct {
	*BaseNode
	methods Node
}

func (n *InterfaceTypeNode) Methods() Node {
	return n.methods
}

func (n *InterfaceTypeNode) SetMethods(v Node) {
	n.methods = v
}

func (n *InterfaceTypeNode) BuildLink() {
	if !n.Methods().IsDummy() {
		methods := n.Methods()
		methods.BuildLink()
		methods.SetParent(n)
		methods.SetSelfField("methods")
		methods.SetReplaceSelf(func(n Node) {
			n.Parent().(*InterfaceTypeNode).SetMethods(n)
		})
	}
}

func (n *InterfaceTypeNode) Fields() []string {
	return []string{
		"methods",
	}
}

func (n *InterfaceTypeNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "methods" {
		return n.Methods()
	}
	return nil
}

func (n *InterfaceTypeNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetMethods(nodes[0])
}

func (n *InterfaceTypeNode) Fork() Node {
	_ret := &InterfaceTypeNode{
		BaseNode: n.BaseNode.fork(),
		methods:  n.methods.Fork(),
	}
	_ret.methods.SetParent(_ret)
	return _ret
}

func (n *InterfaceTypeNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.methods.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *InterfaceTypeNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"interface_type\""
	ret["methods"] = DumpNode(n.Methods(), hook)
	return ret
}

func NewMapTypeNode(filePath string, fileContent []rune, key Node, value Node, start, end Position) Node {
	if key == nil {
		key = DummyNode
	}
	if value == nil {
		value = DummyNode
	}
	_1 := &MapTypeNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeMapType, start, end),
		key:      key,
		value:    value,
	}
	creationHook(_1)
	return _1
}

type MapTypeNode struct {
	*BaseNode
	key   Node
	value Node
}

func (n *MapTypeNode) Key() Node {
	return n.key
}

func (n *MapTypeNode) SetKey(v Node) {
	n.key = v
}

func (n *MapTypeNode) Value() Node {
	return n.value
}

func (n *MapTypeNode) SetValue(v Node) {
	n.value = v
}

func (n *MapTypeNode) BuildLink() {
	if !n.Key().IsDummy() {
		key := n.Key()
		key.BuildLink()
		key.SetParent(n)
		key.SetSelfField("key")
		key.SetReplaceSelf(func(n Node) {
			n.Parent().(*MapTypeNode).SetKey(n)
		})
	}
	if !n.Value().IsDummy() {
		value := n.Value()
		value.BuildLink()
		value.SetParent(n)
		value.SetSelfField("value")
		value.SetReplaceSelf(func(n Node) {
			n.Parent().(*MapTypeNode).SetValue(n)
		})
	}
}

func (n *MapTypeNode) Fields() []string {
	return []string{
		"key",
		"value",
	}
}

func (n *MapTypeNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "key" {
		return n.Key()
	}
	if field == "value" {
		return n.Value()
	}
	return nil
}

func (n *MapTypeNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetKey(nodes[0])
	n.SetValue(nodes[1])
}

func (n *MapTypeNode) Fork() Node {
	_ret := &MapTypeNode{
		BaseNode: n.BaseNode.fork(),
		key:      n.key.Fork(),
		value:    n.value.Fork(),
	}
	_ret.key.SetParent(_ret)
	_ret.value.SetParent(_ret)
	return _ret
}

func (n *MapTypeNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.key.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.value.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *MapTypeNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"map_type\""
	ret["key"] = DumpNode(n.Key(), hook)
	ret["value"] = DumpNode(n.Value(), hook)
	return ret
}

func NewStructTypeNode(filePath string, fileContent []rune, x Node, start, end Position) Node {
	if x == nil {
		x = DummyNode
	}
	_1 := &StructTypeNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeStructType, start, end),
		x:        x,
	}
	creationHook(_1)
	return _1
}

type StructTypeNode struct {
	*BaseNode
	x Node
}

func (n *StructTypeNode) X() Node {
	return n.x
}

func (n *StructTypeNode) SetX(v Node) {
	n.x = v
}

func (n *StructTypeNode) BuildLink() {
	if !n.X().IsDummy() {
		x := n.X()
		x.BuildLink()
		x.SetParent(n)
		x.SetSelfField("x")
		x.SetReplaceSelf(func(n Node) {
			n.Parent().(*StructTypeNode).SetX(n)
		})
	}
}

func (n *StructTypeNode) Fields() []string {
	return []string{
		"x",
	}
}

func (n *StructTypeNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "x" {
		return n.X()
	}
	return nil
}

func (n *StructTypeNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetX(nodes[0])
}

func (n *StructTypeNode) Fork() Node {
	_ret := &StructTypeNode{
		BaseNode: n.BaseNode.fork(),
		x:        n.x.Fork(),
	}
	_ret.x.SetParent(_ret)
	return _ret
}

func (n *StructTypeNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.x.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *StructTypeNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"struct_type\""
	ret["x"] = DumpNode(n.X(), hook)
	return ret
}

func NewBasicLitNode(filePath string, fileContent []rune, value Node, start, end Position) Node {
	if value == nil {
		value = DummyNode
	}
	_1 := &BasicLitNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeBasicLit, start, end),
		value:    value,
	}
	creationHook(_1)
	return _1
}

type BasicLitNode struct {
	*BaseNode
	value Node
}

func (n *BasicLitNode) Value() Node {
	return n.value
}

func (n *BasicLitNode) SetValue(v Node) {
	n.value = v
}

func (n *BasicLitNode) BuildLink() {
	if !n.Value().IsDummy() {
		value := n.Value()
		value.BuildLink()
		value.SetParent(n)
		value.SetSelfField("value")
		value.SetReplaceSelf(func(n Node) {
			n.Parent().(*BasicLitNode).SetValue(n)
		})
	}
}

func (n *BasicLitNode) Fields() []string {
	return []string{
		"value",
	}
}

func (n *BasicLitNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "value" {
		return n.Value()
	}
	return nil
}

func (n *BasicLitNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetValue(nodes[0])
}

func (n *BasicLitNode) Fork() Node {
	_ret := &BasicLitNode{
		BaseNode: n.BaseNode.fork(),
		value:    n.value.Fork(),
	}
	_ret.value.SetParent(_ret)
	return _ret
}

func (n *BasicLitNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.value.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *BasicLitNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"basic_lit\""
	ret["value"] = DumpNode(n.Value(), hook)
	return ret
}

func NewCompositeLitNode(filePath string, fileContent []rune, type_ Node, elts Node, start, end Position) Node {
	if type_ == nil {
		type_ = DummyNode
	}
	if elts == nil {
		elts = DummyNode
	}
	_1 := &CompositeLitNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeCompositeLit, start, end),
		type_:    type_,
		elts:     elts,
	}
	creationHook(_1)
	return _1
}

type CompositeLitNode struct {
	*BaseNode
	type_ Node
	elts  Node
}

func (n *CompositeLitNode) Type() Node {
	return n.type_
}

func (n *CompositeLitNode) SetType(v Node) {
	n.type_ = v
}

func (n *CompositeLitNode) Elts() Node {
	return n.elts
}

func (n *CompositeLitNode) SetElts(v Node) {
	n.elts = v
}

func (n *CompositeLitNode) BuildLink() {
	if !n.Type().IsDummy() {
		type_ := n.Type()
		type_.BuildLink()
		type_.SetParent(n)
		type_.SetSelfField("type_")
		type_.SetReplaceSelf(func(n Node) {
			n.Parent().(*CompositeLitNode).SetType(n)
		})
	}
	if !n.Elts().IsDummy() {
		elts := n.Elts()
		elts.BuildLink()
		elts.SetParent(n)
		elts.SetSelfField("elts")
		elts.SetReplaceSelf(func(n Node) {
			n.Parent().(*CompositeLitNode).SetElts(n)
		})
	}
}

func (n *CompositeLitNode) Fields() []string {
	return []string{
		"type_",
		"elts",
	}
}

func (n *CompositeLitNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "type_" {
		return n.Type()
	}
	if field == "elts" {
		return n.Elts()
	}
	return nil
}

func (n *CompositeLitNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetType(nodes[0])
	n.SetElts(nodes[1])
}

func (n *CompositeLitNode) Fork() Node {
	_ret := &CompositeLitNode{
		BaseNode: n.BaseNode.fork(),
		type_:    n.type_.Fork(),
		elts:     n.elts.Fork(),
	}
	_ret.type_.SetParent(_ret)
	_ret.elts.SetParent(_ret)
	return _ret
}

func (n *CompositeLitNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.type_.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.elts.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *CompositeLitNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"composite_lit\""
	ret["type"] = DumpNode(n.Type(), hook)
	ret["elts"] = DumpNode(n.Elts(), hook)
	return ret
}

func NewFunctionLitNode(filePath string, fileContent []rune, type_ Node, body Node, start, end Position) Node {
	if type_ == nil {
		type_ = DummyNode
	}
	if body == nil {
		body = DummyNode
	}
	_1 := &FunctionLitNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeFunctionLit, start, end),
		type_:    type_,
		body:     body,
	}
	creationHook(_1)
	return _1
}

type FunctionLitNode struct {
	*BaseNode
	type_ Node
	body  Node
}

func (n *FunctionLitNode) Type() Node {
	return n.type_
}

func (n *FunctionLitNode) SetType(v Node) {
	n.type_ = v
}

func (n *FunctionLitNode) Body() Node {
	return n.body
}

func (n *FunctionLitNode) SetBody(v Node) {
	n.body = v
}

func (n *FunctionLitNode) BuildLink() {
	if !n.Type().IsDummy() {
		type_ := n.Type()
		type_.BuildLink()
		type_.SetParent(n)
		type_.SetSelfField("type_")
		type_.SetReplaceSelf(func(n Node) {
			n.Parent().(*FunctionLitNode).SetType(n)
		})
	}
	if !n.Body().IsDummy() {
		body := n.Body()
		body.BuildLink()
		body.SetParent(n)
		body.SetSelfField("body")
		body.SetReplaceSelf(func(n Node) {
			n.Parent().(*FunctionLitNode).SetBody(n)
		})
	}
}

func (n *FunctionLitNode) Fields() []string {
	return []string{
		"type_",
		"body",
	}
}

func (n *FunctionLitNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "type_" {
		return n.Type()
	}
	if field == "body" {
		return n.Body()
	}
	return nil
}

func (n *FunctionLitNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetType(nodes[0])
	n.SetBody(nodes[1])
}

func (n *FunctionLitNode) Fork() Node {
	_ret := &FunctionLitNode{
		BaseNode: n.BaseNode.fork(),
		type_:    n.type_.Fork(),
		body:     n.body.Fork(),
	}
	_ret.type_.SetParent(_ret)
	_ret.body.SetParent(_ret)
	return _ret
}

func (n *FunctionLitNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.type_.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.body.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *FunctionLitNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"function_lit\""
	ret["type"] = DumpNode(n.Type(), hook)
	ret["body"] = DumpNode(n.Body(), hook)
	return ret
}

func NewCaseClauseNode(filePath string, fileContent []rune, list Node, body Node, start, end Position) Node {
	if list == nil {
		list = DummyNode
	}
	if body == nil {
		body = DummyNode
	}
	_1 := &CaseClauseNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeCaseClause, start, end),
		list:     list,
		body:     body,
	}
	creationHook(_1)
	return _1
}

type CaseClauseNode struct {
	*BaseNode
	list Node
	body Node
}

func (n *CaseClauseNode) List() Node {
	return n.list
}

func (n *CaseClauseNode) SetList(v Node) {
	n.list = v
}

func (n *CaseClauseNode) Body() Node {
	return n.body
}

func (n *CaseClauseNode) SetBody(v Node) {
	n.body = v
}

func (n *CaseClauseNode) BuildLink() {
	if !n.List().IsDummy() {
		list := n.List()
		list.BuildLink()
		list.SetParent(n)
		list.SetSelfField("list")
		list.SetReplaceSelf(func(n Node) {
			n.Parent().(*CaseClauseNode).SetList(n)
		})
	}
	if !n.Body().IsDummy() {
		body := n.Body()
		body.BuildLink()
		body.SetParent(n)
		body.SetSelfField("body")
		body.SetReplaceSelf(func(n Node) {
			n.Parent().(*CaseClauseNode).SetBody(n)
		})
	}
}

func (n *CaseClauseNode) Fields() []string {
	return []string{
		"list",
		"body",
	}
}

func (n *CaseClauseNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "list" {
		return n.List()
	}
	if field == "body" {
		return n.Body()
	}
	return nil
}

func (n *CaseClauseNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetList(nodes[0])
	n.SetBody(nodes[1])
}

func (n *CaseClauseNode) Fork() Node {
	_ret := &CaseClauseNode{
		BaseNode: n.BaseNode.fork(),
		list:     n.list.Fork(),
		body:     n.body.Fork(),
	}
	_ret.list.SetParent(_ret)
	_ret.body.SetParent(_ret)
	return _ret
}

func (n *CaseClauseNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.list.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.body.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *CaseClauseNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"case_clause\""
	ret["list"] = DumpNode(n.List(), hook)
	ret["body"] = DumpNode(n.Body(), hook)
	return ret
}

func NewCommonClauseNode(filePath string, fileContent []rune, common Node, body Node, start, end Position) Node {
	if common == nil {
		common = DummyNode
	}
	if body == nil {
		body = DummyNode
	}
	_1 := &CommonClauseNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeCommonClause, start, end),
		common:   common,
		body:     body,
	}
	creationHook(_1)
	return _1
}

type CommonClauseNode struct {
	*BaseNode
	common Node
	body   Node
}

func (n *CommonClauseNode) Common() Node {
	return n.common
}

func (n *CommonClauseNode) SetCommon(v Node) {
	n.common = v
}

func (n *CommonClauseNode) Body() Node {
	return n.body
}

func (n *CommonClauseNode) SetBody(v Node) {
	n.body = v
}

func (n *CommonClauseNode) BuildLink() {
	if !n.Common().IsDummy() {
		common := n.Common()
		common.BuildLink()
		common.SetParent(n)
		common.SetSelfField("common")
		common.SetReplaceSelf(func(n Node) {
			n.Parent().(*CommonClauseNode).SetCommon(n)
		})
	}
	if !n.Body().IsDummy() {
		body := n.Body()
		body.BuildLink()
		body.SetParent(n)
		body.SetSelfField("body")
		body.SetReplaceSelf(func(n Node) {
			n.Parent().(*CommonClauseNode).SetBody(n)
		})
	}
}

func (n *CommonClauseNode) Fields() []string {
	return []string{
		"common",
		"body",
	}
}

func (n *CommonClauseNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "common" {
		return n.Common()
	}
	if field == "body" {
		return n.Body()
	}
	return nil
}

func (n *CommonClauseNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetCommon(nodes[0])
	n.SetBody(nodes[1])
}

func (n *CommonClauseNode) Fork() Node {
	_ret := &CommonClauseNode{
		BaseNode: n.BaseNode.fork(),
		common:   n.common.Fork(),
		body:     n.body.Fork(),
	}
	_ret.common.SetParent(_ret)
	_ret.body.SetParent(_ret)
	return _ret
}

func (n *CommonClauseNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.common.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.body.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *CommonClauseNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"common_clause\""
	ret["common"] = DumpNode(n.Common(), hook)
	ret["body"] = DumpNode(n.Body(), hook)
	return ret
}

func NewFieldListNode(filePath string, fileContent []rune, list Node, start, end Position) Node {
	if list == nil {
		list = DummyNode
	}
	_1 := &FieldListNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeFieldList, start, end),
		list:     list,
	}
	creationHook(_1)
	return _1
}

type FieldListNode struct {
	*BaseNode
	list Node
}

func (n *FieldListNode) List() Node {
	return n.list
}

func (n *FieldListNode) SetList(v Node) {
	n.list = v
}

func (n *FieldListNode) BuildLink() {
	if !n.List().IsDummy() {
		list := n.List()
		list.BuildLink()
		list.SetParent(n)
		list.SetSelfField("list")
		list.SetReplaceSelf(func(n Node) {
			n.Parent().(*FieldListNode).SetList(n)
		})
	}
}

func (n *FieldListNode) Fields() []string {
	return []string{
		"list",
	}
}

func (n *FieldListNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "list" {
		return n.List()
	}
	return nil
}

func (n *FieldListNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetList(nodes[0])
}

func (n *FieldListNode) Fork() Node {
	_ret := &FieldListNode{
		BaseNode: n.BaseNode.fork(),
		list:     n.list.Fork(),
	}
	_ret.list.SetParent(_ret)
	return _ret
}

func (n *FieldListNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.list.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *FieldListNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"field_list\""
	ret["list"] = DumpNode(n.List(), hook)
	return ret
}

func NewFieldNode(filePath string, fileContent []rune, names Node, type_ Node, tag Node, start, end Position) Node {
	if names == nil {
		names = DummyNode
	}
	if type_ == nil {
		type_ = DummyNode
	}
	if tag == nil {
		tag = DummyNode
	}
	_1 := &FieldNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeField, start, end),
		names:    names,
		type_:    type_,
		tag:      tag,
	}
	creationHook(_1)
	return _1
}

type FieldNode struct {
	*BaseNode
	names Node
	type_ Node
	tag   Node
}

func (n *FieldNode) Names() Node {
	return n.names
}

func (n *FieldNode) SetNames(v Node) {
	n.names = v
}

func (n *FieldNode) Type() Node {
	return n.type_
}

func (n *FieldNode) SetType(v Node) {
	n.type_ = v
}

func (n *FieldNode) Tag() Node {
	return n.tag
}

func (n *FieldNode) SetTag(v Node) {
	n.tag = v
}

func (n *FieldNode) BuildLink() {
	if !n.Names().IsDummy() {
		names := n.Names()
		names.BuildLink()
		names.SetParent(n)
		names.SetSelfField("names")
		names.SetReplaceSelf(func(n Node) {
			n.Parent().(*FieldNode).SetNames(n)
		})
	}
	if !n.Type().IsDummy() {
		type_ := n.Type()
		type_.BuildLink()
		type_.SetParent(n)
		type_.SetSelfField("type_")
		type_.SetReplaceSelf(func(n Node) {
			n.Parent().(*FieldNode).SetType(n)
		})
	}
	if !n.Tag().IsDummy() {
		tag := n.Tag()
		tag.BuildLink()
		tag.SetParent(n)
		tag.SetSelfField("tag")
		tag.SetReplaceSelf(func(n Node) {
			n.Parent().(*FieldNode).SetTag(n)
		})
	}
}

func (n *FieldNode) Fields() []string {
	return []string{
		"names",
		"type_",
		"tag",
	}
}

func (n *FieldNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "names" {
		return n.Names()
	}
	if field == "type_" {
		return n.Type()
	}
	if field == "tag" {
		return n.Tag()
	}
	return nil
}

func (n *FieldNode) SetChild(nodes []Node) {
	if len(nodes) != 3 {
		return
	}
	n.SetNames(nodes[0])
	n.SetType(nodes[1])
	n.SetTag(nodes[2])
}

func (n *FieldNode) Fork() Node {
	_ret := &FieldNode{
		BaseNode: n.BaseNode.fork(),
		names:    n.names.Fork(),
		type_:    n.type_.Fork(),
		tag:      n.tag.Fork(),
	}
	_ret.names.SetParent(_ret)
	_ret.type_.SetParent(_ret)
	_ret.tag.SetParent(_ret)
	return _ret
}

func (n *FieldNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.names.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.type_.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.tag.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *FieldNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"field\""
	ret["names"] = DumpNode(n.Names(), hook)
	ret["type"] = DumpNode(n.Type(), hook)
	ret["tag"] = DumpNode(n.Tag(), hook)
	return ret
}

func NewImportDeclNode(filePath string, fileContent []rune, x Node, start, end Position) Node {
	if x == nil {
		x = DummyNode
	}
	_1 := &ImportDeclNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeImportDecl, start, end),
		x:        x,
	}
	creationHook(_1)
	return _1
}

type ImportDeclNode struct {
	*BaseNode
	x Node
}

func (n *ImportDeclNode) X() Node {
	return n.x
}

func (n *ImportDeclNode) SetX(v Node) {
	n.x = v
}

func (n *ImportDeclNode) BuildLink() {
	if !n.X().IsDummy() {
		x := n.X()
		x.BuildLink()
		x.SetParent(n)
		x.SetSelfField("x")
		x.SetReplaceSelf(func(n Node) {
			n.Parent().(*ImportDeclNode).SetX(n)
		})
	}
}

func (n *ImportDeclNode) Fields() []string {
	return []string{
		"x",
	}
}

func (n *ImportDeclNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "x" {
		return n.X()
	}
	return nil
}

func (n *ImportDeclNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetX(nodes[0])
}

func (n *ImportDeclNode) Fork() Node {
	_ret := &ImportDeclNode{
		BaseNode: n.BaseNode.fork(),
		x:        n.x.Fork(),
	}
	_ret.x.SetParent(_ret)
	return _ret
}

func (n *ImportDeclNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.x.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *ImportDeclNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"import_decl\""
	ret["x"] = DumpNode(n.X(), hook)
	return ret
}

func NewImportSpecNode(filePath string, fileContent []rune, name Node, source Node, start, end Position) Node {
	if name == nil {
		name = DummyNode
	}
	if source == nil {
		source = DummyNode
	}
	_1 := &ImportSpecNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeImportSpec, start, end),
		name:     name,
		source:   source,
	}
	creationHook(_1)
	return _1
}

type ImportSpecNode struct {
	*BaseNode
	name   Node
	source Node
}

func (n *ImportSpecNode) Name() Node {
	return n.name
}

func (n *ImportSpecNode) SetName(v Node) {
	n.name = v
}

func (n *ImportSpecNode) Source() Node {
	return n.source
}

func (n *ImportSpecNode) SetSource(v Node) {
	n.source = v
}

func (n *ImportSpecNode) BuildLink() {
	if !n.Name().IsDummy() {
		name := n.Name()
		name.BuildLink()
		name.SetParent(n)
		name.SetSelfField("name")
		name.SetReplaceSelf(func(n Node) {
			n.Parent().(*ImportSpecNode).SetName(n)
		})
	}
	if !n.Source().IsDummy() {
		source := n.Source()
		source.BuildLink()
		source.SetParent(n)
		source.SetSelfField("source")
		source.SetReplaceSelf(func(n Node) {
			n.Parent().(*ImportSpecNode).SetSource(n)
		})
	}
}

func (n *ImportSpecNode) Fields() []string {
	return []string{
		"name",
		"source",
	}
}

func (n *ImportSpecNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "name" {
		return n.Name()
	}
	if field == "source" {
		return n.Source()
	}
	return nil
}

func (n *ImportSpecNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetName(nodes[0])
	n.SetSource(nodes[1])
}

func (n *ImportSpecNode) Fork() Node {
	_ret := &ImportSpecNode{
		BaseNode: n.BaseNode.fork(),
		name:     n.name.Fork(),
		source:   n.source.Fork(),
	}
	_ret.name.SetParent(_ret)
	_ret.source.SetParent(_ret)
	return _ret
}

func (n *ImportSpecNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.name.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.source.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *ImportSpecNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"import_spec\""
	ret["name"] = DumpNode(n.Name(), hook)
	ret["source"] = DumpNode(n.Source(), hook)
	return ret
}

func NewConstSpecNode(filePath string, fileContent []rune, names Node, type_ Node, values Node, start, end Position) Node {
	if names == nil {
		names = DummyNode
	}
	if type_ == nil {
		type_ = DummyNode
	}
	if values == nil {
		values = DummyNode
	}
	_1 := &ConstSpecNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeConstSpec, start, end),
		names:    names,
		type_:    type_,
		values:   values,
	}
	creationHook(_1)
	return _1
}

type ConstSpecNode struct {
	*BaseNode
	names  Node
	type_  Node
	values Node
}

func (n *ConstSpecNode) Names() Node {
	return n.names
}

func (n *ConstSpecNode) SetNames(v Node) {
	n.names = v
}

func (n *ConstSpecNode) Type() Node {
	return n.type_
}

func (n *ConstSpecNode) SetType(v Node) {
	n.type_ = v
}

func (n *ConstSpecNode) Values() Node {
	return n.values
}

func (n *ConstSpecNode) SetValues(v Node) {
	n.values = v
}

func (n *ConstSpecNode) BuildLink() {
	if !n.Names().IsDummy() {
		names := n.Names()
		names.BuildLink()
		names.SetParent(n)
		names.SetSelfField("names")
		names.SetReplaceSelf(func(n Node) {
			n.Parent().(*ConstSpecNode).SetNames(n)
		})
	}
	if !n.Type().IsDummy() {
		type_ := n.Type()
		type_.BuildLink()
		type_.SetParent(n)
		type_.SetSelfField("type_")
		type_.SetReplaceSelf(func(n Node) {
			n.Parent().(*ConstSpecNode).SetType(n)
		})
	}
	if !n.Values().IsDummy() {
		values := n.Values()
		values.BuildLink()
		values.SetParent(n)
		values.SetSelfField("values")
		values.SetReplaceSelf(func(n Node) {
			n.Parent().(*ConstSpecNode).SetValues(n)
		})
	}
}

func (n *ConstSpecNode) Fields() []string {
	return []string{
		"names",
		"type_",
		"values",
	}
}

func (n *ConstSpecNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "names" {
		return n.Names()
	}
	if field == "type_" {
		return n.Type()
	}
	if field == "values" {
		return n.Values()
	}
	return nil
}

func (n *ConstSpecNode) SetChild(nodes []Node) {
	if len(nodes) != 3 {
		return
	}
	n.SetNames(nodes[0])
	n.SetType(nodes[1])
	n.SetValues(nodes[2])
}

func (n *ConstSpecNode) Fork() Node {
	_ret := &ConstSpecNode{
		BaseNode: n.BaseNode.fork(),
		names:    n.names.Fork(),
		type_:    n.type_.Fork(),
		values:   n.values.Fork(),
	}
	_ret.names.SetParent(_ret)
	_ret.type_.SetParent(_ret)
	_ret.values.SetParent(_ret)
	return _ret
}

func (n *ConstSpecNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.names.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.type_.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.values.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *ConstSpecNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"const_spec\""
	ret["names"] = DumpNode(n.Names(), hook)
	ret["type"] = DumpNode(n.Type(), hook)
	ret["values"] = DumpNode(n.Values(), hook)
	return ret
}

func NewVarSpecNode(filePath string, fileContent []rune, names Node, type_ Node, values Node, start, end Position) Node {
	if names == nil {
		names = DummyNode
	}
	if type_ == nil {
		type_ = DummyNode
	}
	if values == nil {
		values = DummyNode
	}
	_1 := &VarSpecNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeVarSpec, start, end),
		names:    names,
		type_:    type_,
		values:   values,
	}
	creationHook(_1)
	return _1
}

type VarSpecNode struct {
	*BaseNode
	names  Node
	type_  Node
	values Node
}

func (n *VarSpecNode) Names() Node {
	return n.names
}

func (n *VarSpecNode) SetNames(v Node) {
	n.names = v
}

func (n *VarSpecNode) Type() Node {
	return n.type_
}

func (n *VarSpecNode) SetType(v Node) {
	n.type_ = v
}

func (n *VarSpecNode) Values() Node {
	return n.values
}

func (n *VarSpecNode) SetValues(v Node) {
	n.values = v
}

func (n *VarSpecNode) BuildLink() {
	if !n.Names().IsDummy() {
		names := n.Names()
		names.BuildLink()
		names.SetParent(n)
		names.SetSelfField("names")
		names.SetReplaceSelf(func(n Node) {
			n.Parent().(*VarSpecNode).SetNames(n)
		})
	}
	if !n.Type().IsDummy() {
		type_ := n.Type()
		type_.BuildLink()
		type_.SetParent(n)
		type_.SetSelfField("type_")
		type_.SetReplaceSelf(func(n Node) {
			n.Parent().(*VarSpecNode).SetType(n)
		})
	}
	if !n.Values().IsDummy() {
		values := n.Values()
		values.BuildLink()
		values.SetParent(n)
		values.SetSelfField("values")
		values.SetReplaceSelf(func(n Node) {
			n.Parent().(*VarSpecNode).SetValues(n)
		})
	}
}

func (n *VarSpecNode) Fields() []string {
	return []string{
		"names",
		"type_",
		"values",
	}
}

func (n *VarSpecNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "names" {
		return n.Names()
	}
	if field == "type_" {
		return n.Type()
	}
	if field == "values" {
		return n.Values()
	}
	return nil
}

func (n *VarSpecNode) SetChild(nodes []Node) {
	if len(nodes) != 3 {
		return
	}
	n.SetNames(nodes[0])
	n.SetType(nodes[1])
	n.SetValues(nodes[2])
}

func (n *VarSpecNode) Fork() Node {
	_ret := &VarSpecNode{
		BaseNode: n.BaseNode.fork(),
		names:    n.names.Fork(),
		type_:    n.type_.Fork(),
		values:   n.values.Fork(),
	}
	_ret.names.SetParent(_ret)
	_ret.type_.SetParent(_ret)
	_ret.values.SetParent(_ret)
	return _ret
}

func (n *VarSpecNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.names.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.type_.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.values.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *VarSpecNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"var_spec\""
	ret["names"] = DumpNode(n.Names(), hook)
	ret["type"] = DumpNode(n.Type(), hook)
	ret["values"] = DumpNode(n.Values(), hook)
	return ret
}

func NewTypeEqSpecNode(filePath string, fileContent []rune, name Node, typeParams Node, type_ Node, start, end Position) Node {
	if name == nil {
		name = DummyNode
	}
	if typeParams == nil {
		typeParams = DummyNode
	}
	if type_ == nil {
		type_ = DummyNode
	}
	_1 := &TypeEqSpecNode{
		BaseNode:   NewBaseNode(filePath, fileContent, NodeTypeTypeEqSpec, start, end),
		name:       name,
		typeParams: typeParams,
		type_:      type_,
	}
	creationHook(_1)
	return _1
}

type TypeEqSpecNode struct {
	*BaseNode
	name       Node
	typeParams Node
	type_      Node
}

func (n *TypeEqSpecNode) Name() Node {
	return n.name
}

func (n *TypeEqSpecNode) SetName(v Node) {
	n.name = v
}

func (n *TypeEqSpecNode) TypeParams() Node {
	return n.typeParams
}

func (n *TypeEqSpecNode) SetTypeParams(v Node) {
	n.typeParams = v
}

func (n *TypeEqSpecNode) Type() Node {
	return n.type_
}

func (n *TypeEqSpecNode) SetType(v Node) {
	n.type_ = v
}

func (n *TypeEqSpecNode) BuildLink() {
	if !n.Name().IsDummy() {
		name := n.Name()
		name.BuildLink()
		name.SetParent(n)
		name.SetSelfField("name")
		name.SetReplaceSelf(func(n Node) {
			n.Parent().(*TypeEqSpecNode).SetName(n)
		})
	}
	if !n.TypeParams().IsDummy() {
		typeParams := n.TypeParams()
		typeParams.BuildLink()
		typeParams.SetParent(n)
		typeParams.SetSelfField("type_params")
		typeParams.SetReplaceSelf(func(n Node) {
			n.Parent().(*TypeEqSpecNode).SetTypeParams(n)
		})
	}
	if !n.Type().IsDummy() {
		type_ := n.Type()
		type_.BuildLink()
		type_.SetParent(n)
		type_.SetSelfField("type_")
		type_.SetReplaceSelf(func(n Node) {
			n.Parent().(*TypeEqSpecNode).SetType(n)
		})
	}
}

func (n *TypeEqSpecNode) Fields() []string {
	return []string{
		"name",
		"type_params",
		"type_",
	}
}

func (n *TypeEqSpecNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "name" {
		return n.Name()
	}
	if field == "type_params" {
		return n.TypeParams()
	}
	if field == "type_" {
		return n.Type()
	}
	return nil
}

func (n *TypeEqSpecNode) SetChild(nodes []Node) {
	if len(nodes) != 3 {
		return
	}
	n.SetName(nodes[0])
	n.SetTypeParams(nodes[1])
	n.SetType(nodes[2])
}

func (n *TypeEqSpecNode) Fork() Node {
	_ret := &TypeEqSpecNode{
		BaseNode:   n.BaseNode.fork(),
		name:       n.name.Fork(),
		typeParams: n.typeParams.Fork(),
		type_:      n.type_.Fork(),
	}
	_ret.name.SetParent(_ret)
	_ret.typeParams.SetParent(_ret)
	_ret.type_.SetParent(_ret)
	return _ret
}

func (n *TypeEqSpecNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.name.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.typeParams.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.type_.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *TypeEqSpecNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"type_eq_spec\""
	ret["name"] = DumpNode(n.Name(), hook)
	ret["type_params"] = DumpNode(n.TypeParams(), hook)
	ret["type"] = DumpNode(n.Type(), hook)
	return ret
}

func NewTypeSpecNode(filePath string, fileContent []rune, name Node, typeParams Node, type_ Node, start, end Position) Node {
	if name == nil {
		name = DummyNode
	}
	if typeParams == nil {
		typeParams = DummyNode
	}
	if type_ == nil {
		type_ = DummyNode
	}
	_1 := &TypeSpecNode{
		BaseNode:   NewBaseNode(filePath, fileContent, NodeTypeTypeSpec, start, end),
		name:       name,
		typeParams: typeParams,
		type_:      type_,
	}
	creationHook(_1)
	return _1
}

type TypeSpecNode struct {
	*BaseNode
	name       Node
	typeParams Node
	type_      Node
}

func (n *TypeSpecNode) Name() Node {
	return n.name
}

func (n *TypeSpecNode) SetName(v Node) {
	n.name = v
}

func (n *TypeSpecNode) TypeParams() Node {
	return n.typeParams
}

func (n *TypeSpecNode) SetTypeParams(v Node) {
	n.typeParams = v
}

func (n *TypeSpecNode) Type() Node {
	return n.type_
}

func (n *TypeSpecNode) SetType(v Node) {
	n.type_ = v
}

func (n *TypeSpecNode) BuildLink() {
	if !n.Name().IsDummy() {
		name := n.Name()
		name.BuildLink()
		name.SetParent(n)
		name.SetSelfField("name")
		name.SetReplaceSelf(func(n Node) {
			n.Parent().(*TypeSpecNode).SetName(n)
		})
	}
	if !n.TypeParams().IsDummy() {
		typeParams := n.TypeParams()
		typeParams.BuildLink()
		typeParams.SetParent(n)
		typeParams.SetSelfField("type_params")
		typeParams.SetReplaceSelf(func(n Node) {
			n.Parent().(*TypeSpecNode).SetTypeParams(n)
		})
	}
	if !n.Type().IsDummy() {
		type_ := n.Type()
		type_.BuildLink()
		type_.SetParent(n)
		type_.SetSelfField("type_")
		type_.SetReplaceSelf(func(n Node) {
			n.Parent().(*TypeSpecNode).SetType(n)
		})
	}
}

func (n *TypeSpecNode) Fields() []string {
	return []string{
		"name",
		"type_params",
		"type_",
	}
}

func (n *TypeSpecNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "name" {
		return n.Name()
	}
	if field == "type_params" {
		return n.TypeParams()
	}
	if field == "type_" {
		return n.Type()
	}
	return nil
}

func (n *TypeSpecNode) SetChild(nodes []Node) {
	if len(nodes) != 3 {
		return
	}
	n.SetName(nodes[0])
	n.SetTypeParams(nodes[1])
	n.SetType(nodes[2])
}

func (n *TypeSpecNode) Fork() Node {
	_ret := &TypeSpecNode{
		BaseNode:   n.BaseNode.fork(),
		name:       n.name.Fork(),
		typeParams: n.typeParams.Fork(),
		type_:      n.type_.Fork(),
	}
	_ret.name.SetParent(_ret)
	_ret.typeParams.SetParent(_ret)
	_ret.type_.SetParent(_ret)
	return _ret
}

func (n *TypeSpecNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.name.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.typeParams.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.type_.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *TypeSpecNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"type_spec\""
	ret["name"] = DumpNode(n.Name(), hook)
	ret["type_params"] = DumpNode(n.TypeParams(), hook)
	ret["type"] = DumpNode(n.Type(), hook)
	return ret
}

func NewConstDeclNode(filePath string, fileContent []rune, specs Node, start, end Position) Node {
	if specs == nil {
		specs = DummyNode
	}
	_1 := &ConstDeclNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeConstDecl, start, end),
		specs:    specs,
	}
	creationHook(_1)
	return _1
}

type ConstDeclNode struct {
	*BaseNode
	specs Node
}

func (n *ConstDeclNode) Specs() Node {
	return n.specs
}

func (n *ConstDeclNode) SetSpecs(v Node) {
	n.specs = v
}

func (n *ConstDeclNode) BuildLink() {
	if !n.Specs().IsDummy() {
		specs := n.Specs()
		specs.BuildLink()
		specs.SetParent(n)
		specs.SetSelfField("specs")
		specs.SetReplaceSelf(func(n Node) {
			n.Parent().(*ConstDeclNode).SetSpecs(n)
		})
	}
}

func (n *ConstDeclNode) Fields() []string {
	return []string{
		"specs",
	}
}

func (n *ConstDeclNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "specs" {
		return n.Specs()
	}
	return nil
}

func (n *ConstDeclNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetSpecs(nodes[0])
}

func (n *ConstDeclNode) Fork() Node {
	_ret := &ConstDeclNode{
		BaseNode: n.BaseNode.fork(),
		specs:    n.specs.Fork(),
	}
	_ret.specs.SetParent(_ret)
	return _ret
}

func (n *ConstDeclNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.specs.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *ConstDeclNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"const_decl\""
	ret["specs"] = DumpNode(n.Specs(), hook)
	return ret
}

func NewVarDeclNode(filePath string, fileContent []rune, specs Node, start, end Position) Node {
	if specs == nil {
		specs = DummyNode
	}
	_1 := &VarDeclNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeVarDecl, start, end),
		specs:    specs,
	}
	creationHook(_1)
	return _1
}

type VarDeclNode struct {
	*BaseNode
	specs Node
}

func (n *VarDeclNode) Specs() Node {
	return n.specs
}

func (n *VarDeclNode) SetSpecs(v Node) {
	n.specs = v
}

func (n *VarDeclNode) BuildLink() {
	if !n.Specs().IsDummy() {
		specs := n.Specs()
		specs.BuildLink()
		specs.SetParent(n)
		specs.SetSelfField("specs")
		specs.SetReplaceSelf(func(n Node) {
			n.Parent().(*VarDeclNode).SetSpecs(n)
		})
	}
}

func (n *VarDeclNode) Fields() []string {
	return []string{
		"specs",
	}
}

func (n *VarDeclNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "specs" {
		return n.Specs()
	}
	return nil
}

func (n *VarDeclNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetSpecs(nodes[0])
}

func (n *VarDeclNode) Fork() Node {
	_ret := &VarDeclNode{
		BaseNode: n.BaseNode.fork(),
		specs:    n.specs.Fork(),
	}
	_ret.specs.SetParent(_ret)
	return _ret
}

func (n *VarDeclNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.specs.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *VarDeclNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"var_decl\""
	ret["specs"] = DumpNode(n.Specs(), hook)
	return ret
}

func NewTypeDeclNode(filePath string, fileContent []rune, specs Node, start, end Position) Node {
	if specs == nil {
		specs = DummyNode
	}
	_1 := &TypeDeclNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeTypeDecl, start, end),
		specs:    specs,
	}
	creationHook(_1)
	return _1
}

type TypeDeclNode struct {
	*BaseNode
	specs Node
}

func (n *TypeDeclNode) Specs() Node {
	return n.specs
}

func (n *TypeDeclNode) SetSpecs(v Node) {
	n.specs = v
}

func (n *TypeDeclNode) BuildLink() {
	if !n.Specs().IsDummy() {
		specs := n.Specs()
		specs.BuildLink()
		specs.SetParent(n)
		specs.SetSelfField("specs")
		specs.SetReplaceSelf(func(n Node) {
			n.Parent().(*TypeDeclNode).SetSpecs(n)
		})
	}
}

func (n *TypeDeclNode) Fields() []string {
	return []string{
		"specs",
	}
}

func (n *TypeDeclNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "specs" {
		return n.Specs()
	}
	return nil
}

func (n *TypeDeclNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetSpecs(nodes[0])
}

func (n *TypeDeclNode) Fork() Node {
	_ret := &TypeDeclNode{
		BaseNode: n.BaseNode.fork(),
		specs:    n.specs.Fork(),
	}
	_ret.specs.SetParent(_ret)
	return _ret
}

func (n *TypeDeclNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.specs.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *TypeDeclNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"type_decl\""
	ret["specs"] = DumpNode(n.Specs(), hook)
	return ret
}

func NewEllipsisNode(filePath string, fileContent []rune, elt Node, start, end Position) Node {
	if elt == nil {
		elt = DummyNode
	}
	_1 := &EllipsisNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeEllipsis, start, end),
		elt:      elt,
	}
	creationHook(_1)
	return _1
}

type EllipsisNode struct {
	*BaseNode
	elt Node
}

func (n *EllipsisNode) Elt() Node {
	return n.elt
}

func (n *EllipsisNode) SetElt(v Node) {
	n.elt = v
}

func (n *EllipsisNode) BuildLink() {
	if !n.Elt().IsDummy() {
		elt := n.Elt()
		elt.BuildLink()
		elt.SetParent(n)
		elt.SetSelfField("elt")
		elt.SetReplaceSelf(func(n Node) {
			n.Parent().(*EllipsisNode).SetElt(n)
		})
	}
}

func (n *EllipsisNode) Fields() []string {
	return []string{
		"elt",
	}
}

func (n *EllipsisNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "elt" {
		return n.Elt()
	}
	return nil
}

func (n *EllipsisNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetElt(nodes[0])
}

func (n *EllipsisNode) Fork() Node {
	_ret := &EllipsisNode{
		BaseNode: n.BaseNode.fork(),
		elt:      n.elt.Fork(),
	}
	_ret.elt.SetParent(_ret)
	return _ret
}

func (n *EllipsisNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.elt.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *EllipsisNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"ellipsis\""
	ret["elt"] = DumpNode(n.Elt(), hook)
	return ret
}

func NewLabeledStmtNode(filePath string, fileContent []rune, label Node, stmt Node, start, end Position) Node {
	if label == nil {
		label = DummyNode
	}
	if stmt == nil {
		stmt = DummyNode
	}
	_1 := &LabeledStmtNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeLabeledStmt, start, end),
		label:    label,
		stmt:     stmt,
	}
	creationHook(_1)
	return _1
}

type LabeledStmtNode struct {
	*BaseNode
	label Node
	stmt  Node
}

func (n *LabeledStmtNode) Label() Node {
	return n.label
}

func (n *LabeledStmtNode) SetLabel(v Node) {
	n.label = v
}

func (n *LabeledStmtNode) Stmt() Node {
	return n.stmt
}

func (n *LabeledStmtNode) SetStmt(v Node) {
	n.stmt = v
}

func (n *LabeledStmtNode) BuildLink() {
	if !n.Label().IsDummy() {
		label := n.Label()
		label.BuildLink()
		label.SetParent(n)
		label.SetSelfField("label")
		label.SetReplaceSelf(func(n Node) {
			n.Parent().(*LabeledStmtNode).SetLabel(n)
		})
	}
	if !n.Stmt().IsDummy() {
		stmt := n.Stmt()
		stmt.BuildLink()
		stmt.SetParent(n)
		stmt.SetSelfField("stmt")
		stmt.SetReplaceSelf(func(n Node) {
			n.Parent().(*LabeledStmtNode).SetStmt(n)
		})
	}
}

func (n *LabeledStmtNode) Fields() []string {
	return []string{
		"label",
		"stmt",
	}
}

func (n *LabeledStmtNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "label" {
		return n.Label()
	}
	if field == "stmt" {
		return n.Stmt()
	}
	return nil
}

func (n *LabeledStmtNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetLabel(nodes[0])
	n.SetStmt(nodes[1])
}

func (n *LabeledStmtNode) Fork() Node {
	_ret := &LabeledStmtNode{
		BaseNode: n.BaseNode.fork(),
		label:    n.label.Fork(),
		stmt:     n.stmt.Fork(),
	}
	_ret.label.SetParent(_ret)
	_ret.stmt.SetParent(_ret)
	return _ret
}

func (n *LabeledStmtNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.label.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.stmt.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *LabeledStmtNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"labeled_stmt\""
	ret["label"] = DumpNode(n.Label(), hook)
	ret["stmt"] = DumpNode(n.Stmt(), hook)
	return ret
}

func NewGenericTypeInstantiationNode(filePath string, fileContent []rune, base Node, args Node, start, end Position) Node {
	if base == nil {
		base = DummyNode
	}
	if args == nil {
		args = DummyNode
	}
	_1 := &GenericTypeInstantiationNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeGenericTypeInstantiation, start, end),
		base:     base,
		args:     args,
	}
	creationHook(_1)
	return _1
}

type GenericTypeInstantiationNode struct {
	*BaseNode
	base Node
	args Node
}

func (n *GenericTypeInstantiationNode) Base() Node {
	return n.base
}

func (n *GenericTypeInstantiationNode) SetBase(v Node) {
	n.base = v
}

func (n *GenericTypeInstantiationNode) Args() Node {
	return n.args
}

func (n *GenericTypeInstantiationNode) SetArgs(v Node) {
	n.args = v
}

func (n *GenericTypeInstantiationNode) BuildLink() {
	if !n.Base().IsDummy() {
		base := n.Base()
		base.BuildLink()
		base.SetParent(n)
		base.SetSelfField("base")
		base.SetReplaceSelf(func(n Node) {
			n.Parent().(*GenericTypeInstantiationNode).SetBase(n)
		})
	}
	if !n.Args().IsDummy() {
		args := n.Args()
		args.BuildLink()
		args.SetParent(n)
		args.SetSelfField("args")
		args.SetReplaceSelf(func(n Node) {
			n.Parent().(*GenericTypeInstantiationNode).SetArgs(n)
		})
	}
}

func (n *GenericTypeInstantiationNode) Fields() []string {
	return []string{
		"base",
		"args",
	}
}

func (n *GenericTypeInstantiationNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "base" {
		return n.Base()
	}
	if field == "args" {
		return n.Args()
	}
	return nil
}

func (n *GenericTypeInstantiationNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetBase(nodes[0])
	n.SetArgs(nodes[1])
}

func (n *GenericTypeInstantiationNode) Fork() Node {
	_ret := &GenericTypeInstantiationNode{
		BaseNode: n.BaseNode.fork(),
		base:     n.base.Fork(),
		args:     n.args.Fork(),
	}
	_ret.base.SetParent(_ret)
	_ret.args.SetParent(_ret)
	return _ret
}

func (n *GenericTypeInstantiationNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.base.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.args.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *GenericTypeInstantiationNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"generic_type_instantiation\""
	ret["base"] = DumpNode(n.Base(), hook)
	ret["args"] = DumpNode(n.Args(), hook)
	return ret
}

func NewIdentNode(filePath string, fileContent []rune, x Node, start, end Position) Node {
	if x == nil {
		x = DummyNode
	}
	_1 := &IdentNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeIdent, start, end),
		x:        x,
	}
	creationHook(_1)
	return _1
}

type IdentNode struct {
	*BaseNode
	x Node
}

func (n *IdentNode) X() Node {
	return n.x
}

func (n *IdentNode) SetX(v Node) {
	n.x = v
}

func (n *IdentNode) BuildLink() {
	if !n.X().IsDummy() {
		x := n.X()
		x.BuildLink()
		x.SetParent(n)
		x.SetSelfField("x")
		x.SetReplaceSelf(func(n Node) {
			n.Parent().(*IdentNode).SetX(n)
		})
	}
}

func (n *IdentNode) Fields() []string {
	return []string{
		"x",
	}
}

func (n *IdentNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "x" {
		return n.X()
	}
	return nil
}

func (n *IdentNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetX(nodes[0])
}

func (n *IdentNode) Fork() Node {
	_ret := &IdentNode{
		BaseNode: n.BaseNode.fork(),
		x:        n.x.Fork(),
	}
	_ret.x.SetParent(_ret)
	return _ret
}

func (n *IdentNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.x.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *IdentNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"ident\""
	ret["x"] = DumpNode(n.X(), hook)
	return ret
}

func NewMakeExprNode(filePath string, fileContent []rune, x Node, start, end Position) Node {
	if x == nil {
		x = DummyNode
	}
	_1 := &MakeExprNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeMakeExpr, start, end),
		x:        x,
	}
	creationHook(_1)
	return _1
}

type MakeExprNode struct {
	*BaseNode
	x Node
}

func (n *MakeExprNode) X() Node {
	return n.x
}

func (n *MakeExprNode) SetX(v Node) {
	n.x = v
}

func (n *MakeExprNode) BuildLink() {
	if !n.X().IsDummy() {
		x := n.X()
		x.BuildLink()
		x.SetParent(n)
		x.SetSelfField("x")
		x.SetReplaceSelf(func(n Node) {
			n.Parent().(*MakeExprNode).SetX(n)
		})
	}
}

func (n *MakeExprNode) Fields() []string {
	return []string{
		"x",
	}
}

func (n *MakeExprNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "x" {
		return n.X()
	}
	return nil
}

func (n *MakeExprNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetX(nodes[0])
}

func (n *MakeExprNode) Fork() Node {
	_ret := &MakeExprNode{
		BaseNode: n.BaseNode.fork(),
		x:        n.x.Fork(),
	}
	_ret.x.SetParent(_ret)
	return _ret
}

func (n *MakeExprNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.x.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *MakeExprNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"make_expr\""
	ret["x"] = DumpNode(n.X(), hook)
	return ret
}

func NewNewExprNode(filePath string, fileContent []rune, x Node, start, end Position) Node {
	if x == nil {
		x = DummyNode
	}
	_1 := &NewExprNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeNewExpr, start, end),
		x:        x,
	}
	creationHook(_1)
	return _1
}

type NewExprNode struct {
	*BaseNode
	x Node
}

func (n *NewExprNode) X() Node {
	return n.x
}

func (n *NewExprNode) SetX(v Node) {
	n.x = v
}

func (n *NewExprNode) BuildLink() {
	if !n.X().IsDummy() {
		x := n.X()
		x.BuildLink()
		x.SetParent(n)
		x.SetSelfField("x")
		x.SetReplaceSelf(func(n Node) {
			n.Parent().(*NewExprNode).SetX(n)
		})
	}
}

func (n *NewExprNode) Fields() []string {
	return []string{
		"x",
	}
}

func (n *NewExprNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "x" {
		return n.X()
	}
	return nil
}

func (n *NewExprNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetX(nodes[0])
}

func (n *NewExprNode) Fork() Node {
	_ret := &NewExprNode{
		BaseNode: n.BaseNode.fork(),
		x:        n.x.Fork(),
	}
	_ret.x.SetParent(_ret)
	return _ret
}

func (n *NewExprNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.x.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *NewExprNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"new_expr\""
	ret["x"] = DumpNode(n.X(), hook)
	return ret
}

func NewPackageIdentNode(filePath string, fileContent []rune, ident Node, start, end Position) Node {
	if ident == nil {
		ident = DummyNode
	}
	_1 := &PackageIdentNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypePackageIdent, start, end),
		ident:    ident,
	}
	creationHook(_1)
	return _1
}

type PackageIdentNode struct {
	*BaseNode
	ident Node
}

func (n *PackageIdentNode) Ident() Node {
	return n.ident
}

func (n *PackageIdentNode) SetIdent(v Node) {
	n.ident = v
}

func (n *PackageIdentNode) BuildLink() {
	if !n.Ident().IsDummy() {
		ident := n.Ident()
		ident.BuildLink()
		ident.SetParent(n)
		ident.SetSelfField("ident")
		ident.SetReplaceSelf(func(n Node) {
			n.Parent().(*PackageIdentNode).SetIdent(n)
		})
	}
}

func (n *PackageIdentNode) Fields() []string {
	return []string{
		"ident",
	}
}

func (n *PackageIdentNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "ident" {
		return n.Ident()
	}
	return nil
}

func (n *PackageIdentNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetIdent(nodes[0])
}

func (n *PackageIdentNode) Fork() Node {
	_ret := &PackageIdentNode{
		BaseNode: n.BaseNode.fork(),
		ident:    n.ident.Fork(),
	}
	_ret.ident.SetParent(_ret)
	return _ret
}

func (n *PackageIdentNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.ident.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *PackageIdentNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"package_ident\""
	ret["ident"] = DumpNode(n.Ident(), hook)
	return ret
}

func NewImportDotNode(filePath string, fileContent []rune, dot Node, start, end Position) Node {
	if dot == nil {
		dot = DummyNode
	}
	_1 := &ImportDotNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeImportDot, start, end),
		dot:      dot,
	}
	creationHook(_1)
	return _1
}

type ImportDotNode struct {
	*BaseNode
	dot Node
}

func (n *ImportDotNode) Dot() Node {
	return n.dot
}

func (n *ImportDotNode) SetDot(v Node) {
	n.dot = v
}

func (n *ImportDotNode) BuildLink() {
	if !n.Dot().IsDummy() {
		dot := n.Dot()
		dot.BuildLink()
		dot.SetParent(n)
		dot.SetSelfField("dot")
		dot.SetReplaceSelf(func(n Node) {
			n.Parent().(*ImportDotNode).SetDot(n)
		})
	}
}

func (n *ImportDotNode) Fields() []string {
	return []string{
		"dot",
	}
}

func (n *ImportDotNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "dot" {
		return n.Dot()
	}
	return nil
}

func (n *ImportDotNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetDot(nodes[0])
}

func (n *ImportDotNode) Fork() Node {
	_ret := &ImportDotNode{
		BaseNode: n.BaseNode.fork(),
		dot:      n.dot.Fork(),
	}
	_ret.dot.SetParent(_ret)
	return _ret
}

func (n *ImportDotNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.dot.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *ImportDotNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"import_dot\""
	ret["dot"] = DumpNode(n.Dot(), hook)
	return ret
}

func NewImportIdentNode(filePath string, fileContent []rune, ident Node, start, end Position) Node {
	if ident == nil {
		ident = DummyNode
	}
	_1 := &ImportIdentNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeImportIdent, start, end),
		ident:    ident,
	}
	creationHook(_1)
	return _1
}

type ImportIdentNode struct {
	*BaseNode
	ident Node
}

func (n *ImportIdentNode) Ident() Node {
	return n.ident
}

func (n *ImportIdentNode) SetIdent(v Node) {
	n.ident = v
}

func (n *ImportIdentNode) BuildLink() {
	if !n.Ident().IsDummy() {
		ident := n.Ident()
		ident.BuildLink()
		ident.SetParent(n)
		ident.SetSelfField("ident")
		ident.SetReplaceSelf(func(n Node) {
			n.Parent().(*ImportIdentNode).SetIdent(n)
		})
	}
}

func (n *ImportIdentNode) Fields() []string {
	return []string{
		"ident",
	}
}

func (n *ImportIdentNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "ident" {
		return n.Ident()
	}
	return nil
}

func (n *ImportIdentNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetIdent(nodes[0])
}

func (n *ImportIdentNode) Fork() Node {
	_ret := &ImportIdentNode{
		BaseNode: n.BaseNode.fork(),
		ident:    n.ident.Fork(),
	}
	_ret.ident.SetParent(_ret)
	return _ret
}

func (n *ImportIdentNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.ident.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *ImportIdentNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"import_ident\""
	ret["ident"] = DumpNode(n.Ident(), hook)
	return ret
}

func NewImportPathNode(filePath string, fileContent []rune, path Node, start, end Position) Node {
	if path == nil {
		path = DummyNode
	}
	_1 := &ImportPathNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeImportPath, start, end),
		path:     path,
	}
	creationHook(_1)
	return _1
}

type ImportPathNode struct {
	*BaseNode
	path Node
}

func (n *ImportPathNode) Path() Node {
	return n.path
}

func (n *ImportPathNode) SetPath(v Node) {
	n.path = v
}

func (n *ImportPathNode) BuildLink() {
	if !n.Path().IsDummy() {
		path := n.Path()
		path.BuildLink()
		path.SetParent(n)
		path.SetSelfField("path")
		path.SetReplaceSelf(func(n Node) {
			n.Parent().(*ImportPathNode).SetPath(n)
		})
	}
}

func (n *ImportPathNode) Fields() []string {
	return []string{
		"path",
	}
}

func (n *ImportPathNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "path" {
		return n.Path()
	}
	return nil
}

func (n *ImportPathNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetPath(nodes[0])
}

func (n *ImportPathNode) Fork() Node {
	_ret := &ImportPathNode{
		BaseNode: n.BaseNode.fork(),
		path:     n.path.Fork(),
	}
	_ret.path.SetParent(_ret)
	return _ret
}

func (n *ImportPathNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.path.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *ImportPathNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"import_path\""
	ret["path"] = DumpNode(n.Path(), hook)
	return ret
}

func NewConstIdentNode(filePath string, fileContent []rune, ident Node, start, end Position) Node {
	if ident == nil {
		ident = DummyNode
	}
	_1 := &ConstIdentNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeConstIdent, start, end),
		ident:    ident,
	}
	creationHook(_1)
	return _1
}

type ConstIdentNode struct {
	*BaseNode
	ident Node
}

func (n *ConstIdentNode) Ident() Node {
	return n.ident
}

func (n *ConstIdentNode) SetIdent(v Node) {
	n.ident = v
}

func (n *ConstIdentNode) BuildLink() {
	if !n.Ident().IsDummy() {
		ident := n.Ident()
		ident.BuildLink()
		ident.SetParent(n)
		ident.SetSelfField("ident")
		ident.SetReplaceSelf(func(n Node) {
			n.Parent().(*ConstIdentNode).SetIdent(n)
		})
	}
}

func (n *ConstIdentNode) Fields() []string {
	return []string{
		"ident",
	}
}

func (n *ConstIdentNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "ident" {
		return n.Ident()
	}
	return nil
}

func (n *ConstIdentNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetIdent(nodes[0])
}

func (n *ConstIdentNode) Fork() Node {
	_ret := &ConstIdentNode{
		BaseNode: n.BaseNode.fork(),
		ident:    n.ident.Fork(),
	}
	_ret.ident.SetParent(_ret)
	return _ret
}

func (n *ConstIdentNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.ident.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *ConstIdentNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"const_ident\""
	ret["ident"] = DumpNode(n.Ident(), hook)
	return ret
}

func NewVarIdentNode(filePath string, fileContent []rune, ident Node, start, end Position) Node {
	if ident == nil {
		ident = DummyNode
	}
	_1 := &VarIdentNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeVarIdent, start, end),
		ident:    ident,
	}
	creationHook(_1)
	return _1
}

type VarIdentNode struct {
	*BaseNode
	ident Node
}

func (n *VarIdentNode) Ident() Node {
	return n.ident
}

func (n *VarIdentNode) SetIdent(v Node) {
	n.ident = v
}

func (n *VarIdentNode) BuildLink() {
	if !n.Ident().IsDummy() {
		ident := n.Ident()
		ident.BuildLink()
		ident.SetParent(n)
		ident.SetSelfField("ident")
		ident.SetReplaceSelf(func(n Node) {
			n.Parent().(*VarIdentNode).SetIdent(n)
		})
	}
}

func (n *VarIdentNode) Fields() []string {
	return []string{
		"ident",
	}
}

func (n *VarIdentNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "ident" {
		return n.Ident()
	}
	return nil
}

func (n *VarIdentNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetIdent(nodes[0])
}

func (n *VarIdentNode) Fork() Node {
	_ret := &VarIdentNode{
		BaseNode: n.BaseNode.fork(),
		ident:    n.ident.Fork(),
	}
	_ret.ident.SetParent(_ret)
	return _ret
}

func (n *VarIdentNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.ident.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *VarIdentNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"var_ident\""
	ret["ident"] = DumpNode(n.Ident(), hook)
	return ret
}

func NewTypeIdentNode(filePath string, fileContent []rune, ident Node, start, end Position) Node {
	if ident == nil {
		ident = DummyNode
	}
	_1 := &TypeIdentNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeTypeIdent, start, end),
		ident:    ident,
	}
	creationHook(_1)
	return _1
}

type TypeIdentNode struct {
	*BaseNode
	ident Node
}

func (n *TypeIdentNode) Ident() Node {
	return n.ident
}

func (n *TypeIdentNode) SetIdent(v Node) {
	n.ident = v
}

func (n *TypeIdentNode) BuildLink() {
	if !n.Ident().IsDummy() {
		ident := n.Ident()
		ident.BuildLink()
		ident.SetParent(n)
		ident.SetSelfField("ident")
		ident.SetReplaceSelf(func(n Node) {
			n.Parent().(*TypeIdentNode).SetIdent(n)
		})
	}
}

func (n *TypeIdentNode) Fields() []string {
	return []string{
		"ident",
	}
}

func (n *TypeIdentNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "ident" {
		return n.Ident()
	}
	return nil
}

func (n *TypeIdentNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetIdent(nodes[0])
}

func (n *TypeIdentNode) Fork() Node {
	_ret := &TypeIdentNode{
		BaseNode: n.BaseNode.fork(),
		ident:    n.ident.Fork(),
	}
	_ret.ident.SetParent(_ret)
	return _ret
}

func (n *TypeIdentNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.ident.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *TypeIdentNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"type_ident\""
	ret["ident"] = DumpNode(n.Ident(), hook)
	return ret
}

func NewFunctionIdentNode(filePath string, fileContent []rune, ident Node, start, end Position) Node {
	if ident == nil {
		ident = DummyNode
	}
	_1 := &FunctionIdentNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeFunctionIdent, start, end),
		ident:    ident,
	}
	creationHook(_1)
	return _1
}

type FunctionIdentNode struct {
	*BaseNode
	ident Node
}

func (n *FunctionIdentNode) Ident() Node {
	return n.ident
}

func (n *FunctionIdentNode) SetIdent(v Node) {
	n.ident = v
}

func (n *FunctionIdentNode) BuildLink() {
	if !n.Ident().IsDummy() {
		ident := n.Ident()
		ident.BuildLink()
		ident.SetParent(n)
		ident.SetSelfField("ident")
		ident.SetReplaceSelf(func(n Node) {
			n.Parent().(*FunctionIdentNode).SetIdent(n)
		})
	}
}

func (n *FunctionIdentNode) Fields() []string {
	return []string{
		"ident",
	}
}

func (n *FunctionIdentNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "ident" {
		return n.Ident()
	}
	return nil
}

func (n *FunctionIdentNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetIdent(nodes[0])
}

func (n *FunctionIdentNode) Fork() Node {
	_ret := &FunctionIdentNode{
		BaseNode: n.BaseNode.fork(),
		ident:    n.ident.Fork(),
	}
	_ret.ident.SetParent(_ret)
	return _ret
}

func (n *FunctionIdentNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.ident.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *FunctionIdentNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"function_ident\""
	ret["ident"] = DumpNode(n.Ident(), hook)
	return ret
}

func NewMethodIdentNode(filePath string, fileContent []rune, ident Node, start, end Position) Node {
	if ident == nil {
		ident = DummyNode
	}
	_1 := &MethodIdentNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeMethodIdent, start, end),
		ident:    ident,
	}
	creationHook(_1)
	return _1
}

type MethodIdentNode struct {
	*BaseNode
	ident Node
}

func (n *MethodIdentNode) Ident() Node {
	return n.ident
}

func (n *MethodIdentNode) SetIdent(v Node) {
	n.ident = v
}

func (n *MethodIdentNode) BuildLink() {
	if !n.Ident().IsDummy() {
		ident := n.Ident()
		ident.BuildLink()
		ident.SetParent(n)
		ident.SetSelfField("ident")
		ident.SetReplaceSelf(func(n Node) {
			n.Parent().(*MethodIdentNode).SetIdent(n)
		})
	}
}

func (n *MethodIdentNode) Fields() []string {
	return []string{
		"ident",
	}
}

func (n *MethodIdentNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "ident" {
		return n.Ident()
	}
	return nil
}

func (n *MethodIdentNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetIdent(nodes[0])
}

func (n *MethodIdentNode) Fork() Node {
	_ret := &MethodIdentNode{
		BaseNode: n.BaseNode.fork(),
		ident:    n.ident.Fork(),
	}
	_ret.ident.SetParent(_ret)
	return _ret
}

func (n *MethodIdentNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.ident.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *MethodIdentNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"method_ident\""
	ret["ident"] = DumpNode(n.Ident(), hook)
	return ret
}

func NewGenericParameterNode(filePath string, fileContent []rune, name Node, type_ Node, start, end Position) Node {
	if name == nil {
		name = DummyNode
	}
	if type_ == nil {
		type_ = DummyNode
	}
	_1 := &GenericParameterNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeGenericParameter, start, end),
		name:     name,
		type_:    type_,
	}
	creationHook(_1)
	return _1
}

type GenericParameterNode struct {
	*BaseNode
	name  Node
	type_ Node
}

func (n *GenericParameterNode) Name() Node {
	return n.name
}

func (n *GenericParameterNode) SetName(v Node) {
	n.name = v
}

func (n *GenericParameterNode) Type() Node {
	return n.type_
}

func (n *GenericParameterNode) SetType(v Node) {
	n.type_ = v
}

func (n *GenericParameterNode) BuildLink() {
	if !n.Name().IsDummy() {
		name := n.Name()
		name.BuildLink()
		name.SetParent(n)
		name.SetSelfField("name")
		name.SetReplaceSelf(func(n Node) {
			n.Parent().(*GenericParameterNode).SetName(n)
		})
	}
	if !n.Type().IsDummy() {
		type_ := n.Type()
		type_.BuildLink()
		type_.SetParent(n)
		type_.SetSelfField("type_")
		type_.SetReplaceSelf(func(n Node) {
			n.Parent().(*GenericParameterNode).SetType(n)
		})
	}
}

func (n *GenericParameterNode) Fields() []string {
	return []string{
		"name",
		"type_",
	}
}

func (n *GenericParameterNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "name" {
		return n.Name()
	}
	if field == "type_" {
		return n.Type()
	}
	return nil
}

func (n *GenericParameterNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetName(nodes[0])
	n.SetType(nodes[1])
}

func (n *GenericParameterNode) Fork() Node {
	_ret := &GenericParameterNode{
		BaseNode: n.BaseNode.fork(),
		name:     n.name.Fork(),
		type_:    n.type_.Fork(),
	}
	_ret.name.SetParent(_ret)
	_ret.type_.SetParent(_ret)
	return _ret
}

func (n *GenericParameterNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.name.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.type_.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *GenericParameterNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"generic_parameter\""
	ret["name"] = DumpNode(n.Name(), hook)
	ret["type"] = DumpNode(n.Type(), hook)
	return ret
}

func NewGenericParameterIdentNode(filePath string, fileContent []rune, ident Node, start, end Position) Node {
	if ident == nil {
		ident = DummyNode
	}
	_1 := &GenericParameterIdentNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeGenericParameterIdent, start, end),
		ident:    ident,
	}
	creationHook(_1)
	return _1
}

type GenericParameterIdentNode struct {
	*BaseNode
	ident Node
}

func (n *GenericParameterIdentNode) Ident() Node {
	return n.ident
}

func (n *GenericParameterIdentNode) SetIdent(v Node) {
	n.ident = v
}

func (n *GenericParameterIdentNode) BuildLink() {
	if !n.Ident().IsDummy() {
		ident := n.Ident()
		ident.BuildLink()
		ident.SetParent(n)
		ident.SetSelfField("ident")
		ident.SetReplaceSelf(func(n Node) {
			n.Parent().(*GenericParameterIdentNode).SetIdent(n)
		})
	}
}

func (n *GenericParameterIdentNode) Fields() []string {
	return []string{
		"ident",
	}
}

func (n *GenericParameterIdentNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "ident" {
		return n.Ident()
	}
	return nil
}

func (n *GenericParameterIdentNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetIdent(nodes[0])
}

func (n *GenericParameterIdentNode) Fork() Node {
	_ret := &GenericParameterIdentNode{
		BaseNode: n.BaseNode.fork(),
		ident:    n.ident.Fork(),
	}
	_ret.ident.SetParent(_ret)
	return _ret
}

func (n *GenericParameterIdentNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.ident.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *GenericParameterIdentNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"generic_parameter_ident\""
	ret["ident"] = DumpNode(n.Ident(), hook)
	return ret
}

func NewGenericUnionConstraintNode(filePath string, fileContent []rune, list Node, start, end Position) Node {
	if list == nil {
		list = DummyNode
	}
	_1 := &GenericUnionConstraintNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeGenericUnionConstraint, start, end),
		list:     list,
	}
	creationHook(_1)
	return _1
}

type GenericUnionConstraintNode struct {
	*BaseNode
	list Node
}

func (n *GenericUnionConstraintNode) List() Node {
	return n.list
}

func (n *GenericUnionConstraintNode) SetList(v Node) {
	n.list = v
}

func (n *GenericUnionConstraintNode) BuildLink() {
	if !n.List().IsDummy() {
		list := n.List()
		list.BuildLink()
		list.SetParent(n)
		list.SetSelfField("list")
		list.SetReplaceSelf(func(n Node) {
			n.Parent().(*GenericUnionConstraintNode).SetList(n)
		})
	}
}

func (n *GenericUnionConstraintNode) Fields() []string {
	return []string{
		"list",
	}
}

func (n *GenericUnionConstraintNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "list" {
		return n.List()
	}
	return nil
}

func (n *GenericUnionConstraintNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetList(nodes[0])
}

func (n *GenericUnionConstraintNode) Fork() Node {
	_ret := &GenericUnionConstraintNode{
		BaseNode: n.BaseNode.fork(),
		list:     n.list.Fork(),
	}
	_ret.list.SetParent(_ret)
	return _ret
}

func (n *GenericUnionConstraintNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.list.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *GenericUnionConstraintNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"generic_union_constraint\""
	ret["list"] = DumpNode(n.List(), hook)
	return ret
}

func NewGenericUnderlyingTypeConstraintNode(filePath string, fileContent []rune, type_ Node, start, end Position) Node {
	if type_ == nil {
		type_ = DummyNode
	}
	_1 := &GenericUnderlyingTypeConstraintNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeGenericUnderlyingTypeConstraint, start, end),
		type_:    type_,
	}
	creationHook(_1)
	return _1
}

type GenericUnderlyingTypeConstraintNode struct {
	*BaseNode
	type_ Node
}

func (n *GenericUnderlyingTypeConstraintNode) Type() Node {
	return n.type_
}

func (n *GenericUnderlyingTypeConstraintNode) SetType(v Node) {
	n.type_ = v
}

func (n *GenericUnderlyingTypeConstraintNode) BuildLink() {
	if !n.Type().IsDummy() {
		type_ := n.Type()
		type_.BuildLink()
		type_.SetParent(n)
		type_.SetSelfField("type_")
		type_.SetReplaceSelf(func(n Node) {
			n.Parent().(*GenericUnderlyingTypeConstraintNode).SetType(n)
		})
	}
}

func (n *GenericUnderlyingTypeConstraintNode) Fields() []string {
	return []string{
		"type_",
	}
}

func (n *GenericUnderlyingTypeConstraintNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "type_" {
		return n.Type()
	}
	return nil
}

func (n *GenericUnderlyingTypeConstraintNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetType(nodes[0])
}

func (n *GenericUnderlyingTypeConstraintNode) Fork() Node {
	_ret := &GenericUnderlyingTypeConstraintNode{
		BaseNode: n.BaseNode.fork(),
		type_:    n.type_.Fork(),
	}
	_ret.type_.SetParent(_ret)
	return _ret
}

func (n *GenericUnderlyingTypeConstraintNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.type_.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *GenericUnderlyingTypeConstraintNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"generic_underlying_type_constraint\""
	ret["type"] = DumpNode(n.Type(), hook)
	return ret
}

func NewGenericTypeConstraintNode(filePath string, fileContent []rune, type_ Node, start, end Position) Node {
	if type_ == nil {
		type_ = DummyNode
	}
	_1 := &GenericTypeConstraintNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeGenericTypeConstraint, start, end),
		type_:    type_,
	}
	creationHook(_1)
	return _1
}

type GenericTypeConstraintNode struct {
	*BaseNode
	type_ Node
}

func (n *GenericTypeConstraintNode) Type() Node {
	return n.type_
}

func (n *GenericTypeConstraintNode) SetType(v Node) {
	n.type_ = v
}

func (n *GenericTypeConstraintNode) BuildLink() {
	if !n.Type().IsDummy() {
		type_ := n.Type()
		type_.BuildLink()
		type_.SetParent(n)
		type_.SetSelfField("type_")
		type_.SetReplaceSelf(func(n Node) {
			n.Parent().(*GenericTypeConstraintNode).SetType(n)
		})
	}
}

func (n *GenericTypeConstraintNode) Fields() []string {
	return []string{
		"type_",
	}
}

func (n *GenericTypeConstraintNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "type_" {
		return n.Type()
	}
	return nil
}

func (n *GenericTypeConstraintNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetType(nodes[0])
}

func (n *GenericTypeConstraintNode) Fork() Node {
	_ret := &GenericTypeConstraintNode{
		BaseNode: n.BaseNode.fork(),
		type_:    n.type_.Fork(),
	}
	_ret.type_.SetParent(_ret)
	return _ret
}

func (n *GenericTypeConstraintNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.type_.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *GenericTypeConstraintNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"generic_type_constraint\""
	ret["type"] = DumpNode(n.Type(), hook)
	return ret
}

func NewEllipsisParameterNode(filePath string, fileContent []rune, name Node, type_ Node, start, end Position) Node {
	if name == nil {
		name = DummyNode
	}
	if type_ == nil {
		type_ = DummyNode
	}
	_1 := &EllipsisParameterNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeEllipsisParameter, start, end),
		name:     name,
		type_:    type_,
	}
	creationHook(_1)
	return _1
}

type EllipsisParameterNode struct {
	*BaseNode
	name  Node
	type_ Node
}

func (n *EllipsisParameterNode) Name() Node {
	return n.name
}

func (n *EllipsisParameterNode) SetName(v Node) {
	n.name = v
}

func (n *EllipsisParameterNode) Type() Node {
	return n.type_
}

func (n *EllipsisParameterNode) SetType(v Node) {
	n.type_ = v
}

func (n *EllipsisParameterNode) BuildLink() {
	if !n.Name().IsDummy() {
		name := n.Name()
		name.BuildLink()
		name.SetParent(n)
		name.SetSelfField("name")
		name.SetReplaceSelf(func(n Node) {
			n.Parent().(*EllipsisParameterNode).SetName(n)
		})
	}
	if !n.Type().IsDummy() {
		type_ := n.Type()
		type_.BuildLink()
		type_.SetParent(n)
		type_.SetSelfField("type_")
		type_.SetReplaceSelf(func(n Node) {
			n.Parent().(*EllipsisParameterNode).SetType(n)
		})
	}
}

func (n *EllipsisParameterNode) Fields() []string {
	return []string{
		"name",
		"type_",
	}
}

func (n *EllipsisParameterNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "name" {
		return n.Name()
	}
	if field == "type_" {
		return n.Type()
	}
	return nil
}

func (n *EllipsisParameterNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetName(nodes[0])
	n.SetType(nodes[1])
}

func (n *EllipsisParameterNode) Fork() Node {
	_ret := &EllipsisParameterNode{
		BaseNode: n.BaseNode.fork(),
		name:     n.name.Fork(),
		type_:    n.type_.Fork(),
	}
	_ret.name.SetParent(_ret)
	_ret.type_.SetParent(_ret)
	return _ret
}

func (n *EllipsisParameterNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.name.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.type_.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *EllipsisParameterNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"ellipsis_parameter\""
	ret["name"] = DumpNode(n.Name(), hook)
	ret["type"] = DumpNode(n.Type(), hook)
	return ret
}

func NewParameterNode(filePath string, fileContent []rune, name Node, type_ Node, start, end Position) Node {
	if name == nil {
		name = DummyNode
	}
	if type_ == nil {
		type_ = DummyNode
	}
	_1 := &ParameterNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeParameter, start, end),
		name:     name,
		type_:    type_,
	}
	creationHook(_1)
	return _1
}

type ParameterNode struct {
	*BaseNode
	name  Node
	type_ Node
}

func (n *ParameterNode) Name() Node {
	return n.name
}

func (n *ParameterNode) SetName(v Node) {
	n.name = v
}

func (n *ParameterNode) Type() Node {
	return n.type_
}

func (n *ParameterNode) SetType(v Node) {
	n.type_ = v
}

func (n *ParameterNode) BuildLink() {
	if !n.Name().IsDummy() {
		name := n.Name()
		name.BuildLink()
		name.SetParent(n)
		name.SetSelfField("name")
		name.SetReplaceSelf(func(n Node) {
			n.Parent().(*ParameterNode).SetName(n)
		})
	}
	if !n.Type().IsDummy() {
		type_ := n.Type()
		type_.BuildLink()
		type_.SetParent(n)
		type_.SetSelfField("type_")
		type_.SetReplaceSelf(func(n Node) {
			n.Parent().(*ParameterNode).SetType(n)
		})
	}
}

func (n *ParameterNode) Fields() []string {
	return []string{
		"name",
		"type_",
	}
}

func (n *ParameterNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "name" {
		return n.Name()
	}
	if field == "type_" {
		return n.Type()
	}
	return nil
}

func (n *ParameterNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetName(nodes[0])
	n.SetType(nodes[1])
}

func (n *ParameterNode) Fork() Node {
	_ret := &ParameterNode{
		BaseNode: n.BaseNode.fork(),
		name:     n.name.Fork(),
		type_:    n.type_.Fork(),
	}
	_ret.name.SetParent(_ret)
	_ret.type_.SetParent(_ret)
	return _ret
}

func (n *ParameterNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.name.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.type_.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *ParameterNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"parameter\""
	ret["name"] = DumpNode(n.Name(), hook)
	ret["type"] = DumpNode(n.Type(), hook)
	return ret
}

func NewParameterIdentNode(filePath string, fileContent []rune, ident Node, start, end Position) Node {
	if ident == nil {
		ident = DummyNode
	}
	_1 := &ParameterIdentNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeParameterIdent, start, end),
		ident:    ident,
	}
	creationHook(_1)
	return _1
}

type ParameterIdentNode struct {
	*BaseNode
	ident Node
}

func (n *ParameterIdentNode) Ident() Node {
	return n.ident
}

func (n *ParameterIdentNode) SetIdent(v Node) {
	n.ident = v
}

func (n *ParameterIdentNode) BuildLink() {
	if !n.Ident().IsDummy() {
		ident := n.Ident()
		ident.BuildLink()
		ident.SetParent(n)
		ident.SetSelfField("ident")
		ident.SetReplaceSelf(func(n Node) {
			n.Parent().(*ParameterIdentNode).SetIdent(n)
		})
	}
}

func (n *ParameterIdentNode) Fields() []string {
	return []string{
		"ident",
	}
}

func (n *ParameterIdentNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "ident" {
		return n.Ident()
	}
	return nil
}

func (n *ParameterIdentNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetIdent(nodes[0])
}

func (n *ParameterIdentNode) Fork() Node {
	_ret := &ParameterIdentNode{
		BaseNode: n.BaseNode.fork(),
		ident:    n.ident.Fork(),
	}
	_ret.ident.SetParent(_ret)
	return _ret
}

func (n *ParameterIdentNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.ident.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *ParameterIdentNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"parameter_ident\""
	ret["ident"] = DumpNode(n.Ident(), hook)
	return ret
}

func NewFunctionResultNode(filePath string, fileContent []rune, name Node, type_ Node, start, end Position) Node {
	if name == nil {
		name = DummyNode
	}
	if type_ == nil {
		type_ = DummyNode
	}
	_1 := &FunctionResultNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeFunctionResult, start, end),
		name:     name,
		type_:    type_,
	}
	creationHook(_1)
	return _1
}

type FunctionResultNode struct {
	*BaseNode
	name  Node
	type_ Node
}

func (n *FunctionResultNode) Name() Node {
	return n.name
}

func (n *FunctionResultNode) SetName(v Node) {
	n.name = v
}

func (n *FunctionResultNode) Type() Node {
	return n.type_
}

func (n *FunctionResultNode) SetType(v Node) {
	n.type_ = v
}

func (n *FunctionResultNode) BuildLink() {
	if !n.Name().IsDummy() {
		name := n.Name()
		name.BuildLink()
		name.SetParent(n)
		name.SetSelfField("name")
		name.SetReplaceSelf(func(n Node) {
			n.Parent().(*FunctionResultNode).SetName(n)
		})
	}
	if !n.Type().IsDummy() {
		type_ := n.Type()
		type_.BuildLink()
		type_.SetParent(n)
		type_.SetSelfField("type_")
		type_.SetReplaceSelf(func(n Node) {
			n.Parent().(*FunctionResultNode).SetType(n)
		})
	}
}

func (n *FunctionResultNode) Fields() []string {
	return []string{
		"name",
		"type_",
	}
}

func (n *FunctionResultNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "name" {
		return n.Name()
	}
	if field == "type_" {
		return n.Type()
	}
	return nil
}

func (n *FunctionResultNode) SetChild(nodes []Node) {
	if len(nodes) != 2 {
		return
	}
	n.SetName(nodes[0])
	n.SetType(nodes[1])
}

func (n *FunctionResultNode) Fork() Node {
	_ret := &FunctionResultNode{
		BaseNode: n.BaseNode.fork(),
		name:     n.name.Fork(),
		type_:    n.type_.Fork(),
	}
	_ret.name.SetParent(_ret)
	_ret.type_.SetParent(_ret)
	return _ret
}

func (n *FunctionResultNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.name.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.type_.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *FunctionResultNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"function_result\""
	ret["name"] = DumpNode(n.Name(), hook)
	ret["type"] = DumpNode(n.Type(), hook)
	return ret
}

func NewFunctionResultIdentNode(filePath string, fileContent []rune, ident Node, start, end Position) Node {
	if ident == nil {
		ident = DummyNode
	}
	_1 := &FunctionResultIdentNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeFunctionResultIdent, start, end),
		ident:    ident,
	}
	creationHook(_1)
	return _1
}

type FunctionResultIdentNode struct {
	*BaseNode
	ident Node
}

func (n *FunctionResultIdentNode) Ident() Node {
	return n.ident
}

func (n *FunctionResultIdentNode) SetIdent(v Node) {
	n.ident = v
}

func (n *FunctionResultIdentNode) BuildLink() {
	if !n.Ident().IsDummy() {
		ident := n.Ident()
		ident.BuildLink()
		ident.SetParent(n)
		ident.SetSelfField("ident")
		ident.SetReplaceSelf(func(n Node) {
			n.Parent().(*FunctionResultIdentNode).SetIdent(n)
		})
	}
}

func (n *FunctionResultIdentNode) Fields() []string {
	return []string{
		"ident",
	}
}

func (n *FunctionResultIdentNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "ident" {
		return n.Ident()
	}
	return nil
}

func (n *FunctionResultIdentNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetIdent(nodes[0])
}

func (n *FunctionResultIdentNode) Fork() Node {
	_ret := &FunctionResultIdentNode{
		BaseNode: n.BaseNode.fork(),
		ident:    n.ident.Fork(),
	}
	_ret.ident.SetParent(_ret)
	return _ret
}

func (n *FunctionResultIdentNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.ident.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *FunctionResultIdentNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"function_result_ident\""
	ret["ident"] = DumpNode(n.Ident(), hook)
	return ret
}

func NewFunctionDeclNode(filePath string, fileContent []rune, name Node, genericParameters Node, parameters Node, results Node, body Node, start, end Position) Node {
	if name == nil {
		name = DummyNode
	}
	if genericParameters == nil {
		genericParameters = DummyNode
	}
	if parameters == nil {
		parameters = DummyNode
	}
	if results == nil {
		results = DummyNode
	}
	if body == nil {
		body = DummyNode
	}
	_1 := &FunctionDeclNode{
		BaseNode:          NewBaseNode(filePath, fileContent, NodeTypeFunctionDecl, start, end),
		name:              name,
		genericParameters: genericParameters,
		parameters:        parameters,
		results:           results,
		body:              body,
	}
	creationHook(_1)
	return _1
}

type FunctionDeclNode struct {
	*BaseNode
	name              Node
	genericParameters Node
	parameters        Node
	results           Node
	body              Node
}

func (n *FunctionDeclNode) Name() Node {
	return n.name
}

func (n *FunctionDeclNode) SetName(v Node) {
	n.name = v
}

func (n *FunctionDeclNode) GenericParameters() Node {
	return n.genericParameters
}

func (n *FunctionDeclNode) SetGenericParameters(v Node) {
	n.genericParameters = v
}

func (n *FunctionDeclNode) Parameters() Node {
	return n.parameters
}

func (n *FunctionDeclNode) SetParameters(v Node) {
	n.parameters = v
}

func (n *FunctionDeclNode) Results() Node {
	return n.results
}

func (n *FunctionDeclNode) SetResults(v Node) {
	n.results = v
}

func (n *FunctionDeclNode) Body() Node {
	return n.body
}

func (n *FunctionDeclNode) SetBody(v Node) {
	n.body = v
}

func (n *FunctionDeclNode) BuildLink() {
	if !n.Name().IsDummy() {
		name := n.Name()
		name.BuildLink()
		name.SetParent(n)
		name.SetSelfField("name")
		name.SetReplaceSelf(func(n Node) {
			n.Parent().(*FunctionDeclNode).SetName(n)
		})
	}
	if !n.GenericParameters().IsDummy() {
		genericParameters := n.GenericParameters()
		genericParameters.BuildLink()
		genericParameters.SetParent(n)
		genericParameters.SetSelfField("generic_parameters")
		genericParameters.SetReplaceSelf(func(n Node) {
			n.Parent().(*FunctionDeclNode).SetGenericParameters(n)
		})
	}
	if !n.Parameters().IsDummy() {
		parameters := n.Parameters()
		parameters.BuildLink()
		parameters.SetParent(n)
		parameters.SetSelfField("parameters")
		parameters.SetReplaceSelf(func(n Node) {
			n.Parent().(*FunctionDeclNode).SetParameters(n)
		})
	}
	if !n.Results().IsDummy() {
		results := n.Results()
		results.BuildLink()
		results.SetParent(n)
		results.SetSelfField("results")
		results.SetReplaceSelf(func(n Node) {
			n.Parent().(*FunctionDeclNode).SetResults(n)
		})
	}
	if !n.Body().IsDummy() {
		body := n.Body()
		body.BuildLink()
		body.SetParent(n)
		body.SetSelfField("body")
		body.SetReplaceSelf(func(n Node) {
			n.Parent().(*FunctionDeclNode).SetBody(n)
		})
	}
}

func (n *FunctionDeclNode) Fields() []string {
	return []string{
		"name",
		"generic_parameters",
		"parameters",
		"results",
		"body",
	}
}

func (n *FunctionDeclNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "name" {
		return n.Name()
	}
	if field == "generic_parameters" {
		return n.GenericParameters()
	}
	if field == "parameters" {
		return n.Parameters()
	}
	if field == "results" {
		return n.Results()
	}
	if field == "body" {
		return n.Body()
	}
	return nil
}

func (n *FunctionDeclNode) SetChild(nodes []Node) {
	if len(nodes) != 5 {
		return
	}
	n.SetName(nodes[0])
	n.SetGenericParameters(nodes[1])
	n.SetParameters(nodes[2])
	n.SetResults(nodes[3])
	n.SetBody(nodes[4])
}

func (n *FunctionDeclNode) Fork() Node {
	_ret := &FunctionDeclNode{
		BaseNode:          n.BaseNode.fork(),
		name:              n.name.Fork(),
		genericParameters: n.genericParameters.Fork(),
		parameters:        n.parameters.Fork(),
		results:           n.results.Fork(),
		body:              n.body.Fork(),
	}
	_ret.name.SetParent(_ret)
	_ret.genericParameters.SetParent(_ret)
	_ret.parameters.SetParent(_ret)
	_ret.results.SetParent(_ret)
	_ret.body.SetParent(_ret)
	return _ret
}

func (n *FunctionDeclNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.name.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.genericParameters.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.parameters.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.results.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.body.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *FunctionDeclNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"function_decl\""
	ret["name"] = DumpNode(n.Name(), hook)
	ret["generic_parameters"] = DumpNode(n.GenericParameters(), hook)
	ret["parameters"] = DumpNode(n.Parameters(), hook)
	ret["results"] = DumpNode(n.Results(), hook)
	ret["body"] = DumpNode(n.Body(), hook)
	return ret
}

func NewMethodDeclNode(filePath string, fileContent []rune, receiver Node, name Node, genericParameters Node, parameters Node, results Node, body Node, start, end Position) Node {
	if receiver == nil {
		receiver = DummyNode
	}
	if name == nil {
		name = DummyNode
	}
	if genericParameters == nil {
		genericParameters = DummyNode
	}
	if parameters == nil {
		parameters = DummyNode
	}
	if results == nil {
		results = DummyNode
	}
	if body == nil {
		body = DummyNode
	}
	_1 := &MethodDeclNode{
		BaseNode:          NewBaseNode(filePath, fileContent, NodeTypeMethodDecl, start, end),
		receiver:          receiver,
		name:              name,
		genericParameters: genericParameters,
		parameters:        parameters,
		results:           results,
		body:              body,
	}
	creationHook(_1)
	return _1
}

type MethodDeclNode struct {
	*BaseNode
	receiver          Node
	name              Node
	genericParameters Node
	parameters        Node
	results           Node
	body              Node
}

func (n *MethodDeclNode) Receiver() Node {
	return n.receiver
}

func (n *MethodDeclNode) SetReceiver(v Node) {
	n.receiver = v
}

func (n *MethodDeclNode) Name() Node {
	return n.name
}

func (n *MethodDeclNode) SetName(v Node) {
	n.name = v
}

func (n *MethodDeclNode) GenericParameters() Node {
	return n.genericParameters
}

func (n *MethodDeclNode) SetGenericParameters(v Node) {
	n.genericParameters = v
}

func (n *MethodDeclNode) Parameters() Node {
	return n.parameters
}

func (n *MethodDeclNode) SetParameters(v Node) {
	n.parameters = v
}

func (n *MethodDeclNode) Results() Node {
	return n.results
}

func (n *MethodDeclNode) SetResults(v Node) {
	n.results = v
}

func (n *MethodDeclNode) Body() Node {
	return n.body
}

func (n *MethodDeclNode) SetBody(v Node) {
	n.body = v
}

func (n *MethodDeclNode) BuildLink() {
	if !n.Receiver().IsDummy() {
		receiver := n.Receiver()
		receiver.BuildLink()
		receiver.SetParent(n)
		receiver.SetSelfField("receiver")
		receiver.SetReplaceSelf(func(n Node) {
			n.Parent().(*MethodDeclNode).SetReceiver(n)
		})
	}
	if !n.Name().IsDummy() {
		name := n.Name()
		name.BuildLink()
		name.SetParent(n)
		name.SetSelfField("name")
		name.SetReplaceSelf(func(n Node) {
			n.Parent().(*MethodDeclNode).SetName(n)
		})
	}
	if !n.GenericParameters().IsDummy() {
		genericParameters := n.GenericParameters()
		genericParameters.BuildLink()
		genericParameters.SetParent(n)
		genericParameters.SetSelfField("generic_parameters")
		genericParameters.SetReplaceSelf(func(n Node) {
			n.Parent().(*MethodDeclNode).SetGenericParameters(n)
		})
	}
	if !n.Parameters().IsDummy() {
		parameters := n.Parameters()
		parameters.BuildLink()
		parameters.SetParent(n)
		parameters.SetSelfField("parameters")
		parameters.SetReplaceSelf(func(n Node) {
			n.Parent().(*MethodDeclNode).SetParameters(n)
		})
	}
	if !n.Results().IsDummy() {
		results := n.Results()
		results.BuildLink()
		results.SetParent(n)
		results.SetSelfField("results")
		results.SetReplaceSelf(func(n Node) {
			n.Parent().(*MethodDeclNode).SetResults(n)
		})
	}
	if !n.Body().IsDummy() {
		body := n.Body()
		body.BuildLink()
		body.SetParent(n)
		body.SetSelfField("body")
		body.SetReplaceSelf(func(n Node) {
			n.Parent().(*MethodDeclNode).SetBody(n)
		})
	}
}

func (n *MethodDeclNode) Fields() []string {
	return []string{
		"receiver",
		"name",
		"generic_parameters",
		"parameters",
		"results",
		"body",
	}
}

func (n *MethodDeclNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "receiver" {
		return n.Receiver()
	}
	if field == "name" {
		return n.Name()
	}
	if field == "generic_parameters" {
		return n.GenericParameters()
	}
	if field == "parameters" {
		return n.Parameters()
	}
	if field == "results" {
		return n.Results()
	}
	if field == "body" {
		return n.Body()
	}
	return nil
}

func (n *MethodDeclNode) SetChild(nodes []Node) {
	if len(nodes) != 6 {
		return
	}
	n.SetReceiver(nodes[0])
	n.SetName(nodes[1])
	n.SetGenericParameters(nodes[2])
	n.SetParameters(nodes[3])
	n.SetResults(nodes[4])
	n.SetBody(nodes[5])
}

func (n *MethodDeclNode) Fork() Node {
	_ret := &MethodDeclNode{
		BaseNode:          n.BaseNode.fork(),
		receiver:          n.receiver.Fork(),
		name:              n.name.Fork(),
		genericParameters: n.genericParameters.Fork(),
		parameters:        n.parameters.Fork(),
		results:           n.results.Fork(),
		body:              n.body.Fork(),
	}
	_ret.receiver.SetParent(_ret)
	_ret.name.SetParent(_ret)
	_ret.genericParameters.SetParent(_ret)
	_ret.parameters.SetParent(_ret)
	_ret.results.SetParent(_ret)
	_ret.body.SetParent(_ret)
	return _ret
}

func (n *MethodDeclNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.receiver.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.name.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.genericParameters.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.parameters.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.results.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.body.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *MethodDeclNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"method_decl\""
	ret["receiver"] = DumpNode(n.Receiver(), hook)
	ret["name"] = DumpNode(n.Name(), hook)
	ret["generic_parameters"] = DumpNode(n.GenericParameters(), hook)
	ret["parameters"] = DumpNode(n.Parameters(), hook)
	ret["results"] = DumpNode(n.Results(), hook)
	ret["body"] = DumpNode(n.Body(), hook)
	return ret
}

func NewReceiverIdentNode(filePath string, fileContent []rune, ident Node, start, end Position) Node {
	if ident == nil {
		ident = DummyNode
	}
	_1 := &ReceiverIdentNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeReceiverIdent, start, end),
		ident:    ident,
	}
	creationHook(_1)
	return _1
}

type ReceiverIdentNode struct {
	*BaseNode
	ident Node
}

func (n *ReceiverIdentNode) Ident() Node {
	return n.ident
}

func (n *ReceiverIdentNode) SetIdent(v Node) {
	n.ident = v
}

func (n *ReceiverIdentNode) BuildLink() {
	if !n.Ident().IsDummy() {
		ident := n.Ident()
		ident.BuildLink()
		ident.SetParent(n)
		ident.SetSelfField("ident")
		ident.SetReplaceSelf(func(n Node) {
			n.Parent().(*ReceiverIdentNode).SetIdent(n)
		})
	}
}

func (n *ReceiverIdentNode) Fields() []string {
	return []string{
		"ident",
	}
}

func (n *ReceiverIdentNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "ident" {
		return n.Ident()
	}
	return nil
}

func (n *ReceiverIdentNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetIdent(nodes[0])
}

func (n *ReceiverIdentNode) Fork() Node {
	_ret := &ReceiverIdentNode{
		BaseNode: n.BaseNode.fork(),
		ident:    n.ident.Fork(),
	}
	_ret.ident.SetParent(_ret)
	return _ret
}

func (n *ReceiverIdentNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.ident.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *ReceiverIdentNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"receiver_ident\""
	ret["ident"] = DumpNode(n.Ident(), hook)
	return ret
}

func NewReceiverTypeIdentNode(filePath string, fileContent []rune, ident Node, start, end Position) Node {
	if ident == nil {
		ident = DummyNode
	}
	_1 := &ReceiverTypeIdentNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeReceiverTypeIdent, start, end),
		ident:    ident,
	}
	creationHook(_1)
	return _1
}

type ReceiverTypeIdentNode struct {
	*BaseNode
	ident Node
}

func (n *ReceiverTypeIdentNode) Ident() Node {
	return n.ident
}

func (n *ReceiverTypeIdentNode) SetIdent(v Node) {
	n.ident = v
}

func (n *ReceiverTypeIdentNode) BuildLink() {
	if !n.Ident().IsDummy() {
		ident := n.Ident()
		ident.BuildLink()
		ident.SetParent(n)
		ident.SetSelfField("ident")
		ident.SetReplaceSelf(func(n Node) {
			n.Parent().(*ReceiverTypeIdentNode).SetIdent(n)
		})
	}
}

func (n *ReceiverTypeIdentNode) Fields() []string {
	return []string{
		"ident",
	}
}

func (n *ReceiverTypeIdentNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "ident" {
		return n.Ident()
	}
	return nil
}

func (n *ReceiverTypeIdentNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetIdent(nodes[0])
}

func (n *ReceiverTypeIdentNode) Fork() Node {
	_ret := &ReceiverTypeIdentNode{
		BaseNode: n.BaseNode.fork(),
		ident:    n.ident.Fork(),
	}
	_ret.ident.SetParent(_ret)
	return _ret
}

func (n *ReceiverTypeIdentNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.ident.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *ReceiverTypeIdentNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"receiver_type_ident\""
	ret["ident"] = DumpNode(n.Ident(), hook)
	return ret
}

func NewReceiverGenericTypeIdentNode(filePath string, fileContent []rune, ident Node, start, end Position) Node {
	if ident == nil {
		ident = DummyNode
	}
	_1 := &ReceiverGenericTypeIdentNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeReceiverGenericTypeIdent, start, end),
		ident:    ident,
	}
	creationHook(_1)
	return _1
}

type ReceiverGenericTypeIdentNode struct {
	*BaseNode
	ident Node
}

func (n *ReceiverGenericTypeIdentNode) Ident() Node {
	return n.ident
}

func (n *ReceiverGenericTypeIdentNode) SetIdent(v Node) {
	n.ident = v
}

func (n *ReceiverGenericTypeIdentNode) BuildLink() {
	if !n.Ident().IsDummy() {
		ident := n.Ident()
		ident.BuildLink()
		ident.SetParent(n)
		ident.SetSelfField("ident")
		ident.SetReplaceSelf(func(n Node) {
			n.Parent().(*ReceiverGenericTypeIdentNode).SetIdent(n)
		})
	}
}

func (n *ReceiverGenericTypeIdentNode) Fields() []string {
	return []string{
		"ident",
	}
}

func (n *ReceiverGenericTypeIdentNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "ident" {
		return n.Ident()
	}
	return nil
}

func (n *ReceiverGenericTypeIdentNode) SetChild(nodes []Node) {
	if len(nodes) != 1 {
		return
	}
	n.SetIdent(nodes[0])
}

func (n *ReceiverGenericTypeIdentNode) Fork() Node {
	_ret := &ReceiverGenericTypeIdentNode{
		BaseNode: n.BaseNode.fork(),
		ident:    n.ident.Fork(),
	}
	_ret.ident.SetParent(_ret)
	return _ret
}

func (n *ReceiverGenericTypeIdentNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.ident.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *ReceiverGenericTypeIdentNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"receiver_generic_type_ident\""
	ret["ident"] = DumpNode(n.Ident(), hook)
	return ret
}

func NewReceiverNode(filePath string, fileContent []rune, name Node, star Node, type_ Node, start, end Position) Node {
	if name == nil {
		name = DummyNode
	}
	if star == nil {
		star = DummyNode
	}
	if type_ == nil {
		type_ = DummyNode
	}
	_1 := &ReceiverNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeReceiver, start, end),
		name:     name,
		star:     star,
		type_:    type_,
	}
	creationHook(_1)
	return _1
}

type ReceiverNode struct {
	*BaseNode
	name  Node
	star  Node
	type_ Node
}

func (n *ReceiverNode) Name() Node {
	return n.name
}

func (n *ReceiverNode) SetName(v Node) {
	n.name = v
}

func (n *ReceiverNode) Star() Node {
	return n.star
}

func (n *ReceiverNode) SetStar(v Node) {
	n.star = v
}

func (n *ReceiverNode) Type() Node {
	return n.type_
}

func (n *ReceiverNode) SetType(v Node) {
	n.type_ = v
}

func (n *ReceiverNode) BuildLink() {
	if !n.Name().IsDummy() {
		name := n.Name()
		name.BuildLink()
		name.SetParent(n)
		name.SetSelfField("name")
		name.SetReplaceSelf(func(n Node) {
			n.Parent().(*ReceiverNode).SetName(n)
		})
	}
	if !n.Star().IsDummy() {
		star := n.Star()
		star.BuildLink()
		star.SetParent(n)
		star.SetSelfField("star")
		star.SetReplaceSelf(func(n Node) {
			n.Parent().(*ReceiverNode).SetStar(n)
		})
	}
	if !n.Type().IsDummy() {
		type_ := n.Type()
		type_.BuildLink()
		type_.SetParent(n)
		type_.SetSelfField("type_")
		type_.SetReplaceSelf(func(n Node) {
			n.Parent().(*ReceiverNode).SetType(n)
		})
	}
}

func (n *ReceiverNode) Fields() []string {
	return []string{
		"name",
		"star",
		"type_",
	}
}

func (n *ReceiverNode) Child(field string) Node {
	if field == "" {
		return nil
	}
	if field == "name" {
		return n.Name()
	}
	if field == "star" {
		return n.Star()
	}
	if field == "type_" {
		return n.Type()
	}
	return nil
}

func (n *ReceiverNode) SetChild(nodes []Node) {
	if len(nodes) != 3 {
		return
	}
	n.SetName(nodes[0])
	n.SetStar(nodes[1])
	n.SetType(nodes[2])
}

func (n *ReceiverNode) Fork() Node {
	_ret := &ReceiverNode{
		BaseNode: n.BaseNode.fork(),
		name:     n.name.Fork(),
		star:     n.star.Fork(),
		type_:    n.type_.Fork(),
	}
	_ret.name.SetParent(_ret)
	_ret.star.SetParent(_ret)
	_ret.type_.SetParent(_ret)
	return _ret
}

func (n *ReceiverNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if n.name.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.star.Visit(beforeChildren, afterChildren) {
		return true
	}
	if n.type_.Visit(beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *ReceiverNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	ret := make(map[string]string)
	ret["kind"] = "\"receiver\""
	ret["name"] = DumpNode(n.Name(), hook)
	ret["star"] = DumpNode(n.Star(), hook)
	ret["type"] = DumpNode(n.Type(), hook)
	return ret
}

func NewTokenizer(filePath string, fileContent []rune) *Tokenizer {
	tk := &Tokenizer{
		_filePath:  filePath,
		_buf:       fileContent,
		_bufSize:   len(fileContent),
		_pos:       Position{},
		_prevPos:   Position{},
		_lookahead: 0,
	}
	tk._lookahead = tk._safeRead()
	tk.initKeywords()
	return tk
}

type Tokenizer struct {
	_filePath  string
	_buf       []rune
	_bufSize   int
	_pos       Position
	_prevPos   Position
	_lookahead rune
	_keywords  map[string]string
}

func (tk *Tokenizer) Parse() (tokens []*Token, err error) {
	tokens = make([]*Token, 0)
	for {
		var tok *Token
		tok, err = tk.next()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, tok)
		if tok.Kind == TokenTypeEndOfFile {
			break
		}
	}
	return tokens, nil
}

func (tk *Tokenizer) _lineEnd(ch rune) bool {
	return ch == '\n' || (ch == '\r' && tk._pos.Offset < len(tk._buf) && tk._buf[tk._pos.Offset] != '\n')
}

func (tk *Tokenizer) _errorMsg(msg string) string {
	return fmt.Sprintf("fail to tokenize %s\n%s", msg, errorContext(tk._filePath, tk._buf, tk._prevPos.Offset, tk._prevPos.LineIdx, tk._prevPos.CharIdx))
}

func (tk *Tokenizer) _stepForward(ch rune) {
	p := &tk._pos
	p.Offset++
	p.CharIdx++
	if tk._lineEnd(ch) {
		p.LineIdx++
		p.CharIdx = 0
	}
}

func (tk *Tokenizer) _forward() {
	tk._stepForward(tk._safeRead())
	tk._lookahead = tk._safeRead()
}

func (tk *Tokenizer) _mark() Position {
	return tk._pos
}

func (tk *Tokenizer) _reset(p Position) {
	tk._pos = p
	tk._lookahead = tk._safeRead()
}

func (tk *Tokenizer) _safeRead() rune {
	if tk._pos.Offset >= tk._bufSize {
		return '\x00'
	} else {
		return tk._buf[tk._pos.Offset]
	}
}

func (tk *Tokenizer) _expect(r rune) bool {
	if equalRune(r, tk._lookahead) {
		tk._forward()
		return true
	}
	return false
}

func (tk *Tokenizer) _expectS(s string) bool {
	pos := tk._pos
	for i := 0; i < len(s); i++ {
		if equalRune(rune(s[i]), tk._lookahead) {
			tk._forward()
		} else {
			tk._reset(pos)
			return false
		}
	}
	return true
}

func (tk *Tokenizer) _expectU(s []rune) bool {
	pos := tk._pos
	for i := 0; i < len(s); i++ {
		if equalRune(s[i], tk._lookahead) {
			tk._forward()
		} else {
			tk._reset(pos)
			return false
		}
	}
	return true
}

func (tk *Tokenizer) _expectR(s, e rune) bool {
	if inRange(tk._lookahead, s, e) {
		tk._forward()
		return true
	}
	return false
}

func (tk *Tokenizer) _anyButEof() bool {
	if tk._lookahead != 0 {
		tk._forward()
		return true
	} else {
		return false
	}
}

func (tk *Tokenizer) _createToken(kind string) *Token {
	val := tk._buf[tk._prevPos.Offset:tk._pos.Offset]
	token := NewToken(kind, tk._prevPos, tk._pos, val)
	tk._prevPos = tk._pos
	return token
}

// newline:
//
//	| '\r\n'
//	| '\n'
//	| '\r'
func (tk *Tokenizer) newline() bool {
	// '\r\n'
	if tk._expectS("\r\n") {
		return true
	}
	// '\n'
	if tk._expectS("\n") {
		return true
	}
	// '\r'
	if tk._expectS("\r") {
		return true
	}
	return false
}

// _any_but_eol:
//
//	| !newline _any_but_eof
func (tk *Tokenizer) _anyButEol() bool {
	// !newline _any_but_eof
	_p := tk._mark()
	_ok := false
	if tk.newline() {
		_ok = true
	}
	tk._reset(_p)
	if !_ok {
		if tk._anyButEof() {
			return true
		}
	}
	return false
}

// _whitespace_ch:
//
//	| [ \t\f\u1680\u180E\u2000-\u200A\u202F\u205F\u3000\uFEFF\u00A0]
func (tk *Tokenizer) _whitespaceCh() bool {
	// [ \t\f\u1680\u180E\u2000-\u200A\u202F\u205F\u3000\uFEFF\u00A0]
	if tk._expect(0x20) || tk._expect(0x9) || tk._expect(0xC) || tk._expect(0x1680) || tk._expect(0x180E) || tk._expectR(0x2000, 0x200A) || tk._expect(0x202F) || tk._expect(0x205F) || tk._expect(0x3000) || tk._expect(0xFEFF) || tk._expect(0xA0) {
		return true
	}
	return false
}

// whitespace:
//
//	| _whitespace_ch+
func (tk *Tokenizer) whitespace() bool {
	// _whitespace_ch+
	if tk._whitespaceCh() {
		for {
			_ok := false
			if tk._whitespaceCh() {
				_ok = true
			}
			if !_ok {
				break
			}
		}
		return true
	}
	return false
}

func (tk *Tokenizer) op() string {
	entered := false
	kind := TokenTypeDummy
	switch tk._lookahead {
	case '!':
		entered = true
		tk._forward()
		if tk._lookahead == '=' {
			tk._forward()
			kind = TokenTypeOpNotEqual
			break
		}
		kind = TokenTypeOpNot
	case '%':
		entered = true
		tk._forward()
		if tk._lookahead == '=' {
			tk._forward()
			kind = TokenTypeOpPercentEqual
			break
		}
		kind = TokenTypeOpPercent
	case '&':
		entered = true
		tk._forward()
		if tk._lookahead == '&' {
			tk._forward()
			kind = TokenTypeOpAndAnd
			break
		}
		if tk._lookahead == '=' {
			tk._forward()
			kind = TokenTypeOpAndEqual
			break
		}
		if tk._lookahead == '^' {
			tk._forward()
			if tk._lookahead == '=' {
				tk._forward()
				kind = TokenTypeOpAndCaretEqual
				break
			}
			kind = TokenTypeOpAndCaret
			break
		}
		kind = TokenTypeOpAnd
	case '(':
		entered = true
		tk._forward()
		kind = TokenTypeOpLeftParen
	case ')':
		entered = true
		tk._forward()
		kind = TokenTypeOpRightParen
	case '*':
		entered = true
		tk._forward()
		if tk._lookahead == '=' {
			tk._forward()
			kind = TokenTypeOpStarEqual
			break
		}
		kind = TokenTypeOpStar
	case '+':
		entered = true
		tk._forward()
		if tk._lookahead == '+' {
			tk._forward()
			kind = TokenTypeOpPlusPlus
			break
		}
		if tk._lookahead == '=' {
			tk._forward()
			kind = TokenTypeOpPlusEqual
			break
		}
		kind = TokenTypeOpPlus
	case ',':
		entered = true
		tk._forward()
		kind = TokenTypeOpComma
	case '-':
		entered = true
		tk._forward()
		if tk._lookahead == '-' {
			tk._forward()
			kind = TokenTypeOpMinusMinus
			break
		}
		if tk._lookahead == '=' {
			tk._forward()
			kind = TokenTypeOpMinusEqual
			break
		}
		kind = TokenTypeOpMinus
	case '.':
		entered = true
		tk._forward()
		if tk._lookahead == '.' {
			_p := tk._mark()
			tk._forward()
			if tk._lookahead == '.' {
				tk._forward()
				kind = TokenTypeOpDotDotDot
				break
			}
			tk._reset(_p)
		}
		kind = TokenTypeOpDot
	case '/':
		entered = true
		tk._forward()
		if tk._lookahead == '=' {
			tk._forward()
			kind = TokenTypeOpSlashEqual
			break
		}
		kind = TokenTypeOpSlash
	case ':':
		entered = true
		tk._forward()
		if tk._lookahead == '=' {
			tk._forward()
			kind = TokenTypeOpColonEqual
			break
		}
		kind = TokenTypeOpColon
	case ';':
		entered = true
		tk._forward()
		kind = TokenTypeOpSemi
	case '<':
		entered = true
		tk._forward()
		if tk._lookahead == '-' {
			tk._forward()
			kind = TokenTypeOpLessMinus
			break
		}
		if tk._lookahead == '<' {
			tk._forward()
			if tk._lookahead == '=' {
				tk._forward()
				kind = TokenTypeOpLessLessEqual
				break
			}
			kind = TokenTypeOpLessLess
			break
		}
		if tk._lookahead == '=' {
			tk._forward()
			kind = TokenTypeOpLessEqual
			break
		}
		kind = TokenTypeOpLess
	case '=':
		entered = true
		tk._forward()
		if tk._lookahead == '=' {
			tk._forward()
			kind = TokenTypeOpEqualEqual
			break
		}
		kind = TokenTypeOpEqual
	case '>':
		entered = true
		tk._forward()
		if tk._lookahead == '=' {
			tk._forward()
			kind = TokenTypeOpGreaterEqual
			break
		}
		if tk._lookahead == '>' {
			tk._forward()
			if tk._lookahead == '=' {
				tk._forward()
				kind = TokenTypeOpGreaterGreaterEqual
				break
			}
			kind = TokenTypeOpGreaterGreater
			break
		}
		kind = TokenTypeOpGreater
	case '[':
		entered = true
		tk._forward()
		kind = TokenTypeOpLeftBracket
	case ']':
		entered = true
		tk._forward()
		kind = TokenTypeOpRightBracket
	case '^':
		entered = true
		tk._forward()
		if tk._lookahead == '=' {
			tk._forward()
			kind = TokenTypeOpCaretEqual
			break
		}
		kind = TokenTypeOpCaret
	case '{':
		entered = true
		tk._forward()
		kind = TokenTypeOpLeftBrace
	case '|':
		entered = true
		tk._forward()
		if tk._lookahead == '=' {
			tk._forward()
			kind = TokenTypeOpBarEqual
			break
		}
		if tk._lookahead == '|' {
			tk._forward()
			kind = TokenTypeOpBarBar
			break
		}
		kind = TokenTypeOpBar
	case '}':
		entered = true
		tk._forward()
		kind = TokenTypeOpRightBrace
	case '~':
		entered = true
		tk._forward()
		kind = TokenTypeOpTilde
	default:
		break
	}
	if entered && kind == TokenTypeDummy {
		panic(tk._errorMsg("op"))
	} else {
		return kind
	}
}

func (tk *Tokenizer) next() (*Token, error) {
	kind := TokenTypeDummy
	if tk._lookahead == '\x00' {
		if tk._pos.Offset > tk._bufSize {
			panic(tk._errorMsg("eof"))
		}
		tk._stepForward('\x00')
		kind = TokenTypeEndOfFile
	} else if tk.whitespace() {
		kind = TokenTypeWhitespace
	} else if tk.newline() {
		kind = TokenTypeNewline
	} else if tk.comment() {
		kind = TokenTypeComment
	} else if tk.ident() {
		kind = TokenTypeIdent
	} else if tk.string_() {
		kind = TokenTypeString
	} else if tk.number() {
		kind = TokenTypeNumber
	} else {
		kind = tk.op()
		if kind == TokenTypeDummy {
			return nil, errors.New(tk._errorMsg(string(tk._buf[tk._prevPos.Offset])))
		}
	}

	var val []rune
	if kind == TokenTypeEndOfFile {
		val = []rune("END_OF_FILE")
	} else {
		val = tk._buf[tk._prevPos.Offset:tk._pos.Offset]
	}
	if kind == TokenTypeIdent {
		k, ok := tk._keywords[string(val)]
		if ok {
			kind = k
		}
	}
	ret := NewToken(kind, tk._prevPos, tk._pos, val)
	tk._prevPos = tk._pos
	return ret, nil
}

func (tk *Tokenizer) initKeywords() {
	tk._keywords = make(map[string]string)
	tk._keywords["break"] = TokenTypeKwBreak
	tk._keywords["case"] = TokenTypeKwCase
	tk._keywords["chan"] = TokenTypeKwChan
	tk._keywords["const"] = TokenTypeKwConst
	tk._keywords["continue"] = TokenTypeKwContinue
	tk._keywords["default"] = TokenTypeKwDefault
	tk._keywords["defer"] = TokenTypeKwDefer
	tk._keywords["else"] = TokenTypeKwElse
	tk._keywords["fallthrough"] = TokenTypeKwFallthrough
	tk._keywords["for"] = TokenTypeKwFor
	tk._keywords["func"] = TokenTypeKwFunc
	tk._keywords["go"] = TokenTypeKwGo
	tk._keywords["goto"] = TokenTypeKwGoto
	tk._keywords["if"] = TokenTypeKwIf
	tk._keywords["import"] = TokenTypeKwImport
	tk._keywords["interface"] = TokenTypeKwInterface
	tk._keywords["map"] = TokenTypeKwMap
	tk._keywords["package"] = TokenTypeKwPackage
	tk._keywords["range"] = TokenTypeKwRange
	tk._keywords["return"] = TokenTypeKwReturn
	tk._keywords["select"] = TokenTypeKwSelect
	tk._keywords["struct"] = TokenTypeKwStruct
	tk._keywords["switch"] = TokenTypeKwSwitch
	tk._keywords["type"] = TokenTypeKwType
	tk._keywords["var"] = TokenTypeKwVar
}

// comment:
//
//	| '//' _any_but_eol*
//	| '/*' (!'*/' _any_but_eof)* '*/'
//	_group_1 <-- (!'*/' _any_but_eof)
func (tk *Tokenizer) comment() bool {
	// '//' _any_but_eol*
	_p := tk._mark()
	if tk._expectS("//") {
		for {
			_ok := false
			if tk._anyButEol() {
				_ok = true
			}
			if !_ok {
				break
			}
		}
		return true
	}
	tk._reset(_p)
	// '/*' (!'*/' _any_but_eof)* '*/'
	_p = tk._mark()
	if tk._expectS("/*") {
		for {
			_ok1 := false
			if tk._group1() {
				_ok1 = true
			}
			if !_ok1 {
				break
			}
		}
		if tk._expectS("*/") {
			return true
		}
	}
	tk._reset(_p)
	return false
}

// _ident_ch:
//
//	| !_whitespace_ch [a-zA-Z_\u0080-\uFFFF]
func (tk *Tokenizer) _identCh() bool {
	// !_whitespace_ch [a-zA-Z_\u0080-\uFFFF]
	_p := tk._mark()
	_ok := false
	if tk._whitespaceCh() {
		_ok = true
	}
	tk._reset(_p)
	if !_ok {
		if tk._expectR(0x61, 0x7A) || tk._expectR(0x41, 0x5A) || tk._expect(0x5F) || tk._expectR(0x80, 0xFFFF) {
			return true
		}
	}
	return false
}

// ident:
//
//	| _ident_ch (_ident_ch | [0-9])*
//	_group_2 <-- (_ident_ch | [0-9])
func (tk *Tokenizer) ident() bool {
	// _ident_ch (_ident_ch | [0-9])*
	_p := tk._mark()
	if tk._identCh() {
		for {
			_ok := false
			if tk._group2() {
				_ok = true
			}
			if !_ok {
				break
			}
		}
		return true
	}
	tk._reset(_p)
	return false
}

// string:
//
//	| '`' (!'`' _any_but_eof)* '`'
//	| '"' ('\\' _any_but_eof | !'"' _any_but_eol)* '"'
//	_group_3 <-- (!'`' _any_but_eof)
//	_group_4 <-- ('\\' _any_but_eof | !'"' _any_but_eol)
func (tk *Tokenizer) string_() bool {
	// '`' (!'`' _any_but_eof)* '`'
	_p := tk._mark()
	if tk._expect(0x60) {
		for {
			_ok := false
			if tk._group3() {
				_ok = true
			}
			if !_ok {
				break
			}
		}
		if tk._expect(0x60) {
			return true
		}
	}
	tk._reset(_p)
	// '"' ('\\' _any_but_eof | !'"' _any_but_eol)* '"'
	_p = tk._mark()
	if tk._expect(0x22) {
		for {
			_ok1 := false
			if tk._group4() {
				_ok1 = true
			}
			if !_ok1 {
				break
			}
		}
		if tk._expect(0x22) {
			return true
		}
	}
	tk._reset(_p)
	return false
}

// number:
//
//	| '0' [oObBxX] [_0-9a-fA-F.]+ ([pP] [+-]? [_0-9]+)? 'i'?
//	| '.' [0-9] ('_'? [0-9])* ([eE] [+-]? [0-9] ('_'? [0-9])*)? 'i'?
//	| [0-9] ('_'? [0-9])* '.'? ([0-9] ('_'? [0-9])*)? ([eE] [+-]? [0-9] ('_'? [0-9])*)? 'i'?
//	| '\'' ('\\' _any_but_eof | !'\'' _any_but_eol)+ '\''
//	_group_5 <-- ([pP] [+-]? [_0-9]+)
//	_group_6 <-- ('_'? [0-9])
//	_group_7 <-- ([eE] [+-]? [0-9] ('_'? [0-9])*)
//	_group_6 <-- ('_'? [0-9])
//	_group_8 <-- ([0-9] ('_'? [0-9])*)
//	_group_7 <-- ([eE] [+-]? [0-9] ('_'? [0-9])*)
//	_group_9 <-- ('\\' _any_but_eof | !'\'' _any_but_eol)
func (tk *Tokenizer) number() bool {
	// '0' [oObBxX] [_0-9a-fA-F.]+ ([pP] [+-]? [_0-9]+)? 'i'?
	_p := tk._mark()
	if tk._expect(0x30) {
		if tk._expect(0x6F) || tk._expect(0x4F) || tk._expect(0x62) || tk._expect(0x42) || tk._expect(0x78) || tk._expect(0x58) {
			if tk._expect(0x5F) || tk._expectR(0x30, 0x39) || tk._expectR(0x61, 0x66) || tk._expectR(0x41, 0x46) || tk._expect(0x2E) {
				for {
					_ok := false
					if tk._expect(0x5F) || tk._expectR(0x30, 0x39) || tk._expectR(0x61, 0x66) || tk._expectR(0x41, 0x46) || tk._expect(0x2E) {
						_ok = true
					}
					if !_ok {
						break
					}
				}
				if tk._group5() {
				}
				if tk._expect(0x69) {
				}
				return true
			}
		}
	}
	tk._reset(_p)
	// '.' [0-9] ('_'? [0-9])* ([eE] [+-]? [0-9] ('_'? [0-9])*)? 'i'?
	_p = tk._mark()
	if tk._expect(0x2E) {
		if tk._expectR(0x30, 0x39) {
			for {
				_ok1 := false
				if tk._group6() {
					_ok1 = true
				}
				if !_ok1 {
					break
				}
			}
			if tk._group7() {
			}
			if tk._expect(0x69) {
			}
			return true
		}
	}
	tk._reset(_p)
	// [0-9] ('_'? [0-9])* '.'? ([0-9] ('_'? [0-9])*)? ([eE] [+-]? [0-9] ('_'? [0-9])*)? 'i'?
	_p = tk._mark()
	if tk._expectR(0x30, 0x39) {
		for {
			_ok2 := false
			if tk._group6() {
				_ok2 = true
			}
			if !_ok2 {
				break
			}
		}
		if tk._expect(0x2E) {
		}
		if tk._group8() {
		}
		if tk._group7() {
		}
		if tk._expect(0x69) {
		}
		return true
	}
	tk._reset(_p)
	// '\'' ('\\' _any_but_eof | !'\'' _any_but_eol)+ '\''
	_p = tk._mark()
	if tk._expectS("'") {
		if tk._group9() {
			for {
				_ok3 := false
				if tk._group9() {
					_ok3 = true
				}
				if !_ok3 {
					break
				}
			}
			if tk._expectS("'") {
				return true
			}
		}
	}
	tk._reset(_p)
	return false
}

// _group_1:
//
//	| !'*/' _any_but_eof
func (tk *Tokenizer) _group1() bool {
	// !'*/' _any_but_eof
	_p := tk._mark()
	_ok := false
	if tk._expectS("*/") {
		_ok = true
	}
	tk._reset(_p)
	if !_ok {
		if tk._anyButEof() {
			return true
		}
	}
	return false
}

// _group_2:
//
//	| _ident_ch
//	| [0-9]
func (tk *Tokenizer) _group2() bool {
	// _ident_ch
	if tk._identCh() {
		return true
	}
	// [0-9]
	if tk._expectR(0x30, 0x39) {
		return true
	}
	return false
}

// _group_3:
//
//	| !'`' _any_but_eof
func (tk *Tokenizer) _group3() bool {
	// !'`' _any_but_eof
	_p := tk._mark()
	_ok := false
	if tk._expect(0x60) {
		_ok = true
	}
	tk._reset(_p)
	if !_ok {
		if tk._anyButEof() {
			return true
		}
	}
	return false
}

// _group_4:
//
//	| '\\' _any_but_eof
//	| !'"' _any_but_eol
func (tk *Tokenizer) _group4() bool {
	// '\\' _any_but_eof
	_p := tk._mark()
	if tk._expectS("\\") {
		if tk._anyButEof() {
			return true
		}
	}
	tk._reset(_p)
	// !'"' _any_but_eol
	_p1 := tk._mark()
	_ok := false
	if tk._expect(0x22) {
		_ok = true
	}
	tk._reset(_p1)
	if !_ok {
		if tk._anyButEol() {
			return true
		}
	}
	return false
}

// _group_5:
//
//	| [pP] [+-]? [_0-9]+
func (tk *Tokenizer) _group5() bool {
	// [pP] [+-]? [_0-9]+
	_p := tk._mark()
	if tk._expect(0x70) || tk._expect(0x50) {
		if tk._expect(0x2B) || tk._expect(0x2D) {
		}
		if tk._expect(0x5F) || tk._expectR(0x30, 0x39) {
			for {
				_ok := false
				if tk._expect(0x5F) || tk._expectR(0x30, 0x39) {
					_ok = true
				}
				if !_ok {
					break
				}
			}
			return true
		}
	}
	tk._reset(_p)
	return false
}

// _group_6:
//
//	| '_'? [0-9]
func (tk *Tokenizer) _group6() bool {
	// '_'? [0-9]
	_p := tk._mark()
	if tk._expect(0x5F) {
	}
	if tk._expectR(0x30, 0x39) {
		return true
	}
	tk._reset(_p)
	return false
}

// _group_7:
//
//	| [eE] [+-]? [0-9] ('_'? [0-9])*
//	_group_6 <-- ('_'? [0-9])
func (tk *Tokenizer) _group7() bool {
	// [eE] [+-]? [0-9] ('_'? [0-9])*
	_p := tk._mark()
	if tk._expect(0x65) || tk._expect(0x45) {
		if tk._expect(0x2B) || tk._expect(0x2D) {
		}
		if tk._expectR(0x30, 0x39) {
			for {
				_ok := false
				if tk._group6() {
					_ok = true
				}
				if !_ok {
					break
				}
			}
			return true
		}
	}
	tk._reset(_p)
	return false
}

// _group_8:
//
//	| [0-9] ('_'? [0-9])*
//	_group_6 <-- ('_'? [0-9])
func (tk *Tokenizer) _group8() bool {
	// [0-9] ('_'? [0-9])*
	_p := tk._mark()
	if tk._expectR(0x30, 0x39) {
		for {
			_ok := false
			if tk._group6() {
				_ok = true
			}
			if !_ok {
				break
			}
		}
		return true
	}
	tk._reset(_p)
	return false
}

// _group_9:
//
//	| '\\' _any_but_eof
//	| !'\'' _any_but_eol
func (tk *Tokenizer) _group9() bool {
	// '\\' _any_but_eof
	_p := tk._mark()
	if tk._expectS("\\") {
		if tk._anyButEof() {
			return true
		}
	}
	tk._reset(_p)
	// !'\'' _any_but_eol
	_p1 := tk._mark()
	_ok := false
	if tk._expectS("'") {
		_ok = true
	}
	tk._reset(_p1)
	if !_ok {
		if tk._anyButEol() {
			return true
		}
	}
	return false
}

const expressionMemoId = 0

type NodeCache struct {
	val Node
	pos int
}

type Parser struct {
	_filePath    string
	_fileContent []rune

	_tokens []*Token
	_max    int
	_pos    int
	_x      int

	_bracketDepth  int
	_bracketDepths []int

	_nodeCache []map[int]*NodeCache

	_any any
}

func NewParser(filePath string, fileContent []rune, tokens []*Token) *Parser {
	ps := Parser{_filePath: filePath, _fileContent: fileContent, _tokens: tokens}
	ps._max = len(ps._tokens)
	ps._pos = 0
	ps._x = 0

	ps._bracketDepths = make([]int, ps._max+1)
	ps._nodeCache = make([]map[int]*NodeCache, ps._max)

	return &ps
}

func (ps *Parser) _mark() int {
	ps._bracketDepths[ps._pos] = ps._bracketDepth
	return ps._pos
}

func (ps *Parser) _reset(pos int) {
	ps._pos = pos
	ps._bracketDepth = ps._bracketDepths[ps._pos]
}

func (ps *Parser) _stepForward(tok *Token) {
	if len(tok.Value) == 1 {
		val := tok.Value[0]
		if val == '(' || val == '[' || val == '{' {
			ps._bracketDepth++
		} else if val == ')' || val == ']' || val == '}' {
			ps._bracketDepth--
		}
	}
	ps._pos++
	if ps._pos >= ps._max {
		ps._pos = ps._max - 1
	}
	if ps._pos > ps._x {
		ps._x = ps._pos
	}
}

func (ps *Parser) _expectK(kind string) Node {
	tok := ps._tokens[ps._pos]
	if tok.Kind == kind {
		ps._stepForward(tok)
		return NewTokenNode(ps._filePath, ps._fileContent, tok)
	}
	return nil
}

func (ps *Parser) _expectV(val string) Node {
	tok := ps._tokens[ps._pos]
	if len(tok.Value) == len(val) && string(tok.Value) == val {
		ps._stepForward(tok)
		return NewTokenNode(ps._filePath, ps._fileContent, tok)
	}
	return nil
}

func (ps *Parser) _anyToken() Node {
	tok := ps._tokens[ps._pos]
	ps._stepForward(tok)
	return NewTokenNode(ps._filePath, ps._fileContent, tok)
}

func (ps *Parser) _pseudoToken(v ...any) Node {
	var start, end *Position
	for _, t := range v {
		switch vv := t.(type) {
		case *Token:
			if vv == nil {
				continue
			}
			if start == nil {
				start = &vv.Start
			}
			end = &vv.End
		case []*Token:
			if vv == nil {
				continue
			}
			if len(vv) > 0 {
				if start == nil {
					start = &vv[0].Start
				}
				end = &vv[len(vv)-1].End
			}
		case Node:
			if vv == nil {
				continue
			}
			if start == nil {
				p := vv.RangeStart()
				start = &p
			}
			p := vv.RangeEnd()
			end = &p
		case []Node:
			if vv == nil {
				continue
			}
			if len(vv) > 0 {
				if start == nil {
					p := vv[0].RangeStart()
					start = &p
				}
				p := vv[len(vv)-1].RangeEnd()
				end = &p
			}
		default:
			return nil
		}
	}
	if start == nil || end == nil {
		return nil
	}
	val := ps._fileContent[start.Offset:end.Offset]
	return NewTokenNode(ps._filePath, ps._fileContent, NewToken(TokenTypePseudo, *start, *end, val))
}

func (ps *Parser) _expectPseudoNewline() Node {
	if ps._pos < 1 || ps._pos >= len(ps._tokens) {
		return nil
	}
	current := ps._tokens[ps._pos-1]
	lookahead := ps._tokens[ps._pos]
	if current.End.LineIdx == lookahead.Start.LineIdx {
		return nil
	}
	return NewTokenNode(ps._filePath, ps._fileContent, lookahead)
}

func (ps *Parser) _visibleTokenBefore(pos int) *Token {
	for i := pos - 1; i >= 0; i-- {
		kind := ps._tokens[i].Kind
		if kind != TokenTypeWhitespace && kind != TokenTypeNewline {
			return ps._tokens[i]
		}
	}
	return nil
}

func (ps *Parser) _mergeNodes(items ...any) Node {
	ret := make([]Node, 0)
	for _, item := range items {
		if item == nil {
			continue
		}
		if n, ok := item.(Node); ok {
			if n != nil && !n.IsDummy() {
				ret = append(ret, n)
			}
		} else if s, ok := item.([]Node); ok {
			if s != nil {
				for _, ss := range s {
					if ss != nil && !ss.IsDummy() {
						ret = append(ret, ss)
					}
				}
			}
		} else {
			panic("misused merge_nodes api")
		}
	}
	return NewNodesNode(ret)
}

func (ps *Parser) Parse() (ret Node, err error) {
	ret = ps.file()
	if ps._expectK(TokenTypeEndOfFile) != nil {
		return ret, nil
	}
	tok := ps._tokens[ps._x]
	return nil, fmt.Errorf("fail to parse: %s\n%s", ps._filePath, errorContext(ps._filePath, ps._fileContent, tok.Start.Offset, tok.Start.LineIdx, tok.Start.CharIdx))
}

/*
file:
| n=package_decl i=import_decl* t=top_level_decl_semi* END_OF_FILE {file(n, i, t)}
*/
func (ps *Parser) file() Node {
	/* n=package_decl i=import_decl* t=top_level_decl_semi* END_OF_FILE {file(n, i, t)}
	 */
	pos := ps._mark()
	for {
		var i Node
		var n Node
		var t Node
		n = ps.packageDecl()
		if n == nil {
			break
		}
		_1 := make([]Node, 0)
		var _2 Node
		for {
			_2 = ps.importDecl()
			if _2 == nil {
				break
			}
			_1 = append(_1, _2)
		}
		i = NewNodesNode(_1)
		_ = i
		_3 := make([]Node, 0)
		var _4 Node
		for {
			_4 = ps.topLevelDeclSemi()
			if _4 == nil {
				break
			}
			_3 = append(_3, _4)
		}
		t = NewNodesNode(_3)
		_ = t
		var _5 Node
		_5 = ps._expectK(TokenTypeEndOfFile)
		if _5 == nil {
			break
		}
		return NewFileNode(ps._filePath, ps._fileContent, n, i, t, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
package_decl:
| 'package' n=package_ident ';' {n}
*/
func (ps *Parser) packageDecl() Node {
	/* 'package' n=package_ident ';' {n}
	 */
	pos := ps._mark()
	for {
		var n Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwPackage)
		if _1 == nil {
			break
		}
		n = ps.packageIdent()
		if n == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpSemi)
		if _2 == nil {
			break
		}
		return n
	}
	ps._reset(pos)
	return nil
}

/*
package_ident:
| n=IDENT {package_ident(n)}
*/
func (ps *Parser) packageIdent() Node {
	/* n=IDENT {package_ident(n)}
	 */
	pos := ps._mark()
	for {
		var n Node
		n = ps._expectK(TokenTypeIdent)
		if n == nil {
			break
		}
		return NewPackageIdentNode(ps._filePath, ps._fileContent, n, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
import_decl:
| 'import' '(' specs=(t=import_spec pseudo_semi {t})* ')' ';'? {import_decl(specs)}
| 'import' specs=import_spec ';'? {import_decl([specs])}
_group_1 <-- (t=import_spec pseudo_semi {t})
*/
func (ps *Parser) importDecl() Node {
	/* 'import' '(' specs=(t=import_spec pseudo_semi {t})* ')' ';'? {import_decl(specs)}
	 */
	pos := ps._mark()
	for {
		var specs Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwImport)
		if _1 == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpLeftParen)
		if _2 == nil {
			break
		}
		_3 := make([]Node, 0)
		var _4 Node
		for {
			_4 = ps._group1()
			if _4 == nil {
				break
			}
			_3 = append(_3, _4)
		}
		specs = NewNodesNode(_3)
		_ = specs
		var _5 Node
		_5 = ps._expectK(TokenTypeOpRightParen)
		if _5 == nil {
			break
		}
		var _6 Node
		_6 = ps._expectK(TokenTypeOpSemi)
		_ = _6
		return NewImportDeclNode(ps._filePath, ps._fileContent, specs, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* 'import' specs=import_spec ';'? {import_decl([specs])}
	 */
	for {
		var specs Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwImport)
		if _1 == nil {
			break
		}
		specs = ps.importSpec()
		if specs == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpSemi)
		_ = _2
		return NewImportDeclNode(ps._filePath, ps._fileContent, NewNodesNode([]Node{specs}), ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
import_spec:
| name=(import_dot | import_ident)? path=import_path {import_spec(name, path)}
_group_2 <-- (import_dot | import_ident)
*/
func (ps *Parser) importSpec() Node {
	/* name=(import_dot | import_ident)? path=import_path {import_spec(name, path)}
	 */
	pos := ps._mark()
	for {
		var name Node
		var path Node
		name = ps._group2()
		_ = name
		path = ps.importPath()
		if path == nil {
			break
		}
		return NewImportSpecNode(ps._filePath, ps._fileContent, name, path, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
import_dot:
| d='.' {import_dot(d)}
*/
func (ps *Parser) importDot() Node {
	/* d='.' {import_dot(d)}
	 */
	pos := ps._mark()
	for {
		var d Node
		d = ps._expectK(TokenTypeOpDot)
		if d == nil {
			break
		}
		return NewImportDotNode(ps._filePath, ps._fileContent, d, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
import_ident:
| n=IDENT {import_ident(n)}
*/
func (ps *Parser) importIdent() Node {
	/* n=IDENT {import_ident(n)}
	 */
	pos := ps._mark()
	for {
		var n Node
		n = ps._expectK(TokenTypeIdent)
		if n == nil {
			break
		}
		return NewImportIdentNode(ps._filePath, ps._fileContent, n, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
import_path:
| s=STRING {import_path(s)}
*/
func (ps *Parser) importPath() Node {
	/* s=STRING {import_path(s)}
	 */
	pos := ps._mark()
	for {
		var s Node
		s = ps._expectK(TokenTypeString)
		if s == nil {
			break
		}
		return NewImportPathNode(ps._filePath, ps._fileContent, s, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
top_level_decl_semi:
| t=top_level_decl ';'? {t}
*/
func (ps *Parser) topLevelDeclSemi() Node {
	/* t=top_level_decl ';'? {t}
	 */
	pos := ps._mark()
	for {
		var t Node
		t = ps.topLevelDecl()
		if t == nil {
			break
		}
		var _1 Node
		_1 = ps._expectK(TokenTypeOpSemi)
		_ = _1
		return t
	}
	ps._reset(pos)
	return nil
}

/*
top_level_decl:
| function_decl
| method_decl
| const_decl
| var_decl
| type_decl
*/
func (ps *Parser) topLevelDecl() Node {
	/* function_decl
	 */
	for {
		var _1 Node
		_1 = ps.functionDecl()
		if _1 == nil {
			break
		}
		return _1
	}
	/* method_decl
	 */
	for {
		var _1 Node
		_1 = ps.methodDecl()
		if _1 == nil {
			break
		}
		return _1
	}
	/* const_decl
	 */
	for {
		var _1 Node
		_1 = ps.constDecl()
		if _1 == nil {
			break
		}
		return _1
	}
	/* var_decl
	 */
	for {
		var _1 Node
		_1 = ps.varDecl()
		if _1 == nil {
			break
		}
		return _1
	}
	/* type_decl
	 */
	for {
		var _1 Node
		_1 = ps.typeDecl()
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
function_decl:
| 'func' n=function_ident t=function_generic_parameters? p=function_parameters r=function_results? b=block? {function_decl(n, t, p, r, b)}
*/
func (ps *Parser) functionDecl() Node {
	/* 'func' n=function_ident t=function_generic_parameters? p=function_parameters r=function_results? b=block? {function_decl(n, t, p, r, b)}
	 */
	pos := ps._mark()
	for {
		var b Node
		var n Node
		var p Node
		var r Node
		var t Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwFunc)
		if _1 == nil {
			break
		}
		n = ps.functionIdent()
		if n == nil {
			break
		}
		t = ps.functionGenericParameters()
		_ = t
		p = ps.functionParameters()
		if p == nil {
			break
		}
		r = ps.functionResults()
		_ = r
		b = ps.block()
		_ = b
		return NewFunctionDeclNode(ps._filePath, ps._fileContent, n, t, p, r, b, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
function_ident:
| n=IDENT {function_ident(n)}
*/
func (ps *Parser) functionIdent() Node {
	/* n=IDENT {function_ident(n)}
	 */
	pos := ps._mark()
	for {
		var n Node
		n = ps._expectK(TokenTypeIdent)
		if n == nil {
			break
		}
		return NewFunctionIdentNode(ps._filePath, ps._fileContent, n, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
function_parameters:
| '(' p=','.parameter* ','? ')' {p}
*/
func (ps *Parser) functionParameters() Node {
	/* '(' p=','.parameter* ','? ')' {p}
	 */
	pos := ps._mark()
	for {
		var p Node
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftParen)
		if _1 == nil {
			break
		}
		_2 := make([]Node, 0)
		var _3 Node
		var _4 Node
		_3 = ps.parameter()
		if _3 != nil {
			_2 = append(_2, _3)
			for {
				_p := ps._mark()
				_4 = ps._expectK(TokenTypeOpComma)
				if _4 == nil {
					break
				}
				_3 = ps.parameter()
				if _3 == nil {
					ps._reset(_p)
					break
				}
				_2 = append(_2, _3)
			}
		}
		p = NewNodesNode(_2)
		_ = p
		var _5 Node
		_5 = ps._expectK(TokenTypeOpComma)
		_ = _5
		var _6 Node
		_6 = ps._expectK(TokenTypeOpRightParen)
		if _6 == nil {
			break
		}
		return p
	}
	ps._reset(pos)
	return nil
}

/*
parameter:
| n=parameter_ident '...' t=type {ellipsis_parameter(n, t)}
| n=parameter_ident t=type? {parameter(n, t)}
*/
func (ps *Parser) parameter() Node {
	/* n=parameter_ident '...' t=type {ellipsis_parameter(n, t)}
	 */
	pos := ps._mark()
	for {
		var n Node
		var t Node
		n = ps.parameterIdent()
		if n == nil {
			break
		}
		var _1 Node
		_1 = ps._expectK(TokenTypeOpDotDotDot)
		if _1 == nil {
			break
		}
		t = ps.type_()
		if t == nil {
			break
		}
		return NewEllipsisParameterNode(ps._filePath, ps._fileContent, n, t, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* n=parameter_ident t=type? {parameter(n, t)}
	 */
	for {
		var n Node
		var t Node
		n = ps.parameterIdent()
		if n == nil {
			break
		}
		t = ps.type_()
		_ = t
		return NewParameterNode(ps._filePath, ps._fileContent, n, t, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
parameter_ident:
| n=IDENT {parameter_ident(n)}
*/
func (ps *Parser) parameterIdent() Node {
	/* n=IDENT {parameter_ident(n)}
	 */
	pos := ps._mark()
	for {
		var n Node
		n = ps._expectK(TokenTypeIdent)
		if n == nil {
			break
		}
		return NewParameterIdentNode(ps._filePath, ps._fileContent, n, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
function_results:
| '(' s=','.function_result+ ','? ')' {s}
| r=type_only_function_result {[r]}
*/
func (ps *Parser) functionResults() Node {
	/* '(' s=','.function_result+ ','? ')' {s}
	 */
	pos := ps._mark()
	for {
		var s Node
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftParen)
		if _1 == nil {
			break
		}
		_2 := make([]Node, 0)
		var _3, _4 Node
		_3 = ps.functionResult()
		if _3 == nil {
			break
		}
		_2 = append(_2, _3)
		for {
			_p := ps._mark()
			_4 = ps._expectK(TokenTypeOpComma)
			if _4 == nil {
				break
			}
			_3 = ps.functionResult()
			if _3 == nil {
				ps._reset(_p)
				break
			}
			_2 = append(_2, _3)
		}
		s = NewNodesNode(_2)
		_ = s
		var _5 Node
		_5 = ps._expectK(TokenTypeOpComma)
		_ = _5
		var _6 Node
		_6 = ps._expectK(TokenTypeOpRightParen)
		if _6 == nil {
			break
		}
		return s
	}
	ps._reset(pos)
	/* r=type_only_function_result {[r]}
	 */
	for {
		var r Node
		r = ps.typeOnlyFunctionResult()
		if r == nil {
			break
		}
		return NewNodesNode([]Node{r})
	}
	ps._reset(pos)
	return nil
}

/*
function_result:
| n=function_result_ident t=type? {function_result(n, t)}
| type_only_function_result
*/
func (ps *Parser) functionResult() Node {
	/* n=function_result_ident t=type? {function_result(n, t)}
	 */
	pos := ps._mark()
	for {
		var n Node
		var t Node
		n = ps.functionResultIdent()
		if n == nil {
			break
		}
		t = ps.type_()
		_ = t
		return NewFunctionResultNode(ps._filePath, ps._fileContent, n, t, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* type_only_function_result
	 */
	for {
		var _1 Node
		_1 = ps.typeOnlyFunctionResult()
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
type_only_function_result:
| t=type {function_result(_, t)}
*/
func (ps *Parser) typeOnlyFunctionResult() Node {
	/* t=type {function_result(_, t)}
	 */
	pos := ps._mark()
	for {
		var t Node
		t = ps.type_()
		if t == nil {
			break
		}
		return NewFunctionResultNode(ps._filePath, ps._fileContent, nil, t, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
function_result_ident:
| n=IDENT {function_result_ident(n)}
*/
func (ps *Parser) functionResultIdent() Node {
	/* n=IDENT {function_result_ident(n)}
	 */
	pos := ps._mark()
	for {
		var n Node
		n = ps._expectK(TokenTypeIdent)
		if n == nil {
			break
		}
		return NewFunctionResultIdentNode(ps._filePath, ps._fileContent, n, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
method_decl:
| 'func' '(' rc=receiver ')' n=method_ident '(' p=','.parameter* ','? ')' rs=result_decl? b=block? {method_decl(rc, n, p, rs, b)}
*/
func (ps *Parser) methodDecl() Node {
	/* 'func' '(' rc=receiver ')' n=method_ident '(' p=','.parameter* ','? ')' rs=result_decl? b=block? {method_decl(rc, n, p, rs, b)}
	 */
	pos := ps._mark()
	for {
		var b Node
		var n Node
		var p Node
		var rc Node
		var rs Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwFunc)
		if _1 == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpLeftParen)
		if _2 == nil {
			break
		}
		rc = ps.receiver()
		if rc == nil {
			break
		}
		var _3 Node
		_3 = ps._expectK(TokenTypeOpRightParen)
		if _3 == nil {
			break
		}
		n = ps.methodIdent()
		if n == nil {
			break
		}
		var _4 Node
		_4 = ps._expectK(TokenTypeOpLeftParen)
		if _4 == nil {
			break
		}
		_5 := make([]Node, 0)
		var _6 Node
		var _7 Node
		_6 = ps.parameter()
		if _6 != nil {
			_5 = append(_5, _6)
			for {
				_p := ps._mark()
				_7 = ps._expectK(TokenTypeOpComma)
				if _7 == nil {
					break
				}
				_6 = ps.parameter()
				if _6 == nil {
					ps._reset(_p)
					break
				}
				_5 = append(_5, _6)
			}
		}
		p = NewNodesNode(_5)
		_ = p
		var _8 Node
		_8 = ps._expectK(TokenTypeOpComma)
		_ = _8
		var _9 Node
		_9 = ps._expectK(TokenTypeOpRightParen)
		if _9 == nil {
			break
		}
		rs = ps.resultDecl()
		_ = rs
		b = ps.block()
		_ = b
		return NewMethodDeclNode(ps._filePath, ps._fileContent, rc, n, p, rs, b, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
method_ident:
| n=IDENT {method_ident(n)}
*/
func (ps *Parser) methodIdent() Node {
	/* n=IDENT {method_ident(n)}
	 */
	pos := ps._mark()
	for {
		var n Node
		n = ps._expectK(TokenTypeIdent)
		if n == nil {
			break
		}
		return NewMethodIdentNode(ps._filePath, ps._fileContent, n, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
receiver:
| n=receiver_ident? star='*'? t=receiver_type_ident ('[' g=','.receiver_generic_type_ident+ ','? ']')? {receiver(n, star, t, g)}
*/
func (ps *Parser) receiver() Node {
	/* n=receiver_ident? star='*'? t=receiver_type_ident ('[' g=','.receiver_generic_type_ident+ ','? ']')? {receiver(n, star, t, g)}
	 */
	pos := ps._mark()
	for {
		var g Node
		var n Node
		var star Node
		var t Node
		n = ps.receiverIdent()
		_ = n
		star = ps._expectK(TokenTypeOpStar)
		_ = star
		t = ps.receiverTypeIdent()
		if t == nil {
			break
		}
		var _1 Node
		for {
			_ok := false
			_p := ps._mark()
			for {
				var _2 Node
				_2 = ps._expectK(TokenTypeOpLeftBracket)
				if _2 == nil {
					break
				}
				_3 := make([]Node, 0)
				var _4, _5 Node
				_4 = ps.receiverGenericTypeIdent()
				if _4 == nil {
					break
				}
				_3 = append(_3, _4)
				for {
					_p1 := ps._mark()
					_5 = ps._expectK(TokenTypeOpComma)
					if _5 == nil {
						break
					}
					_4 = ps.receiverGenericTypeIdent()
					if _4 == nil {
						ps._reset(_p1)
						break
					}
					_3 = append(_3, _4)
				}
				g = NewNodesNode(_3)
				_ = g
				var _6 Node
				_6 = ps._expectK(TokenTypeOpComma)
				_ = _6
				_1 = ps._expectK(TokenTypeOpRightBracket)
				if _1 == nil {
					break
				}
				_ok = true
				break
			}
			if !_ok {
				ps._reset(_p)
				g = nil
			}
			break
		}
		_ = _1
		return NewReceiverNode(ps._filePath, ps._fileContent, n, star, t, g, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
receiver_ident:
| n=IDENT {receiver_ident(n)}
*/
func (ps *Parser) receiverIdent() Node {
	/* n=IDENT {receiver_ident(n)}
	 */
	pos := ps._mark()
	for {
		var n Node
		n = ps._expectK(TokenTypeIdent)
		if n == nil {
			break
		}
		return NewReceiverIdentNode(ps._filePath, ps._fileContent, n, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
receiver_type_ident:
| n=IDENT {receiver_type_ident(n)}
*/
func (ps *Parser) receiverTypeIdent() Node {
	/* n=IDENT {receiver_type_ident(n)}
	 */
	pos := ps._mark()
	for {
		var n Node
		n = ps._expectK(TokenTypeIdent)
		if n == nil {
			break
		}
		return NewReceiverTypeIdentNode(ps._filePath, ps._fileContent, n, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
receiver_generic_type_ident:
| n=IDENT {receiver_generic_type_ident(n)}
*/
func (ps *Parser) receiverGenericTypeIdent() Node {
	/* n=IDENT {receiver_generic_type_ident(n)}
	 */
	pos := ps._mark()
	for {
		var n Node
		n = ps._expectK(TokenTypeIdent)
		if n == nil {
			break
		}
		return NewReceiverGenericTypeIdentNode(ps._filePath, ps._fileContent, n, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
const_decl:
| 'const' '(' c=const_spec_semi* ')' {const_decl(c)}
| 'const' c=const_spec {const_decl([c])}
*/
func (ps *Parser) constDecl() Node {
	/* 'const' '(' c=const_spec_semi* ')' {const_decl(c)}
	 */
	pos := ps._mark()
	for {
		var c Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwConst)
		if _1 == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpLeftParen)
		if _2 == nil {
			break
		}
		_3 := make([]Node, 0)
		var _4 Node
		for {
			_4 = ps.constSpecSemi()
			if _4 == nil {
				break
			}
			_3 = append(_3, _4)
		}
		c = NewNodesNode(_3)
		_ = c
		var _5 Node
		_5 = ps._expectK(TokenTypeOpRightParen)
		if _5 == nil {
			break
		}
		return NewConstDeclNode(ps._filePath, ps._fileContent, c, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* 'const' c=const_spec {const_decl([c])}
	 */
	for {
		var c Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwConst)
		if _1 == nil {
			break
		}
		c = ps.constSpec()
		if c == nil {
			break
		}
		return NewConstDeclNode(ps._filePath, ps._fileContent, NewNodesNode([]Node{c}), ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
const_spec_semi:
| c=const_spec pseudo_semi {c}
*/
func (ps *Parser) constSpecSemi() Node {
	/* c=const_spec pseudo_semi {c}
	 */
	pos := ps._mark()
	for {
		var c Node
		c = ps.constSpec()
		if c == nil {
			break
		}
		var _1 Node
		_1 = ps.pseudoSemi()
		if _1 == nil {
			break
		}
		return c
	}
	ps._reset(pos)
	return nil
}

/*
const_spec:
| i=','.const_ident+ (t=type? '=' e=expression_list)? {const_spec(i, t, e)}
*/
func (ps *Parser) constSpec() Node {
	/* i=','.const_ident+ (t=type? '=' e=expression_list)? {const_spec(i, t, e)}
	 */
	pos := ps._mark()
	for {
		var e Node
		var i Node
		var t Node
		_1 := make([]Node, 0)
		var _2, _3 Node
		_2 = ps.constIdent()
		if _2 == nil {
			break
		}
		_1 = append(_1, _2)
		for {
			_p := ps._mark()
			_3 = ps._expectK(TokenTypeOpComma)
			if _3 == nil {
				break
			}
			_2 = ps.constIdent()
			if _2 == nil {
				ps._reset(_p)
				break
			}
			_1 = append(_1, _2)
		}
		i = NewNodesNode(_1)
		_ = i
		var _4 Node
		for {
			_ok := false
			_p1 := ps._mark()
			for {
				t = ps.type_()
				_ = t
				var _5 Node
				_5 = ps._expectK(TokenTypeOpEqual)
				if _5 == nil {
					break
				}
				e = ps.expressionList()
				if e == nil {
					break
				}
				_4 = e
				_ok = true
				break
			}
			if !_ok {
				ps._reset(_p1)
				t = nil
			}
			break
		}
		_ = _4
		return NewConstSpecNode(ps._filePath, ps._fileContent, i, t, e, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
const_ident:
| n=IDENT {const_ident(n)}
*/
func (ps *Parser) constIdent() Node {
	/* n=IDENT {const_ident(n)}
	 */
	pos := ps._mark()
	for {
		var n Node
		n = ps._expectK(TokenTypeIdent)
		if n == nil {
			break
		}
		return NewConstIdentNode(ps._filePath, ps._fileContent, n, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
var_decl:
| 'var' '(' x=var_spec_semi* ')' {var_decl(x)}
| 'var' x=var_spec {var_decl([x])}
*/
func (ps *Parser) varDecl() Node {
	/* 'var' '(' x=var_spec_semi* ')' {var_decl(x)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwVar)
		if _1 == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpLeftParen)
		if _2 == nil {
			break
		}
		_3 := make([]Node, 0)
		var _4 Node
		for {
			_4 = ps.varSpecSemi()
			if _4 == nil {
				break
			}
			_3 = append(_3, _4)
		}
		x = NewNodesNode(_3)
		_ = x
		var _5 Node
		_5 = ps._expectK(TokenTypeOpRightParen)
		if _5 == nil {
			break
		}
		return NewVarDeclNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* 'var' x=var_spec {var_decl([x])}
	 */
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwVar)
		if _1 == nil {
			break
		}
		x = ps.varSpec()
		if x == nil {
			break
		}
		return NewVarDeclNode(ps._filePath, ps._fileContent, NewNodesNode([]Node{x}), ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
var_spec_semi:
| x=var_spec pseudo_semi {x}
*/
func (ps *Parser) varSpecSemi() Node {
	/* x=var_spec pseudo_semi {x}
	 */
	pos := ps._mark()
	for {
		var x Node
		x = ps.varSpec()
		if x == nil {
			break
		}
		var _1 Node
		_1 = ps.pseudoSemi()
		if _1 == nil {
			break
		}
		return x
	}
	ps._reset(pos)
	return nil
}

/*
var_spec:
| i=','.var_ident+  (t=type? '=' e=expression_list)? {var_spec(i, t, e)}
*/
func (ps *Parser) varSpec() Node {
	/* i=','.var_ident+ (t=type? '=' e=expression_list)? {var_spec(i, t, e)}
	 */
	pos := ps._mark()
	for {
		var e Node
		var i Node
		var t Node
		_1 := make([]Node, 0)
		var _2, _3 Node
		_2 = ps.varIdent()
		if _2 == nil {
			break
		}
		_1 = append(_1, _2)
		for {
			_p := ps._mark()
			_3 = ps._expectK(TokenTypeOpComma)
			if _3 == nil {
				break
			}
			_2 = ps.varIdent()
			if _2 == nil {
				ps._reset(_p)
				break
			}
			_1 = append(_1, _2)
		}
		i = NewNodesNode(_1)
		_ = i
		var _4 Node
		for {
			_ok := false
			_p1 := ps._mark()
			for {
				t = ps.type_()
				_ = t
				var _5 Node
				_5 = ps._expectK(TokenTypeOpEqual)
				if _5 == nil {
					break
				}
				e = ps.expressionList()
				if e == nil {
					break
				}
				_4 = e
				_ok = true
				break
			}
			if !_ok {
				ps._reset(_p1)
				t = nil
			}
			break
		}
		_ = _4
		return NewVarSpecNode(ps._filePath, ps._fileContent, i, t, e, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
var_ident:
| n=IDENT {var_ident(n)}
*/
func (ps *Parser) varIdent() Node {
	/* n=IDENT {var_ident(n)}
	 */
	pos := ps._mark()
	for {
		var n Node
		n = ps._expectK(TokenTypeIdent)
		if n == nil {
			break
		}
		return NewVarIdentNode(ps._filePath, ps._fileContent, n, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
type_decl:
| 'type' '(' x=type_spec_semi* ')' {type_decl(x)}
| 'type' x=type_spec {type_decl([x])}
*/
func (ps *Parser) typeDecl() Node {
	/* 'type' '(' x=type_spec_semi* ')' {type_decl(x)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwType)
		if _1 == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpLeftParen)
		if _2 == nil {
			break
		}
		_3 := make([]Node, 0)
		var _4 Node
		for {
			_4 = ps.typeSpecSemi()
			if _4 == nil {
				break
			}
			_3 = append(_3, _4)
		}
		x = NewNodesNode(_3)
		_ = x
		var _5 Node
		_5 = ps._expectK(TokenTypeOpRightParen)
		if _5 == nil {
			break
		}
		return NewTypeDeclNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* 'type' x=type_spec {type_decl([x])}
	 */
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwType)
		if _1 == nil {
			break
		}
		x = ps.typeSpec()
		if x == nil {
			break
		}
		return NewTypeDeclNode(ps._filePath, ps._fileContent, NewNodesNode([]Node{x}), ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
type_spec_semi:
| x=type_spec pseudo_semi {x}
*/
func (ps *Parser) typeSpecSemi() Node {
	/* x=type_spec pseudo_semi {x}
	 */
	pos := ps._mark()
	for {
		var x Node
		x = ps.typeSpec()
		if x == nil {
			break
		}
		var _1 Node
		_1 = ps.pseudoSemi()
		if _1 == nil {
			break
		}
		return x
	}
	ps._reset(pos)
	return nil
}

/*
type_spec:
| x=type_ident t=generic_parameter_decl? '=' y=type {type_eq_spec(x, t, y)}
| x=type_ident t=generic_parameter_decl? y=type {type_spec(x, t, y)}
*/
func (ps *Parser) typeSpec() Node {
	/* x=type_ident t=generic_parameter_decl? '=' y=type {type_eq_spec(x, t, y)}
	 */
	pos := ps._mark()
	for {
		var t Node
		var x Node
		var y Node
		x = ps.typeIdent()
		if x == nil {
			break
		}
		t = ps.genericParameterDecl()
		_ = t
		var _1 Node
		_1 = ps._expectK(TokenTypeOpEqual)
		if _1 == nil {
			break
		}
		y = ps.type_()
		if y == nil {
			break
		}
		return NewTypeEqSpecNode(ps._filePath, ps._fileContent, x, t, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* x=type_ident t=generic_parameter_decl? y=type {type_spec(x, t, y)}
	 */
	for {
		var t Node
		var x Node
		var y Node
		x = ps.typeIdent()
		if x == nil {
			break
		}
		t = ps.genericParameterDecl()
		_ = t
		y = ps.type_()
		if y == nil {
			break
		}
		return NewTypeSpecNode(ps._filePath, ps._fileContent, x, t, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
type_ident:
| n=IDENT {type_ident(n)}
*/
func (ps *Parser) typeIdent() Node {
	/* n=IDENT {type_ident(n)}
	 */
	pos := ps._mark()
	for {
		var n Node
		n = ps._expectK(TokenTypeIdent)
		if n == nil {
			break
		}
		return NewTypeIdentNode(ps._filePath, ps._fileContent, n, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
function_generic_parameters:
| generic_parameter_decl
*/
func (ps *Parser) functionGenericParameters() Node {
	/* generic_parameter_decl
	 */
	for {
		var _1 Node
		_1 = ps.genericParameterDecl()
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
generic_parameter_decl:
| '[' x=','.generic_parameter+ ','? ']' {x}
*/
func (ps *Parser) genericParameterDecl() Node {
	/* '[' x=','.generic_parameter+ ','? ']' {x}
	 */
	pos := ps._mark()
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftBracket)
		if _1 == nil {
			break
		}
		_2 := make([]Node, 0)
		var _3, _4 Node
		_3 = ps.genericParameter()
		if _3 == nil {
			break
		}
		_2 = append(_2, _3)
		for {
			_p := ps._mark()
			_4 = ps._expectK(TokenTypeOpComma)
			if _4 == nil {
				break
			}
			_3 = ps.genericParameter()
			if _3 == nil {
				ps._reset(_p)
				break
			}
			_2 = append(_2, _3)
		}
		x = NewNodesNode(_2)
		_ = x
		var _5 Node
		_5 = ps._expectK(TokenTypeOpComma)
		_ = _5
		var _6 Node
		_6 = ps._expectK(TokenTypeOpRightBracket)
		if _6 == nil {
			break
		}
		return x
	}
	ps._reset(pos)
	return nil
}

/*
generic_parameter:
| n=generic_parameter_ident t=generic_union_constraint? {generic_parameter(n, t)}
*/
func (ps *Parser) genericParameter() Node {
	/* n=generic_parameter_ident t=generic_union_constraint? {generic_parameter(n, t)}
	 */
	pos := ps._mark()
	for {
		var n Node
		var t Node
		n = ps.genericParameterIdent()
		if n == nil {
			break
		}
		t = ps.genericUnionConstraint()
		_ = t
		return NewGenericParameterNode(ps._filePath, ps._fileContent, n, t, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
generic_parameter_ident:
| n=IDENT {generic_parameter_ident(n)}
*/
func (ps *Parser) genericParameterIdent() Node {
	/* n=IDENT {generic_parameter_ident(n)}
	 */
	pos := ps._mark()
	for {
		var n Node
		n = ps._expectK(TokenTypeIdent)
		if n == nil {
			break
		}
		return NewGenericParameterIdentNode(ps._filePath, ps._fileContent, n, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
generic_union_constraint:
| x=type_constraint !'|' {x}
| x='|'.type_constraint+ {generic_union_constraint(x)}
*/
func (ps *Parser) genericUnionConstraint() Node {
	/* x=type_constraint !'|' {x}
	 */
	pos := ps._mark()
	for {
		var x Node
		x = ps.typeConstraint()
		if x == nil {
			break
		}
		var _1 Node
		_p := ps._mark()
		_1 = ps._expectK(TokenTypeOpBar)
		if _1 != nil {
			ps._reset(_p)
		}
		if _1 != nil {
			break
		}
		return x
	}
	ps._reset(pos)
	/* x='|'.type_constraint+ {generic_union_constraint(x)}
	 */
	for {
		var x Node
		_1 := make([]Node, 0)
		var _2, _3 Node
		_2 = ps.typeConstraint()
		if _2 == nil {
			break
		}
		_1 = append(_1, _2)
		for {
			_p := ps._mark()
			_3 = ps._expectK(TokenTypeOpBar)
			if _3 == nil {
				break
			}
			_2 = ps.typeConstraint()
			if _2 == nil {
				ps._reset(_p)
				break
			}
			_1 = append(_1, _2)
		}
		x = NewNodesNode(_1)
		_ = x
		return NewGenericUnionConstraintNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
type_constraint:
| '~' x=type {generic_underlying_type_constraint(x)}
| x=type {generic_type_constraint(x)}
*/
func (ps *Parser) typeConstraint() Node {
	/* '~' x=type {generic_underlying_type_constraint(x)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeOpTilde)
		if _1 == nil {
			break
		}
		x = ps.type_()
		if x == nil {
			break
		}
		return NewGenericUnderlyingTypeConstraintNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* x=type {generic_type_constraint(x)}
	 */
	for {
		var x Node
		x = ps.type_()
		if x == nil {
			break
		}
		return NewGenericTypeConstraintNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
block:
| '{' x=statement_semi_list? '}' {block_stmt(x)}
*/
func (ps *Parser) block() Node {
	/* '{' x=statement_semi_list? '}' {block_stmt(x)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftBrace)
		if _1 == nil {
			break
		}
		x = ps.statementSemiList()
		_ = x
		var _2 Node
		_2 = ps._expectK(TokenTypeOpRightBrace)
		if _2 == nil {
			break
		}
		return NewBlockStmtNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
statement_semi_list:
| statement_semi+
*/
func (ps *Parser) statementSemiList() Node {
	/* statement_semi+
	 */
	for {
		var _1 Node
		_2 := make([]Node, 0)
		var _3 Node
		_3 = ps.statementSemi()
		if _3 == nil {
			break
		}
		_2 = append(_2, _3)
		for {
			_3 = ps.statementSemi()
			if _3 == nil {
				break
			}
			_2 = append(_2, _3)
		}
		_1 = NewNodesNode(_2)
		_ = _1
		return _1
	}
	return nil
}

/*
statement_semi:
| x=statement pseudo_semi {x}
*/
func (ps *Parser) statementSemi() Node {
	/* x=statement pseudo_semi {x}
	 */
	pos := ps._mark()
	for {
		var x Node
		x = ps.statement()
		if x == nil {
			break
		}
		var _1 Node
		_1 = ps.pseudoSemi()
		if _1 == nil {
			break
		}
		return x
	}
	ps._reset(pos)
	return nil
}

/*
statement:
| var_decl
| const_decl
| type_decl
| labeled_stmt
| simple_stmt
| if_stmt
| type_switch_stmt
| expr_switch_stmt
| 'select' b=select_body {select_stmt(b)}
| for_stmt
| 'go' x=expression {go_stmt(x)}
| 'return' x=expression_list? {return_stmt(x)}
| x='break' y=IDENT? {branch_stmt(x,y)}
| x='continue' y=IDENT? {branch_stmt(x,y)}
| x='goto' y=IDENT {branch_stmt(x,y)}
| x='fallthrough' {branch_stmt(x,_)}
| 'defer' x=expression {defer_stmt(x)}
| block
*/
func (ps *Parser) statement() Node {
	/* var_decl
	 */
	for {
		var _1 Node
		_1 = ps.varDecl()
		if _1 == nil {
			break
		}
		return _1
	}
	/* const_decl
	 */
	for {
		var _1 Node
		_1 = ps.constDecl()
		if _1 == nil {
			break
		}
		return _1
	}
	/* type_decl
	 */
	for {
		var _1 Node
		_1 = ps.typeDecl()
		if _1 == nil {
			break
		}
		return _1
	}
	/* labeled_stmt
	 */
	for {
		var _1 Node
		_1 = ps.labeledStmt()
		if _1 == nil {
			break
		}
		return _1
	}
	/* simple_stmt
	 */
	for {
		var _1 Node
		_1 = ps.simpleStmt()
		if _1 == nil {
			break
		}
		return _1
	}
	/* if_stmt
	 */
	for {
		var _1 Node
		_1 = ps.ifStmt()
		if _1 == nil {
			break
		}
		return _1
	}
	/* type_switch_stmt
	 */
	for {
		var _1 Node
		_1 = ps.typeSwitchStmt()
		if _1 == nil {
			break
		}
		return _1
	}
	/* expr_switch_stmt
	 */
	for {
		var _1 Node
		_1 = ps.exprSwitchStmt()
		if _1 == nil {
			break
		}
		return _1
	}
	/* 'select' b=select_body {select_stmt(b)}
	 */
	pos := ps._mark()
	for {
		var b Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwSelect)
		if _1 == nil {
			break
		}
		b = ps.selectBody()
		if b == nil {
			break
		}
		return NewSelectStmtNode(ps._filePath, ps._fileContent, b, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* for_stmt
	 */
	for {
		var _1 Node
		_1 = ps.forStmt()
		if _1 == nil {
			break
		}
		return _1
	}
	/* 'go' x=expression {go_stmt(x)}
	 */
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwGo)
		if _1 == nil {
			break
		}
		x = ps.expression()
		if x == nil {
			break
		}
		return NewGoStmtNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* 'return' x=expression_list? {return_stmt(x)}
	 */
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwReturn)
		if _1 == nil {
			break
		}
		x = ps.expressionList()
		_ = x
		return NewReturnStmtNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* x='break' y=IDENT? {branch_stmt(x,y)}
	 */
	for {
		var x Node
		var y Node
		x = ps._expectK(TokenTypeKwBreak)
		if x == nil {
			break
		}
		y = ps._expectK(TokenTypeIdent)
		_ = y
		return NewBranchStmtNode(ps._filePath, ps._fileContent, x, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* x='continue' y=IDENT? {branch_stmt(x,y)}
	 */
	for {
		var x Node
		var y Node
		x = ps._expectK(TokenTypeKwContinue)
		if x == nil {
			break
		}
		y = ps._expectK(TokenTypeIdent)
		_ = y
		return NewBranchStmtNode(ps._filePath, ps._fileContent, x, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* x='goto' y=IDENT {branch_stmt(x,y)}
	 */
	for {
		var x Node
		var y Node
		x = ps._expectK(TokenTypeKwGoto)
		if x == nil {
			break
		}
		y = ps._expectK(TokenTypeIdent)
		if y == nil {
			break
		}
		return NewBranchStmtNode(ps._filePath, ps._fileContent, x, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* x='fallthrough' {branch_stmt(x,_)}
	 */
	for {
		var x Node
		x = ps._expectK(TokenTypeKwFallthrough)
		if x == nil {
			break
		}
		return NewBranchStmtNode(ps._filePath, ps._fileContent, x, nil, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* 'defer' x=expression {defer_stmt(x)}
	 */
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwDefer)
		if _1 == nil {
			break
		}
		x = ps.expression()
		if x == nil {
			break
		}
		return NewDeferStmtNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* block
	 */
	for {
		var _1 Node
		_1 = ps.block()
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
select_body:
| '{' cases=common_clause* '}' {block_stmt(cases)}
*/
func (ps *Parser) selectBody() Node {
	/* '{' cases=common_clause* '}' {block_stmt(cases)}
	 */
	pos := ps._mark()
	for {
		var cases Node
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftBrace)
		if _1 == nil {
			break
		}
		_2 := make([]Node, 0)
		var _3 Node
		for {
			_3 = ps.commonClause()
			if _3 == nil {
				break
			}
			_2 = append(_2, _3)
		}
		cases = NewNodesNode(_2)
		_ = cases
		var _4 Node
		_4 = ps._expectK(TokenTypeOpRightBrace)
		if _4 == nil {
			break
		}
		return NewBlockStmtNode(ps._filePath, ps._fileContent, cases, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
for_stmt:
| 'for' [ c=expression? ] b=block {for_stmt(_,c,_,b)}
| 'for' [ i=simple_stmt? ';' c=expression? ';' post=simple_stmt? ] b=block {for_stmt(i,c,post,b)}
| 'for' [ 'range' x=expression ] b=block {range_stmt(_,_,x,b,_)}
| 'for' [ k=expression ',' v=expression tok=(':='|'=') 'range' x=expression ] b=block {range_stmt(k,v,x,b,tok)}
| 'for' [ k=expression tok=(':='|'=') 'range' x=expression ] b=block {range_stmt(k,_,x,b,tok)}
_group_3 <-- (':='|'=')
_group_3 <-- (':='|'=')
*/
func (ps *Parser) forStmt() Node {
	/* 'for' [ c=expression? ] b=block {for_stmt(_,c,_,b)}
	 */
	pos := ps._mark()
	for {
		var b Node
		var c Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwFor)
		if _1 == nil {
			break
		}
		_break := true
		ps._enter()
		for {
			c = ps.expression()
			_ = c
			_break = false
			break
		}
		ps._leave()
		if _break {
			break
		}
		b = ps.block()
		if b == nil {
			break
		}
		return NewForStmtNode(ps._filePath, ps._fileContent, nil, c, nil, b, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* 'for' [ i=simple_stmt? ';' c=expression? ';' post=simple_stmt? ] b=block {for_stmt(i,c,post,b)}
	 */
	for {
		var b Node
		var c Node
		var i Node
		var post Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwFor)
		if _1 == nil {
			break
		}
		_break := true
		ps._enter()
		for {
			i = ps.simpleStmt()
			_ = i
			var _2 Node
			_2 = ps._expectK(TokenTypeOpSemi)
			if _2 == nil {
				break
			}
			c = ps.expression()
			_ = c
			var _3 Node
			_3 = ps._expectK(TokenTypeOpSemi)
			if _3 == nil {
				break
			}
			post = ps.simpleStmt()
			_ = post
			_break = false
			break
		}
		ps._leave()
		if _break {
			break
		}
		b = ps.block()
		if b == nil {
			break
		}
		return NewForStmtNode(ps._filePath, ps._fileContent, i, c, post, b, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* 'for' [ 'range' x=expression ] b=block {range_stmt(_,_,x,b,_)}
	 */
	for {
		var b Node
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwFor)
		if _1 == nil {
			break
		}
		_break := true
		ps._enter()
		for {
			var _2 Node
			_2 = ps._expectK(TokenTypeKwRange)
			if _2 == nil {
				break
			}
			x = ps.expression()
			if x == nil {
				break
			}
			_break = false
			break
		}
		ps._leave()
		if _break {
			break
		}
		b = ps.block()
		if b == nil {
			break
		}
		return NewRangeStmtNode(ps._filePath, ps._fileContent, nil, nil, x, b, nil, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* 'for' [ k=expression ',' v=expression tok=(':='|'=') 'range' x=expression ] b=block {range_stmt(k,v,x,b,tok)}
	 */
	for {
		var b Node
		var k Node
		var tok Node
		var v Node
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwFor)
		if _1 == nil {
			break
		}
		_break := true
		ps._enter()
		for {
			k = ps.expression()
			if k == nil {
				break
			}
			var _2 Node
			_2 = ps._expectK(TokenTypeOpComma)
			if _2 == nil {
				break
			}
			v = ps.expression()
			if v == nil {
				break
			}
			tok = ps._group3()
			if tok == nil {
				break
			}
			var _3 Node
			_3 = ps._expectK(TokenTypeKwRange)
			if _3 == nil {
				break
			}
			x = ps.expression()
			if x == nil {
				break
			}
			_break = false
			break
		}
		ps._leave()
		if _break {
			break
		}
		b = ps.block()
		if b == nil {
			break
		}
		return NewRangeStmtNode(ps._filePath, ps._fileContent, k, v, x, b, tok, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* 'for' [ k=expression tok=(':='|'=') 'range' x=expression ] b=block {range_stmt(k,_,x,b,tok)}
	 */
	for {
		var b Node
		var k Node
		var tok Node
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwFor)
		if _1 == nil {
			break
		}
		_break := true
		ps._enter()
		for {
			k = ps.expression()
			if k == nil {
				break
			}
			tok = ps._group3()
			if tok == nil {
				break
			}
			var _2 Node
			_2 = ps._expectK(TokenTypeKwRange)
			if _2 == nil {
				break
			}
			x = ps.expression()
			if x == nil {
				break
			}
			_break = false
			break
		}
		ps._leave()
		if _break {
			break
		}
		b = ps.block()
		if b == nil {
			break
		}
		return NewRangeStmtNode(ps._filePath, ps._fileContent, k, nil, x, b, tok, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
common_clause:
| 'case' x=send_stmt ':' y=statement_semi_list? {common_clause(x,y)}
| 'case' x=recv_stmt ':' y=statement_semi_list? {common_clause(x,y)}
| 'default' ':' x=statement_semi_list? {common_clause(_,x)}
*/
func (ps *Parser) commonClause() Node {
	/* 'case' x=send_stmt ':' y=statement_semi_list? {common_clause(x,y)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var y Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwCase)
		if _1 == nil {
			break
		}
		x = ps.sendStmt()
		if x == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpColon)
		if _2 == nil {
			break
		}
		y = ps.statementSemiList()
		_ = y
		return NewCommonClauseNode(ps._filePath, ps._fileContent, x, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* 'case' x=recv_stmt ':' y=statement_semi_list? {common_clause(x,y)}
	 */
	for {
		var x Node
		var y Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwCase)
		if _1 == nil {
			break
		}
		x = ps.recvStmt()
		if x == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpColon)
		if _2 == nil {
			break
		}
		y = ps.statementSemiList()
		_ = y
		return NewCommonClauseNode(ps._filePath, ps._fileContent, x, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* 'default' ':' x=statement_semi_list? {common_clause(_,x)}
	 */
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwDefault)
		if _1 == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpColon)
		if _2 == nil {
			break
		}
		x = ps.statementSemiList()
		_ = x
		return NewCommonClauseNode(ps._filePath, ps._fileContent, nil, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
recv_stmt:
| x=expression_list op='=' y=expression {assign_stmt(x, op, [y])}
| x=identifier_list op=':=' y=expression {assign_stmt(x, op, [y])}
| x=expression {expr_stmt(x)}
*/
func (ps *Parser) recvStmt() Node {
	/* x=expression_list op='=' y=expression {assign_stmt(x, op, [y])}
	 */
	pos := ps._mark()
	for {
		var op Node
		var x Node
		var y Node
		x = ps.expressionList()
		if x == nil {
			break
		}
		op = ps._expectK(TokenTypeOpEqual)
		if op == nil {
			break
		}
		y = ps.expression()
		if y == nil {
			break
		}
		return NewAssignStmtNode(ps._filePath, ps._fileContent, x, op, NewNodesNode([]Node{y}), ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* x=identifier_list op=':=' y=expression {assign_stmt(x, op, [y])}
	 */
	for {
		var op Node
		var x Node
		var y Node
		x = ps.identifierList()
		if x == nil {
			break
		}
		op = ps._expectK(TokenTypeOpColonEqual)
		if op == nil {
			break
		}
		y = ps.expression()
		if y == nil {
			break
		}
		return NewAssignStmtNode(ps._filePath, ps._fileContent, x, op, NewNodesNode([]Node{y}), ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* x=expression {expr_stmt(x)}
	 */
	for {
		var x Node
		x = ps.expression()
		if x == nil {
			break
		}
		return NewExprStmtNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
type_switch_body:
| '{' cases=type_case_clause* '}' {block_stmt(cases)}
*/
func (ps *Parser) typeSwitchBody() Node {
	/* '{' cases=type_case_clause* '}' {block_stmt(cases)}
	 */
	pos := ps._mark()
	for {
		var cases Node
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftBrace)
		if _1 == nil {
			break
		}
		_2 := make([]Node, 0)
		var _3 Node
		for {
			_3 = ps.typeCaseClause()
			if _3 == nil {
				break
			}
			_2 = append(_2, _3)
		}
		cases = NewNodesNode(_2)
		_ = cases
		var _4 Node
		_4 = ps._expectK(TokenTypeOpRightBrace)
		if _4 == nil {
			break
		}
		return NewBlockStmtNode(ps._filePath, ps._fileContent, cases, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
type_switch_stmt:
| 'switch' [ (init=simple_stmt ';')? assign=type_switch_guard ] b=type_switch_body {type_switch_stmt(init,assign,b)}
*/
func (ps *Parser) typeSwitchStmt() Node {
	/* 'switch' [ (init=simple_stmt ';')? assign=type_switch_guard ] b=type_switch_body {type_switch_stmt(init,assign,b)}
	 */
	pos := ps._mark()
	for {
		var assign Node
		var b Node
		var init Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwSwitch)
		if _1 == nil {
			break
		}
		_break := true
		ps._enter()
		for {
			var _2 Node
			for {
				_ok := false
				_p := ps._mark()
				for {
					init = ps.simpleStmt()
					if init == nil {
						break
					}
					_2 = ps._expectK(TokenTypeOpSemi)
					if _2 == nil {
						break
					}
					_ok = true
					break
				}
				if !_ok {
					ps._reset(_p)
					init = nil
				}
				break
			}
			_ = _2
			assign = ps.typeSwitchGuard()
			if assign == nil {
				break
			}
			_break = false
			break
		}
		ps._leave()
		if _break {
			break
		}
		b = ps.typeSwitchBody()
		if b == nil {
			break
		}
		return NewTypeSwitchStmtNode(ps._filePath, ps._fileContent, init, assign, b, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
type_assert_expr:
| r=primary_expr '.' '(' 'type' ')' {type_assert_expr(r,_)}
*/
func (ps *Parser) typeAssertExpr() Node {
	/* r=primary_expr '.' '(' 'type' ')' {type_assert_expr(r,_)}
	 */
	pos := ps._mark()
	for {
		var r Node
		r = ps.primaryExpr()
		if r == nil {
			break
		}
		var _1 Node
		_1 = ps._expectK(TokenTypeOpDot)
		if _1 == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpLeftParen)
		if _2 == nil {
			break
		}
		var _3 Node
		_3 = ps._expectK(TokenTypeKwType)
		if _3 == nil {
			break
		}
		var _4 Node
		_4 = ps._expectK(TokenTypeOpRightParen)
		if _4 == nil {
			break
		}
		return NewTypeAssertExprNode(ps._filePath, ps._fileContent, r, nil, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
type_switch_guard:
| l=IDENT op=':=' t=type_assert_expr {assign_stmt([l], op, [t])}
| t=type_assert_expr {expr_stmt(t)}
*/
func (ps *Parser) typeSwitchGuard() Node {
	/* l=IDENT op=':=' t=type_assert_expr {assign_stmt([l], op, [t])}
	 */
	pos := ps._mark()
	for {
		var l Node
		var op Node
		var t Node
		l = ps._expectK(TokenTypeIdent)
		if l == nil {
			break
		}
		op = ps._expectK(TokenTypeOpColonEqual)
		if op == nil {
			break
		}
		t = ps.typeAssertExpr()
		if t == nil {
			break
		}
		return NewAssignStmtNode(ps._filePath, ps._fileContent, NewNodesNode([]Node{l}), op, NewNodesNode([]Node{t}), ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* t=type_assert_expr {expr_stmt(t)}
	 */
	for {
		var t Node
		t = ps.typeAssertExpr()
		if t == nil {
			break
		}
		return NewExprStmtNode(ps._filePath, ps._fileContent, t, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
type_case_clause:
| 'case' x=','.type+ ':' y=statement_semi_list? {case_clause(x,y)}
| 'default' ':' x=statement_semi_list? {case_clause(_,x)}
*/
func (ps *Parser) typeCaseClause() Node {
	/* 'case' x=','.type+ ':' y=statement_semi_list? {case_clause(x,y)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var y Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwCase)
		if _1 == nil {
			break
		}
		_2 := make([]Node, 0)
		var _3, _4 Node
		_3 = ps.type_()
		if _3 == nil {
			break
		}
		_2 = append(_2, _3)
		for {
			_p := ps._mark()
			_4 = ps._expectK(TokenTypeOpComma)
			if _4 == nil {
				break
			}
			_3 = ps.type_()
			if _3 == nil {
				ps._reset(_p)
				break
			}
			_2 = append(_2, _3)
		}
		x = NewNodesNode(_2)
		_ = x
		var _5 Node
		_5 = ps._expectK(TokenTypeOpColon)
		if _5 == nil {
			break
		}
		y = ps.statementSemiList()
		_ = y
		return NewCaseClauseNode(ps._filePath, ps._fileContent, x, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* 'default' ':' x=statement_semi_list? {case_clause(_,x)}
	 */
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwDefault)
		if _1 == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpColon)
		if _2 == nil {
			break
		}
		x = ps.statementSemiList()
		_ = x
		return NewCaseClauseNode(ps._filePath, ps._fileContent, nil, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
expr_switch_body:
| '{' cases=expr_case_clause* '}' {block_stmt(cases)}
*/
func (ps *Parser) exprSwitchBody() Node {
	/* '{' cases=expr_case_clause* '}' {block_stmt(cases)}
	 */
	pos := ps._mark()
	for {
		var cases Node
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftBrace)
		if _1 == nil {
			break
		}
		_2 := make([]Node, 0)
		var _3 Node
		for {
			_3 = ps.exprCaseClause()
			if _3 == nil {
				break
			}
			_2 = append(_2, _3)
		}
		cases = NewNodesNode(_2)
		_ = cases
		var _4 Node
		_4 = ps._expectK(TokenTypeOpRightBrace)
		if _4 == nil {
			break
		}
		return NewBlockStmtNode(ps._filePath, ps._fileContent, cases, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
expr_switch_stmt:
| 'switch' [ (init=simple_stmt ';')? tag=expression? ]  b=expr_switch_body {switch_stmt(init,tag,b)}
*/
func (ps *Parser) exprSwitchStmt() Node {
	/* 'switch' [ (init=simple_stmt ';')? tag=expression? ] b=expr_switch_body {switch_stmt(init,tag,b)}
	 */
	pos := ps._mark()
	for {
		var b Node
		var init Node
		var tag Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwSwitch)
		if _1 == nil {
			break
		}
		_break := true
		ps._enter()
		for {
			var _2 Node
			for {
				_ok := false
				_p := ps._mark()
				for {
					init = ps.simpleStmt()
					if init == nil {
						break
					}
					_2 = ps._expectK(TokenTypeOpSemi)
					if _2 == nil {
						break
					}
					_ok = true
					break
				}
				if !_ok {
					ps._reset(_p)
					init = nil
				}
				break
			}
			_ = _2
			tag = ps.expression()
			_ = tag
			_break = false
			break
		}
		ps._leave()
		if _break {
			break
		}
		b = ps.exprSwitchBody()
		if b == nil {
			break
		}
		return NewSwitchStmtNode(ps._filePath, ps._fileContent, init, tag, b, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
expr_case_clause:
| 'case' x=expression_list ':' y=statement_semi_list? {case_clause(x,y)}
| 'default' ':' x=statement_semi_list? {case_clause(_,x)}
*/
func (ps *Parser) exprCaseClause() Node {
	/* 'case' x=expression_list ':' y=statement_semi_list? {case_clause(x,y)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var y Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwCase)
		if _1 == nil {
			break
		}
		x = ps.expressionList()
		if x == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpColon)
		if _2 == nil {
			break
		}
		y = ps.statementSemiList()
		_ = y
		return NewCaseClauseNode(ps._filePath, ps._fileContent, x, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* 'default' ':' x=statement_semi_list? {case_clause(_,x)}
	 */
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwDefault)
		if _1 == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpColon)
		if _2 == nil {
			break
		}
		x = ps.statementSemiList()
		_ = x
		return NewCaseClauseNode(ps._filePath, ps._fileContent, nil, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
if_stmt:
| 'if' [ (init=simple_stmt ';')? cond=expression ] body=block 'else' else_=if_stmt {if_stmt(init, cond, body, else_)}
| 'if' [ (init=simple_stmt ';')? cond=expression ] body=block 'else' else_=block {if_stmt(init, cond, body, else_)}
| 'if' [ init=simple_stmt ';' cond=expression ] body=block {if_stmt(init, cond, body, _)}
| 'if' [ cond=expression ] body=block {if_stmt(_, cond, body, _)}
*/
func (ps *Parser) ifStmt() Node {
	/* 'if' [ (init=simple_stmt ';')? cond=expression ] body=block 'else' else_=if_stmt {if_stmt(init, cond, body, else_)}
	 */
	pos := ps._mark()
	for {
		var body Node
		var cond Node
		var else_ Node
		var init Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwIf)
		if _1 == nil {
			break
		}
		_break := true
		ps._enter()
		for {
			var _2 Node
			for {
				_ok := false
				_p := ps._mark()
				for {
					init = ps.simpleStmt()
					if init == nil {
						break
					}
					_2 = ps._expectK(TokenTypeOpSemi)
					if _2 == nil {
						break
					}
					_ok = true
					break
				}
				if !_ok {
					ps._reset(_p)
					init = nil
				}
				break
			}
			_ = _2
			cond = ps.expression()
			if cond == nil {
				break
			}
			_break = false
			break
		}
		ps._leave()
		if _break {
			break
		}
		body = ps.block()
		if body == nil {
			break
		}
		var _3 Node
		_3 = ps._expectK(TokenTypeKwElse)
		if _3 == nil {
			break
		}
		else_ = ps.ifStmt()
		if else_ == nil {
			break
		}
		return NewIfStmtNode(ps._filePath, ps._fileContent, init, cond, body, else_, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* 'if' [ (init=simple_stmt ';')? cond=expression ] body=block 'else' else_=block {if_stmt(init, cond, body, else_)}
	 */
	for {
		var body Node
		var cond Node
		var else_ Node
		var init Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwIf)
		if _1 == nil {
			break
		}
		_break := true
		ps._enter()
		for {
			var _2 Node
			for {
				_ok := false
				_p := ps._mark()
				for {
					init = ps.simpleStmt()
					if init == nil {
						break
					}
					_2 = ps._expectK(TokenTypeOpSemi)
					if _2 == nil {
						break
					}
					_ok = true
					break
				}
				if !_ok {
					ps._reset(_p)
					init = nil
				}
				break
			}
			_ = _2
			cond = ps.expression()
			if cond == nil {
				break
			}
			_break = false
			break
		}
		ps._leave()
		if _break {
			break
		}
		body = ps.block()
		if body == nil {
			break
		}
		var _3 Node
		_3 = ps._expectK(TokenTypeKwElse)
		if _3 == nil {
			break
		}
		else_ = ps.block()
		if else_ == nil {
			break
		}
		return NewIfStmtNode(ps._filePath, ps._fileContent, init, cond, body, else_, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* 'if' [ init=simple_stmt ';' cond=expression ] body=block {if_stmt(init, cond, body, _)}
	 */
	for {
		var body Node
		var cond Node
		var init Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwIf)
		if _1 == nil {
			break
		}
		_break := true
		ps._enter()
		for {
			init = ps.simpleStmt()
			if init == nil {
				break
			}
			var _2 Node
			_2 = ps._expectK(TokenTypeOpSemi)
			if _2 == nil {
				break
			}
			cond = ps.expression()
			if cond == nil {
				break
			}
			_break = false
			break
		}
		ps._leave()
		if _break {
			break
		}
		body = ps.block()
		if body == nil {
			break
		}
		return NewIfStmtNode(ps._filePath, ps._fileContent, init, cond, body, nil, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* 'if' [ cond=expression ] body=block {if_stmt(_, cond, body, _)}
	 */
	for {
		var body Node
		var cond Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwIf)
		if _1 == nil {
			break
		}
		_break := true
		ps._enter()
		for {
			cond = ps.expression()
			if cond == nil {
				break
			}
			_break = false
			break
		}
		ps._leave()
		if _break {
			break
		}
		body = ps.block()
		if body == nil {
			break
		}
		return NewIfStmtNode(ps._filePath, ps._fileContent, nil, cond, body, nil, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
simple_stmt:
| assignment
| short_val_decl
| inc_dec_stmt
| send_stmt
| expression_stmt
*/
func (ps *Parser) simpleStmt() Node {
	/* assignment
	 */
	for {
		var _1 Node
		_1 = ps.assignment()
		if _1 == nil {
			break
		}
		return _1
	}
	/* short_val_decl
	 */
	for {
		var _1 Node
		_1 = ps.shortValDecl()
		if _1 == nil {
			break
		}
		return _1
	}
	/* inc_dec_stmt
	 */
	for {
		var _1 Node
		_1 = ps.incDecStmt()
		if _1 == nil {
			break
		}
		return _1
	}
	/* send_stmt
	 */
	for {
		var _1 Node
		_1 = ps.sendStmt()
		if _1 == nil {
			break
		}
		return _1
	}
	/* expression_stmt
	 */
	for {
		var _1 Node
		_1 = ps.expressionStmt()
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
expression_stmt:
| x=expression {expr_stmt(x)}
*/
func (ps *Parser) expressionStmt() Node {
	/* x=expression {expr_stmt(x)}
	 */
	pos := ps._mark()
	for {
		var x Node
		x = ps.expression()
		if x == nil {
			break
		}
		return NewExprStmtNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
send_stmt:
| x=expression '<-' y=expression {send_stmt(x,y)}
*/
func (ps *Parser) sendStmt() Node {
	/* x=expression '<-' y=expression {send_stmt(x,y)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var y Node
		x = ps.expression()
		if x == nil {
			break
		}
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLessMinus)
		if _1 == nil {
			break
		}
		y = ps.expression()
		if y == nil {
			break
		}
		return NewSendStmtNode(ps._filePath, ps._fileContent, x, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
inc_dec_stmt:
| x=expression y=('++'|'--') {inc_dec_stmt(x,y)}
_group_4 <-- ('++'|'--')
*/
func (ps *Parser) incDecStmt() Node {
	/* x=expression y=('++'|'--') {inc_dec_stmt(x,y)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var y Node
		x = ps.expression()
		if x == nil {
			break
		}
		y = ps._group4()
		if y == nil {
			break
		}
		return NewIncDecStmtNode(ps._filePath, ps._fileContent, x, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
assignment:
| l=expression_list op=assign_op r=expression_list {assign_stmt(l, op, r)}
*/
func (ps *Parser) assignment() Node {
	/* l=expression_list op=assign_op r=expression_list {assign_stmt(l, op, r)}
	 */
	pos := ps._mark()
	for {
		var l Node
		var op Node
		var r Node
		l = ps.expressionList()
		if l == nil {
			break
		}
		op = ps.assignOp()
		if op == nil {
			break
		}
		r = ps.expressionList()
		if r == nil {
			break
		}
		return NewAssignStmtNode(ps._filePath, ps._fileContent, l, op, r, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
assign_op:
| '='
| '+='
| '-='
| '|='
| '^='
| '*='
| '/='
| '%='
| '<<='
| '>>='
| '&='
| '&^='
*/
func (ps *Parser) assignOp() Node {
	/* '='
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpEqual)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '+='
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpPlusEqual)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '-='
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpMinusEqual)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '|='
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpBarEqual)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '^='
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpCaretEqual)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '*='
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpStarEqual)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '/='
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpSlashEqual)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '%='
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpPercentEqual)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '<<='
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLessLessEqual)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '>>='
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpGreaterGreaterEqual)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '&='
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpAndEqual)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '&^='
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpAndCaretEqual)
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
short_val_decl:
| l=identifier_list op=':=' r=expression_list {assign_stmt(l, op, r)}
*/
func (ps *Parser) shortValDecl() Node {
	/* l=identifier_list op=':=' r=expression_list {assign_stmt(l, op, r)}
	 */
	pos := ps._mark()
	for {
		var l Node
		var op Node
		var r Node
		l = ps.identifierList()
		if l == nil {
			break
		}
		op = ps._expectK(TokenTypeOpColonEqual)
		if op == nil {
			break
		}
		r = ps.expressionList()
		if r == nil {
			break
		}
		return NewAssignStmtNode(ps._filePath, ps._fileContent, l, op, r, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
empty_block:
| '{' '}' {block_stmt(_)}
*/
func (ps *Parser) emptyBlock() Node {
	/* '{' '}' {block_stmt(_)}
	 */
	pos := ps._mark()
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftBrace)
		if _1 == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpRightBrace)
		if _2 == nil {
			break
		}
		return NewBlockStmtNode(ps._filePath, ps._fileContent, nil, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
labeled_stmt:
| x=IDENT ':' b=empty_block {labeled_stmt(x,b)}
| x=IDENT ':' y=statement {labeled_stmt(x,y)}
| x=IDENT ':' {labeled_stmt(x,_)}
*/
func (ps *Parser) labeledStmt() Node {
	/* x=IDENT ':' b=empty_block {labeled_stmt(x,b)}
	 */
	pos := ps._mark()
	for {
		var b Node
		var x Node
		x = ps._expectK(TokenTypeIdent)
		if x == nil {
			break
		}
		var _1 Node
		_1 = ps._expectK(TokenTypeOpColon)
		if _1 == nil {
			break
		}
		b = ps.emptyBlock()
		if b == nil {
			break
		}
		return NewLabeledStmtNode(ps._filePath, ps._fileContent, x, b, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* x=IDENT ':' y=statement {labeled_stmt(x,y)}
	 */
	for {
		var x Node
		var y Node
		x = ps._expectK(TokenTypeIdent)
		if x == nil {
			break
		}
		var _1 Node
		_1 = ps._expectK(TokenTypeOpColon)
		if _1 == nil {
			break
		}
		y = ps.statement()
		if y == nil {
			break
		}
		return NewLabeledStmtNode(ps._filePath, ps._fileContent, x, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* x=IDENT ':' {labeled_stmt(x,_)}
	 */
	for {
		var x Node
		x = ps._expectK(TokenTypeIdent)
		if x == nil {
			break
		}
		var _1 Node
		_1 = ps._expectK(TokenTypeOpColon)
		if _1 == nil {
			break
		}
		return NewLabeledStmtNode(ps._filePath, ps._fileContent, x, nil, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
type:
| type_name_or_generic_type_instantiation
| type_lit
| '(' x=type ')' {paren_expr(x)}
*/
func (ps *Parser) type_() Node {
	/* type_name_or_generic_type_instantiation
	 */
	for {
		var _1 Node
		_1 = ps.typeNameOrGenericTypeInstantiation()
		if _1 == nil {
			break
		}
		return _1
	}
	/* type_lit
	 */
	for {
		var _1 Node
		_1 = ps.typeLit()
		if _1 == nil {
			break
		}
		return _1
	}
	/* '(' x=type ')' {paren_expr(x)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftParen)
		if _1 == nil {
			break
		}
		x = ps.type_()
		if x == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpRightParen)
		if _2 == nil {
			break
		}
		return NewParenExprNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
type_name_or_generic_type_instantiation:
| x=type_name y=generic_args {generic_type_instantiation(x, y)}
| type_name
*/
func (ps *Parser) typeNameOrGenericTypeInstantiation() Node {
	/* x=type_name y=generic_args {generic_type_instantiation(x, y)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var y Node
		x = ps.typeName()
		if x == nil {
			break
		}
		y = ps.genericArgs()
		if y == nil {
			break
		}
		return NewGenericTypeInstantiationNode(ps._filePath, ps._fileContent, x, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* type_name
	 */
	for {
		var _1 Node
		_1 = ps.typeName()
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
type_name:
| qualified_ident
| IDENT
*/
func (ps *Parser) typeName() Node {
	/* qualified_ident
	 */
	for {
		var _1 Node
		_1 = ps.qualifiedIdent()
		if _1 == nil {
			break
		}
		return _1
	}
	/* IDENT
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeIdent)
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
generic_args:
| '[' s=','.type+ ']' {s}
*/
func (ps *Parser) genericArgs() Node {
	/* '[' s=','.type+ ']' {s}
	 */
	pos := ps._mark()
	for {
		var s Node
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftBracket)
		if _1 == nil {
			break
		}
		_2 := make([]Node, 0)
		var _3, _4 Node
		_3 = ps.type_()
		if _3 == nil {
			break
		}
		_2 = append(_2, _3)
		for {
			_p := ps._mark()
			_4 = ps._expectK(TokenTypeOpComma)
			if _4 == nil {
				break
			}
			_3 = ps.type_()
			if _3 == nil {
				ps._reset(_p)
				break
			}
			_2 = append(_2, _3)
		}
		s = NewNodesNode(_2)
		_ = s
		var _5 Node
		_5 = ps._expectK(TokenTypeOpRightBracket)
		if _5 == nil {
			break
		}
		return s
	}
	ps._reset(pos)
	return nil
}

/*
type_lit:
| '*' x=type {star_expr(x)}
| array_type
| struct_type
| 'func' x=signature {x}
| 'interface' b=interface_body {interface_type(b)}
| map_type
| channel_type
*/
func (ps *Parser) typeLit() Node {
	/* '*' x=type {star_expr(x)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeOpStar)
		if _1 == nil {
			break
		}
		x = ps.type_()
		if x == nil {
			break
		}
		return NewStarExprNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* array_type
	 */
	for {
		var _1 Node
		_1 = ps.arrayType()
		if _1 == nil {
			break
		}
		return _1
	}
	/* struct_type
	 */
	for {
		var _1 Node
		_1 = ps.structType()
		if _1 == nil {
			break
		}
		return _1
	}
	/* 'func' x=signature {x}
	 */
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwFunc)
		if _1 == nil {
			break
		}
		x = ps.signature()
		if x == nil {
			break
		}
		return x
	}
	ps._reset(pos)
	/* 'interface' b=interface_body {interface_type(b)}
	 */
	for {
		var b Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwInterface)
		if _1 == nil {
			break
		}
		b = ps.interfaceBody()
		if b == nil {
			break
		}
		return NewInterfaceTypeNode(ps._filePath, ps._fileContent, b, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* map_type
	 */
	for {
		var _1 Node
		_1 = ps.mapType()
		if _1 == nil {
			break
		}
		return _1
	}
	/* channel_type
	 */
	for {
		var _1 Node
		_1 = ps.channelType()
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
interface_body:
| '{' x=method_spec_and_interface_type_name_semi* '}' {field_list(x)}
*/
func (ps *Parser) interfaceBody() Node {
	/* '{' x=method_spec_and_interface_type_name_semi* '}' {field_list(x)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftBrace)
		if _1 == nil {
			break
		}
		_2 := make([]Node, 0)
		var _3 Node
		for {
			_3 = ps.methodSpecAndInterfaceTypeNameSemi()
			if _3 == nil {
				break
			}
			_2 = append(_2, _3)
		}
		x = NewNodesNode(_2)
		_ = x
		var _4 Node
		_4 = ps._expectK(TokenTypeOpRightBrace)
		if _4 == nil {
			break
		}
		return NewFieldListNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
method_spec_and_interface_type_name_semi:
| method_spec_semi
| interface_type_name_semi
| '|'.('~'? t=type {t})+ pseudo_semi {field(_,_,_)}
_group_5 <-- ('~'? t=type {t})
*/
func (ps *Parser) methodSpecAndInterfaceTypeNameSemi() Node {
	/* method_spec_semi
	 */
	for {
		var _1 Node
		_1 = ps.methodSpecSemi()
		if _1 == nil {
			break
		}
		return _1
	}
	/* interface_type_name_semi
	 */
	for {
		var _1 Node
		_1 = ps.interfaceTypeNameSemi()
		if _1 == nil {
			break
		}
		return _1
	}
	/* '|'.('~'? t=type {t})+ pseudo_semi {field(_,_,_)}
	 */
	pos := ps._mark()
	for {
		var _1 Node
		_2 := make([]Node, 0)
		var _3, _4 Node
		_3 = ps._group5()
		if _3 == nil {
			break
		}
		_2 = append(_2, _3)
		for {
			_p := ps._mark()
			_4 = ps._expectK(TokenTypeOpBar)
			if _4 == nil {
				break
			}
			_3 = ps._group5()
			if _3 == nil {
				ps._reset(_p)
				break
			}
			_2 = append(_2, _3)
		}
		_1 = NewNodesNode(_2)
		_ = _1
		var _5 Node
		_5 = ps.pseudoSemi()
		if _5 == nil {
			break
		}
		return NewFieldNode(ps._filePath, ps._fileContent, nil, nil, nil, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
method_spec_semi:
| x=method_spec pseudo_semi {x}
*/
func (ps *Parser) methodSpecSemi() Node {
	/* x=method_spec pseudo_semi {x}
	 */
	pos := ps._mark()
	for {
		var x Node
		x = ps.methodSpec()
		if x == nil {
			break
		}
		var _1 Node
		_1 = ps.pseudoSemi()
		if _1 == nil {
			break
		}
		return x
	}
	ps._reset(pos)
	return nil
}

/*
interface_type_name_semi:
| x=type_name pseudo_semi {field(_,x,_)}
*/
func (ps *Parser) interfaceTypeNameSemi() Node {
	/* x=type_name pseudo_semi {field(_,x,_)}
	 */
	pos := ps._mark()
	for {
		var x Node
		x = ps.typeName()
		if x == nil {
			break
		}
		var _1 Node
		_1 = ps.pseudoSemi()
		if _1 == nil {
			break
		}
		return NewFieldNode(ps._filePath, ps._fileContent, nil, x, nil, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
channel_type:
| t=(a='chan' b='<-' {_pseudo_token(a, b)} | a='<-' b='chan' {_pseudo_token(a, b)} | 'chan') x=type {chan_type(t, x)}
_group_6 <-- (a='chan' b='<-' {_pseudo_token(a, b)} | a='<-' b='chan' {_pseudo_token(a, b)} | 'chan')
*/
func (ps *Parser) channelType() Node {
	/* t=(a='chan' b='<-' {_pseudo_token(a, b)} | a='<-' b='chan' {_pseudo_token(a, b)} | 'chan') x=type {chan_type(t, x)}
	 */
	pos := ps._mark()
	for {
		var t Node
		var x Node
		t = ps._group6()
		if t == nil {
			break
		}
		x = ps.type_()
		if x == nil {
			break
		}
		return NewChanTypeNode(ps._filePath, ps._fileContent, t, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
map_type:
| 'map' '[' x=type ']' y=type {map_type(x,y)}
*/
func (ps *Parser) mapType() Node {
	/* 'map' '[' x=type ']' y=type {map_type(x,y)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var y Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwMap)
		if _1 == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpLeftBracket)
		if _2 == nil {
			break
		}
		x = ps.type_()
		if x == nil {
			break
		}
		var _3 Node
		_3 = ps._expectK(TokenTypeOpRightBracket)
		if _3 == nil {
			break
		}
		y = ps.type_()
		if y == nil {
			break
		}
		return NewMapTypeNode(ps._filePath, ps._fileContent, x, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
qualified_ident:
| x=IDENT '.' y=IDENT { selector_expr(x, y) }
*/
func (ps *Parser) qualifiedIdent() Node {
	/* x=IDENT '.' y=IDENT { selector_expr(x, y) }
	 */
	pos := ps._mark()
	for {
		var x Node
		var y Node
		x = ps._expectK(TokenTypeIdent)
		if x == nil {
			break
		}
		var _1 Node
		_1 = ps._expectK(TokenTypeOpDot)
		if _1 == nil {
			break
		}
		y = ps._expectK(TokenTypeIdent)
		if y == nil {
			break
		}
		return NewSelectorExprNode(ps._filePath, ps._fileContent, x, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
receiver:
| parameters
*/
func (ps *Parser) receiver() Node {
	/* parameters
	 */
	for {
		var _1 Node
		_1 = ps.parameters()
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
identifier_list:
| x=','.IDENT+ {x}
*/
func (ps *Parser) identifierList() Node {
	/* x=','.IDENT+ {x}
	 */
	pos := ps._mark()
	for {
		var x Node
		_1 := make([]Node, 0)
		var _2, _3 Node
		_2 = ps._expectK(TokenTypeIdent)
		if _2 == nil {
			break
		}
		_1 = append(_1, _2)
		for {
			_p := ps._mark()
			_3 = ps._expectK(TokenTypeOpComma)
			if _3 == nil {
				break
			}
			_2 = ps._expectK(TokenTypeIdent)
			if _2 == nil {
				ps._reset(_p)
				break
			}
			_1 = append(_1, _2)
		}
		x = NewNodesNode(_1)
		_ = x
		return x
	}
	ps._reset(pos)
	return nil
}

/*
expression_list:
| ','.expression+
*/
func (ps *Parser) expressionList() Node {
	/* ','.expression+
	 */
	for {
		var _1 Node
		_2 := make([]Node, 0)
		var _3, _4 Node
		_3 = ps.expression()
		if _3 == nil {
			break
		}
		_2 = append(_2, _3)
		for {
			_p := ps._mark()
			_4 = ps._expectK(TokenTypeOpComma)
			if _4 == nil {
				break
			}
			_3 = ps.expression()
			if _3 == nil {
				ps._reset(_p)
				break
			}
			_2 = append(_2, _3)
		}
		_1 = NewNodesNode(_2)
		_ = _1
		return _1
	}
	return nil
}

func (ps *Parser) expression() Node {
	pos := ps._mark()
	var ok bool
	var cache *NodeCache
	cacheAtPos := ps._nodeCache[pos]
	if cacheAtPos != nil {
		if cache, ok = cacheAtPos[expressionMemoId]; ok {
			if cache.val == nil {
				return nil
			}
			ps._reset(cache.pos)
			return cache.val
		}
	} else {
		cacheAtPos = make(map[int]*NodeCache)
		ps._nodeCache[pos] = cacheAtPos
	}
	t := ps.expression_()
	cacheAtPos[expressionMemoId] = &NodeCache{t, ps._mark()}
	return t
}

/*
expression!:
| conditional_or_expression
*/
func (ps *Parser) expression_() Node {
	/* conditional_or_expression
	 */
	for {
		var _1 Node
		_1 = ps.conditionalOrExpression()
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
conditional_or_expression:
| x=conditional_or_expression z='||' y=conditional_and_expression {binary_expr(x,y,z)}
| conditional_and_expression
*/
func (ps *Parser) conditionalOrExpression() Node {
	_left := ps.conditionalOrExpressionLeftMost()
	if _left == nil {
		return nil
	}
	_ret := ps.conditionalOrExpressionRightPart(_left)
	for _ret != nil {
		_left = _ret
		_ret = ps.conditionalOrExpressionRightPart(_left)
	}
	return _left
}

func (ps *Parser) conditionalOrExpressionLeftMost() Node {
	/* conditional_and_expression
	 */
	for {
		var _1 Node
		_1 = ps.conditionalAndExpression()
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

func (ps *Parser) conditionalOrExpressionRightPart(_left Node) Node {
	/* x=conditional_or_expression z='||' y=conditional_and_expression {binary_expr(x,y,z)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var y Node
		var z Node
		x = _left
		z = ps._expectK(TokenTypeOpBarBar)
		if z == nil {
			break
		}
		y = ps.conditionalAndExpression()
		if y == nil {
			break
		}
		return NewBinaryExprNode(ps._filePath, ps._fileContent, x, y, z, _left.RangeStart(), ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
conditional_and_expression:
| x=conditional_and_expression z='&&' y=rel_op_expression {binary_expr(x,y,z)}
| rel_op_expression
*/
func (ps *Parser) conditionalAndExpression() Node {
	_left := ps.conditionalAndExpressionLeftMost()
	if _left == nil {
		return nil
	}
	_ret := ps.conditionalAndExpressionRightPart(_left)
	for _ret != nil {
		_left = _ret
		_ret = ps.conditionalAndExpressionRightPart(_left)
	}
	return _left
}

func (ps *Parser) conditionalAndExpressionLeftMost() Node {
	/* rel_op_expression
	 */
	for {
		var _1 Node
		_1 = ps.relOpExpression()
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

func (ps *Parser) conditionalAndExpressionRightPart(_left Node) Node {
	/* x=conditional_and_expression z='&&' y=rel_op_expression {binary_expr(x,y,z)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var y Node
		var z Node
		x = _left
		z = ps._expectK(TokenTypeOpAndAnd)
		if z == nil {
			break
		}
		y = ps.relOpExpression()
		if y == nil {
			break
		}
		return NewBinaryExprNode(ps._filePath, ps._fileContent, x, y, z, _left.RangeStart(), ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
rel_op_expression:
| x=rel_op_expression z=rel_op y=add_op_expression {binary_expr(x,y,z)}
| add_op_expression
*/
func (ps *Parser) relOpExpression() Node {
	_left := ps.relOpExpressionLeftMost()
	if _left == nil {
		return nil
	}
	_ret := ps.relOpExpressionRightPart(_left)
	for _ret != nil {
		_left = _ret
		_ret = ps.relOpExpressionRightPart(_left)
	}
	return _left
}

func (ps *Parser) relOpExpressionLeftMost() Node {
	/* add_op_expression
	 */
	for {
		var _1 Node
		_1 = ps.addOpExpression()
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

func (ps *Parser) relOpExpressionRightPart(_left Node) Node {
	/* x=rel_op_expression z=rel_op y=add_op_expression {binary_expr(x,y,z)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var y Node
		var z Node
		x = _left
		z = ps.relOp()
		if z == nil {
			break
		}
		y = ps.addOpExpression()
		if y == nil {
			break
		}
		return NewBinaryExprNode(ps._filePath, ps._fileContent, x, y, z, _left.RangeStart(), ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
add_op_expression:
| x=add_op_expression z=add_op y=mul_op_expression {binary_expr(x,y,z)}
| mul_op_expression
*/
func (ps *Parser) addOpExpression() Node {
	_left := ps.addOpExpressionLeftMost()
	if _left == nil {
		return nil
	}
	_ret := ps.addOpExpressionRightPart(_left)
	for _ret != nil {
		_left = _ret
		_ret = ps.addOpExpressionRightPart(_left)
	}
	return _left
}

func (ps *Parser) addOpExpressionLeftMost() Node {
	/* mul_op_expression
	 */
	for {
		var _1 Node
		_1 = ps.mulOpExpression()
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

func (ps *Parser) addOpExpressionRightPart(_left Node) Node {
	/* x=add_op_expression z=add_op y=mul_op_expression {binary_expr(x,y,z)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var y Node
		var z Node
		x = _left
		z = ps.addOp()
		if z == nil {
			break
		}
		y = ps.mulOpExpression()
		if y == nil {
			break
		}
		return NewBinaryExprNode(ps._filePath, ps._fileContent, x, y, z, _left.RangeStart(), ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
mul_op_expression:
| x=mul_op_expression z=mul_op y=unary_expr {binary_expr(x,y,z)}
| unary_expr
*/
func (ps *Parser) mulOpExpression() Node {
	_left := ps.mulOpExpressionLeftMost()
	if _left == nil {
		return nil
	}
	_ret := ps.mulOpExpressionRightPart(_left)
	for _ret != nil {
		_left = _ret
		_ret = ps.mulOpExpressionRightPart(_left)
	}
	return _left
}

func (ps *Parser) mulOpExpressionLeftMost() Node {
	/* unary_expr
	 */
	for {
		var _1 Node
		_1 = ps.unaryExpr()
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

func (ps *Parser) mulOpExpressionRightPart(_left Node) Node {
	/* x=mul_op_expression z=mul_op y=unary_expr {binary_expr(x,y,z)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var y Node
		var z Node
		x = _left
		z = ps.mulOp()
		if z == nil {
			break
		}
		y = ps.unaryExpr()
		if y == nil {
			break
		}
		return NewBinaryExprNode(ps._filePath, ps._fileContent, x, y, z, _left.RangeStart(), ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
unary_expr:
| '*' x=unary_expr {star_expr(x)}
| x=unary_op y=unary_expr {unary_expr(x,y)}
| primary_expr
*/
func (ps *Parser) unaryExpr() Node {
	/* '*' x=unary_expr {star_expr(x)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeOpStar)
		if _1 == nil {
			break
		}
		x = ps.unaryExpr()
		if x == nil {
			break
		}
		return NewStarExprNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* x=unary_op y=unary_expr {unary_expr(x,y)}
	 */
	for {
		var x Node
		var y Node
		x = ps.unaryOp()
		if x == nil {
			break
		}
		y = ps.unaryExpr()
		if y == nil {
			break
		}
		return NewUnaryExprNode(ps._filePath, ps._fileContent, x, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* primary_expr
	 */
	for {
		var _1 Node
		_1 = ps.primaryExpr()
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
primary_expr:
| make_ident '(' t=type (',' expression_list)? ','? ')' {make_expr(t)}
| new_ident '(' t=type ')' {new_expr(t)}
| x=primary_expr g=generic_args? '(' ')' {call_expr(x, g, _)}
| x=primary_expr g=generic_args? '(' y=expression_list '...'? ','? ')' {call_expr(x,g,y)}
| x=primary_expr '.' '(' y=type ')' {type_assert_expr(x,y)}
| e=primary_expr '[' l=expression? ':' h=expression ':' m=expression ']' {slice_expr(e,l,h,m)}
| e=primary_expr '[' l=expression? ':' h=expression? ']' {slice_expr(e,l,h,_)}
| x=primary_expr '[' y=expression ']' {index_expr(x,y)}
| x=primary_expr '.' y=IDENT {selector_expr(x, y)}
| x=type g=generic_args? '(' y=expression ','? ')' {call_expr(x,g,[y])}
| '(' x=expression ')' {paren_expr(x)}
| x=NUMBER {basic_lit(x)}
| x=STRING {basic_lit(x)}
| x=literal_type '{' y=','.keyed_element* ','? '}' {composite_lit(x, y)}
| _hack_composite_lit_node
| 'func' x=signature y=block {function_lit(x,y)}
| x=type '.' y=IDENT {selector_expr(x,y)}
| i=IDENT {ident(i)}
*/
func (ps *Parser) primaryExpr() Node {
	_left := ps.primaryExprLeftMost()
	if _left == nil {
		return nil
	}
	_ret := ps.primaryExprRightPart(_left)
	for _ret != nil {
		_left = _ret
		_ret = ps.primaryExprRightPart(_left)
	}
	return _left
}

func (ps *Parser) primaryExprLeftMost() Node {
	/* make_ident '(' t=type (',' expression_list)? ','? ')' {make_expr(t)}
	 */
	pos := ps._mark()
	for {
		var t Node
		var _1 Node
		_1 = ps.makeIdent()
		if _1 == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpLeftParen)
		if _2 == nil {
			break
		}
		t = ps.type_()
		if t == nil {
			break
		}
		var _3 Node
		for {
			_ok := false
			_p := ps._mark()
			for {
				var _4 Node
				_4 = ps._expectK(TokenTypeOpComma)
				if _4 == nil {
					break
				}
				_3 = ps.expressionList()
				if _3 == nil {
					break
				}
				_ok = true
				break
			}
			if !_ok {
				ps._reset(_p)
			}
			break
		}
		_ = _3
		var _5 Node
		_5 = ps._expectK(TokenTypeOpComma)
		_ = _5
		var _6 Node
		_6 = ps._expectK(TokenTypeOpRightParen)
		if _6 == nil {
			break
		}
		return NewMakeExprNode(ps._filePath, ps._fileContent, t, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* new_ident '(' t=type ')' {new_expr(t)}
	 */
	for {
		var t Node
		var _1 Node
		_1 = ps.newIdent()
		if _1 == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpLeftParen)
		if _2 == nil {
			break
		}
		t = ps.type_()
		if t == nil {
			break
		}
		var _3 Node
		_3 = ps._expectK(TokenTypeOpRightParen)
		if _3 == nil {
			break
		}
		return NewNewExprNode(ps._filePath, ps._fileContent, t, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* x=type g=generic_args? '(' y=expression ','? ')' {call_expr(x,g,[y])}
	 */
	for {
		var g Node
		var x Node
		var y Node
		x = ps.type_()
		if x == nil {
			break
		}
		g = ps.genericArgs()
		_ = g
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftParen)
		if _1 == nil {
			break
		}
		y = ps.expression()
		if y == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpComma)
		_ = _2
		var _3 Node
		_3 = ps._expectK(TokenTypeOpRightParen)
		if _3 == nil {
			break
		}
		return NewCallExprNode(ps._filePath, ps._fileContent, x, g, NewNodesNode([]Node{y}), ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* '(' x=expression ')' {paren_expr(x)}
	 */
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftParen)
		if _1 == nil {
			break
		}
		x = ps.expression()
		if x == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpRightParen)
		if _2 == nil {
			break
		}
		return NewParenExprNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* x=NUMBER {basic_lit(x)}
	 */
	for {
		var x Node
		x = ps._expectK(TokenTypeNumber)
		if x == nil {
			break
		}
		return NewBasicLitNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* x=STRING {basic_lit(x)}
	 */
	for {
		var x Node
		x = ps._expectK(TokenTypeString)
		if x == nil {
			break
		}
		return NewBasicLitNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* x=literal_type '{' y=','.keyed_element* ','? '}' {composite_lit(x, y)}
	 */
	for {
		var x Node
		var y Node
		x = ps.literalType()
		if x == nil {
			break
		}
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftBrace)
		if _1 == nil {
			break
		}
		_2 := make([]Node, 0)
		var _3 Node
		var _4 Node
		_3 = ps.keyedElement()
		if _3 != nil {
			_2 = append(_2, _3)
			for {
				_p := ps._mark()
				_4 = ps._expectK(TokenTypeOpComma)
				if _4 == nil {
					break
				}
				_3 = ps.keyedElement()
				if _3 == nil {
					ps._reset(_p)
					break
				}
				_2 = append(_2, _3)
			}
		}
		y = NewNodesNode(_2)
		_ = y
		var _5 Node
		_5 = ps._expectK(TokenTypeOpComma)
		_ = _5
		var _6 Node
		_6 = ps._expectK(TokenTypeOpRightBrace)
		if _6 == nil {
			break
		}
		return NewCompositeLitNode(ps._filePath, ps._fileContent, x, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* _hack_composite_lit_node
	 */
	for {
		var _1 Node
		_1 = ps._hackCompositeLitNode()
		if _1 == nil {
			break
		}
		return _1
	}
	/* 'func' x=signature y=block {function_lit(x,y)}
	 */
	for {
		var x Node
		var y Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwFunc)
		if _1 == nil {
			break
		}
		x = ps.signature()
		if x == nil {
			break
		}
		y = ps.block()
		if y == nil {
			break
		}
		return NewFunctionLitNode(ps._filePath, ps._fileContent, x, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* x=type '.' y=IDENT {selector_expr(x,y)}
	 */
	for {
		var x Node
		var y Node
		x = ps.type_()
		if x == nil {
			break
		}
		var _1 Node
		_1 = ps._expectK(TokenTypeOpDot)
		if _1 == nil {
			break
		}
		y = ps._expectK(TokenTypeIdent)
		if y == nil {
			break
		}
		return NewSelectorExprNode(ps._filePath, ps._fileContent, x, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* i=IDENT {ident(i)}
	 */
	for {
		var i Node
		i = ps._expectK(TokenTypeIdent)
		if i == nil {
			break
		}
		return NewIdentNode(ps._filePath, ps._fileContent, i, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

func (ps *Parser) primaryExprRightPart(_left Node) Node {
	/* x=primary_expr g=generic_args? '(' ')' {call_expr(x, g, _)}
	 */
	pos := ps._mark()
	for {
		var g Node
		var x Node
		x = _left
		g = ps.genericArgs()
		_ = g
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftParen)
		if _1 == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpRightParen)
		if _2 == nil {
			break
		}
		return NewCallExprNode(ps._filePath, ps._fileContent, x, g, nil, _left.RangeStart(), ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* x=primary_expr g=generic_args? '(' y=expression_list '...'? ','? ')' {call_expr(x,g,y)}
	 */
	for {
		var g Node
		var x Node
		var y Node
		x = _left
		g = ps.genericArgs()
		_ = g
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftParen)
		if _1 == nil {
			break
		}
		y = ps.expressionList()
		if y == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpDotDotDot)
		_ = _2
		var _3 Node
		_3 = ps._expectK(TokenTypeOpComma)
		_ = _3
		var _4 Node
		_4 = ps._expectK(TokenTypeOpRightParen)
		if _4 == nil {
			break
		}
		return NewCallExprNode(ps._filePath, ps._fileContent, x, g, y, _left.RangeStart(), ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* x=primary_expr '.' '(' y=type ')' {type_assert_expr(x,y)}
	 */
	for {
		var x Node
		var y Node
		x = _left
		var _1 Node
		_1 = ps._expectK(TokenTypeOpDot)
		if _1 == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpLeftParen)
		if _2 == nil {
			break
		}
		y = ps.type_()
		if y == nil {
			break
		}
		var _3 Node
		_3 = ps._expectK(TokenTypeOpRightParen)
		if _3 == nil {
			break
		}
		return NewTypeAssertExprNode(ps._filePath, ps._fileContent, x, y, _left.RangeStart(), ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* e=primary_expr '[' l=expression? ':' h=expression ':' m=expression ']' {slice_expr(e,l,h,m)}
	 */
	for {
		var e Node
		var h Node
		var l Node
		var m Node
		e = _left
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftBracket)
		if _1 == nil {
			break
		}
		l = ps.expression()
		_ = l
		var _2 Node
		_2 = ps._expectK(TokenTypeOpColon)
		if _2 == nil {
			break
		}
		h = ps.expression()
		if h == nil {
			break
		}
		var _3 Node
		_3 = ps._expectK(TokenTypeOpColon)
		if _3 == nil {
			break
		}
		m = ps.expression()
		if m == nil {
			break
		}
		var _4 Node
		_4 = ps._expectK(TokenTypeOpRightBracket)
		if _4 == nil {
			break
		}
		return NewSliceExprNode(ps._filePath, ps._fileContent, e, l, h, m, _left.RangeStart(), ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* e=primary_expr '[' l=expression? ':' h=expression? ']' {slice_expr(e,l,h,_)}
	 */
	for {
		var e Node
		var h Node
		var l Node
		e = _left
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftBracket)
		if _1 == nil {
			break
		}
		l = ps.expression()
		_ = l
		var _2 Node
		_2 = ps._expectK(TokenTypeOpColon)
		if _2 == nil {
			break
		}
		h = ps.expression()
		_ = h
		var _3 Node
		_3 = ps._expectK(TokenTypeOpRightBracket)
		if _3 == nil {
			break
		}
		return NewSliceExprNode(ps._filePath, ps._fileContent, e, l, h, nil, _left.RangeStart(), ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* x=primary_expr '[' y=expression ']' {index_expr(x,y)}
	 */
	for {
		var x Node
		var y Node
		x = _left
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftBracket)
		if _1 == nil {
			break
		}
		y = ps.expression()
		if y == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpRightBracket)
		if _2 == nil {
			break
		}
		return NewIndexExprNode(ps._filePath, ps._fileContent, x, y, _left.RangeStart(), ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* x=primary_expr '.' y=IDENT {selector_expr(x, y)}
	 */
	for {
		var x Node
		var y Node
		x = _left
		var _1 Node
		_1 = ps._expectK(TokenTypeOpDot)
		if _1 == nil {
			break
		}
		y = ps._expectK(TokenTypeIdent)
		if y == nil {
			break
		}
		return NewSelectorExprNode(ps._filePath, ps._fileContent, x, y, _left.RangeStart(), ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
make_ident:
| x='make' {ident(x)}
*/
func (ps *Parser) makeIdent() Node {
	/* x='make' {ident(x)}
	 */
	pos := ps._mark()
	for {
		var x Node
		x = ps._expectV("\u006Da\u006Be")
		if x == nil {
			break
		}
		return NewIdentNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
new_ident:
| x='new' {ident(x)}
*/
func (ps *Parser) newIdent() Node {
	/* x='new' {ident(x)}
	 */
	pos := ps._mark()
	for {
		var x Node
		x = ps._expectV("\u006Ee\u0077")
		if x == nil {
			break
		}
		return NewIdentNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
rel_op:
| '=='
| '!='
| '<'
| '<='
| '>'
| '>='
*/
func (ps *Parser) relOp() Node {
	/* '=='
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpEqualEqual)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '!='
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpNotEqual)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '<'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLess)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '<='
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLessEqual)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '>'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpGreater)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '>='
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpGreaterEqual)
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
add_op:
| '+'
| '-'
| '|'
| '^'
*/
func (ps *Parser) addOp() Node {
	/* '+'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpPlus)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '-'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpMinus)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '|'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpBar)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '^'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpCaret)
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
mul_op:
| '*'
| '/'
| '%'
| '<<'
| '>>'
| '&'
| '&^'
*/
func (ps *Parser) mulOp() Node {
	/* '*'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpStar)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '/'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpSlash)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '%'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpPercent)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '<<'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLessLess)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '>>'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpGreaterGreater)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '&'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpAnd)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '&^'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpAndCaret)
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
unary_op:
| '+'
| '-'
| '!'
| '^'
| '*'
| '&'
| '<-'
*/
func (ps *Parser) unaryOp() Node {
	/* '+'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpPlus)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '-'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpMinus)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '!'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpNot)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '^'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpCaret)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '*'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpStar)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '&'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpAnd)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '<-'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLessMinus)
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
composite_lit:
| x=type_name_or_generic_type_instantiation? '{' y=','.keyed_element* ','? '}' {composite_lit(x, y)}
*/
func (ps *Parser) compositeLit() Node {
	/* x=type_name_or_generic_type_instantiation? '{' y=','.keyed_element* ','? '}' {composite_lit(x, y)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var y Node
		x = ps.typeNameOrGenericTypeInstantiation()
		_ = x
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftBrace)
		if _1 == nil {
			break
		}
		_2 := make([]Node, 0)
		var _3 Node
		var _4 Node
		_3 = ps.keyedElement()
		if _3 != nil {
			_2 = append(_2, _3)
			for {
				_p := ps._mark()
				_4 = ps._expectK(TokenTypeOpComma)
				if _4 == nil {
					break
				}
				_3 = ps.keyedElement()
				if _3 == nil {
					ps._reset(_p)
					break
				}
				_2 = append(_2, _3)
			}
		}
		y = NewNodesNode(_2)
		_ = y
		var _5 Node
		_5 = ps._expectK(TokenTypeOpComma)
		_ = _5
		var _6 Node
		_6 = ps._expectK(TokenTypeOpRightBrace)
		if _6 == nil {
			break
		}
		return NewCompositeLitNode(ps._filePath, ps._fileContent, x, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
ellipsis:
| '...' {ellipsis(_)}
*/
func (ps *Parser) ellipsis() Node {
	/* '...' {ellipsis(_)}
	 */
	pos := ps._mark()
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpDotDotDot)
		if _1 == nil {
			break
		}
		return NewEllipsisNode(ps._filePath, ps._fileContent, nil, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
literal_type:
| struct_type
| '[' e=ellipsis ']' x=type {array_type(e,x)}
| array_type
| map_type
*/
func (ps *Parser) literalType() Node {
	/* struct_type
	 */
	for {
		var _1 Node
		_1 = ps.structType()
		if _1 == nil {
			break
		}
		return _1
	}
	/* '[' e=ellipsis ']' x=type {array_type(e,x)}
	 */
	pos := ps._mark()
	for {
		var e Node
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftBracket)
		if _1 == nil {
			break
		}
		e = ps.ellipsis()
		if e == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpRightBracket)
		if _2 == nil {
			break
		}
		x = ps.type_()
		if x == nil {
			break
		}
		return NewArrayTypeNode(ps._filePath, ps._fileContent, e, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* array_type
	 */
	for {
		var _1 Node
		_1 = ps.arrayType()
		if _1 == nil {
			break
		}
		return _1
	}
	/* map_type
	 */
	for {
		var _1 Node
		_1 = ps.mapType()
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
keyed_element:
| x=expression ':' y=expression {key_value_expr(x,y)}
| expression
*/
func (ps *Parser) keyedElement() Node {
	/* x=expression ':' y=expression {key_value_expr(x,y)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var y Node
		x = ps.expression()
		if x == nil {
			break
		}
		var _1 Node
		_1 = ps._expectK(TokenTypeOpColon)
		if _1 == nil {
			break
		}
		y = ps.expression()
		if y == nil {
			break
		}
		return NewKeyValueExprNode(ps._filePath, ps._fileContent, x, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* expression
	 */
	for {
		var _1 Node
		_1 = ps.expression()
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
array_type:
| '[' ']' x=type {array_type(_,x)}
| '[' e=ellipsis ']' x=type {array_type(e,x)}
| '[' x=expression ']' y=type {array_type(x,y)}
*/
func (ps *Parser) arrayType() Node {
	/* '[' ']' x=type {array_type(_,x)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftBracket)
		if _1 == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpRightBracket)
		if _2 == nil {
			break
		}
		x = ps.type_()
		if x == nil {
			break
		}
		return NewArrayTypeNode(ps._filePath, ps._fileContent, nil, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* '[' e=ellipsis ']' x=type {array_type(e,x)}
	 */
	for {
		var e Node
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftBracket)
		if _1 == nil {
			break
		}
		e = ps.ellipsis()
		if e == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpRightBracket)
		if _2 == nil {
			break
		}
		x = ps.type_()
		if x == nil {
			break
		}
		return NewArrayTypeNode(ps._filePath, ps._fileContent, e, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* '[' x=expression ']' y=type {array_type(x,y)}
	 */
	for {
		var x Node
		var y Node
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftBracket)
		if _1 == nil {
			break
		}
		x = ps.expression()
		if x == nil {
			break
		}
		var _2 Node
		_2 = ps._expectK(TokenTypeOpRightBracket)
		if _2 == nil {
			break
		}
		y = ps.type_()
		if y == nil {
			break
		}
		return NewArrayTypeNode(ps._filePath, ps._fileContent, x, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
method_spec:
| x=IDENT y=signature {field([x],y,_)}
*/
func (ps *Parser) methodSpec() Node {
	/* x=IDENT y=signature {field([x],y,_)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var y Node
		x = ps._expectK(TokenTypeIdent)
		if x == nil {
			break
		}
		y = ps.signature()
		if y == nil {
			break
		}
		return NewFieldNode(ps._filePath, ps._fileContent, NewNodesNode([]Node{x}), y, nil, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
struct_body:
| '{' x=field_decl_semi* '}' {field_list(x)}
*/
func (ps *Parser) structBody() Node {
	/* '{' x=field_decl_semi* '}' {field_list(x)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeOpLeftBrace)
		if _1 == nil {
			break
		}
		_2 := make([]Node, 0)
		var _3 Node
		for {
			_3 = ps.fieldDeclSemi()
			if _3 == nil {
				break
			}
			_2 = append(_2, _3)
		}
		x = NewNodesNode(_2)
		_ = x
		var _4 Node
		_4 = ps._expectK(TokenTypeOpRightBrace)
		if _4 == nil {
			break
		}
		return NewFieldListNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
struct_type:
| 'struct' b=struct_body {struct_type(b)}
*/
func (ps *Parser) structType() Node {
	/* 'struct' b=struct_body {struct_type(b)}
	 */
	pos := ps._mark()
	for {
		var b Node
		var _1 Node
		_1 = ps._expectK(TokenTypeKwStruct)
		if _1 == nil {
			break
		}
		b = ps.structBody()
		if b == nil {
			break
		}
		return NewStructTypeNode(ps._filePath, ps._fileContent, b, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
field_decl_semi:
| x=field_decl pseudo_semi {x}
*/
func (ps *Parser) fieldDeclSemi() Node {
	/* x=field_decl pseudo_semi {x}
	 */
	pos := ps._mark()
	for {
		var x Node
		x = ps.fieldDecl()
		if x == nil {
			break
		}
		var _1 Node
		_1 = ps.pseudoSemi()
		if _1 == nil {
			break
		}
		return x
	}
	ps._reset(pos)
	return nil
}

/*
field_decl:
| x=identifier_list y=type z=tag? {field(x,y,z)}
| x=embedded_field y=tag? {field(_,x,y)}
*/
func (ps *Parser) fieldDecl() Node {
	/* x=identifier_list y=type z=tag? {field(x,y,z)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var y Node
		var z Node
		x = ps.identifierList()
		if x == nil {
			break
		}
		y = ps.type_()
		if y == nil {
			break
		}
		z = ps.tag()
		_ = z
		return NewFieldNode(ps._filePath, ps._fileContent, x, y, z, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* x=embedded_field y=tag? {field(_,x,y)}
	 */
	for {
		var x Node
		var y Node
		x = ps.embeddedField()
		if x == nil {
			break
		}
		y = ps.tag()
		_ = y
		return NewFieldNode(ps._filePath, ps._fileContent, nil, x, y, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
embedded_field:
| '*' x=type_name_or_generic_type_instantiation {star_expr(x)}
| t=type_name_or_generic_type_instantiation {t}
*/
func (ps *Parser) embeddedField() Node {
	/* '*' x=type_name_or_generic_type_instantiation {star_expr(x)}
	 */
	pos := ps._mark()
	for {
		var x Node
		var _1 Node
		_1 = ps._expectK(TokenTypeOpStar)
		if _1 == nil {
			break
		}
		x = ps.typeNameOrGenericTypeInstantiation()
		if x == nil {
			break
		}
		return NewStarExprNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	/* t=type_name_or_generic_type_instantiation {t}
	 */
	for {
		var t Node
		t = ps.typeNameOrGenericTypeInstantiation()
		if t == nil {
			break
		}
		return t
	}
	ps._reset(pos)
	return nil
}

/*
tag:
| x=STRING {basic_lit(x)}
*/
func (ps *Parser) tag() Node {
	/* x=STRING {basic_lit(x)}
	 */
	pos := ps._mark()
	for {
		var x Node
		x = ps._expectK(TokenTypeString)
		if x == nil {
			break
		}
		return NewBasicLitNode(ps._filePath, ps._fileContent, x, ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End)
	}
	ps._reset(pos)
	return nil
}

/*
pseudo_semi:
| ';'
| &')'
| &'}'
*/
func (ps *Parser) pseudoSemi() Node {
	/* ';'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpSemi)
		if _1 == nil {
			break
		}
		return _1
	}
	/* &')'
	 */
	for {
		var _1 Node
		_p := ps._mark()
		_1 = ps._expectK(TokenTypeOpRightParen)
		if _1 != nil {
			ps._reset(_p)
		}
		if _1 == nil {
			break
		}
		return _1
	}
	/* &'}'
	 */
	for {
		var _1 Node
		_p := ps._mark()
		_1 = ps._expectK(TokenTypeOpRightBrace)
		if _1 != nil {
			ps._reset(_p)
		}
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
_group_1:
| t=import_spec pseudo_semi {t}
*/
func (ps *Parser) _group1() Node {
	/* t=import_spec pseudo_semi {t}
	 */
	pos := ps._mark()
	for {
		var t Node
		t = ps.importSpec()
		if t == nil {
			break
		}
		var _1 Node
		_1 = ps.pseudoSemi()
		if _1 == nil {
			break
		}
		return t
	}
	ps._reset(pos)
	return nil
}

/*
_group_2:
| import_dot
| import_ident
*/
func (ps *Parser) _group2() Node {
	/* import_dot
	 */
	for {
		var _1 Node
		_1 = ps.importDot()
		if _1 == nil {
			break
		}
		return _1
	}
	/* import_ident
	 */
	for {
		var _1 Node
		_1 = ps.importIdent()
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
_group_3:
| ':='
| '='
*/
func (ps *Parser) _group3() Node {
	/* ':='
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpColonEqual)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '='
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpEqual)
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
_group_4:
| '++'
| '--'
*/
func (ps *Parser) _group4() Node {
	/* '++'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpPlusPlus)
		if _1 == nil {
			break
		}
		return _1
	}
	/* '--'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeOpMinusMinus)
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

/*
_group_5:
| '~'? t=type {t}
*/
func (ps *Parser) _group5() Node {
	/* '~'? t=type {t}
	 */
	pos := ps._mark()
	for {
		var t Node
		var _1 Node
		_1 = ps._expectK(TokenTypeOpTilde)
		_ = _1
		t = ps.type_()
		if t == nil {
			break
		}
		return t
	}
	ps._reset(pos)
	return nil
}

/*
_group_6:
| a='chan' b='<-' {_pseudo_token(a, b)}
| a='<-' b='chan' {_pseudo_token(a, b)}
| 'chan'
*/
func (ps *Parser) _group6() Node {
	/* a='chan' b='<-' {_pseudo_token(a, b)}
	 */
	pos := ps._mark()
	for {
		var a Node
		var b Node
		a = ps._expectK(TokenTypeKwChan)
		if a == nil {
			break
		}
		b = ps._expectK(TokenTypeOpLessMinus)
		if b == nil {
			break
		}
		return ps._pseudoToken(a, b)
	}
	ps._reset(pos)
	/* a='<-' b='chan' {_pseudo_token(a, b)}
	 */
	for {
		var a Node
		var b Node
		a = ps._expectK(TokenTypeOpLessMinus)
		if a == nil {
			break
		}
		b = ps._expectK(TokenTypeKwChan)
		if b == nil {
			break
		}
		return ps._pseudoToken(a, b)
	}
	ps._reset(pos)
	/* 'chan'
	 */
	for {
		var _1 Node
		_1 = ps._expectK(TokenTypeKwChan)
		if _1 == nil {
			break
		}
		return _1
	}
	return nil
}

func (tk *Tokenizer) Clean(tokens []*Token) []*Token {
	ret := make([]*Token, 0)
	var last *Token
	for _, tok := range tokens {
		// insert optional semicolon
		// The formal grammar uses semicolons ";" as terminators in a number of productions. Go programs may omit most of these semicolons using the following two rules:
		//
		// When the input is broken into tokens, a semicolon is automatically inserted into the token stream immediately after a line's final token if that token is
		// an identifier
		// an integer, floating-point, imaginary, rune, or string literal
		// one of the keywords break, continue, fallthrough, or return
		// one of the operators and punctuation ++, --, ), ], or }
		// To allow complex statements to occupy a single line, a semicolon may be omitted before a closing ")" or "}".
		if tok.Kind == TokenTypeNewline {
			if last != nil && last.Kind != TokenTypeOpSemi {
				insertSemi := false
				switch last.Kind {
				case TokenTypeIdent:
					insertSemi = true
				case TokenTypeString:
					insertSemi = true
				case TokenTypeOpRightParen, TokenTypeOpRightBracket, TokenTypeOpRightBrace:
					// ),],}
					insertSemi = true
				case TokenTypeOpPlusPlus, TokenTypeOpMinusMinus:
					// ++,--
					insertSemi = true
				case TokenTypeNumber:
					insertSemi = true
				case TokenTypeKwFallthrough, TokenTypeKwReturn, TokenTypeKwBreak, TokenTypeKwContinue:
					insertSemi = true
				}

				if insertSemi {
					last = NewToken(TokenTypeOpSemi, last.Start, last.End, []rune(";"))
					ret = append(ret, last)
				}
			}
		}

		if tok.Kind == TokenTypeWhitespace || tok.Kind == TokenTypeNewline || tok.Kind == TokenTypeComment {
			continue
		}

		ret = append(ret, tok)
		last = tok
		if tok.Kind == TokenTypeEndOfFile {
			break
		}
	}
	return ret
}

func (ps *Parser) _setDepth(d int) {
    ps._any = d
}

func (ps *Parser) _getDepth() int {
    return ps._any.(int)
}

func (ps *Parser) _enter() {
	ps._setDepth(ps._bracketDepth + 1)
}

func (ps *Parser) _leave() {
	ps._setDepth(0)
}

func (ps *Parser) _hackCompositeLitNode() Node {
	if ps._bracketDepth >= ps._getDepth() {
		return ps.compositeLit()
	}
	return nil
}
func DumpNode(n Node, hook func(Node, map[string]string) string) string {
	return CustomDumpNode(n, hook)
}

func DumpNodeIndent(node Node) string {
	result := SimpleDumpNode(node)
	var v any
	err := json.Unmarshal([]byte(result), &v)
	if err != nil {
		panic(err)
	}
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func CustomDumpNode(node Node, hook func(Node, map[string]string) string) string {
	if node.IsDummy() {
		return "null"
	}
	itemMap := node.Dump(hook)
	ret := hook(node, itemMap)
	if ret != "" {
		return ret
	}
	items := make([]string, 0)
	for k, v := range itemMap {
		if k == "kind" {
			continue
		}
		items = append(items, fmt.Sprintf("\"%s\": %s", k, v))
	}
	sort.Strings(items)
	items = append([]string{fmt.Sprintf("\"kind\": %s", itemMap["kind"])}, items...)
	return fmt.Sprintf("{%s}", strings.Join(items, ", "))
}

func SimpleDumpNode(node Node) string {
	return CustomDumpNode(node, func(n Node, m map[string]string) string {
		return ""
	})
}

func QueryNode(node Node, path string) (any, error) {
	if path == "" {
		return node, nil
	}

	items := strings.Split(path, "/")
	var base any
	base = node
	for _, item := range items {
		var name, nodeType string
		if strings.Contains(item, ":") {
			subs := strings.Split(item, ":")
			name = toCamelCase(subs[0])
			nodeType = subs[1]
		} else {
			name = toCamelCase(item)
		}

		switch base.(type) {
		case Node:
			node = base.(Node)
			if name == "." {
				base = node
			} else if name == ".." {
				base = node.Parent()
				if base == nil {
					return nil, errors.New("query error: node has no parent")
				}
			} else {
				t := reflect.TypeOf(node)
				m, ok := t.MethodByName(name)
				if !ok {
					methods := make([]string, 0)
					for i := 0; i < t.NumMethod(); i++ {
						tmp := t.Method(i).Name
						methods = append(methods, tmp)
					}
					return nil, errors.New(fmt.Sprintf("query error: %v has no method '%s', available: %s", t, name, strings.Join(methods, ", ")))
				}
				result := m.Func.Call([]reflect.Value{
					reflect.ValueOf(node),
				})
				base = result[0].Interface()
			}
		case []Node:
			nodes := base.([]Node)
			index, err := strconv.Atoi(name)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("query error: index should be an integer: '%s'", name))
			}
			if index < 0 || index >= len(nodes) {
				return nil, errors.New("index error")
			}
			base = nodes[index]
		default:
			return nil, errors.New(fmt.Sprintf("query error: neither Node nor []Node: '%s'", name))
		}

		// type assertion
		if nodeType != "" {
			if cast, isNode := base.(Node); isNode {
				t := TypeNameOf(cast)
				if strings.ToLower(t) != nodeType {
					return nil, errors.New(fmt.Sprintf("type assertion error, expect: %s, actual: %s", nodeType, t))
				}
			} else {
				return nil, errors.New(fmt.Sprintf("type assertion error, not node"))
			}
		}
	}
	return base, nil
}

func ParseFile(filePath string) (Node, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	r, _ := DecodeBytes(b)
	tokenizer := NewTokenizer(filePath, r)
	var tokens []*Token
	tokens, err = tokenizer.Parse()
	if err != nil {
		return nil, err
	}
	tokens = tokenizer.Clean(tokens)
	parser := NewParser(filePath, r, tokens)
	var ret Node
	ret, err = parser.Parse()
	if err != nil {
		return nil, err
	}
	if ret != nil {
		ret.BuildLink()
	}
	return ret, nil
}

func ParseBytes(filePath string, b []byte) (Node, error) {
	var err error
	r, _ := DecodeBytes(b)
	tokenizer := NewTokenizer(filePath, r)
	var tokens []*Token
	tokens, err = tokenizer.Parse()
	if err != nil {
		return nil, err
	}
	tokens = tokenizer.Clean(tokens)
	parser := NewParser(filePath, r, tokens)
	var ret Node
	ret, err = parser.Parse()
	if err != nil {
		return nil, err
	}
	if ret != nil {
		ret.BuildLink()
	}
	return ret, nil
}
