/*
 * Package utils
 *
 * Part of XPMC.
 * Contains utility functions and other miscellaneous functions
 * and variables.
 *
 * /Mic, 2012-2013
 */
 
package utils

import (
    "errors"
    "fmt"
    "os"
    "strconv"
    "strings"
    "container/list"
    "io/ioutil"
)

type ParserState struct {
    LineNum int
    ShortFileName string
    WorkDir string
    fileData []byte
    fileDataPos int
    UserDefinedBase int
    currentBase int
    allowFloats bool
    wtListOk bool
    listDelimiter string    
}

type ParamList struct {
    MainPart []int
    LoopedPart []int
}


type ParserStateStack struct {
    data *list.List
}

func (s *ParserStateStack) Push(e *ParserState) {
    _ = s.data.PushBack(e)
}

func (s *ParserStateStack) Pop() *ParserState {
    e := s.data.Back()
    return s.data.Remove(e).(*ParserState)
}

func (s *ParserStateStack) Peek() *ParserState {
    e := s.data.Back()
    return e.Value.(*ParserState)
}

func (s *ParserStateStack) Len() int {
    return s.data.Len()
}

func NewParserStateStack() *ParserStateStack {
    return &ParserStateStack{list.New()}
}

func (s *ParserState) Init(fileName string) error {
    var err error
    s.fileData, err = ioutil.ReadFile(fileName)
    s.fileDataPos = 0
    s.LineNum = 1
    s.UserDefinedBase = 10
    s.currentBase = 10
    s.wtListOk = false
    s.allowFloats = false
    
    s.WorkDir = ""
    lastSlash := strings.LastIndexAny(fileName, "\\/")
    if lastSlash >= 0 {
        s.WorkDir = fileName[:lastSlash]
        s.ShortFileName = fileName[lastSlash+1:]
    } else {
        s.ShortFileName = fileName
    }
    s.listDelimiter = "{}"
    return err
}

func NewParserState(fileName string) (parser *ParserState, err error) {
    parser = &ParserState{}
    err = parser.Init(fileName)
    return
}

// Global variables

var Parser *ParserState
var OldParsers *ParserStateStack

// Local variables

var warningsAreErrors bool = false
var verboseMode bool = false
var definedSymbols map[string]bool

// Compiler messages

func ERROR(msg string) {
    fmt.Printf("%s@%d, Error: %s\n", Parser.ShortFileName, Parser.LineNum, msg)
    os.Exit(1)
}

func WARNING(msg string) {
    fmt.Printf("%s@%d, Warning: %s\n", Parser.ShortFileName, Parser.LineNum, msg)
    if warningsAreErrors {
        os.Exit(1)
    }
}

func INFO(msg string) {
    if verboseMode {
        fmt.Printf("Info: %s\n", msg)
    }
}

func Verbose(flag bool) {
    verboseMode = flag
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

////////

func (p *ParserState) Getch() int {
    c := -1
    if p.fileDataPos < len(p.fileData) {
        c = int(p.fileData[p.fileDataPos])
        p.fileDataPos++
    }
    
    return c
}


// Unget the last read character  
func (p *ParserState) Ungetch() {
    if p.fileDataPos > 0 {
        p.fileDataPos--
    }
}


func IsNumeric(c int) bool {
    return strings.ContainsRune("0123456789", rune(c))
}


// Consume whitespace
func (p *ParserState) SkipWhitespace() {
    c := 0
    
    for c != -1 {
        c = p.Getch()
        if c == ' ' || c == '\t' || c == 13 || c == 10 {
            if c == 10 {
                p.LineNum++
            }
        } else {
            break
        }
    }
    
    p.Ungetch()
}


// Read a string (anything but whitespace or EOF)
func (p *ParserState) GetString() string {
    var c int
    
    p.SkipWhitespace()
    
    s := ""
    c = 0
    for c != -1 {
        c = p.Getch()
        if c == -1 || c == ' ' || c == '\t' || c == 13 || c == 10 {
            break
        } else {
            s += string(byte(c))
        }
    }
    
    p.Ungetch()
    
    return s
}


// Read a string of characters that belong to the set specified in validChars
func (p *ParserState) GetStringInRange(validChars string) string {
    var c int
    
    p.SkipWhitespace()

    s := ""
    c = 0
    for c != -1 {
        c = p.Getch()
        if !strings.ContainsRune(validChars, rune(c)) {
            break
        }
        s += string(byte(c))
    }
    
    p.Ungetch()
    
    return s
}


// Read a string of characters until some character in endChars is found.
func (p *ParserState) GetStringUntil(endChars string) string {
    var c int
    
    p.SkipWhitespace()
    s := ""
    c = 0
    for c != -1 {
        c = p.Getch()
        if strings.ContainsRune(endChars, rune(c)) {
            break
        } else {
            s += string(byte(c))
        }
    }
    
    p.Ungetch()
    
    return s
}

func (p *ParserState) GetAlphaString() string {
    var c int
    
    p.SkipWhitespace()

    s := ""
    c = 0
    for c != -1 {
        c = p.Getch()
        if (c >= 'A' && c <= 'Z') ||
           (c >= 'a' && c <= 'z') {
            s += string(byte(c))
        } else {
            break
        }
    }
    
    p.Ungetch()
    
    return s
}


func (p *ParserState) SetNumericBase(newBase int) {
    p.currentBase = newBase
}


func (p *ParserState) AllowFloatsInNumericStrings() {
    p.allowFloats = true
}


func (p *ParserState) GetNumericString() string {
    p.SkipWhitespace()

    s := ""
    c := 0
    
    prefixOk := false
    prefix   := 0
    sign     := 0

    for {
        c = p.Getch()
        if (c >= '0' && c <= '9') || 
           (c == '-' && len(s) == 0) ||
           (c == '.' && p.allowFloats && len(s) > 0 && (!(prefix == 'x' || p.currentBase == 16))) ||
           ((c == 'x' || c == 'd') && prefixOk) ||
           ((prefix == 'x' || p.currentBase == 16) && ((c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F'))) {
            if c == '0' && len(s) == 0 && prefix == 0 {
                prefixOk = true
            } else {
                if ((c == 'x' || c == 'd') && prefixOk) {
                    prefix = c
                }
                prefixOk = false
            }
            if c >= 'a' && c <= 'f' {
                c -= ' '
            }
            if c == '-' {
                if sign == 0 {
                    sign = -1
                }
            } else {
                s += string(byte(c))
            }
        } else {
            break
        }
    }
    
    p.Ungetch()
    
    if prefix == 'x' {
        if len(s) > 2 {
            //s = '#' & s[3..length(s)]
        } else {
            //s = ""
        }
    } else if prefix == 'd' {
        if len(s) > 2 {
            s = s[2:]
        } else {
            s = ""
        }
    } else if p.currentBase == 16 {
        if len(s) > 0 {
            //s = '#' & s
        }
    }
    
    p.allowFloats = false
    p.currentBase = 10
    
    if sign == -1 {
        s = "-" + s
    }
    
    return s
}


func (p *ParserState) GetList() (*ParamList,error) {
    var startVal, stopVal, stepVal int
    
    lst := &ParamList{[]int{}, []int{}}
    err := errors.New("Bad list")
    
    commaOk := false            // Not ok to read a comma
    pipeOk  := true             // Ok to read a |
    gotPipe := false
    endOk   := false            // Not ok to read a }
    concatTo := []int{}     
    
    p.SkipWhitespace()
    
    c := p.Getch()
    
    if byte(c) == p.listDelimiter[0] {
        for {
            p.SkipWhitespace()
            p.SetNumericBase(p.UserDefinedBase)
            t := p.GetNumericString()
            
            if len(t) > 0 {
                num, e := strconv.ParseInt(t, p.UserDefinedBase, 0)
                if e == nil {
                    p.SkipWhitespace()
                    c = p.Getch()
                    if c == ':' {
                        startVal = int(num)
                        p.SetNumericBase(p.UserDefinedBase)
                        t = p.GetNumericString()
                        num, e = strconv.ParseInt(t, p.UserDefinedBase, 0)
                        rept := 0
                        if e == nil {
                            stopVal = int(num)
                            p.SkipWhitespace()
                            c = p.Getch()
                            if c == ':' {
                                if p.UserDefinedBase == 10 {
                                    p.AllowFloatsInNumericStrings()
                                }
                                p.SetNumericBase(p.UserDefinedBase)
                                t = p.GetNumericString()
                                num, e = strconv.ParseInt(t, p.UserDefinedBase, 0)
                                if e == nil {
                                    stepVal = int(num)
                                } else {
                                    ERROR("Malformed interval: " + t)
                                }
                                c = p.Getch()
                                if c == '\''{
                                    t = p.GetNumericString()
                                    num, e = strconv.ParseInt(t, p.UserDefinedBase, 0)
                                    if e == nil {
                                        if num < 1 {
                                            ERROR("Repeat value must be >= 1",)
                                        }
                                        rept = int(num)
                                    } else {
                                        ERROR("Expected a repeat value, got " + t)
                                    }
                                } else {
                                    p.Ungetch()
                                }
                            } else if c == '\'' {
                                if rept != 0 {
                                    ERROR("Found more than one repeat value for the same interval")
                                }
                                if stopVal < startVal {
                                    stepVal = -1
                                } else {
                                    stepVal = 1
                                }
                                t = p.GetNumericString()
                                num, e = strconv.ParseInt(t, p.UserDefinedBase, 0)
                                if e == nil {
                                    if num < 1 {
                                        ERROR("Repeat value must be >= 1")
                                    }
                                    rept = int(num)
                                } else {
                                    ERROR("Expected a repeat value, got " + t)
                                }
                                                                
                            } else {
                                p.Ungetch()
                                if stopVal < startVal {
                                    stepVal = -1
                                } else {
                                    stepVal = 1
                                }
                            }
                        } else {
                            ERROR("Malformed interval: "  + t)
                        }
                        
                        if stopVal < startVal && stepVal >= 0 {
                            WARNING(fmt.Sprintf("Auto-negating step value for interval %d:%d", startVal, stopVal))
                            stepVal = -stepVal
                        } else if stopVal > startVal && stepVal <= 0 {
                            WARNING(fmt.Sprintf("Auto-negating step value for interval %d:%d", startVal, stopVal))
                            stepVal = -stepVal
                        }

                        if stepVal == 0 {
                            ERROR("Step value must be non-zero")
                        }
                        
                        if rept == 0 {
                            rept = 1
                        }
                        
                        for i := startVal; i <= stopVal; i+=stepVal {
                            for j := 1; j <= rept; j++ {
                                concatTo = append(concatTo, int(i))
                            }
                        }
                    
                    } else if c == '\'' {
                        startVal = int(num)
                        p.SetNumericBase(p.UserDefinedBase)
                        t = p.GetNumericString()
                        num, e = strconv.ParseInt(t, p.UserDefinedBase, 0)
                        if e == nil {
                            if num < 1 {
                                ERROR("Repeat value must be >= 1")
                            } else {
                                if num > 100 {
                                    WARNING("Ignoring repeat values > 100")
                                    num = 1
                                }
                                for i := 1; i <= int(num); i++ {
                                    concatTo = append(concatTo, startVal)
                                }
                            }
                        } else {
                            ERROR("Expected a repeat value, got " + t)
                        }
                    } else {
                        p.Ungetch()
                        concatTo = append(concatTo, int(num))
                    }
                    commaOk = true
                    pipeOk  = true
                    endOk   = true
                } else {
                    ERROR("Syntax error: " + t)
                }
            } else {
                c = p.Getch()
                if c == ',' {
                    if !commaOk {
                        ERROR("Unexpected comma")
                    }
                    commaOk = false
                    pipeOk  = false
                    endOk   = false
                } else if c == '|' {
                    if pipeOk && !gotPipe {
                        lst.MainPart = append(lst.MainPart, concatTo...)
                        concatTo = []int{}
                        commaOk = false
                        pipeOk  = false
                        endOk   = false
                        gotPipe = true
                    } else {
                        ERROR("Unexpected |")
                    }
                } else if byte(c) == p.listDelimiter[1] {
                    if endOk {
                        if gotPipe {
                            lst.LoopedPart = append(lst.LoopedPart, concatTo...)
                        } else {
                            lst.MainPart = append(lst.MainPart, concatTo...)
                        }
                        err = nil
                        break
                    } else {
                        ERROR("Malformed list")
                    }
                } else if c == ';' {
                    for c != 10 && c != -1 {
                        c = p.Getch()
                    }
                } else if c == '"' {
                    t = ""
                    c = p.Getch()
                    for c != '"' && c != -1 {
                        t += string(byte(c))
                        c = p.Getch()
                    }
                    //ToDo: fix:  s[concatTo] &= {t}
                } else if c == '\'' || c == ':' {
                    ERROR("Unexpected " + string(byte(c)))
                } else if c == 'W' && p.wtListOk {
                    c = p.Getch()
                    if c == 'T' {
                        t = p.GetNumericString()
                        num, e := strconv.ParseInt(t, p.UserDefinedBase, 0)
                        if e == nil {
                            if num==0 { num++ }  //temp
                            //ToDo: fix: s[concatTo] &= {{-1, 'W', 'T', o[2]}}
                            commaOk = true
                            pipeOk  = true
                            endOk   = true
                        } else {
                            ERROR("Expected a number, got " + t)
                        }
                    } else {
                        ERROR("Expected WT, got W" + string(byte(c)))
                    }
                } else if c == -1 {
                    break
                }
            }
        }
    } else {
        ERROR("Expected {, got " + string(byte(c)))
    }
    
    if err == nil {
        for {
            p.SkipWhitespace()
            c = p.Getch()
            if c == '+' || c == '-' || c == '*' || c == '\'' {
                p.SkipWhitespace()
                p.AllowFloatsInNumericStrings()
                t := p.GetNumericString()
                if len(t) > 0 {
                    num, e := strconv.ParseInt(t, p.UserDefinedBase, 0)
                    if e == nil {
                        for i, _ := range lst.MainPart {
                            if c == '+' {
                                lst.MainPart[i] += int(num)
                            } else if c == '-' {
                                lst.MainPart[i] -= int(num)
                            } else if c == '*' {
                                lst.MainPart[i] *= int(num)
                            } else if c == '\'' {
                                if num < 1 {
                                    ERROR("Repeat value must be >= 1")
                                }
                                if num > 100 {
                                    WARNING("Repeat values > 100 are ignored")
                                } else {
                                    t := []int{}
                                    for j, _ := range lst.MainPart {
                                        for k := 0; k < int(num); k++ {
                                            t = append(t, lst.MainPart[j])
                                        }
                                    }
                                    lst.MainPart = t
                                }
                            }
                        }
                        for i, _ := range lst.LoopedPart {
                            if c == '+' {
                                lst.LoopedPart[i] += int(num)
                            } else if c == '-' {
                                lst.LoopedPart[i] -= int(num)
                            } else if c == '*' {
                                lst.LoopedPart[i] *= int(num)
                            } else if c == '\'' {
                                if num < 1 {
                                    ERROR("Repeat value must be >= 1")
                                }
                                if num > 100 {
                                    WARNING("Repeat values > 100 are ignored")
                                } else {
                                    t := []int{}
                                    for j, _ := range lst.LoopedPart {
                                        for k := 0; k < int(num); k++ {
                                            t = append(t, lst.LoopedPart[j])
                                        }
                                    }
                                    lst.LoopedPart = t
                                }
                            }
                        }
                    } else {
                        ERROR("Syntax error: " + t)
                    }
                } else {
                    ERROR("Expected a numeric constant after " + string(byte(c)))
                }
            } else {
                p.Ungetch()
                break
            }
        }
    }
    
    // Reset these configurations
    p.wtListOk = false
    p.listDelimiter = "{}"
    
    return lst, err
}


func (p *ParserState) SetListDelimiters(delim string) {
    p.listDelimiter = delim
}


// Return the position of i within x 
func PositionOfInt(x []int, i int) int {
    for j, _ := range x {
        if x[j] == i {
            return j
        }
    }
    return -1
}

// Return the position of s within x 
func PositionOfString(x []string, s string) int {
    for j, _ := range x {
        if x[j] == s {
            return j
        }
    }
    return -1
}


func (lst *ParamList) Format() string {
    str := "{"
    
    for i, x := range lst.MainPart {
        str += fmt.Sprintf("%d", x)
        if i < len(lst.MainPart)-1 {
            str += " "
        }
    }
    
    if len(lst.LoopedPart) > 0 {
        str += " | "
        for i, x := range lst.LoopedPart {
            str += fmt.Sprintf("%d", x)
            if i < len(lst.LoopedPart)-1 {
                str += " "
            }
        }
    }

    str += "}"
    return str
}


func (lst *ParamList) MoveToStart() {
    // ToDo: implement
}

func (lst *ParamList) Step() {
    // ToDo: implement
}

func (lst *ParamList) Peek() int {
    // ToDo: implement
    return 0
}

func (lst *ParamList) IsEmpty() bool {
    return len(lst.MainPart) == 0 && len(lst.LoopedPart) == 0
}
