/*
 * Package utils
 *
 * Part of XPMC.
 * Contains utility functions and other miscellaneous functions
 * and variables.
 *
 * /Mic, 2012
 */
 
package utils

import (
    "errors"
    "fmt"
    "strconv"
    "strings"
)

var LineNum int
var ShortFileName string
var fileData []byte
var fileDataPos int
var UserDefinedBase int

var allowFloats, wtListOk bool
var currentBase int
var listDelimiter string


type ParamList struct {
    MainPart []int
    LoopedPart []int
}


func Init() {
    LineNum = 1
    UserDefinedBase = 10
}

// Compiler messages

func ERROR(msg string) {
    fmt.Printf("%s@%d, Error: %s\n", ShortFileName, LineNum, msg)
}

func WARNING(msg string) {
    fmt.Printf("%s@%d, Warning: %s\n", ShortFileName, LineNum, msg)
}

func INFO(msg string) {
    fmt.Printf("Info: %s\n", msg)
}


func Getch() int {
    c := -1
    if fileDataPos < len(fileData) {
        c = int(fileData[fileDataPos])
        fileDataPos++
    }
    
    return c
}


// Unget the last read character  
func Ungetch() {
    if fileDataPos > 0 {
        fileDataPos--
    }
}


func IsNumeric(c int) bool {
    return strings.ContainsRune("0123456789", rune(c))
}


// Consume whitespace
func SkipWhitespace() {
    c := 0
    
    for c != -1 {
        c = Getch()
        if c == ' ' || c == '\t' || c == 13 || c == 10 {
            if c == 10 {
                LineNum++
            }
        } else {
            break
        }
    }
    
    Ungetch()
}


// Read a string (anything but whitespace or EOF)
func GetString() string {
    var c int
    
    SkipWhitespace()
    
    s := ""
    c = 0
    for c != -1 {
        c = Getch()
        if c == -1 || c == ' ' || c == '\t' || c == 13 || c == 10 {
            break
        } else {
            s += string(byte(c))
        }
    }
    
    Ungetch()
    
    return s
}


// Read a string of characters that belong to the set specified in validChars
func GetStringInRange(validChars string) string {
    var c int
    
    SkipWhitespace()

    s := ""
    c = 0
    for c != -1 {
        c = Getch()
        if !strings.ContainsRune(validChars, rune(c)) {
            break
        }
        s += string(byte(c))
    }
    
    Ungetch()
    
    return s
}


// Read a string of characters until some character in endChars is found.
func GetStringUntil(endChars string) string {
    var c int
    
    SkipWhitespace()
    s := ""
    c = 0
    for c != -1 {
        c = Getch()
        if strings.ContainsRune(endChars, rune(c)) {
            break
        } else {
            s += string(byte(c))
        }
    }
    
    Ungetch()
    
    return s
}

func GetAlphaString() string {
    var c int
    
    SkipWhitespace()

    s := ""
    c = 0
    for c != -1 {
        c = Getch()
        if (c >= 'A' && c <= 'Z') ||
           (c >= 'a' && c <= 'z') {
            s += string(byte(c))
        } else {
            break
        }
    }
    
    Ungetch()
    
    return s
}


func SetNumericBase(newBase int) {
    currentBase = newBase
}


func AllowFloatsInNumericStrings() {
    allowFloats = true
}


func GetNumericString() string {
    SkipWhitespace()

    s := ""
    c := 0
    prefixOk := false
    prefix := 0
    sign := 0

    for {
        c = Getch()
        if (c >= '0' && c <= '9') || 
           (c == '-' && len(s) == 0) ||
           (c == '.' && allowFloats && len(s) > 0 && (!(prefix == 'x' || currentBase == 16))) ||
           ((c == 'x' || c == 'd') && prefixOk) ||
           ((prefix == 'x' || currentBase == 16) && ((c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F'))) {
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
    
    Ungetch()
    
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
    } else if currentBase == 16 {
        if len(s) > 0 {
            //s = '#' & s
        }
    }
    
    allowFloats = false
    currentBase = 10
    
    if sign == -1 {
        s = "-" + s
    }
    
    return s
}


func GetList() (*ParamList,error) {
    var startVal, stopVal, stepVal int
    
    lst := &ParamList{}
    err := errors.New("Bad list")
    
    commaOk := false            // Not ok to read a comma
    pipeOk := true              // Ok to read a |
    endOk := false              // Not ok to read a }
    concatTo := lst.MainPart    // Concatenate to s[2]
    
    SkipWhitespace()
    
    c := Getch()

    if byte(c) == listDelimiter[0] {
        for {
            SkipWhitespace()
            //set_numeric_base(UserDefinedBase)
            t := GetNumericString()
            
            if len(t) > 0 {
                num, e := strconv.ParseInt(t, UserDefinedBase, 0)
                if e == nil {
                    SkipWhitespace()
                    c = Getch()
                    if c == ':' {
                        startVal = int(num)
                        SetNumericBase(UserDefinedBase)
                        t = GetNumericString()
                        num, e = strconv.ParseInt(t, UserDefinedBase, 0)
                        rept := 0
                        if e == nil {
                            stopVal = int(num)
                            SkipWhitespace()
                            c = Getch()
                            if c == ':' {
                                if UserDefinedBase == 10 {
                                    AllowFloatsInNumericStrings()
                                }
                                SetNumericBase(UserDefinedBase)
                                t = GetNumericString()
                                num, e = strconv.ParseInt(t, UserDefinedBase, 0)
                                if e == nil {
                                    stepVal = int(num)
                                } else {
                                    ERROR("Malformed interval: " + t)
                                }
                                c = Getch()
                                if c == '\''{
                                    t = GetNumericString()
                                    num, e = strconv.ParseInt(t, UserDefinedBase, 0)
                                    if e == nil {
                                        if num < 1 {
                                            ERROR("Repeat value must be >= 1",)
                                        }
                                        rept = int(num)
                                    } else {
                                        ERROR("Expected a repeat value, got " + t)
                                    }
                                } else {
                                    Ungetch()
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
                                t = GetNumericString()
                                num, e = strconv.ParseInt(t, UserDefinedBase, 0)
                                if e == nil {
                                    if num < 1 {
                                        ERROR("Repeat value must be >= 1")
                                    }
                                    rept = int(num)
                                } else {
                                    ERROR("Expected a repeat value, got " + t)
                                }
                                                                
                            } else {
                                Ungetch()
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
                        SetNumericBase(UserDefinedBase)
                        t = GetNumericString()
                        num, e = strconv.ParseInt(t, UserDefinedBase, 0)
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
                        Ungetch()
                        concatTo = append(concatTo, int(num))
                    }
                    commaOk = true
                    pipeOk = true
                    endOk = true
                } else {
                    ERROR("Syntax error: " + t)
                }
            } else {
                c = Getch()
                if c == ',' {
                    if !commaOk {
                        ERROR("Unexpected comma")
                    }
                    commaOk = false
                    pipeOk = false
                    endOk = false
                } else if c == '|' {
                    if lst.LoopedPart == nil && pipeOk {
                        concatTo = lst.LoopedPart
                        commaOk = false
                        pipeOk = false
                        endOk = false
                    } else {
                        ERROR("Unexpected |")
                    }
                } else if byte(c) == listDelimiter[1] {
                    if endOk {
                        err = error(nil)
                        break
                    } else {
                        ERROR("Malformed list")
                    }
                } else if c == ';' {
                    for c != 10 && c != -1 {
                        c = Getch()
                    }
                } else if c == '"' {
                    t = ""
                    c = Getch()
                    for c != '"' && c != -1 {
                        t += string(byte(c))
                        c = Getch()
                    }
                    //ToDo: fix:  s[concatTo] &= {t}
                } else if c == '\'' || c == ':' {
                    ERROR("Unexpected " + string(byte(c)))
                } else if c == 'W' && wtListOk {
                    c = Getch()
                    if c == 'T' {
                        t = GetNumericString()
                        num, e := strconv.ParseInt(t, UserDefinedBase, 0)
                        if e == nil {
                            if num==0 { num++ }  //temp
                            //ToDo: fix: s[concatTo] &= {{-1, 'W', 'T', o[2]}}
                            commaOk = true
                            pipeOk = true
                            endOk = true
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
            SkipWhitespace()
            c = Getch()
            if c == '+' || c == '-' || c == '*' || c == '\'' {
                SkipWhitespace()
                AllowFloatsInNumericStrings()
                t := GetNumericString()
                if len(t) > 0 {
                    num, e := strconv.ParseInt(t, UserDefinedBase, 0)
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
                Ungetch()
                break
            }
        }
    }
    
    // Reset these configurations
    wtListOk = false
    listDelimiter = "{}"
    
    return lst, err
}


func SetListDelimiters(delim string) {
    listDelimiter = delim
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


func (lst *ParamList) Format() string {
    // ToDo: implement
    return ""
}


func (lst *ParamList) IsEmpty() bool {
    return len(lst.MainPart) == 0 && len(lst.LoopedPart) == 0
}
