/*
 * Package utils
 * ParserState functions
 *
 * Part of XPMC.
 * Contains utility functions and other miscellaneous functions
 * and variables.
 *
 * /Mic, 2012-2014
 */
 
package utils

import (
    "errors"
    "fmt"
    "os"
    "strconv"
    "strings"
    "io/ioutil"
)

type ParserState struct {
    LineNum int
    Column int
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


func (s *ParserState) Init(fileName string) error {
    var err error
    s.fileData, err = ioutil.ReadFile(fileName)
    DEBUG("Parsing " + fileName)
    s.fileDataPos = 0
    s.LineNum = 1
    s.Column = 0
    s.UserDefinedBase = 10
    s.currentBase = 10
    s.wtListOk = false
    s.allowFloats = false
    
    s.WorkDir = ""
    lastSlash := strings.LastIndexAny(fileName, "\\/")
    if lastSlash >= 0 {
        s.WorkDir = fileName[:lastSlash] + string(os.PathSeparator)
        s.ShortFileName = fileName[lastSlash+1:]
        DEBUG("WorkDir = " + s.WorkDir)
    } else {
        s.ShortFileName = fileName
        s.WorkDir, _ = os.Getwd()
        s.WorkDir += string(os.PathSeparator)
        DEBUG("WorkDir = " + s.WorkDir)
    }
    s.listDelimiter = "{}"
    return err
}


func NewParserState(fileName string) (parser *ParserState, err error) {
    parser = &ParserState{}
    err = parser.Init(fileName)
    return
}


/* Inserts the MML code in the given string into the parser's data blob
 * at the current position.
 */
func (p *ParserState) InsertString(s string) {
    if p.fileDataPos < len(p.fileData) {
        tail := make([]byte, len(p.fileData[p.fileDataPos:]))
        copy(tail, p.fileData[p.fileDataPos:])
        p.fileData = append(p.fileData[:p.fileDataPos], []byte(s)...)
        p.fileData = append(p.fileData, tail...)
    } else {
        p.fileData = append(p.fileData, []byte(s)...)
    }
}


/* Returns the next character from the fileData slice.
 */
func (p *ParserState) Getch() int {
    c := -1
    if p.fileDataPos < len(p.fileData) {
        c = int(p.fileData[p.fileDataPos])
        p.fileDataPos++
        p.Column++
    }
    
    return c
}


/* Ungets the last read character.
 */
func (p *ParserState) Ungetch() {
    if p.fileDataPos > 0 {
        p.fileDataPos--
        p.Column--
    }
}


/* Returns the next character from the fileData slice without
 * modifying the position.
 */
func (p *ParserState) Peekch() int {
    c := -1
    if p.fileDataPos < len(p.fileData) {
        c = int(p.fileData[p.fileDataPos])
    }
    
    return c
}

/* Returns a string with the given maximum length from the fileData slice
 * without changing the fileData position.
 */
func (p *ParserState) PeekString(maxChars int) string {
    str := ""
    for i := 0; i < maxChars; i++ {
        if (p.fileDataPos + i) >= len(p.fileData) {
            break
        }
        str += string(p.fileData[p.fileDataPos + i])
    }
    return str
}


/* Skips n characters ahead in the fileData slice.
 */
func (p *ParserState) SkipN(n int) {
    p.fileDataPos += n
    p.Column += n
}

func (p *ParserState) AdvanceLine() {
    p.LineNum++
    p.Column = 0
}


/* Consumes whitespace.
 */
func (p *ParserState) SkipWhitespace() {
    c := 0
    
    for c != -1 {
        c = p.Getch()
        if c == ' ' || c == '\t' || c == 13 || c == 10 {
            if c == 10 {
                p.AdvanceLine()
            }
        } else {
            break
        }
    }
    
    p.Ungetch()
}


/* Rolls back whitespace consumption.
 */
func (p *ParserState) UnskipWhitespace() {
    c := 0
    
    for c != -1 && p.fileDataPos > 0 {
        p.fileDataPos--
        c = int(p.fileData[p.fileDataPos])
        if c == ' ' || c == '\t' || c == 13 || c == 10 {
            if c == 10 {
                p.LineNum--
            }
        } else {
            p.fileDataPos++
            break
        }
    }
}


/* Reads and returns a string (anything but whitespace or EOF).
 */
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


/* Reads and returns a string of characters that belong to the set specified
 * in validChars.
 */
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


/* Reads and returns  a string of characters until some character in endChars
 * is found.
 */
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
    
    if c != -1 {
        p.Ungetch()
    }
    
    if prefix == 'x' {
        if len(s) > 2 {
            // s = "0x" + s[2:]
        } else {
            s = ""
        }
    } else if prefix == 'd' {
        if len(s) > 2 {
            s = s[2:]
        } else {
            s = ""
        }
    } else if p.currentBase == 16 {
        if len(s) > 0 {
            s = "0x" + s
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
       
    lst := NewParamList()
    err := errors.New("Bad list")
    
    commaOk := false            // Not ok to read a comma
    pipeOk  := true             // Ok to read a |
    gotPipe := false
    endOk   := false            // Not ok to read a }
    concatTo := []interface{}{}     
    
    p.SkipWhitespace()
    
    c := p.Getch()
    
    if byte(c) == p.listDelimiter[0] {
        for {
            p.SkipWhitespace()
            p.SetNumericBase(p.UserDefinedBase)
            t := p.GetNumericString()
            
            if len(t) > 0 {
                num, e := strconv.ParseInt(t, 0, 0)
                if e == nil {
                    p.SkipWhitespace()
                    c = p.Getch()
                    if c == ':' {
                        startVal = int(num)
                        p.SetNumericBase(p.UserDefinedBase)
                        t = p.GetNumericString()
                        num, e = strconv.ParseInt(t, 0, 0)
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
                                num, e = strconv.ParseInt(t, 0, 0)
                                if e == nil {
                                    stepVal = int(num)
                                } else {
                                    ERROR("Malformed interval: " + t)
                                }
                                c = p.Getch()
                                if c == '\''{
                                    t = p.GetNumericString()
                                    num, e = strconv.ParseInt(t, 0, 0)
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
                                num, e = strconv.ParseInt(t, 0, 0)
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
                                             
                        i := startVal
                        for {
                            for j := 1; j <= rept; j++ {
                                concatTo = append(concatTo, int(i))
                            }
                            if i == stopVal {
                                break
                            }
                            i += stepVal
                        }
                    
                    } else if c == '\'' {
                        startVal = int(num)
                        p.SetNumericBase(p.UserDefinedBase)
                        t = p.GetNumericString()
                        num, e = strconv.ParseInt(t, 0, 0)
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
                    ERROR("Syntax error while parsing list: " + t)
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
                        concatTo = []interface{}{}
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
                    concatTo = append(concatTo, t)
                } else if c == '\'' || c == ':' {
                    ERROR("Unexpected " + string(byte(c)))
                } else if c == 'W' && p.wtListOk {
                    c = p.Getch()
                    if c == 'T' {
                        t = p.GetNumericString()
                        num, e := strconv.ParseInt(t, 0, 0)
                        if e == nil {
                            wtElem := []interface{}{"WT", int(num)}
                            concatTo = append(concatTo, wtElem)
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
    
    // The list body has been parsed. Check for any trailing operators
    if err == nil {
        for {
            p.SkipWhitespace()
            c = p.Getch()
            
            if c == '+' || c == '-' || c == '*' || c == '\'' {
                p.SkipWhitespace()
                p.AllowFloatsInNumericStrings()
                t := p.GetNumericString()
                if len(t) > 0 {
                    num, e := strconv.ParseInt(t, 0, 0)
                    if e == nil {
                        if c == '+' {
                            for i, _ := range lst.MainPart {
                                curr := lst.MainPart[i].(int)
                                lst.MainPart[i] = curr + int(num)
                            }
                            for i, _ := range lst.LoopedPart {
                                curr := lst.LoopedPart[i].(int)
                                lst.LoopedPart[i] = curr + int(num)
                            }
                        
                        } else if c == '-' {
                            for i, _ := range lst.MainPart {
                                curr := lst.MainPart[i].(int)
                                lst.MainPart[i] = curr - int(num)
                            }
                            for i, _ := range lst.LoopedPart {
                                curr := lst.LoopedPart[i].(int)
                                lst.LoopedPart[i] = curr - int(num)
                            }
                        
                        } else if c == '*' {
                            for i, _ := range lst.MainPart {
                                curr := lst.MainPart[i].(int)
                                lst.MainPart[i] = curr * int(num)
                            }                                
                            for i, _ := range lst.LoopedPart {
                                curr := lst.LoopedPart[i].(int)
                                lst.LoopedPart[i] = curr * int(num)
                            }                                
                        
                        } else if c == '\'' {
                            if num < 1 {
                                ERROR("Repeat value must be >= 1")
                            }
                            if num > 100 {
                                WARNING("Repeat values > 100 are ignored")
                            } else {
                                t := []interface{}{}
                                for j, _ := range lst.MainPart {
                                    for k := 0; k < int(num); k++ {
                                        t = append(t, lst.MainPart[j])
                                    }
                                }
                                lst.MainPart = t

                                t = []interface{}{}
                                for j, _ := range lst.LoopedPart {
                                    for k := 0; k < int(num); k++ {
                                        t = append(t, lst.LoopedPart[j])
                                    }
                                }
                                lst.LoopedPart = t
                            }
                        }
                    } else {
                        ERROR("Syntax error while parsing list: " + t)
                    }
                } else {
                    ERROR("Expected a numeric constant after " + string(byte(c)))
                }
            } else {
                p.Ungetch()
                p.UnskipWhitespace()
                break
            }
        }
    }
    
    // Reset these configurations
    p.wtListOk = false
    p.listDelimiter = "{}"
  
    return lst, err
}


func (p *ParserState) AllowWTList() {
    p.wtListOk = true
}


func (p *ParserState) SetListDelimiters(delim string) {
    p.listDelimiter = delim
}

