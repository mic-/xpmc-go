/*
 * Package utils
 *
 * Part of XPMC.
 * Contains utility functions and other miscellaneous functions
 * and variables.
 *
 * /Mic, 2012-2014
 */
 
package utils

import (
    "fmt"
    "math"
    "os"
    "strings"
    "container/list"
)


type GenericStack struct {
    data *list.List
}

func (s *GenericStack) Push(x interface{}) {
    _ = s.data.PushBack(x)
}

func (s *GenericStack) PopBool() bool {
    e := s.data.Back()
    return s.data.Remove(e).(bool)
}

func (s *GenericStack) PeekBool() bool {
    e := s.data.Back()
    return e.Value.(bool)
}

func (s *GenericStack) PopInt() int {
    e := s.data.Back()
    return s.data.Remove(e).(int)
}

func (s *GenericStack) PeekInt() int {
    e := s.data.Back()
    return e.Value.(int)
}

func (s *GenericStack) PopParserState() *ParserState {
    e := s.data.Back()
    return s.data.Remove(e).(*ParserState)
}

func (s *GenericStack) PeekParserState() *ParserState {
    e := s.data.Back()
    return e.Value.(*ParserState)
}

func (s *GenericStack) Len() int {
    return s.data.Len()
}

func NewGenericStack() *GenericStack {
    return &GenericStack{list.New()}
}


// Global variables

var Parser *ParserState
var OldParsers *GenericStack

// Local variables

var warningsAreErrors bool = false
var verboseMode bool = false
var debugMode bool = false
var definedSymbols map[string]bool

// Compiler messages

func ERROR(msg string, args ...interface{}) {
    fmt.Printf(fmt.Sprintf("[%s:%d,%d] Error: ", Parser.ShortFileName, Parser.LineNum, Parser.Column) +
               fmt.Sprintf(msg + "\n", args...))
    os.Exit(1)
}

func WARNING(msg string, args ...interface{}) {
    fmt.Printf(fmt.Sprintf("[%s:%d,%d] Warning: ", Parser.ShortFileName, Parser.LineNum, Parser.Column) +
               fmt.Sprintf(msg + "\n", args...))
    if warningsAreErrors {
        os.Exit(1)
    }
}

func INFO(msg string, args ...interface{}) {
    if verboseMode {
        fmt.Printf("Info: " + msg + "\n", args...)
    }
}

func DEBUG(msg string, args ...interface{}) {
    if debugMode {
        fmt.Printf("Debug: " + msg + "\n", args...)
    }
}

func Verbose(flag bool) {
    verboseMode = flag
}

func DebugMode(flag bool) {
    debugMode = flag
}

func WarningsAreErrors(flag bool) {
    warningsAreErrors = flag
}

////////

func DefineSymbol(sym string, val int) {
    if definedSymbols == nil {
        definedSymbols = map[string]bool{}
    }
    definedSymbols[sym] = true
}

func IsDefined(sym string) int {
    if (definedSymbols[sym]) {
        return 1
    }
    return 0
}


func IsNumeric(c int) bool {
    return strings.ContainsRune("0123456789", rune(c))
}


/* Returns the position of int i within []int x
 */
func PositionOfInt(x []int, i int) int {
    for j, _ := range x {
        if x[j] == i {
            return j
        }
    }
    return -1
}

/* Returns the position of string s within []string x
 */
func PositionOfString(x []string, s string) int {
    for j, _ := range x {
        if x[j] == s {
            return j
        }
    }
    return -1
}


func Round2(x float64) int {
    if math.Abs(x) - math.Floor(math.Abs(x)) > 0.5 {
        return int(math.Ceil(x))
    }
    return int(math.Floor(x))
}
