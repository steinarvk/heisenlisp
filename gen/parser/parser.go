
  package parser

  import (
    "strings"
    "io"
    "unicode"
    "strconv"
    "errors"
    "os"
    "io/ioutil"
    "bytes"
    "fmt"
    "unicode/utf8"

    "github.com/steinarvk/heisenlisp/value/str"
    "github.com/steinarvk/heisenlisp/value/unknowns/fullyunknown"
    "github.com/steinarvk/heisenlisp/types"
    "github.com/steinarvk/heisenlisp/number"

    sexpr "github.com/steinarvk/heisenlisp/expr"
  )

var g = &grammar {
	rules: []*rule{
{
	name: "MultiExpr",
	pos: position{line: 25, col: 1, offset: 422},
	expr: &actionExpr{
	pos: position{line: 25, col: 14, offset: 435},
	run: (*parser).callonMultiExpr1,
	expr: &seqExpr{
	pos: position{line: 25, col: 14, offset: 435},
	exprs: []interface{}{
&labeledExpr{
	pos: position{line: 25, col: 14, offset: 435},
	label: "rv",
	expr: &oneOrMoreExpr{
	pos: position{line: 25, col: 17, offset: 438},
	expr: &ruleRefExpr{
	pos: position{line: 25, col: 17, offset: 438},
	name: "WhitespaceThenExpr",
},
},
},
&ruleRefExpr{
	pos: position{line: 25, col: 37, offset: 458},
	name: "_",
},
&ruleRefExpr{
	pos: position{line: 25, col: 39, offset: 460},
	name: "EOF",
},
	},
},
},
},
{
	name: "SingleExpr",
	pos: position{line: 29, col: 1, offset: 486},
	expr: &actionExpr{
	pos: position{line: 29, col: 15, offset: 500},
	run: (*parser).callonSingleExpr1,
	expr: &seqExpr{
	pos: position{line: 29, col: 15, offset: 500},
	exprs: []interface{}{
&labeledExpr{
	pos: position{line: 29, col: 15, offset: 500},
	label: "rv",
	expr: &ruleRefExpr{
	pos: position{line: 29, col: 18, offset: 503},
	name: "Expr",
},
},
&ruleRefExpr{
	pos: position{line: 29, col: 23, offset: 508},
	name: "_",
},
&ruleRefExpr{
	pos: position{line: 29, col: 25, offset: 510},
	name: "EOF",
},
	},
},
},
},
{
	name: "Expr",
	pos: position{line: 33, col: 1, offset: 536},
	expr: &choiceExpr{
	pos: position{line: 34, col: 5, offset: 550},
	alternatives: []interface{}{
&ruleRefExpr{
	pos: position{line: 34, col: 5, offset: 550},
	name: "ListExpr",
},
&ruleRefExpr{
	pos: position{line: 35, col: 5, offset: 563},
	name: "String",
},
&ruleRefExpr{
	pos: position{line: 36, col: 5, offset: 574},
	name: "QuotingExpr",
},
&ruleRefExpr{
	pos: position{line: 37, col: 5, offset: 590},
	name: "Real",
},
&ruleRefExpr{
	pos: position{line: 38, col: 5, offset: 599},
	name: "Rational",
},
&ruleRefExpr{
	pos: position{line: 39, col: 5, offset: 612},
	name: "Integer",
},
&ruleRefExpr{
	pos: position{line: 40, col: 5, offset: 624},
	name: "Identifier",
},
&ruleRefExpr{
	pos: position{line: 41, col: 5, offset: 639},
	name: "Unknown",
},
	},
},
},
{
	name: "Unknown",
	pos: position{line: 44, col: 1, offset: 650},
	expr: &actionExpr{
	pos: position{line: 44, col: 12, offset: 661},
	run: (*parser).callonUnknown1,
	expr: &litMatcher{
	pos: position{line: 44, col: 12, offset: 661},
	val: "#unknown",
	ignoreCase: false,
},
},
},
{
	name: "LPAREN",
	pos: position{line: 47, col: 1, offset: 709},
	expr: &litMatcher{
	pos: position{line: 47, col: 11, offset: 719},
	val: "(",
	ignoreCase: false,
},
},
{
	name: "RPAREN",
	pos: position{line: 48, col: 1, offset: 723},
	expr: &litMatcher{
	pos: position{line: 48, col: 11, offset: 733},
	val: ")",
	ignoreCase: false,
},
},
{
	name: "oneWhitespace",
	pos: position{line: 50, col: 1, offset: 738},
	expr: &choiceExpr{
	pos: position{line: 50, col: 20, offset: 757},
	alternatives: []interface{}{
&charClassMatcher{
	pos: position{line: 50, col: 20, offset: 757},
	val: "[ \\t\\r\\n]",
	chars: []rune{' ','\t','\r','\n',},
	ignoreCase: false,
	inverted: false,
},
&ruleRefExpr{
	pos: position{line: 50, col: 32, offset: 769},
	name: "comment",
},
	},
},
},
{
	name: "comment",
	pos: position{line: 51, col: 1, offset: 779},
	expr: &seqExpr{
	pos: position{line: 51, col: 12, offset: 790},
	exprs: []interface{}{
&litMatcher{
	pos: position{line: 51, col: 12, offset: 790},
	val: ";;",
	ignoreCase: false,
},
&zeroOrMoreExpr{
	pos: position{line: 51, col: 17, offset: 795},
	expr: &charClassMatcher{
	pos: position{line: 51, col: 17, offset: 795},
	val: "[^\\n]",
	chars: []rune{'\n',},
	ignoreCase: false,
	inverted: true,
},
},
	},
},
},
{
	name: "sp",
	displayName: "\"mandatory whitespace\"",
	pos: position{line: 53, col: 1, offset: 803},
	expr: &oneOrMoreExpr{
	pos: position{line: 53, col: 30, offset: 832},
	expr: &ruleRefExpr{
	pos: position{line: 53, col: 30, offset: 832},
	name: "oneWhitespace",
},
},
},
{
	name: "_",
	displayName: "\"whitespace\"",
	pos: position{line: 54, col: 1, offset: 847},
	expr: &zeroOrMoreExpr{
	pos: position{line: 54, col: 19, offset: 865},
	expr: &ruleRefExpr{
	pos: position{line: 54, col: 19, offset: 865},
	name: "oneWhitespace",
},
},
},
{
	name: "EscapedChar",
	pos: position{line: 56, col: 1, offset: 881},
	expr: &charClassMatcher{
	pos: position{line: 56, col: 16, offset: 896},
	val: "[\\x00-\\x1f\"\\\\]",
	chars: []rune{'"','\\',},
	ranges: []rune{'\x00','\x1f',},
	ignoreCase: false,
	inverted: false,
},
},
{
	name: "EscapeSequence",
	pos: position{line: 57, col: 1, offset: 911},
	expr: &choiceExpr{
	pos: position{line: 57, col: 19, offset: 929},
	alternatives: []interface{}{
&ruleRefExpr{
	pos: position{line: 57, col: 19, offset: 929},
	name: "SingleCharEscape",
},
&ruleRefExpr{
	pos: position{line: 57, col: 38, offset: 948},
	name: "UnicodeEscape",
},
	},
},
},
{
	name: "SingleCharEscape",
	pos: position{line: 58, col: 1, offset: 962},
	expr: &charClassMatcher{
	pos: position{line: 58, col: 21, offset: 982},
	val: "[\"\\\\/bfnrt]",
	chars: []rune{'"','\\','/','b','f','n','r','t',},
	ignoreCase: false,
	inverted: false,
},
},
{
	name: "UnicodeEscape",
	pos: position{line: 59, col: 1, offset: 994},
	expr: &seqExpr{
	pos: position{line: 59, col: 18, offset: 1011},
	exprs: []interface{}{
&litMatcher{
	pos: position{line: 59, col: 18, offset: 1011},
	val: "u",
	ignoreCase: false,
},
&ruleRefExpr{
	pos: position{line: 59, col: 22, offset: 1015},
	name: "HexDigit",
},
&ruleRefExpr{
	pos: position{line: 59, col: 31, offset: 1024},
	name: "HexDigit",
},
&ruleRefExpr{
	pos: position{line: 59, col: 40, offset: 1033},
	name: "HexDigit",
},
&ruleRefExpr{
	pos: position{line: 59, col: 49, offset: 1042},
	name: "HexDigit",
},
	},
},
},
{
	name: "HexDigit",
	pos: position{line: 60, col: 1, offset: 1051},
	expr: &charClassMatcher{
	pos: position{line: 60, col: 13, offset: 1063},
	val: "[0-9a-f]i",
	ranges: []rune{'0','9','a','f',},
	ignoreCase: true,
	inverted: false,
},
},
{
	name: "Identifier",
	pos: position{line: 62, col: 1, offset: 1074},
	expr: &actionExpr{
	pos: position{line: 62, col: 15, offset: 1088},
	run: (*parser).callonIdentifier1,
	expr: &seqExpr{
	pos: position{line: 62, col: 15, offset: 1088},
	exprs: []interface{}{
&charClassMatcher{
	pos: position{line: 62, col: 15, offset: 1088},
	val: "[a-zA-Z?!+/*.=_&<>-]",
	chars: []rune{'?','!','+','/','*','.','=','_','&','<','>','-',},
	ranges: []rune{'a','z','A','Z',},
	ignoreCase: false,
	inverted: false,
},
&zeroOrMoreExpr{
	pos: position{line: 62, col: 36, offset: 1109},
	expr: &charClassMatcher{
	pos: position{line: 62, col: 36, offset: 1109},
	val: "[a-zA-Z0-9?!+/*.=&<>-]",
	chars: []rune{'?','!','+','/','*','.','=','&','<','>','-',},
	ranges: []rune{'a','z','A','Z','0','9',},
	ignoreCase: false,
	inverted: false,
},
},
	},
},
},
},
{
	name: "String",
	pos: position{line: 66, col: 1, offset: 1183},
	expr: &actionExpr{
	pos: position{line: 66, col: 11, offset: 1193},
	run: (*parser).callonString1,
	expr: &seqExpr{
	pos: position{line: 66, col: 11, offset: 1193},
	exprs: []interface{}{
&litMatcher{
	pos: position{line: 66, col: 11, offset: 1193},
	val: "\"",
	ignoreCase: false,
},
&zeroOrMoreExpr{
	pos: position{line: 66, col: 15, offset: 1197},
	expr: &choiceExpr{
	pos: position{line: 66, col: 17, offset: 1199},
	alternatives: []interface{}{
&seqExpr{
	pos: position{line: 66, col: 17, offset: 1199},
	exprs: []interface{}{
&notExpr{
	pos: position{line: 66, col: 17, offset: 1199},
	expr: &ruleRefExpr{
	pos: position{line: 66, col: 18, offset: 1200},
	name: "EscapedChar",
},
},
&anyMatcher{
	line: 66, col: 30, offset: 1212,
},
	},
},
&seqExpr{
	pos: position{line: 66, col: 34, offset: 1216},
	exprs: []interface{}{
&litMatcher{
	pos: position{line: 66, col: 34, offset: 1216},
	val: "\\",
	ignoreCase: false,
},
&ruleRefExpr{
	pos: position{line: 66, col: 39, offset: 1221},
	name: "EscapeSequence",
},
	},
},
	},
},
},
&litMatcher{
	pos: position{line: 66, col: 57, offset: 1239},
	val: "\"",
	ignoreCase: false,
},
	},
},
},
},
{
	name: "Real",
	pos: position{line: 74, col: 1, offset: 1359},
	expr: &actionExpr{
	pos: position{line: 74, col: 9, offset: 1367},
	run: (*parser).callonReal1,
	expr: &choiceExpr{
	pos: position{line: 75, col: 5, offset: 1373},
	alternatives: []interface{}{
&seqExpr{
	pos: position{line: 75, col: 7, offset: 1375},
	exprs: []interface{}{
&zeroOrOneExpr{
	pos: position{line: 75, col: 7, offset: 1375},
	expr: &litMatcher{
	pos: position{line: 75, col: 7, offset: 1375},
	val: "-",
	ignoreCase: false,
},
},
&charClassMatcher{
	pos: position{line: 75, col: 12, offset: 1380},
	val: "[0-9]",
	ranges: []rune{'0','9',},
	ignoreCase: false,
	inverted: false,
},
&zeroOrMoreExpr{
	pos: position{line: 75, col: 18, offset: 1386},
	expr: &charClassMatcher{
	pos: position{line: 75, col: 18, offset: 1386},
	val: "[0-9]",
	ranges: []rune{'0','9',},
	ignoreCase: false,
	inverted: false,
},
},
&litMatcher{
	pos: position{line: 75, col: 25, offset: 1393},
	val: ".",
	ignoreCase: false,
},
&zeroOrMoreExpr{
	pos: position{line: 75, col: 29, offset: 1397},
	expr: &charClassMatcher{
	pos: position{line: 75, col: 29, offset: 1397},
	val: "[0-9]",
	ranges: []rune{'0','9',},
	ignoreCase: false,
	inverted: false,
},
},
	},
},
&seqExpr{
	pos: position{line: 76, col: 7, offset: 1412},
	exprs: []interface{}{
&zeroOrOneExpr{
	pos: position{line: 76, col: 7, offset: 1412},
	expr: &litMatcher{
	pos: position{line: 76, col: 7, offset: 1412},
	val: "-",
	ignoreCase: false,
},
},
&charClassMatcher{
	pos: position{line: 76, col: 12, offset: 1417},
	val: "[0-9]",
	ranges: []rune{'0','9',},
	ignoreCase: false,
	inverted: false,
},
&zeroOrMoreExpr{
	pos: position{line: 76, col: 18, offset: 1423},
	expr: &charClassMatcher{
	pos: position{line: 76, col: 18, offset: 1423},
	val: "[0-9]",
	ranges: []rune{'0','9',},
	ignoreCase: false,
	inverted: false,
},
},
&zeroOrOneExpr{
	pos: position{line: 76, col: 25, offset: 1430},
	expr: &seqExpr{
	pos: position{line: 76, col: 26, offset: 1431},
	exprs: []interface{}{
&litMatcher{
	pos: position{line: 76, col: 26, offset: 1431},
	val: ".",
	ignoreCase: false,
},
&zeroOrMoreExpr{
	pos: position{line: 76, col: 30, offset: 1435},
	expr: &charClassMatcher{
	pos: position{line: 76, col: 30, offset: 1435},
	val: "[0-9]",
	ranges: []rune{'0','9',},
	ignoreCase: false,
	inverted: false,
},
},
	},
},
},
&charClassMatcher{
	pos: position{line: 76, col: 39, offset: 1444},
	val: "[eE]",
	chars: []rune{'e','E',},
	ignoreCase: false,
	inverted: false,
},
&zeroOrOneExpr{
	pos: position{line: 76, col: 44, offset: 1449},
	expr: &litMatcher{
	pos: position{line: 76, col: 44, offset: 1449},
	val: "-",
	ignoreCase: false,
},
},
&oneOrMoreExpr{
	pos: position{line: 76, col: 49, offset: 1454},
	expr: &charClassMatcher{
	pos: position{line: 76, col: 49, offset: 1454},
	val: "[0-9]",
	ranges: []rune{'0','9',},
	ignoreCase: false,
	inverted: false,
},
},
	},
},
	},
},
},
},
{
	name: "Rational",
	pos: position{line: 81, col: 1, offset: 1513},
	expr: &actionExpr{
	pos: position{line: 81, col: 13, offset: 1525},
	run: (*parser).callonRational1,
	expr: &seqExpr{
	pos: position{line: 81, col: 15, offset: 1527},
	exprs: []interface{}{
&zeroOrOneExpr{
	pos: position{line: 81, col: 15, offset: 1527},
	expr: &litMatcher{
	pos: position{line: 81, col: 15, offset: 1527},
	val: "-",
	ignoreCase: false,
},
},
&oneOrMoreExpr{
	pos: position{line: 81, col: 20, offset: 1532},
	expr: &charClassMatcher{
	pos: position{line: 81, col: 20, offset: 1532},
	val: "[0-9]",
	ranges: []rune{'0','9',},
	ignoreCase: false,
	inverted: false,
},
},
&litMatcher{
	pos: position{line: 81, col: 27, offset: 1539},
	val: "/",
	ignoreCase: false,
},
&oneOrMoreExpr{
	pos: position{line: 81, col: 31, offset: 1543},
	expr: &charClassMatcher{
	pos: position{line: 81, col: 31, offset: 1543},
	val: "[0-9]",
	ranges: []rune{'0','9',},
	ignoreCase: false,
	inverted: false,
},
},
	},
},
},
},
{
	name: "Integer",
	pos: position{line: 85, col: 1, offset: 1600},
	expr: &actionExpr{
	pos: position{line: 85, col: 12, offset: 1611},
	run: (*parser).callonInteger1,
	expr: &choiceExpr{
	pos: position{line: 85, col: 14, offset: 1613},
	alternatives: []interface{}{
&litMatcher{
	pos: position{line: 85, col: 14, offset: 1613},
	val: "0",
	ignoreCase: false,
},
&seqExpr{
	pos: position{line: 85, col: 20, offset: 1619},
	exprs: []interface{}{
&zeroOrOneExpr{
	pos: position{line: 85, col: 20, offset: 1619},
	expr: &litMatcher{
	pos: position{line: 85, col: 20, offset: 1619},
	val: "-",
	ignoreCase: false,
},
},
&charClassMatcher{
	pos: position{line: 85, col: 25, offset: 1624},
	val: "[1-9]",
	ranges: []rune{'1','9',},
	ignoreCase: false,
	inverted: false,
},
&zeroOrMoreExpr{
	pos: position{line: 85, col: 31, offset: 1630},
	expr: &charClassMatcher{
	pos: position{line: 85, col: 31, offset: 1630},
	val: "[0-9]",
	ranges: []rune{'0','9',},
	ignoreCase: false,
	inverted: false,
},
},
	},
},
	},
},
},
},
{
	name: "WhitespaceThenExpr",
	pos: position{line: 89, col: 1, offset: 1687},
	expr: &actionExpr{
	pos: position{line: 89, col: 23, offset: 1709},
	run: (*parser).callonWhitespaceThenExpr1,
	expr: &seqExpr{
	pos: position{line: 89, col: 23, offset: 1709},
	exprs: []interface{}{
&ruleRefExpr{
	pos: position{line: 89, col: 23, offset: 1709},
	name: "_",
},
&labeledExpr{
	pos: position{line: 89, col: 25, offset: 1711},
	label: "rv",
	expr: &ruleRefExpr{
	pos: position{line: 89, col: 28, offset: 1714},
	name: "Expr",
},
},
	},
},
},
},
{
	name: "ListExpr",
	pos: position{line: 93, col: 1, offset: 1741},
	expr: &actionExpr{
	pos: position{line: 93, col: 13, offset: 1753},
	run: (*parser).callonListExpr1,
	expr: &seqExpr{
	pos: position{line: 93, col: 13, offset: 1753},
	exprs: []interface{}{
&ruleRefExpr{
	pos: position{line: 93, col: 13, offset: 1753},
	name: "_",
},
&ruleRefExpr{
	pos: position{line: 93, col: 15, offset: 1755},
	name: "LPAREN",
},
&ruleRefExpr{
	pos: position{line: 93, col: 22, offset: 1762},
	name: "_",
},
&labeledExpr{
	pos: position{line: 93, col: 24, offset: 1764},
	label: "more",
	expr: &zeroOrMoreExpr{
	pos: position{line: 93, col: 29, offset: 1769},
	expr: &ruleRefExpr{
	pos: position{line: 93, col: 29, offset: 1769},
	name: "WhitespaceThenExpr",
},
},
},
&ruleRefExpr{
	pos: position{line: 93, col: 49, offset: 1789},
	name: "_",
},
&ruleRefExpr{
	pos: position{line: 93, col: 51, offset: 1791},
	name: "RPAREN",
},
	},
},
},
},
{
	name: "QuotingExpr",
	pos: position{line: 103, col: 1, offset: 1957},
	expr: &choiceExpr{
	pos: position{line: 104, col: 5, offset: 1978},
	alternatives: []interface{}{
&ruleRefExpr{
	pos: position{line: 104, col: 5, offset: 1978},
	name: "QuotedExpr",
},
&ruleRefExpr{
	pos: position{line: 105, col: 5, offset: 1993},
	name: "QuasiQuotedExpr",
},
&ruleRefExpr{
	pos: position{line: 106, col: 5, offset: 2013},
	name: "SplicingUnquotedExpr",
},
&ruleRefExpr{
	pos: position{line: 107, col: 5, offset: 2038},
	name: "UnquotedExpr",
},
	},
},
},
{
	name: "QuotedExpr",
	pos: position{line: 110, col: 1, offset: 2054},
	expr: &actionExpr{
	pos: position{line: 110, col: 15, offset: 2068},
	run: (*parser).callonQuotedExpr1,
	expr: &seqExpr{
	pos: position{line: 110, col: 15, offset: 2068},
	exprs: []interface{}{
&litMatcher{
	pos: position{line: 110, col: 15, offset: 2068},
	val: "'",
	ignoreCase: false,
},
&ruleRefExpr{
	pos: position{line: 110, col: 19, offset: 2072},
	name: "_",
},
&labeledExpr{
	pos: position{line: 110, col: 21, offset: 2074},
	label: "rv",
	expr: &ruleRefExpr{
	pos: position{line: 110, col: 24, offset: 2077},
	name: "Expr",
},
},
	},
},
},
},
{
	name: "QuasiQuotedExpr",
	pos: position{line: 114, col: 1, offset: 2146},
	expr: &actionExpr{
	pos: position{line: 114, col: 20, offset: 2165},
	run: (*parser).callonQuasiQuotedExpr1,
	expr: &seqExpr{
	pos: position{line: 114, col: 20, offset: 2165},
	exprs: []interface{}{
&litMatcher{
	pos: position{line: 114, col: 20, offset: 2165},
	val: "`",
	ignoreCase: false,
},
&ruleRefExpr{
	pos: position{line: 114, col: 24, offset: 2169},
	name: "_",
},
&labeledExpr{
	pos: position{line: 114, col: 26, offset: 2171},
	label: "rv",
	expr: &ruleRefExpr{
	pos: position{line: 114, col: 29, offset: 2174},
	name: "Expr",
},
},
	},
},
},
},
{
	name: "UnquotedExpr",
	pos: position{line: 118, col: 1, offset: 2248},
	expr: &actionExpr{
	pos: position{line: 118, col: 17, offset: 2264},
	run: (*parser).callonUnquotedExpr1,
	expr: &seqExpr{
	pos: position{line: 118, col: 17, offset: 2264},
	exprs: []interface{}{
&litMatcher{
	pos: position{line: 118, col: 17, offset: 2264},
	val: ",",
	ignoreCase: false,
},
&ruleRefExpr{
	pos: position{line: 118, col: 21, offset: 2268},
	name: "_",
},
&labeledExpr{
	pos: position{line: 118, col: 23, offset: 2270},
	label: "rv",
	expr: &ruleRefExpr{
	pos: position{line: 118, col: 26, offset: 2273},
	name: "Expr",
},
},
	},
},
},
},
{
	name: "SplicingUnquotedExpr",
	pos: position{line: 122, col: 1, offset: 2344},
	expr: &actionExpr{
	pos: position{line: 122, col: 25, offset: 2368},
	run: (*parser).callonSplicingUnquotedExpr1,
	expr: &seqExpr{
	pos: position{line: 122, col: 25, offset: 2368},
	exprs: []interface{}{
&litMatcher{
	pos: position{line: 122, col: 25, offset: 2368},
	val: ",@",
	ignoreCase: false,
},
&ruleRefExpr{
	pos: position{line: 122, col: 30, offset: 2373},
	name: "_",
},
&labeledExpr{
	pos: position{line: 122, col: 32, offset: 2375},
	label: "rv",
	expr: &ruleRefExpr{
	pos: position{line: 122, col: 35, offset: 2378},
	name: "Expr",
},
},
	},
},
},
},
{
	name: "EOF",
	pos: position{line: 126, col: 1, offset: 2458},
	expr: &notExpr{
	pos: position{line: 126, col: 8, offset: 2465},
	expr: &anyMatcher{
	line: 126, col: 9, offset: 2466,
},
},
},
	},
}
func (c *current) onMultiExpr1(rv interface{}) (interface{}, error) {
  return rv, nil
}

func (p *parser) callonMultiExpr1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onMultiExpr1(stack["rv"])
}

func (c *current) onSingleExpr1(rv interface{}) (interface{}, error) {
  return rv, nil
}

func (p *parser) callonSingleExpr1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onSingleExpr1(stack["rv"])
}

func (c *current) onUnknown1() (interface{}, error) {
 return fullyunknown.Value, nil 
}

func (p *parser) callonUnknown1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onUnknown1()
}

func (c *current) onIdentifier1() (interface{}, error) {
  return sexpr.ToSymbol(string(c.text)), nil
}

func (p *parser) callonIdentifier1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onIdentifier1()
}

func (c *current) onString1() (interface{}, error) {
  s, err := strconv.Unquote(string(c.text))
  if err != nil {
    return nil, err
  }
  return str.New(s), nil
}

func (p *parser) callonString1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onString1()
}

func (c *current) onReal1() (interface{}, error) {
  return number.FromString(string(c.text))
}

func (p *parser) callonReal1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onReal1()
}

func (c *current) onRational1() (interface{}, error) {
  return number.FromString(string(c.text))
}

func (p *parser) callonRational1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onRational1()
}

func (c *current) onInteger1() (interface{}, error) {
  return number.FromString(string(c.text))
}

func (p *parser) callonInteger1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onInteger1()
}

func (c *current) onWhitespaceThenExpr1(rv interface{}) (interface{}, error) {
  return rv, nil
}

func (p *parser) callonWhitespaceThenExpr1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onWhitespaceThenExpr1(stack["rv"])
}

func (c *current) onListExpr1(more interface{}) (interface{}, error) {
  var rv []types.Value

  for _, another := range more.([]interface{}) {
    rv = append(rv, another.(types.Value))
  }

  return sexpr.WrapList(rv), nil
}

func (p *parser) callonListExpr1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onListExpr1(stack["more"])
}

func (c *current) onQuotedExpr1(rv interface{}) (interface{}, error) {
  return sexpr.WrapInUnary("quote", rv.(types.Value)), nil
}

func (p *parser) callonQuotedExpr1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onQuotedExpr1(stack["rv"])
}

func (c *current) onQuasiQuotedExpr1(rv interface{}) (interface{}, error) {
  return sexpr.WrapInUnary("quasiquote", rv.(types.Value)), nil
}

func (p *parser) callonQuasiQuotedExpr1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onQuasiQuotedExpr1(stack["rv"])
}

func (c *current) onUnquotedExpr1(rv interface{}) (interface{}, error) {
  return sexpr.WrapInUnary("unquote", rv.(types.Value)), nil
}

func (p *parser) callonUnquotedExpr1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onUnquotedExpr1(stack["rv"])
}

func (c *current) onSplicingUnquotedExpr1(rv interface{}) (interface{}, error) {
  return sexpr.WrapInUnary("unquote-splicing", rv.(types.Value)), nil
}

func (p *parser) callonSplicingUnquotedExpr1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onSplicingUnquotedExpr1(stack["rv"])
}


var (
	// errNoRule is returned when the grammar to parse has no rule.
	errNoRule          = errors.New("grammar has no rule")

	// errInvalidEncoding is returned when the source is not properly
	// utf8-encoded.
	errInvalidEncoding = errors.New("invalid encoding")

	// errNoMatch is returned if no match could be found.
	errNoMatch         = errors.New("no match found")
)

// Option is a function that can set an option on the parser. It returns
// the previous setting as an Option.
type Option func(*parser) Option

// Debug creates an Option to set the debug flag to b. When set to true,
// debugging information is printed to stdout while parsing.
//
// The default is false.
func Debug(b bool) Option {
	return func(p *parser) Option {
		old := p.debug
		p.debug = b
		return Debug(old)
	}
}

// Memoize creates an Option to set the memoize flag to b. When set to true,
// the parser will cache all results so each expression is evaluated only
// once. This guarantees linear parsing time even for pathological cases,
// at the expense of more memory and slower times for typical cases.
//
// The default is false.
func Memoize(b bool) Option {
	return func(p *parser) Option {
		old := p.memoize
		p.memoize = b
		return Memoize(old)
	}
}

// Recover creates an Option to set the recover flag to b. When set to
// true, this causes the parser to recover from panics and convert it
// to an error. Setting it to false can be useful while debugging to
// access the full stack trace.
//
// The default is true.
func Recover(b bool) Option {
	return func(p *parser) Option {
		old := p.recover
		p.recover = b
		return Recover(old)
	}
}

// ParseFile parses the file identified by filename.
func ParseFile(filename string, opts ...Option) (interface{}, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseReader(filename, f, opts...)
}

// ParseReader parses the data from r using filename as information in the
// error messages.
func ParseReader(filename string, r io.Reader, opts ...Option) (interface{}, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return Parse(filename, b, opts...)
}

// Parse parses the data from b using filename as information in the
// error messages.
func Parse(filename string, b []byte, opts ...Option) (interface{}, error) {
	return newParser(filename, b, opts...).parse(g)
}

// position records a position in the text.
type position struct {
	line, col, offset int
}

func (p position) String() string {
	return fmt.Sprintf("%d:%d [%d]", p.line, p.col, p.offset)
}

// savepoint stores all state required to go back to this point in the
// parser.
type savepoint struct {
	position
	rn rune
	w  int
}

type current struct {
	pos  position // start position of the match
	text []byte   // raw text of the match
}

// the AST types...

type grammar struct {
	pos   position
	rules []*rule
}

type rule struct {
	pos         position
	name        string
	displayName string
	expr        interface{}
}

type choiceExpr struct {
	pos          position
	alternatives []interface{}
}

type actionExpr struct {
	pos    position
	expr   interface{}
	run    func(*parser) (interface{}, error)
}

type seqExpr struct {
	pos   position
	exprs []interface{}
}

type labeledExpr struct {
	pos   position
	label string
	expr  interface{}
}

type expr struct {
	pos  position
	expr interface{}
}

type andExpr expr
type notExpr expr
type zeroOrOneExpr expr
type zeroOrMoreExpr expr
type oneOrMoreExpr expr

type ruleRefExpr struct {
	pos  position
	name string
}

type andCodeExpr struct {
	pos position
	run func(*parser) (bool, error)
}

type notCodeExpr struct {
	pos position
	run func(*parser) (bool, error)
}

type litMatcher struct {
	pos        position
	val        string
	ignoreCase bool
}

type charClassMatcher struct {
	pos        position
	val        string
	chars      []rune
	ranges     []rune
	classes    []*unicode.RangeTable
	ignoreCase bool
	inverted   bool
}

type anyMatcher position

// errList cumulates the errors found by the parser.
type errList []error

func (e *errList) add(err error) {
	*e = append(*e, err)
}

func (e errList) err() error {
	if len(e) == 0 {
		return nil
	}
	e.dedupe()
	return e
}

func (e *errList) dedupe() {
	var cleaned []error
	set := make(map[string]bool)
	for _, err := range *e {
		if msg := err.Error(); !set[msg] {
			set[msg] = true
			cleaned = append(cleaned, err)
		}
	}
	*e = cleaned
}

func (e errList) Error() string {
	switch len(e) {
	case 0:
		return ""
	case 1:
		return e[0].Error()
	default:
		var buf bytes.Buffer

		for i, err := range e {
			if i > 0 {
				buf.WriteRune('\n')
			}
			buf.WriteString(err.Error())
		}
		return buf.String()
	}
}

// parserError wraps an error with a prefix indicating the rule in which
// the error occurred. The original error is stored in the Inner field.
type parserError struct {
	Inner  error
	pos    position
	prefix string
}

// Error returns the error message.
func (p *parserError) Error() string {
	return p.prefix + ": " + p.Inner.Error()
}

// newParser creates a parser with the specified input source and options.
func newParser(filename string, b []byte, opts ...Option) *parser {
	p := &parser{
		filename: filename,
		errs: new(errList),
		data: b,
		pt: savepoint{position: position{line: 1}},
		recover: true,
	}
	p.setOptions(opts)
	return p
}

// setOptions applies the options to the parser.
func (p *parser) setOptions(opts []Option) {
	for _, opt := range opts {
		opt(p)
	}
}

type resultTuple struct {
	v interface{}
	b bool
	end savepoint
}

type parser struct {
	filename string
	pt       savepoint
	cur      current

	data []byte
	errs *errList

	recover bool
	debug bool
	depth  int

	memoize bool
	// memoization table for the packrat algorithm:
	// map[offset in source] map[expression or rule] {value, match}
	memo map[int]map[interface{}]resultTuple

	// rules table, maps the rule identifier to the rule node
	rules  map[string]*rule
	// variables stack, map of label to value
	vstack []map[string]interface{}
	// rule stack, allows identification of the current rule in errors
	rstack []*rule

	// stats
	exprCnt int
}

// push a variable set on the vstack.
func (p *parser) pushV() {
	if cap(p.vstack) == len(p.vstack) {
		// create new empty slot in the stack
		p.vstack = append(p.vstack, nil)
	} else {
		// slice to 1 more
		p.vstack = p.vstack[:len(p.vstack)+1]
	}

	// get the last args set
	m := p.vstack[len(p.vstack)-1]
	if m != nil && len(m) == 0 {
		// empty map, all good
		return
	}

	m = make(map[string]interface{})
	p.vstack[len(p.vstack)-1] = m
}

// pop a variable set from the vstack.
func (p *parser) popV() {
	// if the map is not empty, clear it
	m := p.vstack[len(p.vstack)-1]
	if len(m) > 0 {
		// GC that map
		p.vstack[len(p.vstack)-1] = nil
	}
	p.vstack = p.vstack[:len(p.vstack)-1]
}

func (p *parser) print(prefix, s string) string {
	if !p.debug {
		return s
	}

	fmt.Printf("%s %d:%d:%d: %s [%#U]\n",
		prefix, p.pt.line, p.pt.col, p.pt.offset, s, p.pt.rn)
	return s
}

func (p *parser) in(s string) string {
	p.depth++
	return p.print(strings.Repeat(" ", p.depth) + ">", s)
}

func (p *parser) out(s string) string {
	p.depth--
	return p.print(strings.Repeat(" ", p.depth) + "<", s)
}

func (p *parser) addErr(err error) {
	p.addErrAt(err, p.pt.position)
}

func (p *parser) addErrAt(err error, pos position) {
	var buf bytes.Buffer
	if p.filename != "" {
		buf.WriteString(p.filename)
	}
	if buf.Len() > 0 {
		buf.WriteString(":")
	}
	buf.WriteString(fmt.Sprintf("%d:%d (%d)", pos.line, pos.col, pos.offset))
	if len(p.rstack) > 0 {
		if buf.Len() > 0 {
			buf.WriteString(": ")
		}
		rule := p.rstack[len(p.rstack)-1]
		if rule.displayName != "" {
			buf.WriteString("rule " + rule.displayName)
		} else {
			buf.WriteString("rule " + rule.name)
		}
	}
	pe := &parserError{Inner: err, pos: pos, prefix: buf.String()}
	p.errs.add(pe)
}

// read advances the parser to the next rune.
func (p *parser) read() {
	p.pt.offset += p.pt.w
	rn, n := utf8.DecodeRune(p.data[p.pt.offset:])
	p.pt.rn = rn
	p.pt.w = n
	p.pt.col++
	if rn == '\n' {
		p.pt.line++
		p.pt.col = 0
	}

	if rn == utf8.RuneError {
		if n == 1 {
			p.addErr(errInvalidEncoding)
		}
	}
}

// restore parser position to the savepoint pt.
func (p *parser) restore(pt savepoint) {
	if p.debug {
		defer p.out(p.in("restore"))
	}
	if pt.offset == p.pt.offset {
		return
	}
	p.pt = pt
}

// get the slice of bytes from the savepoint start to the current position.
func (p *parser) sliceFrom(start savepoint) []byte {
	return p.data[start.position.offset:p.pt.position.offset]
}

func (p *parser) getMemoized(node interface{}) (resultTuple, bool) {
	if len(p.memo) == 0 {
		return resultTuple{}, false
	}
	m := p.memo[p.pt.offset]
	if len(m) == 0 {
		return resultTuple{}, false
	}
	res, ok := m[node]
	return res, ok
}

func (p *parser) setMemoized(pt savepoint, node interface{}, tuple resultTuple) {
	if p.memo == nil {
		p.memo = make(map[int]map[interface{}]resultTuple)
	}
	m := p.memo[pt.offset]
	if m == nil {
		m = make(map[interface{}]resultTuple)
		p.memo[pt.offset] = m
	}
	m[node] = tuple
}

func (p *parser) buildRulesTable(g *grammar) {
	p.rules = make(map[string]*rule, len(g.rules))
	for _, r := range g.rules {
		p.rules[r.name] = r
	}
}

func (p *parser) parse(g *grammar) (val interface{}, err error) {
	if len(g.rules) == 0 {
		p.addErr(errNoRule)
		return nil, p.errs.err()
	}

	// TODO : not super critical but this could be generated
	p.buildRulesTable(g)

	if p.recover {
		// panic can be used in action code to stop parsing immediately
		// and return the panic as an error.
		defer func() {
			if e := recover(); e != nil {
				if p.debug {
					defer p.out(p.in("panic handler"))
				}
				val = nil
				switch e := e.(type) {
				case error:
					p.addErr(e)
				default:
					p.addErr(fmt.Errorf("%v", e))
				}
				err = p.errs.err()
			}
		}()
	}

	// start rule is rule [0]
	p.read() // advance to first rune
	val, ok := p.parseRule(g.rules[0])
	if !ok {
		if len(*p.errs) == 0 {
			// make sure this doesn't go out silently
			p.addErr(errNoMatch)
		}
		return nil, p.errs.err()
	}
	return val, p.errs.err()
}

func (p *parser) parseRule(rule *rule) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseRule " + rule.name))
	}

	if p.memoize {
		res, ok := p.getMemoized(rule)
		if ok {
			p.restore(res.end)
			return res.v, res.b
		}
	}

	start := p.pt
	p.rstack = append(p.rstack, rule)
	p.pushV()
	val, ok := p.parseExpr(rule.expr)
	p.popV()
	p.rstack = p.rstack[:len(p.rstack)-1]
	if ok && p.debug {
		p.print(strings.Repeat(" ", p.depth) + "MATCH", string(p.sliceFrom(start)))
	}

	if p.memoize {
		p.setMemoized(start, rule, resultTuple{val, ok, p.pt})
	}
	return val, ok
}

func (p *parser) parseExpr(expr interface{}) (interface{}, bool) {
	var pt savepoint
	var ok bool

	if p.memoize {
		res, ok := p.getMemoized(expr)
		if ok {
			p.restore(res.end)
			return res.v, res.b
		}
		pt = p.pt
	}

	p.exprCnt++
	var val interface{}
	switch expr := expr.(type) {
	case *actionExpr:
		val, ok = p.parseActionExpr(expr)
	case *andCodeExpr:
		val, ok = p.parseAndCodeExpr(expr)
	case *andExpr:
		val, ok = p.parseAndExpr(expr)
	case *anyMatcher:
		val, ok = p.parseAnyMatcher(expr)
	case *charClassMatcher:
		val, ok = p.parseCharClassMatcher(expr)
	case *choiceExpr:
		val, ok = p.parseChoiceExpr(expr)
	case *labeledExpr:
		val, ok = p.parseLabeledExpr(expr)
	case *litMatcher:
		val, ok = p.parseLitMatcher(expr)
	case *notCodeExpr:
		val, ok = p.parseNotCodeExpr(expr)
	case *notExpr:
		val, ok = p.parseNotExpr(expr)
	case *oneOrMoreExpr:
		val, ok = p.parseOneOrMoreExpr(expr)
	case *ruleRefExpr:
		val, ok = p.parseRuleRefExpr(expr)
	case *seqExpr:
		val, ok = p.parseSeqExpr(expr)
	case *zeroOrMoreExpr:
		val, ok = p.parseZeroOrMoreExpr(expr)
	case *zeroOrOneExpr:
		val, ok = p.parseZeroOrOneExpr(expr)
	default:
		panic(fmt.Sprintf("unknown expression type %T", expr))
	}
	if p.memoize {
		p.setMemoized(pt, expr, resultTuple{val, ok, p.pt})
	}
	return val, ok
}

func (p *parser) parseActionExpr(act *actionExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseActionExpr"))
	}

	start := p.pt
	val, ok := p.parseExpr(act.expr)
	if ok {
		p.cur.pos = start.position
		p.cur.text = p.sliceFrom(start)
		actVal, err := act.run(p)
		if err != nil {
			p.addErrAt(err, start.position)
		}
		val = actVal
	}
	if ok && p.debug {
		p.print(strings.Repeat(" ", p.depth) + "MATCH", string(p.sliceFrom(start)))
	}
	return val, ok
}

func (p *parser) parseAndCodeExpr(and *andCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAndCodeExpr"))
	}

	ok, err := and.run(p)
	if err != nil {
		p.addErr(err)
	}
	return nil, ok
}

func (p *parser) parseAndExpr(and *andExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAndExpr"))
	}

	pt := p.pt
	p.pushV()
	_, ok := p.parseExpr(and.expr)
	p.popV()
	p.restore(pt)
	return nil, ok
}

func (p *parser) parseAnyMatcher(any *anyMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAnyMatcher"))
	}

	if p.pt.rn != utf8.RuneError {
		start := p.pt
		p.read()
		return p.sliceFrom(start), true
	}
	return nil, false
}

func (p *parser) parseCharClassMatcher(chr *charClassMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseCharClassMatcher"))
	}

	cur := p.pt.rn
	// can't match EOF
	if cur == utf8.RuneError {
		return nil, false
	}
	start := p.pt
	if chr.ignoreCase {
		cur = unicode.ToLower(cur)
	}

	// try to match in the list of available chars
	for _, rn := range chr.chars {
		if rn == cur {
			if chr.inverted {
				return nil, false
			}
			p.read()
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of ranges
	for i := 0; i < len(chr.ranges); i += 2 {
		if cur >= chr.ranges[i] && cur <= chr.ranges[i+1] {
			if chr.inverted {
				return nil, false
			}
			p.read()
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of Unicode classes
	for _, cl := range chr.classes {
		if unicode.Is(cl, cur) {
			if chr.inverted {
				return nil, false
			}
			p.read()
			return p.sliceFrom(start), true
		}
	}

	if chr.inverted {
		p.read()
		return p.sliceFrom(start), true
	}
	return nil, false
}

func (p *parser) parseChoiceExpr(ch *choiceExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseChoiceExpr"))
	}

	for _, alt := range ch.alternatives {
		p.pushV()
		val, ok := p.parseExpr(alt)
		p.popV()
		if ok {
			return val, ok
		}
	}
	return nil, false
}

func (p *parser) parseLabeledExpr(lab *labeledExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseLabeledExpr"))
	}

	p.pushV()
	val, ok := p.parseExpr(lab.expr)
	p.popV()
	if ok && lab.label != "" {
		m := p.vstack[len(p.vstack)-1]
		m[lab.label] = val
	}
	return val, ok
}

func (p *parser) parseLitMatcher(lit *litMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseLitMatcher"))
	}

	start := p.pt
	for _, want := range lit.val {
		cur := p.pt.rn
		if lit.ignoreCase {
			cur = unicode.ToLower(cur)
		}
		if cur != want {
			p.restore(start)
			return nil, false
		}
		p.read()
	}
	return p.sliceFrom(start), true
}

func (p *parser) parseNotCodeExpr(not *notCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseNotCodeExpr"))
	}

	ok, err := not.run(p)
	if err != nil {
		p.addErr(err)
	}
	return nil, !ok
}

func (p *parser) parseNotExpr(not *notExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseNotExpr"))
	}

	pt := p.pt
	p.pushV()
	_, ok := p.parseExpr(not.expr)
	p.popV()
	p.restore(pt)
	return nil, !ok
}

func (p *parser) parseOneOrMoreExpr(expr *oneOrMoreExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseOneOrMoreExpr"))
	}

	var vals []interface{}

	for {
		p.pushV()
		val, ok := p.parseExpr(expr.expr)
		p.popV()
		if !ok {
			if len(vals) == 0 {
				// did not match once, no match
				return nil, false
			}
			return vals, true
		}
		vals = append(vals, val)
	}
}

func (p *parser) parseRuleRefExpr(ref *ruleRefExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseRuleRefExpr " + ref.name))
	}

	if ref.name == "" {
		panic(fmt.Sprintf("%s: invalid rule: missing name", ref.pos))
	}

	rule := p.rules[ref.name]
	if rule == nil {
		p.addErr(fmt.Errorf("undefined rule: %s", ref.name))
		return nil, false
	}
	return p.parseRule(rule)
}

func (p *parser) parseSeqExpr(seq *seqExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseSeqExpr"))
	}

	var vals []interface{}

	pt := p.pt
	for _, expr := range seq.exprs {
		val, ok := p.parseExpr(expr)
		if !ok {
			p.restore(pt)
			return nil, false
		}
		vals = append(vals, val)
	}
	return vals, true
}

func (p *parser) parseZeroOrMoreExpr(expr *zeroOrMoreExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseZeroOrMoreExpr"))
	}

	var vals []interface{}

	for {
		p.pushV()
		val, ok := p.parseExpr(expr.expr)
		p.popV()
		if !ok {
			return vals, true
		}
		vals = append(vals, val)
	}
}

func (p *parser) parseZeroOrOneExpr(expr *zeroOrOneExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseZeroOrOneExpr"))
	}

	p.pushV()
	val, _ := p.parseExpr(expr.expr)
	p.popV()
	// whether it matched or not, consider it a match
	return val, true
}

func rangeTable(class string) *unicode.RangeTable {
	if rt, ok := unicode.Categories[class]; ok {
		return rt
	}
	if rt, ok := unicode.Properties[class]; ok {
		return rt
	}
	if rt, ok := unicode.Scripts[class]; ok {
		return rt
	}

	// cannot happen
	panic(fmt.Sprintf("invalid Unicode class: %s", class))
}

