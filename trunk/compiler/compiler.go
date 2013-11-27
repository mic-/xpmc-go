package compiler

import (
    "container/list"
    "fmt"
    "strconv"
    "strings"
    "sync"
    "../channel"
    "../defs"
    "../effects"
    "../song"
    "../targets"
    "../timing"
    "../utils"
)

import . "../utils"


const (
    POLARITY_POSITIVE = 0
    POLARITY_NEGATIVE = 1
    
    ELSIFDEF_TAKEN = 2
)

const ALPHANUM = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrtsuvwxyz"


var CurrSong *song.Song
var SongNum uint
var userDefinedBase int
var implicitAdsrId int
var gbVolCtrl, gbNoise int
var enRev, octaveRev int
var patName string
var slur, tie bool
var lastWasChannelSelect bool
var workDir string


type MmlPattern struct {
    Name string
    Cmds []int
    HasAnyNote bool
    NumTicks int
}

type MmlPatternMap struct {
    data []*MmlPattern
}

func (m *MmlPatternMap) FindKey(key string) int {
    // ToDo: implement
    return -1
}

func (m *MmlPatternMap) HasAnyNote(key string) bool {
    pos := m.FindKey(key)
    if pos >= 0 {
        return m.data[pos].HasAnyNote
    }
    return false
}


type BoolStack struct {
    data *list.List
}

func (s *BoolStack) Push(b bool) {
    _ = s.data.PushBack(b)
}

func (s *BoolStack) Pop() bool {
    e := s.data.Back()
    return s.data.Remove(e).(bool)
}

func (s *BoolStack) Peek() bool {
    e := s.data.Back()
    return e.Value.(bool)
}

func (s *BoolStack) Len() int {
    return s.data.Len()
}

func NewBoolStack() *BoolStack {
    return &BoolStack{list.New()}
}


type IntStack struct {
    data *list.List
}

func (s *IntStack) Push(i int) {
    _ = s.data.PushBack(i)
}

func (s *IntStack) Pop() int {
    e := s.data.Back()
    return s.data.Remove(e).(int)
}

func (s *IntStack) Peek() int {
    e := s.data.Back()
    return e.Value.(int)
}

func (s *IntStack) Len() int {
    return s.data.Len()
}

func NewIntStack() *IntStack {
    return &IntStack{list.New()}
}

                    
var dontCompile *IntStack
var hasElse *BoolStack
var pattern *MmlPattern
var patterns *MmlPatternMap
var keepChannelsActive bool
var callbacks []string


func Init() {
    dontCompile = NewIntStack()
    hasElse = NewBoolStack()
    
    dontCompile.Push(0)
    hasElse.Push(false)
    
    lastWasChannelSelect = false
    
    effects.Init()
}

func Verbose(flag bool) {
    // ToDo: implement
}

func WarningsAreErrors(flag bool) {
    // ToDo: implement
}


func getEffectFrequency() int {
    var n, retVal int
    
    retVal = 0
    
    Parser.SkipWhitespace()
    n = Parser.Getch()
    if n == '(' {
        t := Parser.GetStringUntil(")\t\r\n ")
        if t == "EVERY-FRAME" {
            retVal = 0
        } else if t == "EVERY-NOTE" {
            retVal = 1
        } else {
            ERROR("Unsupported effect frequency: " + t)
        }
        
        if Parser.Getch() != ')' {
            ERROR("Syntax error: expected )")
        }
    } else {
        Parser.Ungetch()
    }
    
    return retVal
}


/* Check if o is within the range of min and max

 Examples:
   inRange(1, 0, 5) -> 1
   inRange([]int{1,2,3}, 1, 10) -> 1
   inRange([]int{1,2}, []int{0,0}, 5) -> 1
   inRange([]int{1,2}, []int{2,0}, 5) -> 0     (min[0]>o[0])
   inRange([]int{1,2}, []int{0,0,0}, 5) -> 0   (too many elements in min)
*/
func inRange(o interface{}, minimum interface{}, maximum interface{}) bool {
    i, isValScalar := o.(int)
    lo, isMinScalar := minimum.(int)
    hi, isMaxScalar := maximum.(int)
    values, isValSlice := o.([]int)
    los, isMinSlice := minimum.([]int)
    his, isMaxSlice := maximum.([]int)
    
    if isValScalar {
        if isMinScalar && isMaxScalar {
            return (i >= lo && i <= hi)
        } else {
            return false
        }
    } else {
        if isMinScalar && isMaxScalar && isValSlice {
            for _, val := range values {
                if val < lo || val > hi {
                    return false
                }
            }
        } else if isMinScalar && isMaxSlice && isValSlice {
            if len(his) == len(values) {
                for j, _ := range values {
                    if values[j] < lo || values[j] > his[j] {
                        return false
                    }
                }
            } else {
                return false
            }
        } else if isMinSlice && isMaxScalar && isValSlice {
            if len(los) == len(values) {
                for j, _ := range values {
                    if values[j] < los[j] || values[j] > hi {
                        return false
                    }
                }
            } else {
                return false
            }
        } else if isMinSlice && isMaxSlice && isValSlice {
            if len(los) == len(his) && len(his) == len(values) {
                for j, _ := range values {
                    if values[j] < los[j] || values[j] > his[j] {
                        return false
                    }
                }
            } else {
                return false
            }
        } else {
            // Unsupported combination of types
            return false
        }
    }
    
    return true
}


func abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}


func writeAllPendingNotes(forceOctChange bool) {
    w := &sync.WaitGroup{}
    for _, chn := range CurrSong.Channels {
        w.Add(1)
        go func() {
            chn.WriteNote(forceOctChange)
            w.Done()
        }()
    }
    w.Wait()
}


// Parse an expression on the form  SYM op SYM op ...
// Where op is either | or &
// The polarity specifies if this expression is for an IF or IFN, and determines
// how the ops should be interpreted.
// Returns 1 if the expression is true, otherwise 0
func evalIfdefExpr(polarity int) int {
    s := Parser.GetStringUntil("&|\r\n")
    expr := IsDefined(s)
    for {
        s = Parser.GetStringInRange("&|")
        if s == "&" {
            s = Parser.GetStringUntil("&|\r\n")
            if polarity == POLARITY_POSITIVE {
                expr = expr & IsDefined(s)
            } else {
                expr = expr | IsDefined(s)
            }
        } else if s == "|" {
            s = Parser.GetStringUntil("&|\r\n")
            if polarity == POLARITY_POSITIVE {
                expr = expr | IsDefined(s)
            } else {
                expr = expr & IsDefined(s)
            }
        } else {
            break
        }
    }
    
    return expr
}


// Handle commands starting with '#', i.e. a meta command 
func handleMetaCommand() {
    if dontCompile.Peek() != 0 {
        s := Parser.GetString()
        switch s {
        case "IFDEF":
            expr := evalIfdefExpr(POLARITY_POSITIVE)
            dontCompile.Push((^expr) | dontCompile.Peek())
            hasElse.Push(false)

        case "IFNDEF":
            expr := evalIfdefExpr(POLARITY_NEGATIVE)
            dontCompile.Push(expr | dontCompile.Peek())
            hasElse.Push(false)             

        case "ELSIFDEF":
            expr := evalIfdefExpr(POLARITY_POSITIVE)
            if dontCompile.Len() > 1 {
                if !hasElse.Peek() {
                    if (dontCompile.Peek() & ELSIFDEF_TAKEN) != ELSIFDEF_TAKEN {
                        _ = dontCompile.Pop()
                        dontCompile.Push((^expr) | dontCompile.Peek())
                    }
                } else {
                    ERROR("ELSIFDEF found after ELSE")
                }
            } else {
                ERROR("ELSIFDEF with no matching IFDEF")
            }                   


        case "ELSE":
            if dontCompile.Len() > 1 {
                if !hasElse.Peek() {
                    if (dontCompile.Peek() & ELSIFDEF_TAKEN) != ELSIFDEF_TAKEN {
                        x := dontCompile.Pop()
                        dontCompile.Push(x | dontCompile.Peek())
                    }
                    _ = hasElse.Pop()
                    hasElse.Push(true)
                } else {
                    ERROR("Only one ELSE allowed per IFDEF")
                }
            } else {
                ERROR("ELSE with no matching IFDEF")
            }

        case "ENDIF":
            if dontCompile.Len() > 1 {
                _ = dontCompile.Pop()
                _ = hasElse.Pop()
            } else {
                ERROR("ENDIF with no matching IFDEF")
            }
        }
    } else {
        for _, chn := range CurrSong.Channels {
            chn.WriteNote(true)
        }

        cmd := Parser.GetString()

        switch cmd {
        case "IFDEF":
            expr := evalIfdefExpr(POLARITY_POSITIVE)
            dontCompile.Push((^expr) | dontCompile.Peek())
            hasElse.Push(false)

        case "IFNDEF":
            expr := evalIfdefExpr(POLARITY_NEGATIVE)
            dontCompile.Push(expr | dontCompile.Peek())
            hasElse.Push(false)

        case "ELSIFDEF":
            if dontCompile.Len() > 1 {
                if !hasElse.Peek() {
                    _ = Parser.GetStringUntil("\r\n")
                    _ = dontCompile.Pop()
                    // Getting here means that the current IFDEF/ELSIFDEF was true,
                    // so whatever is in subsequent ELSIFDEF/ELSE clauses should not
                    // be compiled.
                    dontCompile.Push(ELSIFDEF_TAKEN)
                } else {
                    ERROR("ELSIFDEF found after ELSE")
                }
            } else {
                ERROR("ELSIFDEF with no matching IFDEF")
            }

        case "ELSE":
            if dontCompile.Len() > 1 {
                if !hasElse.Peek() {
                    _ = dontCompile.Pop()
                    dontCompile.Push(1)
                    _ = hasElse.Pop()
                    hasElse.Push(true)
                } else {
                    ERROR("Only one ELSE allowed per IFDEF")
                }
            } else {
                ERROR("ELSE with no matching IFDEF")
            }

        case "ENDIF":
            if dontCompile.Len() > 1 {
                _ = dontCompile.Pop()
                _ = hasElse.Pop()
            } else {
                ERROR("ENDIF with no matching IFDEF")
            }
                    
        case "TITLE":
            CurrSong.Title = Parser.GetStringUntil("\r\n")

        case "TUNE":
            CurrSong.TuneSmsPitch = true

        case "COMPOSER":
            CurrSong.Composer = Parser.GetStringUntil("\r\n")

        case "PROGRAMER","PROGRAMMER":
            CurrSong.Programmer = Parser.GetStringUntil("\r\n")

        case "GAME":
            CurrSong.Game = Parser.GetStringUntil("\r\n")

        case "ALBUM":
            CurrSong.Album = Parser.GetStringUntil("\r\n")

        case "ERROR":
            Parser.SkipWhitespace()
            if Parser.Getch() == '"' {
                s := Parser.GetStringUntil("\"")
                if Parser.Getch() == '"' {
                    ERROR(s)
                } else {
                    ERROR("Malformed #ERROR, missing ending \"")
                }
            } else {
                ERROR("Malformed #ERROR, missing starting \"")
            }

        case "WARNING":
            Parser.SkipWhitespace()
            if Parser.Getch() == '"' {
                s := Parser.GetStringUntil("\"")
                if Parser.Getch() == '"' {
                    WARNING(s)
                } else {
                    ERROR("Malformed #WARNING, missing ending \"")
                }
            } else {
                ERROR("Malformed #WARNING, missing starting \"")
            }
                    
        case "INCLUDE":
            Parser.SkipWhitespace()
            if Parser.Getch() == '"' {
                s := Parser.GetStringUntil("\"")
                if Parser.Getch() == '"' {
                    if len(s) > 0 {
                        if !strings.ContainsRune(s, ':') && s[0] != '\\' {
                            s = workDir + s
                        }
                        CompileFile(s)
                    }
                } else {
                    ERROR("Malformed #INCLUDE, missing ending \"")
                }
            } else {
                ERROR("Malformed #INCLUDE, missing starting \"")
            }

        case "PAL":
            if CurrSong.Target.SupportsPAL() {
                timing.UpdateFreq = 50.0
            }

        case "NTSC":
            timing.UpdateFreq = 60.0

        case "SONG":
            s := Parser.GetString()
            num, err := strconv.ParseInt(s, Parser.UserDefinedBase, 0)
            if err == nil {
                if num > 1 && num < 100 {
                    // ToDo: fix
                    if true { //integer(songs[o[2]]) {
                        /*m = -1
                        l = 0
                        for _, chn := range CurrSong.Channels {
                            songLoopLen[songNum][i] = songLen[songNum][i] - songLoopLen[songNum][i]

                            if chn.LoopPoint == -1 {
                                chn.AddCmd([]int{defs.CMD_END})
                            } else {
                                if !chn.HasAnyNote {
                                    chn.AddCmd([]int{defs.CMD_END})
                                } else {
                                    chn.AddCmd([]int{defs.CMD_JMP, (chn.LoopPoint & 0xFF), (chn.LoopPoint / 0x100)}
                                }
                            }
                            loopPoint[i] = -1
                        end for
                        if keepChannelsActive or len(patName) {
                            compiler.ERROR("Missing }")
                        }
                        songNum = o[2]
                        songs[songNum] = repeat({}, length(supportedChannels))
                        songLen[songNum] = repeat(0, length(supportedChannels))
                        hasAnyNote = repeat(0, length(supportedChannels))
                        songLoopLen[songNum] = songLen[songNum]
                        for i = 1 to length(loopStack) do
                            if length(loopStack[i]) {
                                ERROR("Open [ on channel ", supportedChannels[i])
                            }
                        }*/
                    } else {
                        ERROR("Song " + s + " already defined")
                    }
                } else {
                    ERROR("Bad song number: " + s)
                }
            } else {
                ERROR("Bad song number: " + s)
            }

        case "BASE":
            s := Parser.GetString()
            newBase, err := strconv.ParseInt(s, Parser.UserDefinedBase, 0)
            if err == nil {
                if newBase == 10 || newBase == 16 {
                    userDefinedBase = int(newBase)
                } else {
                    WARNING(cmd +": Expected 10 or 16, got: " + s)
                }
            } else {
                ERROR(cmd +": Expected 10 or 16, got: " + s)
            }

        case "UNIFORM-VOLUME":
            s := Parser.GetString()
            vol, err := strconv.Atoi(s)
            if err == nil {
                if vol > 1 {
                    for _, chn := range CurrSong.Channels {
                        chn.SetMaxVolume(vol)
                    }
                    if vol > 255 {
                        WARNING("Very large max volume specified: " + s)
                    }
                } else {
                    ERROR("Volume must be >= 1: " + s)
                }
            } else {
                ERROR(cmd + ": Expected a positive integer: " + s)
            }

        case "GB-VOLUME-CONTROL":
            s := Parser.GetString()
            ctl, err := strconv.Atoi(s)
            if err == nil {
                if ctl == 1 {
                    gbVolCtrl = 1
                } else if ctl == 0 {
                    gbVolCtrl = 0
                } else {
                    WARNING(cmd + ": Expected 0 or 1, got: " + s)
                }
            } else {
                ERROR(cmd + ": Expected 0 or 1, got: " + s)
            }

        case "GB-NOISE":
            s := Parser.GetString()
            val, err := strconv.Atoi(s)
            if err == nil {
                if val == 1 {
                    gbNoise = 1
                    if CurrSong.Target.GetID() == targets.TARGET_GBC {
                        CurrSong.Channels[3].SetMaxOctave(5)
                    }
                } else if val == 0 {
                    gbNoise = 0
                    if CurrSong.Target.GetID() == targets.TARGET_GBC {
                        CurrSong.Channels[3].SetMaxOctave(11)
                    }
                } else {
                    WARNING(cmd + ": Expected 0 or 1, defaulting to 0: " + s)
                }
            } else {
                ERROR(cmd + ": Expected 0 or 1, got: " + s)
            }

        case "EN-REV":
            s := Parser.GetString()
            rev, err := strconv.Atoi(s)
            if err == nil {
                if rev == 1 {
                    enRev = 1
                    if CurrSong.Target.GetID() == targets.TARGET_C64 ||
                       CurrSong.Target.GetID() == targets.TARGET_AT8 {
                        ERROR("#EN-REV 1 is not supported for this target")
                    }
                } else if rev == 0 {
                    enRev = 0
                } else {
                    WARNING(cmd + ": Expected 0 or 1, defaulting to 0: " + s)
                }
            } else {
                ERROR(cmd + ": Expected 0 or 1, got: " + s)
            }

        case "OCTAVE-REV":
            s := Parser.GetString()
            rev, err := strconv.Atoi(s)
            if err == nil {
                if rev == 1 {
                    octaveRev = -1
                } else if rev == 0 {
                } else {
                    WARNING(cmd + ": Expected 0 or 1, defaulting to 0:" + s)
                }
            } else {
                ERROR(cmd + ": Expected 0 or 1, got:" + s)
            }

        case "AUTO-BANKSWITCH":
            WARNING("Unsupported command: AUTO-BANKSWITCH")
            _ = Parser.GetString() 

        default:
            ERROR("Unknown command: " + cmd)
        }
    }
}



func handleDutyMacDef(num int) {
    idx := effects.DutyMacros.FindKey(num) 
    if CurrSong.GetNumActiveChannels() == 0 {
        if idx < 0 {
            t := Parser.GetString()
            if t == "=" {
                lst, err := Parser.GetList()
                if err == nil {
                    if len(lst.MainPart) != 0 || len(lst.LoopedPart) != 0 {
                        effects.DutyMacros.Append(num, lst)
                        effects.DutyMacros.PutInt(num, getEffectFrequency())
                    } else {
                        ERROR("Empty list for @")
                    }
                } else {
                    ERROR("Syntax error, unable to parse list")
                }
            } else {
                ERROR("Expected '=', got: " + t)
            }
        } else {
            ERROR("Redefinition of @" + strconv.FormatInt(int64(num), 10))
        }
    } else {
        numChannels := 0
        for _, chn := range CurrSong.Channels {
            if chn.Active {
                numChannels++
                if num >= 0 && num <= chn.SupportsDutyChange() {
                    if !chn.Tuple.Active {
                        chn.AddCmd([]int{defs.CMD_DUTY | num})
                    } else {
                        chn.Tuple.Cmds = append(chn.Tuple.Cmds, channel.Note{0xFFFF, float64(defs.CMD_DUTY | num), true})
                    }
                } else {
                    if chn.SupportsDutyChange() == -1 {
                        WARNING("Unsupported command for channel " + chn.GetName() + ": @")
                    } else {
                        ERROR("@ out of range: " + strconv.FormatInt(int64(num), 10))
                    }
                }
            }
        }
        if numChannels == 0 {
            WARNING("Use of @ with no channels active")
        }
    }
                            
}

// Handle definitions of panning macros ("@CS<xy> = {...}")
func handlePanMacDef(cmd string) {
    num, err := strconv.Atoi(cmd[2:])
    if err == nil {
        // Already defined?
        idx := effects.PanMacros.FindKey(num)
        if idx < 0 {
            // ..no. Get the '=' sign and then the list of values
            t := Parser.GetString()
            if t == "=" {
                lst, err := Parser.GetList()
                if err == nil {
                    // The list must contain at least one value
                    if !lst.IsEmpty() {
                        // PC-Engine supports any values, Super Famicom supports -127..+127, all other
                        // targets support -63..+63
                        if (inRange(lst.MainPart, -63, 63) && inRange(lst.LoopedPart, -63, 63)) ||
                           (inRange(lst.MainPart, -127, 127) &&
                              inRange(lst.LoopedPart, -127, 127) &&
                              CurrSong.Target.GetID() == targets.TARGET_SFC) ||
                           CurrSong.Target.GetID() == targets.TARGET_PCE {
                            // Store this effect among the pan macros
                            effects.PanMacros.Append(num, lst)
                            // And store the effect frequency for this particular effect (whether it should
                            // be applied once per frame or once per note).
                            effects.PanMacros.PutInt(num, getEffectFrequency())
                        } else {
                            ERROR("Value of out range (allowed: -63-63): " + lst.Format())
                        }
                    } else {
                        ERROR("Empty list for CS")
                    }
                } else {
                    ERROR("Bad CS: " + t)
                }
            } else {
                ERROR("Expected '='")
            }
        } else {
            ERROR("Redefinition of @" + cmd)
        }
    } else {
        ERROR("Syntax error: @" + cmd)
    }
}


// Handle definitions of arpeggio macros ("@EN<xy> = {...}")
func handleArpeggioDef(cmd string) {
    num, err := strconv.Atoi(cmd[2:])
    if err == nil {
        idx := effects.Arpeggios.FindKey(num)
        if idx < 0 {
            t := Parser.GetString()
            if t == "=" {
                lst, err := Parser.GetList()
                if err == nil {
                    if len(lst.MainPart) != 0 || len(lst.LoopedPart) != 0 {
                        if inRange(lst.MainPart, -63, 63) && inRange(lst.LoopedPart, -63, 63) {
                            effects.Arpeggios.Append(num, lst)
                            effects.Arpeggios.PutInt(num, getEffectFrequency())
                        } else {
                            ERROR("Value of out range (allowed: -63-63): " + lst.Format())
                        }
                    } else {
                        ERROR("Empty list for EN")
                    }
                } else {
                    ERROR("Bad EN: " + t)
                }
            } else {
                ERROR("Expected '='")
            }
        } else {
            ERROR("Redefinition of @" + cmd)
        }
    } else {
        ERROR("Syntax error: @" + cmd)
    }   
}


// Handle definitions of pitch macros ("@EP<xy> = {...}")
func handlePitchMacDef(cmd string) {
    num, err := strconv.Atoi(cmd[2:])
    if err == nil {
        idx := effects.PitchMacros.FindKey(num) 
        if idx < 0 {
            t := Parser.GetString()
            if t == "=" {
                lst, err := Parser.GetList()
                if err == nil {
                    effects.PitchMacros.Append(num, lst)
                    effects.PitchMacros.PutInt(num, getEffectFrequency())
                } else {
                    ERROR("Bad EP: " + t)
                }
            } else {
                ERROR("Expected '='")
            }
        } else {
            ERROR("Redefinition of @" + cmd)
        }
    } else {
        ERROR("Syntax error: @" + cmd)
    }   
}


// Handle definitions of modulation macros ("@MOD<xy> = {...}")
func handleModMacDef(cmd string) {
    num, err := strconv.Atoi(cmd[3:])
    if err == nil {
        idx := effects.MODs.FindKey(num)
        if idx < 0 {
            t := Parser.GetString()
            if t == "=" {
                lst, err := Parser.GetList()
                if err == nil {
                    switch CurrSong.Target.GetID() {
                    case targets.TARGET_SMD:
                        // 3 parameter version: Genesis/Megadrive (YM2612)
                        if len(lst.MainPart) == 3 && len(lst.LoopedPart) == 0 {
                            if inRange(lst.MainPart, 0, []int{7, 7, 3}) {
                                effects.MODs.Append(num, lst)
                            } else {
                                ERROR("Value of out range: " + lst.Format())
                            }
                        } else {
                            ERROR("Bad MOD, expected 3 parameters: " + lst.Format())
                        }
                    case targets.TARGET_CPS, targets.TARGET_X68:
                        // 6 parameter version: CPS-1, X68000 (YM2151)
                        if len(lst.MainPart) == 6 && len(lst.LoopedPart) == 0 {
                            if inRange(lst.MainPart, 0, []int{255, 127, 127, 3, 7, 3}) {
                                effects.MODs.Append(num, lst)
                            } else {
                                ERROR("Value of out range: " + lst.Format())
                            }
                        } else {
                            ERROR("Bad MOD, expected 6 parameters: " + lst.Format())
                        }
                    case targets.TARGET_PCE:
                        // 2 parameter version: PC-Engine (HuC6280)
                        if len(lst.MainPart) == 2 && len(lst.LoopedPart) == 0 {
                            if inRange(lst.MainPart, 0, []int{255, 3}) {
                                effects.MODs.Append(num, lst)
                            } else {
                                ERROR("Value of out range: " + lst.Format())
                            }
                        } else {
                            ERROR("Bad MOD, expected 2 parameters: " + lst.Format())
                        }                                           
                    default:
                        effects.MODs.Append(num, &utils.ParamList{})
                    }
                } else {
                    ERROR("Bad MOD: " + t)
                }
            } else {
                ERROR("Expected '='")
            }
        } else {
            ERROR("Redefinition of @" + cmd)
        }
    } else {
        ERROR("Syntax error: @" + cmd)
    }   
}


func handleFeedbackMacDef(cmd string) {
    num, err := strconv.Atoi(cmd[3:])
    if err == nil {
        idx := effects.FeedbackMacros.FindKey(num)
        if idx < 0 {
            t := Parser.GetString()
            if t == "=" {
                lst, err := Parser.GetList()
                if err == nil {
                    if len(lst.MainPart) != 0 || len(lst.LoopedPart) != 0 {
                        if inRange(lst.MainPart, 0, 7) && inRange(lst.LoopedPart, 0, 7) {
                            effects.FeedbackMacros.Append(num, lst)
                            effects.FeedbackMacros.PutInt(num, getEffectFrequency())
                        } else {
                            ERROR("Value of out range (allowed: 0-7): " + lst.Format())
                        }
                    } else {
                        ERROR("Empty list for FBM")
                    }
                } else {
                    ERROR("Bad FBM: " + t)
                }
            } else {
                ERROR("Expected '='")
            }
        } else {
            ERROR("Redefinition of @" + cmd)
        }
    } else {
        ERROR("Syntax error: @" + cmd)
    }   
}


func handleVibratoMacDef(cmd string) {
    num, err := strconv.Atoi(cmd[2:])
    if err == nil {
        idx := effects.Vibratos.FindKey(num)
        if idx < 0 {
            t := Parser.GetString()
            if t == "=" {
                lst, err := Parser.GetList()
                if err == nil {
                    if len(lst.MainPart) == 3 && len(lst.LoopedPart) == 0 {
                        if inRange(lst.MainPart, []int{0, 1, 0}, []int{127, 127, 63}) {
                            effects.Vibratos.Append(num, lst)
                            effects.Vibratos.PutInt(num, getEffectFrequency())
                        } else {
                            ERROR("Value of out range: " + lst.Format())
                        }
                    } else {
                        ERROR("Bad MP: " + lst.Format())
                    }
                } else {
                    ERROR("Bad MP: " + lst.Format())
                }
            } else {
                ERROR("Expected '='")
            }
        } else {
            ERROR("Redefinition of @" + cmd)
        }
    } else {
        ERROR("Syntax error: @" + cmd)
    }                       
}



func applyCmdOnAllActive(cmdName string, cmd []int) {
    w := &sync.WaitGroup{}
    for _, chn := range CurrSong.Channels {
        w.Add(1)
        go func() {
            if chn.Active {
                chn.AddCmd(cmd)
            }
            w.Done()
        }()
    }
    w.Wait()
}


func applyCmdOnAllActiveFM(cmdName string, cmd []int) {
    for _, chn := range CurrSong.Channels {
        if chn.Active {
            if chn.SupportsFM() {
                chn.AddCmd(cmd)
            } else {
                WARNING(cmdName + " is not supported for channel " + chn.GetName())
            }
        }
    }
}


func applyEffectOnAllActiveSupported(cmdName string, cmd []int, pred func(*channel.Channel) bool,
                                     effMap *effects.EffectMap, num int) {
    for _, chn := range CurrSong.Channels {
        if chn.Active {
            if pred(chn) {
                chn.AddCmd(cmd)
                go effMap.AddRef(num)
            } else {
                WARNING(cmdName + " is not supported for channel " + chn.GetName())
            }
        }
    }
}
                                                        
                                                                        
/*
 * Generates an error if the rune in c isn't a valid channel name.
 */
func assertIsChannelName(c int) {
    if !strings.ContainsRune(CurrSong.Target.GetChannelNames(), rune(c)) {
        ERROR("Unexpected character: " + string(byte(c)))
    }
}


func assertEffectIdExistsAndChannelsActive(name string, eff *effects.EffectMap) (int, int, error) {
    s := Parser.GetNumericString()
    num, err := strconv.Atoi(s)
    idx := -1
    if err == nil {
        idx = eff.FindKey(num)
        if CurrSong.GetNumActiveChannels() == 0 {
            ERROR(name + " requires at least one active channel")
        } else if idx < 0 {
            ERROR("Undefined macro: " + name + s)
        }
    }
    return num, idx, err
}


func assertDisablingEffect(name string, cmd int) {
    c := Parser.Getch()
    s := string(byte(c)) + string(byte(Parser.Getch()))
    if s == "OF" {
        applyCmdOnAllActive(name, []int{cmd, 0})
    } else {
        ERROR("Syntax error: " + name + s)
    }
}


func CompileFile(fileName string) {
    var prevLine int
    var dotOff, tieOff, slurOff bool
    var parserCreationError error
    
    OldParsers.Push(Parser)
    
    Parser,parserCreationError = NewParserState(fileName)
    if parserCreationError != nil {
        ERROR("Failed to read file: " + fileName);
    }
    
    for {
        characterHandled := false

        c := Parser.Getch()
        if c == -1 {
            break
        }
        
        c2 := c
        
        if c == '\n' {
            Parser.LineNum++
        }
        
        if Parser.LineNum > prevLine {
            if !keepChannelsActive {
                for i, _ := range CurrSong.Channels {
                    if i < len(CurrSong.Channels)-1 {
                        CurrSong.Channels[i].Active = false
                    }
                }
            }
            prevLine = Parser.LineNum
        }

        /* Meta-commands */
        
        if dontCompile.Peek() != 0 {
            if c == '#' {
                handleMetaCommand()
            } 
        } else {
            if c == '#' {
                handleMetaCommand()

            } else if c == '@' {
                for _, chn := range CurrSong.Channels {
                    chn.WriteNote(true)
                }
                m := Parser.Getch()
                s := ""
                if m == '@' {
                    s = "@" + Parser.GetNumericString()
                } else if IsNumeric(m) {
                    Parser.Ungetch()
                    s = Parser.GetNumericString()
                } else {
                    Parser.Ungetch()
                    s = Parser.GetAlphaString()
                    s += Parser.GetNumericString()
                }
                
                if len(s) > 0 {
                    if strings.HasPrefix(s, "@") {
                        num, err := strconv.Atoi(s[1:])
                        if err == nil {
                            idx := effects.DutyMacros.FindKey(num) 
                            numChannels := 0
                            for _, chn := range CurrSong.Channels {
                                if chn.Active {
                                    numChannels++
                                    if chn.SupportsDutyChange() > 0 {
                                        idx |= effects.DutyMacros.GetInt(num) * 0x80
                                        chn.AddCmd([]int{defs.CMD_DUTMAC, idx})
                                        effects.DutyMacros.AddRef(num)
                                        chn.UsesEffect["@@"] = true
                                    } else {
                                        WARNING("Unsupported command for channel " +
                                                      chn.GetName() + ": @@")
                                    }
                                }
                            }
                            if numChannels == 0 {
                                WARNING("Use of @@ with no channels active")
                            }
                        } else {
                            ERROR("Expected a number: " + s)
                        }
                    } else {
                        num, err := strconv.Atoi(s)
                        // Duty cycle macro definition
                        if err == nil {
                            handleDutyMacDef(num)

                        // Pan macro definition
                        } else if strings.HasPrefix(s, "CS") {
                            handlePanMacDef(s)
                            
                        // Arpeggio definition
                        } else if strings.HasPrefix(s, "EN") {
                            handleArpeggioDef(s)                    

                        // Feedback macro definition
                        } else if strings.HasPrefix(s, "FBM") {
                            handleFeedbackMacDef(s)
                            
                        // Vibrato definition
                        } else if strings.HasPrefix(s, "MP") {
                            handleVibratoMacDef(s)

                        // Amplitude/frequency modulator
                        } else if strings.HasPrefix(s, "MOD") {
                            handleModMacDef(s)                  

                        // Filter definition
                        } else if strings.HasPrefix(s, "FT") {
                            num, err := strconv.Atoi(s[2:])
                            if err == nil {
                                idx := effects.Filters.FindKey(num)
                                if idx < 0 {
                                    t := Parser.GetString()
                                    if t == "=" {
                                        lst, err := Parser.GetList()
                                        if err == nil {
                                            if CurrSong.Target.GetID() == targets.TARGET_C64 {
                                                if len(lst.MainPart) == 3 && len(lst.LoopedPart) == 0 {
                                                    if inRange(lst.MainPart, []int{0, 0, 0}, []int{3, 2047, 15}) {
                                                        effects.Filters.Append(num, lst)
                                                    }
                                                }               
                                            } else {
                                                effects.Filters.Append(num, &utils.ParamList{})
                                            }
                                        } else {
                                            ERROR("Bad FT: " + t)
                                        }
                                    } else {
                                        ERROR("Expected '='")
                                    }
                                } else {
                                    ERROR("Redefinition of @" + s)
                                }    
                            } else {
                                ERROR("Syntax error: @" + s)
                            }       
                            
                        // Pitch macro definition
                        } else if strings.HasPrefix(s, "EP") {
                            handlePitchMacDef(s)    

                        // Portamento definition
                        } else if strings.HasPrefix(s, "PT") {
                            num, err := strconv.Atoi(s[2:])
                            if err == nil {
                                idx := effects.Portamentos.FindKey(num)
                                if idx < 0 {
                                    t := Parser.GetString()
                                    if t == "=" {
                                        lst, err := Parser.GetList()
                                        if err == nil {
                                            if len(lst.MainPart) == 2 && len(lst.LoopedPart) == 0 {
                                                if inRange(lst.MainPart, []int{0, 1}, []int{127, 127}) {
                                                    effects.Portamentos.Append(num, lst)
                                                } else {
                                                    ERROR("Value out of range: " + lst.Format())
                                                }
                                            } else {
                                                ERROR("Bad PT: " + lst.Format())
                                            }
                                        } else {
                                            ERROR("Bad PT: " + t)
                                        }
                                    } else {
                                        ERROR("Expected '='")
                                    }
                                } else {
                                    ERROR("Redefinition of @" + s)
                                }
                            } else {
                                ERROR("Syntax error: @" + s)
                            }       

                        // Waveform definition (@WT / @WTM)
                        } else if strings.HasPrefix(s, "WT") {
                            num := 0
                            err := error(nil)
                            isWTM := false
                            if strings.HasPrefix(s, "WTM") {
                                isWTM = true
                                num, err = strconv.Atoi(s[3:])
                            } else {
                                num, err = strconv.Atoi(s[2:])
                            }
                            if err == nil {
                                idx := 0
                                if !isWTM {
                                    idx = effects.Waveforms.FindKey(num)
                                } else {
                                    idx = effects.WaveformMacros.FindKey(num)
                                }
                                if idx < 0 {
                                    t := Parser.GetString()
                                    if t == "=" {
                                        if isWTM {
                                            //ToDo: fix:  allow_wt_list()
                                        }
                                        lst, err := Parser.GetList()
                                        if CurrSong.Target.SupportsWaveTable() {
                                            if err == nil {
                                                if !isWTM { // Regular @WT
                                                    if len(lst.LoopedPart) == 0 {
                                                        if inRange(lst.MainPart,
                                                                   CurrSong.Target.GetMinWavSample(),
                                                                   CurrSong.Target.GetMaxWavSample()) {
                                                            if len(lst.MainPart) < CurrSong.Target.GetMinWavLength() {
                                                                WARNING("Padding waveform with zeroes")
                                                                //ToDo: fix:  lst.MainPart &= repeat(0, minWavLength - length(t[2]))
                                                            } else if len(lst.MainPart) > CurrSong.Target.GetMaxWavLength() {
                                                                WARNING("Truncating waveform")
                                                                lst.MainPart = lst.MainPart[0:CurrSong.Target.GetMaxWavLength()]
                                                            }
                                                            //waveforms[2] = append(waveforms[2], t[2])
                                                            effects.Waveforms.Append(num, lst)
                                                        } else {
                                                            ERROR("Waveform data out of range: " + lst.Format())
                                                        }
                                                    } else {
                                                        ERROR("Loops not supported in waveform: " + lst.Format())
                                                    }
                                                } else {            // @WTM
                                                    // ToDo: fix
                                                    /*u = {0,{},{}}
                                                    for j = 2 to 3 do
                                                        for i = 1 to length(t[j]) do
                                                            if and_bits(i, 1) then
                                                                if sequence(t[j][i]) then
                                                                    if length(t[j][i]) = 4 and t[j][i][1] = -1 then
                                                                        u[j] &= t[j][i][4]
                                                                        if assoc_find_key(waveforms, t[j][i][4]) >= 0 then
                                                                            waveforms = assoc_reference(waveforms, t[j][i][4])
                                                                            t[j][i] = assoc_find_key(waveforms, t[j][i][4])
                                                                        } else {
                                                                            ERROR(sprintf("WT%d has not been declared", t[j][i][4]), lineNum)
                                                                        }
                                                                    } else {
                                                                        compiler.ERROR("Expected WT<num>, got " & t[j][i])
                                                                    }
                                                                } else {
                                                                    ERROR("Expected WT<num>, got " & t[j][i])
                                                                }
                                                            } else {
                                                                u[j] &= t[j][i]
                                                                if !inRange(t[j][i], 1, 127) {
                                                                    ERROR(sprintf("Expected a number in the range 1..127, got %d", t[j][i]), lineNum)
                                                                }
                                                            }
                                                        }
                                                    }
                                                    effects.WaveformMacros.Append(num, lst)
                                                    waveformMacros = assoc_insert_extra_data(waveformMacros, o[2], u)*/
                                                }
                                             }else {
                                                ERROR("Bad waveform: " + t)
                                            }
                                        } else {
                                            WARNING("Unsupported command for this target: @WT")
                                        }
                                    } else {
                                        ERROR("Expected '='")
                                    }
                                } else {
                                    ERROR("Redefinition of @" + s)
                                }
                            } else {
                                ERROR("Syntax error: @" + s)
                            }       

                        } else if strings.HasPrefix(s, "XPCM") {
                            num, err := strconv.Atoi(s[4:])
                            if err == nil {
                                idx := effects.PCMs.FindKey(num)
                                if idx < 0 {
                                    t := Parser.GetString()
                                    if t == "=" {
                                        lst, err := Parser.GetList()
                                        if CurrSong.Target.SupportsPCM() {
                                            if err == nil {
                                                if len(lst.LoopedPart) == 0 {
                                                    // ToDo: fix
                                                    /*if len(lst.MainPart) > 1 and sequence(t[2][1]) then
                                                        if not find(':', t[2][1]) and t[2][1][1] != '\\' then
                                                            t[2][1] = workDir & t[2][1]
                                                        }
                                                        if len(lst.MainPart) > 2 {
                                                            t[3] = convert_wav(t[2][1], t[2][2], t[2][3])
                                                        } else {
                                                            t[3] = convert_wav(t[2][1], t[2][2], 100)
                                                        }
                                                        pcms = assoc_append(pcms, o[2], t)
                                                        t = {}
                                                    } else {
                                                        ERROR("Bad XPCM: " & sprint_list(t))
                                                    }*/
                                                } else {
                                                    ERROR("Loops not supported in XPCM: " + lst.Format())
                                                }
                                            } else {
                                                ERROR("Bad XPCM: " + t)
                                            }
                                        } else {
                                            WARNING("Unsupported command for this target: @XPCM")
                                        }
                                    } else {
                                        ERROR("Expected '='")
                                    }
                                } else {
                                    ERROR("Redefinition of @" + s)
                                }
                            } else {
                                ERROR("Syntax error: @" + s)
                            }       
                        
                        } else if strings.HasPrefix(s, "ADSR") {
                            num, err := strconv.Atoi(s[4:])
                            if err == nil {
                                idx := effects.ADSRs.FindKey(num)
                                if idx < 0 {
                                    t := Parser.GetString()
                                    if t == "=" {
                                        lst, err := Parser.GetList()
                                        if err == nil {
                                            if len(lst.LoopedPart) == 0 {
                                                if len(lst.MainPart) == CurrSong.Target.GetAdsrLen() {
                                                    if inRange(lst.MainPart, 0, CurrSong.Target.GetAdsrMax()) {
                                                        effects.ADSRs.Append(num, lst)
                                                    } else {
                                                        ERROR("ADSR parameters out of range: " + lst.Format())
                                                    }
                                                } else {
                                                    ERROR("Bad number of ADSR parameters: " + lst.Format())
                                                }
                                            } else {
                                                ERROR("Bad ADSR: " + lst.Format())
                                            }
                                        } else {
                                            ERROR("Bad ADSR: " + t)
                                        }
                                    } else {
                                        ERROR("Expected '='")
                                    }
                                } else {
                                    ERROR("Redefinition of @" + s)
                                }
                            } else {
                                ERROR("Syntax error: @" + s)
                            }
                            
    
                        } else if strings.HasPrefix(s, "te") {
                            //t := s[2:]
                            m = Parser.Getch()
                            num1, err1 := strconv.Atoi(s[2:])
                            num2 := 7
                            err2 := error(nil)
                            if m == ',' {
                                num2, err2 = strconv.Atoi(Parser.GetNumericString()) 
                            } else {
                                Parser.Ungetch()
                            }
                                
                            if err1 == nil && err2 == nil {
                                for _, chn := range CurrSong.Channels {
                                    chn.WriteNote(true)
                                }
                                if CurrSong.GetNumActiveChannels() == 0 {
                                    WARNING("Trying to set tone envelope with no active channels")
                                } else {
                                    for _, chn := range CurrSong.Channels {
                                        if chn.Active {
                                            if chn.SupportsHwToneEnv() != 0 {
                                                if inRange(num1, -chn.SupportsHwToneEnv(), chn.SupportsHwToneEnv()) {
                                                    switch CurrSong.Target.GetID() {
                                                    case targets.TARGET_GBC: 
                                                        if num1 < 0 {
                                                            num1 = ((abs(num1) - 1) ^ 7) * 16 + 8 + (num2 & 7)
                                                        } else if num1 > 0 {
                                                            num1 = ((num1 - 1) ^ 7) * 16 + (num2 & 7)
                                                        } else {
                                                            num1 = 8
                                                        }
                                                    case targets.TARGET_AST, targets.TARGET_KSS, targets.TARGET_SMS, targets.TARGET_CPC:
                                                        num1 = abs(num1)
                                                    }
                                                    chn.AddCmd([]int{defs.CMD_HWTE, num1})
                                                } else {
                                                    ERROR("Tone envelope number out of range: " + s)
                                                }
                                            } else {
                                                WARNING("Unsupported command for channel " + chn.GetName() + ": @" + s)
                                            }
                                        }
                                    }
                                } /*else {
                                    compiler.ERROR("Bad tone envelope:", s)
                                }*/
                            } else {
                                ERROR("Bad tone envelope: " + s)
                            }

                        } else if strings.HasPrefix(s, "es") {
                            num, err := strconv.Atoi(s[2:])
                            if err == nil {
                                for _, chn := range CurrSong.Channels {
                                    chn.WriteNote(true)
                                }
                                if CurrSong.GetNumActiveChannels() == 0 {
                                    WARNING("Trying to set envelope speed with no active channels")
                                } else if inRange(num, 0, 65535) {
                                    for _, chn := range CurrSong.Channels {
                                        if chn.Active {
                                            switch CurrSong.Target.GetID() {
                                            case targets.TARGET_AST, targets.TARGET_KSS, targets.TARGET_CPC:
                                                num ^= 0xFFFF
                                                chn.AddCmd([]int{defs.CMD_HWES, (num & 0xFF), (num / 0x100)})
                                            default:
                                                WARNING("Unsupported command for this target: es")
                                            }
                                        }
                                    }
                                } else {
                                    ERROR("Bad envelope speed: " + s)
                                }
                            } else {
                                ERROR("Bad tone envelope: " + s)
                            }
                            
                        } else if strings.HasPrefix(s, "ve") {
                            num, err := strconv.Atoi(s[2:])
                            if err == nil {
                                for _, chn := range CurrSong.Channels {
                                    chn.WriteNote(true)
                                }
                                m = 0
                                if CurrSong.Target.GetID() == targets.TARGET_GBC {
                                    if num < 0 {
                                        num = (abs(num) - 1) ^ 7
                                    } else if num > 0 {
                                        num = (num - 1) ^ 7
                                        m = 8
                                    } else {
                                        num = 0
                                    }
                                }
                                if CurrSong.GetNumActiveChannels() == 0 {
                                    WARNING("Trying to set volume envelope with no active channels")
                                } else if err == nil {
                                    for _, chn := range CurrSong.Channels {
                                        if chn.Active {
                                            if chn.SupportsHwVolEnv() != 0 {
                                                if inRange(num, -chn.SupportsHwVolEnv(), chn.SupportsHwVolEnv()) {
                                                    switch CurrSong.Target.GetID() {
                                                    case targets.TARGET_SMS, targets.TARGET_KSS:
                                                        num = abs(num) ^ chn.SupportsHwVolEnv()
                                                    case targets.TARGET_SFC:
                                                        if num == -2 {
                                                            num =  0xA0
                                                        } else if num == -1 {
                                                            num = 0x80
                                                        } else if num == 1 {
                                                            num = 0xC0
                                                        } else if num == 2 {
                                                            num = 0xE0
                                                        }
                                                    }
                                                    chn.AddCmd([]int{defs.CMD_HWVE, num | m})
                                                } else {
                                                    ERROR("Volume envelope value out of range: " + s[2:])
                                                }
                                            } else {
                                                WARNING("Unsupported command for channel " + chn.GetName() + ": @" + s)
                                            }
                                        }
                                    }
                                } else {
                                    ERROR("Bad volume envelope: " + s)
                                }
                            } else {
                                ERROR("Bad volume envelope: " + s)
                            }

                        // Volume macro definition
                        } else if strings.HasPrefix(s, "v") {
                            num, err := strconv.Atoi(s[1:])
                            if err == nil {
                                idx := effects.VolumeMacros.FindKey(num)
                                if CurrSong.GetNumActiveChannels() == 0 {
                                    if len(patName) > 0 {
                                    } else {
                                        if idx < 0 {
                                            t := Parser.GetString()
                                            if t == "=" {
                                                lst, err := Parser.GetList()
                                                if err == nil {
                                                    if inRange(lst.MainPart,
                                                               CurrSong.Target.GetMinVolume(),
                                                               CurrSong.Target.GetMaxVolume()) &&
                                                       inRange(lst.LoopedPart,
                                                               CurrSong.Target.GetMinVolume(),
                                                               CurrSong.Target.GetMaxVolume()) {
                                                        effects.VolumeMacros.Append(num, lst)
                                                        effects.VolumeMacros.PutInt(num, getEffectFrequency())
                                                    } else {
                                                        ERROR("Value out of range: " + lst.Format())
                                                    }
                                                } else {
                                                    ERROR("Bad volume macro: " + t)
                                                }
                                            } else {
                                                ERROR("Expected '='")
                                            }
                                        } else {
                                            ERROR("Redefinition of @" + s)
                                        }
                                    }
                                } else {
                                    if idx < 0 {
                                        ERROR("Undefined macro: @" + s)
                                    } else {
                                        for _, chn := range CurrSong.Channels {
                                            if chn.Active {
                                                if !effects.VolumeMacros.IsEmpty(num) {
                                                    if effects.VolumeMacros.GetInt(num) == 0 {
                                                        chn.AddCmd([]int{defs.CMD_VOLMAC, idx})
                                                    } else {
                                                        chn.AddCmd([]int{defs.CMD_VOLMAC, idx | 0x80})
                                                    }
                                                    effects.VolumeMacros.AddRef(num)
                                                }
                                            }
                                        }
                                    }
                                }
                            } else {
                                ERROR("Syntax error: @" + s)
                            }       
                            

                        // Pulse width macro
                        } else if strings.HasPrefix(s, "pw") {
                            num, err := strconv.Atoi(s[2:])
                            if err == nil {
                                idx := effects.PulseMacros.FindKey(num)
                                if CurrSong.GetNumActiveChannels() == 0 {
                                    if len(patName) > 0 {
                                    } else {
                                        if idx < 0 {
                                            t := Parser.GetString()
                                            if t == "=" {
                                                lst, err := Parser.GetList()
                                                if err == nil {
                                                    if inRange(lst.MainPart, 0, 15) &&
                                                       inRange(lst.LoopedPart, 0, 15) {
                                                        effects.PulseMacros.Append(num, lst)
                                                        effects.PulseMacros.PutInt(num, getEffectFrequency())
                                                    } else {
                                                        ERROR("Value out of range: " + lst.Format())
                                                    }
                                                } else {
                                                    ERROR("Bad pulse width macro: " + t)
                                                }
                                            } else {
                                                ERROR("Expected '='")
                                            }
                                        } else {
                                            ERROR("Redefinition of @" + s)
                                        }
                                    }
                                } else {
                                    if idx < 0 {
                                        ERROR("Undefined macro: @" + s)
                                    } else {
                                        for _, chn := range CurrSong.Channels {
                                            if chn.Active {
                                                if !effects.PulseMacros.IsEmpty(num) {
                                                    chn.AddCmd([]int{defs.CMD_PULMAC, idx})
                                                    effects.PulseMacros.AddRef(num)
                                                    chn.UsesEffect["pw"] = true
                                                }
                                            }
                                        }
                                    }
                                }
                            } else {
                                ERROR("Syntax error: @" + s)
                            }       

                        } else if strings.HasPrefix(s, "q") {
                            num, err := strconv.Atoi(s[1:])
                            if err == nil {
                                for _, chn := range CurrSong.Channels {
                                    chn.WriteNote(true)
                                }
                                if CurrSong.GetNumActiveChannels() == 0 {
                                    WARNING("Trying to set cutoff with no active channels")
                                } else if num >= 0 && num <= 15 {
                                    for _, chn := range CurrSong.Channels {
                                        if chn.Active {
                                            chn.CurrentCutoff.Val = num
                                            chn.CurrentCutoff.Typ = defs.CT_FRAMES
                                            /*s = note_length(i, currentLength[i])
                                            if s[1] != currentNoteFrames[i][1] then
                                                currentNoteFrames[i] = s
                                                chn.AddCmd([]int{defs.CMD_LEN})
                                                chn.WriteLength()
                                            end if*/                                            
                                        }
                                    }
                                } else if num < 0 && num >= -15 {
                                    for _, chn := range CurrSong.Channels {
                                        if chn.Active {
                                            chn.CurrentCutoff.Val = -num
                                            chn.CurrentCutoff.Typ = defs.CT_NEG_FRAMES
                                            /*s = note_length(i, currentLength[i])
                                            if s[1] != currentNoteFrames[i][1] then
                                                currentNoteFrames[i] = s
                                                chn.AddCmd([]int{defs.CMD_LEN})
                                                chn.WriteLength()
                                            }*/                                         
                                        }
                                    }
                                } else {
                                    ERROR("Bad cutoff: " + s)
                                }
                            } else {
                                ERROR("Bad cutoff: " + s)
                            }
                
                        } else {
                            ERROR("Syntax error: @" + s)
                        }
                    }
                } else {
                    ERROR("Unexpected end of file")
                }


            // Pattern definition/invokation
            } else if c == '\\' {
                for _, chn := range CurrSong.Channels {
                    chn.WriteNote(true)
                }
                s := Parser.GetStringInRange(ALPHANUM)
                Parser.SkipWhitespace()
                m := Parser.Getch()
                if len(s) > 0 {
                    pattern = &MmlPattern{}
                    if m == '(' {
                        if CurrSong.GetNumActiveChannels() == 0 {
                            ERROR("Pattern invokation with no active channels")
                        } else {
                            if CurrSong.Channels[len(CurrSong.Channels)-1].Active {
                                ERROR("Pattern invokation found inside pattern")
                            } else {
                                t := Parser.GetStringUntil(")")
                                if len(t) > 0 {
                                }
                                Parser.SkipWhitespace()
                                n := Parser.Getch()
                                if n != ')' {
                                    ERROR("Expected ), got " + string(byte(n)))
                                }

                                // Pattern invokation
                                idx := patterns.FindKey(s)
                                if idx >= 0 {
                                    for _, chn := range CurrSong.Channels {
                                        if chn.Active {
                                            chn.AddCmd([]int{defs.CMD_JSR, idx - 1})
                                            chn.HasAnyNote = chn.HasAnyNote || patterns.HasAnyNote(s)
                                            /*usesEffect[i] = or_bits(usesEffect[i], {1,1,1,1,1,1})
                                            songLen[songNum][i] += patterns[4][idx]*/
                                        }
                                    }
                                } else {
                                    ERROR("Undefined pattern: " + s)
                                }
                            }
                        }
                    } else if m == '{' {
                        if CurrSong.GetNumActiveChannels() == 0 {
                            patName = s
                            CurrSong.Channels[len(CurrSong.Channels)-1].Active = true
                            //songs[songNum][length(songs[songNum])] = {}
                        } else {
                            ERROR("Pattern definitions are not allowed while channels are active")
                        }
                    } else {
                        ERROR("Expected ( or {, got " + string(byte(m)))
                    }
                } else {
                    ERROR("Syntax error")
                }
            
            } else if c == '(' {
                if len(patName) > 0 {
                    ERROR("Found ( inside pattern")
                } else if !keepChannelsActive {
                    if CurrSong.GetNumActiveChannels() > 0 && !CurrSong.Channels[len(CurrSong.Channels)-1].Active {
                        keepChannelsActive = true
                    } else {
                        ERROR("( requires at least one active channel")
                    }
                } else {
                    ERROR("Unexpected (")
                }
                
            } else if c == ')' {
                if keepChannelsActive {
                    keepChannelsActive = false
                } else {
                    ERROR("Unexpected )")
                }
            
            // Macro definition/invokation
            } else if c == '$' {
                writeAllPendingNotes(true)
                s := Parser.GetStringInRange(ALPHANUM)
                Parser.SkipWhitespace()
                m := Parser.Getch()
                t := ""
                if len(s) > 0 {
                    // ToDo: macro := []int{}
                    n := 1
                    if m == '(' {
                        if CurrSong.GetNumActiveChannels() == 0 {
                            l := 1
                            // Read default parameters
                            for {
                                t = Parser.GetStringUntil(",)\t\r\n ")
                                // ToDo: fix
                                //macro = append(macro, {2, t})
                                Parser.SkipWhitespace()
                                n = Parser.Getch()
                                if n == -1 || n == ')' {
                                    break
                                }
                                l += 1
                            }
                            n = 2
                        } else {
                            // Macro invokation
                            for {
                                t = Parser.GetStringUntil(",)\t\r\n ")
                                if len(t) > 0 {
                                    // ToDo: fix
                                    //macro = append(macro, t)
                                }
                                Parser.SkipWhitespace()
                                n = Parser.Getch()
                                if n == -1 || n == ')' {
                                    break
                                }
                            }
                            //ToDo: u := ""
                            //ToDo: defaultParm := []int{}
                            idx := 0 // ToDo: fix:  macros.FindKey(s)
                            if idx >= 0 {
                                // ToDo: fix
                                // Expand the macro
                                /*for i = 1 to length(macros[2][idx]) do
                                    if macros[2][idx][i][1] = 0 then
                                        // This character should be appended as-is
                                        u &= macros[2][idx][i][2]
                                    } else if macros[2][idx][i][1] = 1 then
                                        // This is an argument reference
                                        if macros[2][idx][i][2] > 0 and
                                           macros[2][idx][i][2] <= length(macro) then
                                            u &= macro[macros[2][idx][i][2]]
                                        } else if macros[2][idx][i][2] > 0 and
                                              macros[2][idx][i][2] <= length(defaultParm) then
                                            u &= defaultParm[macros[2][idx][i][2]]
                                        end if
                                    } else if macros[2][idx][i][1] = 2 then
                                        defaultParm = append(defaultParm, macros[2][idx][i][2])
                                    }
                                }
                                if fileDataPos <= len(fileData) then
                                    fileData = fileData[1..fileDataPos-1] & u & fileData[fileDataPos..length(fileData)]
                                } else {
                                    fileData &= u
                                }*/

                                if CurrSong.GetNumActiveChannels() == 0 {
                                    ERROR("Trying to invoke a macro with no channels active")
                                }
                            } else {
                                ERROR("Undefined macro: " + s)
                            }
                            n = 0
                        }
                    }

                    if n == 2 {                     
                        Parser.SkipWhitespace()
                        m = Parser.Getch()
                    }
                    
                    // Macro definition
                    if m == '{' && n > 0 {
                        idx := 0  // ToDo: fix:  macros.FindKey(s)
                        if idx < 0 {
                            if CurrSong.GetNumActiveChannels() == 0 {
                                for n != '}' {
                                    n = Parser.Getch()
                                    if n == '%' {
                                        t = Parser.GetNumericString()
                                        _, err := strconv.Atoi(t)
                                        if err == nil {
                                            // ToDo: fix
                                            //macro = append(macro, {1, num})
                                        } else {
                                            ERROR("Syntax error: " + t)
                                        }
                                        n = Parser.Getch()
                                        if n != '%' {
                                            ERROR("Missing %%")
                                        }
                                    } else if n == '}' || n == -1 {
                                        break
                                    } else if n == '\n' {
                                        Parser.LineNum++
                                    } else if n != '\r' {
                                        // ToDo: fix
                                        //macro = append(macro, {0, n})
                                    }
                                }
                            } else {
                                ERROR("Macro definitions are not allowed while channels are active")
                            }
                            // ToDo: fix
                            //macros[1] = append(macros[1], s)
                            //macros[2] = append(macros[2], macro)
                        } else {
                            ERROR("Macro already defined: " + s)
                        }

                    } else if n != 0 {
                        ERROR("Syntax error: " + string(byte(m)))
                    }
                } else {
                    ERROR("Expected an identifier following $")
                }

            // Callback
            } else if c == '!' {
                writeAllPendingNotes(true)
                s := Parser.GetStringUntil("(\t\r\n ")
                Parser.SkipWhitespace()
                n := Parser.Getch()
                if n == '(' {
                    idx := -1
                    for i, cb := range callbacks {
                        if cb == s {
                            idx = i
                            break
                        }
                    }
                    if idx < 0 {
                        callbacks = append(callbacks, s)
                        idx = len(callbacks) - 1
                    }

                    t := Parser.GetStringUntil(")\t\r\n ")
                    num, err := strconv.Atoi(t)
                    if err == nil {
                        if num == 0 {
                            // Never (Off)
                            numChannels := 0
                            for _, chn := range CurrSong.Channels {
                                if chn.Active {
                                    numChannels++
                                    chn.AddCmd([]int{defs.CMD_CBOFF})
                                }
                            }
                            if numChannels == 0 {
                                ERROR("Trying to set a callback with no channels active")
                            }
                        } else if num == 1 {
                            // Once
                            numChannels := 0
                            for _, chn := range CurrSong.Channels {
                                if chn.Active {
                                    numChannels++
                                    chn.AddCmd([]int{defs.CMD_CBONCE, idx})
                                }
                            }
                            if numChannels == 0 {
                                ERROR("Trying to set a callback with no channels active")
                            }
                        } else {
                            ERROR("Bad callback frequency: " + s)
                        }
                    } else if t == "EVERY-NOTE" {
                        if CurrSong.GetNumActiveChannels() == 0 {
                            ERROR("Trying to set a callback with no channels active")
                        } else {
                            applyCmdOnAllActive("Callback", []int{defs.CMD_CBEVNT, idx})
                        }

                    } else if t == "EVERY-VOL-CHANGE" {
                        if CurrSong.GetNumActiveChannels() == 0 {
                            ERROR("Trying to set a callback with no channels active")
                        } else {
                            applyCmdOnAllActive("Callback", []int{defs.CMD_CBEVVC, idx})
                        }

                    } else if t == "EVERY-VOL-MIN" {
                        if CurrSong.GetNumActiveChannels() == 0 {
                            ERROR("Trying to set a callback with no channels active")
                        } else {
                            applyCmdOnAllActive("Callback", []int{defs.CMD_CBEVVM, idx})
                        }

                    } else if t == "EVERY-OCT-CHANGE" {
                        if CurrSong.GetNumActiveChannels() == 0 {
                            ERROR("Trying to set a callback with no channels active")
                        } else {
                            applyCmdOnAllActive("Callback", []int{defs.CMD_CBEVOC, idx})
                        }

                    } else {
                        ERROR("Bad callback frequency: " + s)
                    }
                    n = Parser.Getch()
                    if n != ')' {
                        ERROR("Expected ): " + string(byte(n)))
                    }
                } else {
                    ERROR("Expected (: " + string(byte(n)))
                }

            // Multiline comment
            } else if c == '/' {
                writeAllPendingNotes(true)
                m := Parser.Getch()
                if m == '*' {
                    n := 0
                    for n != -1 {
                        n = Parser.Getch()
                        if n == '*' {
                            m = Parser.Getch()
                            if m == '/' {
                                break
                            } else {
                                Parser.Ungetch()
                            }
                        } else if n == '\n' {
                            Parser.LineNum++
                        }
                    }
                } else {
                    ERROR("Syntax error: " + string(byte(c)))
                }

            // Beginning of a [..|..]<num> loop
            } else if c == '[' {
                writeAllPendingNotes(true)
                if CurrSong.GetNumActiveChannels() == 0 {
                    // ToDo: treat as error?
                } else {
                    for _, chn := range CurrSong.Channels {
                        if chn.Active {
                            if chn.Tuple.Active {
                                ERROR("Loops not allowed inside {}")
                            }
                            chn.Loops.Push(channel.LoopStackElem{
                                StartPos:   len(chn.Cmds) + 2,
                                StartTicks: chn.Ticks,
                                Unknown:    -1,
                                Skip1Pos:   -1,
                                Skip1Ticks: -1,
                                OrigOctave: chn.CurrentOctave,
                                OctChange:  0,
                                HasOctCmd:  -1,
                                Skip1OctChg:0,
                                Skip1OctCmd:-1,
                            })
                            
                            if chn.Loops.Len() > CurrSong.Target.GetMaxLoopDepth() {
                                ERROR("Maximum nesting level for loops is " +
                                            strconv.FormatInt(int64(CurrSong.Target.GetMaxLoopDepth()), 10))
                            }
                            chn.AddCmd([]int{defs.CMD_LOPCNT, 0})
                        }
                    }
                }
            
            } else if c == '|' {
                writeAllPendingNotes(true)
                if CurrSong.GetNumActiveChannels() == 0 {
                    // ToDo: treat as error?
                } else {
                    for _, chn := range CurrSong.Channels {
                        if chn.Active {
                            if chn.Loops.Len() > 0 {
                                pElem := chn.Loops.Peek()
                                if pElem.Skip1Pos != -1 {
                                    ERROR("Only one | allowed per repeat loop")
                                }
                                pElem.Skip1Pos = len(chn.Cmds) + 2
                                pElem.Skip1Ticks = chn.Ticks
                                chn.AddCmd([]int{defs.CMD_J1, 0, 0})
                            } else {
                                ERROR("Unexpected character: |")
                            }
                        }
                    }
                }
            
            // End of a [..|..]<num> loop
            } else if c == ']' {
                writeAllPendingNotes(true)
                t := Parser.GetNumericString()
                loopCount, err := strconv.Atoi(t)
                for _, chn := range CurrSong.Channels {
                    if chn.Active {
                        elem := chn.Loops.Pop()
                        if elem.StartPos != 0 {
                            if err == nil {
                                if loopCount > 0 {
                                    chn.Cmds[elem.StartPos - 1] = loopCount
                                    chn.AddCmd([]int{defs.CMD_DJNZ, (elem.StartPos & 0xFF), (elem.StartPos / 0x100)})
                                    if elem.Skip1Ticks == -1 {
                                        chn.Ticks += (chn.Ticks - elem.StartTicks) * (loopCount - 1)
                                        if elem.HasOctCmd == -1 {
                                            chn.CurrentOctave = elem.OrigOctave + elem.OctChange * loopCount
                                        } else {
                                            chn.CurrentOctave = elem.HasOctCmd + elem.OctChange
                                        }
                                    } else {
                                        if loopCount < 2 {
                                            ERROR("Loop count must be >= 2 when | is used")
                                        }
                                        chn.Ticks += (chn.Ticks - elem.StartTicks) * (loopCount - 2) +
                                                     (elem.Skip1Ticks - elem.StartTicks)
                                        chn.Cmds[elem.Skip1Pos - 1] = len(chn.Cmds) & 0xFF
                                        chn.Cmds[elem.Skip1Pos] = len(chn.Cmds) / 0x100

                                        if elem.Skip1OctCmd == -1 {
                                            if elem.HasOctCmd == -1 {
                                                chn.CurrentOctave = elem.OrigOctave + elem.OctChange * loopCount +
                                                                                      elem.Skip1OctChg * (loopCount - 1)
                                            } else {
                                                chn.CurrentOctave = elem.HasOctCmd + elem.OctChange
                                            }
                                        } else {
                                            if elem.HasOctCmd == -1 {
                                                chn.CurrentOctave = elem.Skip1OctCmd + elem.OctChange + elem.Skip1OctChg
                                            } else {
                                                chn.CurrentOctave = elem.HasOctCmd + elem.OctChange
                                            }
                                        }
                                    }
                                } else {
                                    ERROR("Bad loop count: " + t)
                                }
                            } else {
                                ERROR("Expected a loop count: " + t)
                            }
                        } else {
                            ERROR("Use of ] with no matching [ on channel " + chn.GetName())
                        }
                    }
                }

            // Octave up/down       
            } else if c == '>' || c == '<' {
                for _, chn := range CurrSong.Channels {
                    chn.WriteNote(true)
                }
                delta := 1
                if c == '<' {
                    delta = -1
                }
                if CurrSong.GetNumActiveChannels() == 0 {
                    WARNING("Trying to change octave with no active channels: " + string(byte(c)))
                } else {
                    for _, chn := range CurrSong.Channels {
                        if chn.Active {
                            if chn.CurrentOctave + delta >= chn.GetMinOctave() &&
                               chn.CurrentOctave + delta <= chn.GetMaxOctave() {
                                chn.CurrentOctave += delta
                                chn.PendingOctChange = delta

                                if chn.Loops.Len() > 0 {
                                    pElem := chn.Loops.Peek()
                                    if pElem.Skip1Pos == -1 {
                                        pElem.OctChange += delta
                                    } else {
                                        pElem.Skip1OctChg += delta
                                    }
                                }
                            
                            } else {
                                ERROR("Octave out of range: " + string(byte(c)) + " (" + 
                                            strconv.FormatInt(int64(chn.CurrentOctave + delta), 10) + ")")
                            }
                        }
                    }
                }
                
            } else if c == 'k' {
                writeAllPendingNotes(true)
                s := Parser.GetNumericString()
                num, err := strconv.Atoi(s)
                if err == nil {
                    if CurrSong.GetNumActiveChannels() == 0 {
                        WARNING("Trying to set length with no active channels")
                    } else if inRange(num, 1, 256) {
                        for _, chn := range CurrSong.Channels {
                            if chn.Active {
                                chn.CurrentLength = float64(num)
                                chn.CurrentNoteFrames.Active, chn.CurrentNoteFrames.Cutoff, _ = chn.NoteLength(chn.CurrentLength)
                                chn.AddCmd([]int{defs.CMD_LEN})
                                chn.WriteLength()
                            }
                        }
                    } else {
                        ERROR("Bad length: " + s)
                    }
                } else {
                    ERROR("Bad length: " + s)
                }
                        
            } else if c == 'l' {
                writeAllPendingNotes(true)
                s := Parser.GetNumericString()
                num, err := strconv.Atoi(s)
                if err == nil {
                    if CurrSong.GetNumActiveChannels() == 0 {
                        WARNING("Trying to set length with no active channels")
                    } else if utils.PositionOfInt(timing.SupportedLengths, num) >= 0 {
                        for _, chn := range CurrSong.Channels {
                            if chn.Active {
                                chn.CurrentLength = 32.0 / float64(num)
                                chn.CurrentNoteFrames.Active, chn.CurrentNoteFrames.Cutoff, _ = chn.NoteLength(chn.CurrentLength)
                                fmt.Printf("The new length is %d. %f active frames, %f cutoff frames\n", num, chn.CurrentNoteFrames.Active, chn.CurrentNoteFrames.Cutoff)
                                chn.AddCmd([]int{defs.CMD_LEN})
                                chn.WriteLength()
                            }
                        }
                    } else {
                        ERROR("Bad length: " + s)
                    }
                }else {
                    ERROR("Bad length: " + s)
                }

            // Set octave
            } else if c == 'o' {
                writeAllPendingNotes(false)
                s := Parser.GetNumericString()
                num, err := strconv.Atoi(s)
                if err == nil {
                    if CurrSong.GetNumActiveChannels() == 0 {
                        WARNING("Trying to set octave with no active channels")
                    } else {
                        for _, chn := range CurrSong.Channels {
                            if chn.Active {
                                if inRange(num, chn.GetMinOctave(), chn.GetMaxOctave()) {
                                    chn.CurrentOctave = num
                                    if !chn.Tuple.Active {
                                        chn.AddCmd([]int{defs.CMD_OCTAVE | chn.CurrentOctave})
                                    } else {
                                        chn.Tuple.Cmds = append(chn.Tuple.Cmds, channel.Note{defs.NON_NOTE_TUPLE_CMD,
                                                                                             float64(defs.CMD_OCTAVE | chn.CurrentOctave),
                                                                                             true})
                                    }
                                    if chn.Loops.Len() > 0 {
                                        pElem := chn.Loops.Peek()
                                        if pElem.Skip1Ticks == -1 {
                                            pElem.HasOctCmd = chn.CurrentOctave
                                            pElem.OctChange = 0
                                        } else {
                                            pElem.Skip1OctCmd = chn.CurrentOctave
                                            pElem.Skip1OctChg = 0
                                        }
                                    }
                                
                                } else {
                                    ERROR("Octave out of range (" +
                                                fmt.Sprintf(" (%d vs [%d,%d])", num, chn.GetMinOctave(), chn.GetMaxOctave()))
                                }
                            }
                        }
                    } /*else {
                        ERROR("Bad octave:", s)
                    }*/
                } else {
                    ERROR("Bad octave: " + s)
                }

            // Set cutoff
            } else if c == 'q' {
                writeAllPendingNotes(true)
                s := Parser.GetNumericString()
                num, err := strconv.Atoi(s)
                if err == nil {
                    if CurrSong.GetNumActiveChannels() == 0 {
                        WARNING("Trying to set cutoff with no active channels")
                    } else if num >= 0 && num <= 8 {
                        for _, chn := range CurrSong.Channels {
                            if chn.Active {
                                chn.CurrentCutoff.Val = num
                                chn.CurrentCutoff.Typ = defs.CT_NORMAL
                                active, cutoff, _ := chn.NoteLength(chn.CurrentLength)
                                if active != chn.CurrentNoteFrames.Active {
                                    chn.CurrentNoteFrames.Active, chn.CurrentNoteFrames.Cutoff = active, cutoff
                                    chn.AddCmd([]int{defs.CMD_LEN})
                                    chn.WriteLength()
                                }
                            }
                        }
                    } else if num < 0 && num >= -8 {
                        for _, chn := range CurrSong.Channels {
                            if chn.Active {
                                chn.CurrentCutoff.Val = -num
                                chn.CurrentCutoff.Typ = defs.CT_NEG
                                active, cutoff, _ := chn.NoteLength(chn.CurrentLength)
                                if active != chn.CurrentNoteFrames.Active {
                                    chn.CurrentNoteFrames.Active, chn.CurrentNoteFrames.Cutoff = active, cutoff
                                    chn.AddCmd([]int{defs.CMD_LEN})
                                    chn.WriteLength()
                                }
                            }
                        }
                    } else {
                        ERROR("Bad cutoff: " + s)
                    }
                } else {
                    ERROR("Bad cutoff: " + s)
                }

            // Set tempo
            } else if c == 't' {
                writeAllPendingNotes(true)
                s := Parser.GetNumericString()
                num, err := strconv.Atoi(s)
                if err == nil {
                    if CurrSong.GetNumActiveChannels() == 0 {
                        WARNING("Trying to set tempo with no active channels")
                    } else if inRange(num, 0, CurrSong.Target.GetMaxTempo()) {
                    	fmt.Printf("New tempo: %d\n", num)
                        for _, chn := range CurrSong.Channels {
                            if chn.Active {
                            	fmt.Printf("Setting tempo on channel %s\n", chn.Name)
                                chn.CurrentTempo = num
                            }
                        }
                    } else {
                        ERROR("Bad tempo: " + s)
                    }
                } else {
                    ERROR("Bad tempo: " + s)
                }

            // Set volume
            } else if c == 'v' {
                writeAllPendingNotes(true)
                
                volType := defs.CMD_VOL2
                volDelta := 0
                num := 0
                err := error(nil)
                s := ""
                
                m := Parser.Getch()
                if m == '+' {
                    volType = defs.CMD_VOLUP
                    if Parser.Getch() == '+' {
                        volType = defs.CMD_VOLUPC
                    } else {
                        Parser.Ungetch()
                    }
                    volDelta = 1
                    s = Parser.GetNumericString()
                    if len(s) > 0 {
                        num, err = strconv.Atoi(s)
                        if err == nil {
                            volDelta = num
                        }
                    }
                } else if m == '-' {
                    volType = defs.CMD_VOLDN
                    if Parser.Getch() == '-' {
                        volType = defs.CMD_VOLDNC
                    } else {
                        Parser.Ungetch()
                    }
                    //l = 1   Needed?
                    s = Parser.GetNumericString()
                    if len(s) > 0 {
                        num, err = strconv.Atoi(s)
                        if err == nil {
                            volDelta = num
                        }
                    }
                } else {
                    Parser.Ungetch()
                    s = Parser.GetNumericString()
                    if len(s) > 0 {
                        num, err = strconv.Atoi(s)
                    } else {
                        ERROR("Expected +, - or a number: " + string(byte(m)))
                    }
                }

                if err == nil {
                    if CurrSong.GetNumActiveChannels() == 0 {
                        if len(patName) > 0 {
                            if num >= CurrSong.Target.GetMinVolume() {
                            } else {
                                ERROR("Bad volume: " + strconv.FormatInt(int64(num), 10))
                            }
                        } else {
                            WARNING("Trying to set volume with no active channels")
                        }
                    } else if num >= CurrSong.Target.GetMinVolume() {
                        for _, chn := range CurrSong.Channels {
                            if chn.Active {
                                if volType == defs.CMD_VOL2 {
                                    if num <= chn.GetMaxVolume() {
                                        chn.CurrentVolume = num
                                        if chn.SupportsVolumeChange() > 0 {
                                            vol := int((chn.CurrentVolume / chn.GetMaxVolume()) * chn.MachineVolLimit())
                                            if vol <= 0x0F {
                                                chn.AddCmd([]int{defs.CMD_VOL2 | vol})
                                            } else {
                                                chn.AddCmd([]int{defs.CMD_VOLSET, vol})
                                            }
                                        } else {
                                            // TODO: Handle this case (e.g. NES triangle channel)
                                            WARNING("Setting volume on channel " + chn.GetName() + " is not supported")
                                        }
                                    } else {
                                        ERROR("Bad volume: " + strconv.FormatInt(int64(num), 10))
                                    }
                                } else {
                                    if volType == defs.CMD_VOLDN {
                                        volDelta = (-volDelta & 0xFF)
                                        volType = defs.CMD_VOLUP
                                    }
                                    if chn.SupportsVolumeChange() > 0 {
                                        chn.AddCmd([]int{volType, volDelta})
                                    }
                                }
                            }
                        }
                    } else {
                        ERROR("Bad volume: " + strconv.FormatInt(int64(num), 10))
                    }
                } else {
                    ERROR("Bad volume: " + s)
                }

            // Single line comment
            } else if c == ';' {
                writeAllPendingNotes(true)
                n := 0
                for n != -1 {
                    n = Parser.Getch()
                    if n == '\n' {
                        Parser.LineNum++
                        break
                    } else if n == '\r' {
                        break
                    } else {
                    }
                }


            // Reads notes (e.g. c d+ a+4 g-1.. e1^8^32 f&f16&f32)
            } else if defs.NoteIndex(c) >= 0 {
                if !slur {
                    writeAllPendingNotes(false)
                }

                tie     = false
                slur    = false
                hasTie  := false
                hasSlur := false
                hasDot  := false
                firstNote := -1
                frames := -1.0
                flatSharp := 0
                
                extraChars := 0
                n := c
                note := 0
                
                for n != -1 {
                    if defs.NoteIndex(n) >= 0 {
                        if extraChars > 0 && !slur {
                            Parser.Ungetch()
                            break
                        }
                        note = defs.NoteIndex(n)
                    } else if n == '&' {
                        if hasTie {
                            ERROR("Trying to use & and ^ in same expression")
                        }
                        for _, chn := range CurrSong.Channels {
                            if chn.Active && chn.Tuple.Active {
                                WARNING("Command ignored inside tuple: &")
                            }
                        }
                        hasSlur = true
                        slur = true
                        note = -1
                    } else if n == '^' {
                        if hasSlur {
                            ERROR("Trying to use & and ^ in same expression")
                        }
                        for _, chn := range CurrSong.Channels {
                            if chn.Active && chn.Tuple.Active {
                                WARNING("Command ignored inside tuple: ^")
                            }
                        }
                        hasTie = true
                        tie = true
                    } else if n == '.' {
                        for _, chn := range CurrSong.Channels {
                            if chn.Active && chn.Tuple.Active {
                                WARNING("Command ignored inside tuple: .")
                            }
                        }
                        hasDot = true
                    } else {
                        Parser.Ungetch()
                        break
                    }

                    if (defs.NoteIndex(n) >= 0 && n != 'r' && n != 's') || (extraChars > 0 && note != -1 && slur) {
                        flatSharp = 0
                        frames = -1
                        m := Parser.Getch()
                        if m == '+' {
                            flatSharp = 1
                            if defs.NoteVal(note + flatSharp) != -2 {
                                ERROR("Bad note: " + string(byte(n)) + "+")
                            }
                        } else if m == '-' {
                            flatSharp = -1
                            if (note + flatSharp < 1) || defs.NoteVal(note + flatSharp) != -2 {
                                ERROR("Bad note: " + string(byte(n)) + "-")
                            }
                        } else {
                            Parser.Ungetch()
                        }
                    } else if n == 'r' || n == 's' {
                        flatSharp = 0
                    }

                    if firstNote == -1 {
                        firstNote = note + flatSharp
                    } else {
                        if (firstNote != note + flatSharp) && note != -1 {
                            ERROR("Trying to concatenate different notes")
                        }
                    }

                    if defs.NoteIndex(n) >= 0 || tie {
                        s := Parser.GetNumericString()
                        if len(s) > 0 {
                            noteLen, err := strconv.Atoi(s)
                            if err == nil {
                                if utils.PositionOfInt(timing.SupportedLengths, noteLen) >= 0 { 
                                    frames = 32.0 / float64(noteLen) 
                                    for _, chn := range CurrSong.Channels {
                                        if chn.Active && chn.Tuple.Active {
                                            WARNING("Note length ignored inside tuple")
                                        }
                                    }
                                } else {
                                    ERROR("Unsupported length: " + s)
                                }
                            } else {
                                ERROR("Bad length: " + s)
                            }
                        }
                    }
                    if frames == -1 && tie {
                        ERROR("Expected a length")
                    }

                    tieOff  = false
                    slurOff = false
                    dotOff  = false
                    
                    numChannels := 0
                    for _, chn := range CurrSong.Channels {
                        if chn.Active {
                            if chn.CurrentOctave <= chn.GetMinOctave() &&
                               note > 0 &&
                               note < chn.GetMinNote() {
                                WARNING("The frequency of the note can not be reproduced for this octave: " +
                                              string(byte(defs.NoteVal(note))))
                            }
                            numChannels++
                            if frames == -1 {
                                frames = chn.CurrentLength
                            }
                            if extraChars == 0 {
                                if n != 'r' && n != 's' {
                                    chn.CurrentNote = channel.Note{(chn.CurrentOctave - chn.GetMinOctave()) * 12 + note + flatSharp - 1,
                                                                   float64(frames),
                                                                   true}
                                } else if n == 'r' {
                                    chn.CurrentNote = channel.Note{defs.Rest, float64(frames), true}
                                } else {
                                    chn.CurrentNote = channel.Note{defs.Rest2, float64(frames), true}
                                }
                                chn.LastSetLength = frames
                            } else {
                                if hasDot {
                                    frames = float64(chn.LastSetLength) / 2
                                    chn.LastSetLength = frames
                                    if frames >= 1 {
                                        chn.CurrentNote.Frames += frames
                                    } else {
                                        ERROR("Note length out of range due to dot command")
                                    }
                                    dotOff = true
                                } else if tie {
                                    chn.CurrentNote.Frames += frames
                                    tieOff = true
                                } else if slur {
                                    if note != -1 {
                                        if ((chn.CurrentOctave - chn.GetMinOctave()) * 12 + note + flatSharp - 1 == chn.CurrentNote.Num) ||
                                            (chn.CurrentNote.Num == defs.Rest && n == 'r') ||
                                            (chn.CurrentNote.Num == defs.Rest2 && n == 's') {
                                            chn.CurrentNote.Frames += frames
                                            chn.LastSetLength = frames
                                            slurOff = true
                                        } else {
                                            ERROR("Bad note: " + string(byte(n)))
                                        }
                                    }
                                } else {
                                    ERROR("Bad note: " + string(byte(n)))
                                }
                            }
                        }
                    }
                    
                    if dotOff {
                        hasDot = false
                    }
                    if slurOff {
                        slur = false
                    }
                    if tieOff {
                        tie = false
                    }
                    
                    if numChannels == 0 {
                        ERROR("Trying to play a note with no active channels")
                    }

                    n = Parser.Getch()
                    extraChars++
                }
                tie = false
                slur = false
                hasDot = false
                for _, chn := range CurrSong.Channels {
                    chn.LastSetLength = 0
                }

            } else if c == '{' {
                writeAllPendingNotes(true)
                if CurrSong.GetNumActiveChannels() == 0 {
                    ERROR("{ requires at least one active channel")
                } else {
                    for _, chn := range CurrSong.Channels {
                        if chn.Active {
                            if !chn.Tuple.Active {
                                chn.Tuple.Active = true
                            } else {
                                //  ERROR("Nested tuples are not allowed", lineNum)
                            }
                        }
                    }
                }
                
            } else if c == '}' {
                writeAllPendingNotes(true)
                s := Parser.GetNumericString()
                tupleLen := -1.0
                if len(s) > 0 {
                    num, err := strconv.Atoi(s)
                    if err == nil {
                        if utils.PositionOfInt(timing.SupportedLengths, num) >= 0 { 
                            tupleLen = 32.0 / float64(num) 
                        } else {
                            ERROR("Unsupported length: " + s)
                        }
                    } else {
                        ERROR("Bad length: " + s)
                    }
                }
                if tupleLen == -1 {
                    if len(patName) > 0 {
                        CurrSong.Channels[len(CurrSong.Channels)-1].AddCmd([]int{defs.CMD_RTS})
                        /*patterns[1] = append(patterns[1], patName)
                        patterns[2] = append(patterns[2], songs[songNum][length(songs[songNum])])
                        patterns[3] &= hasAnyNote[length(supportedChannels)]
                        patterns[4] &= songLen[songNum][length(supportedChannels)]*/
                        patName = ""
                        //songs[songNum][length(songs[songNum])] = {}
                        CurrSong.Channels[len(CurrSong.Channels)-1].Active = false
                    } else {
                        ERROR("Syntax error: }")
                    }
                } else {
                    if CurrSong.GetNumActiveChannels() == 0 {
                        ERROR("Trying to close a tuple with no active channels")
                    } else {
                        for _, chn := range CurrSong.Channels {
                            if chn.Active {
                                if chn.Tuple.Active {
                                    chn.WriteTuple(tupleLen)
                                    chn.Tuple.Active = false
                                }
                            }
                        }
                    }
                }   
                
            } else if strings.ContainsRune(CurrSong.Target.GetChannelNames(), rune(c)) ||
                      strings.ContainsRune("ACDEFKLMOPRSWnpw", rune(c)) {
                writeAllPendingNotes(true)
                
                if c == 'A' {
                    m := Parser.Getch()
                    s := string(byte(m))
                    s += string(byte(Parser.Getch()))
                    s += string(byte(Parser.Getch()))
                    
                    if s == "DSR" {
                        characterHandled = true
                        Parser.Ungetch()
                        num := 0
                        err := error(nil)
                        m = Parser.Getch()
                        if m == '(' {
                            // Implicit ADSR declaration
                            Parser.SetListDelimiters("()")
                            lst, err := Parser.GetList()
                            if err == nil {
                                if len(lst.LoopedPart) == 0 {
                                    if len(lst.MainPart) == CurrSong.Target.GetAdsrLen() {
                                        if inRange(lst.MainPart, 0, CurrSong.Target.GetAdsrMax()) {
                                            effects.ADSRs.Append(implicitAdsrId, lst)
                                        } else {
                                            ERROR("ADSR parameters out of range: " + lst.Format())
                                        }
                                    } else {
                                        ERROR("Bad number of ADSR parameters: " + lst.Format())
                                    }
                                } else {
                                    ERROR("Bad ADSR: " + lst.Format())
                                }
                            } else {
                                ERROR("Bad ADSR: Unable to parse parameter list")
                            }
                            num = implicitAdsrId
                            implicitAdsrId++
                        } else {
                            // Use a previously declared ADSR
                            s = Parser.GetNumericString()
                            num, err = strconv.Atoi(s)
                        }
                        
                        if err == nil {
                            idx := effects.ADSRs.FindKey(num)
                            if CurrSong.GetNumActiveChannels() == 0 {
                                ERROR("ADSR requires at least one active channel")
                            } else if idx >= 0 {
                                for _, chn := range CurrSong.Channels {
                                    if chn.Active {
                                        if chn.SupportsADSR() {
                                            chn.AddCmd([]int{defs.CMD_ADSR, idx})
                                            effects.ADSRs.AddRef(num)
                                        } else {
                                            WARNING("Unsupported command for this channel: ADSR")
                                        }
                                    }
                                }
                            } else {
                                ERROR("Undefined macro: ADSR" + s)
                            }
                        } else {
                            ERROR("Expected a number: " + s)
                        }
                    } else if m == 'M' {
                        Parser.Ungetch()
                        Parser.Ungetch()
                        s = Parser.GetNumericString()
                        if len(s) > 0 {
                            characterHandled = true
                            num, err := strconv.Atoi(s)
                            if err == nil {
                                if inRange(num, 0, 1) {
                                    if CurrSong.GetNumActiveChannels() == 0 {
                                        ERROR("AM requires at least one active channel")
                                    } else {
                                        applyCmdOnAllActiveFM("AM", []int{defs.CMD_HWAM, num})
                                    }
                                } else {
                                    ERROR("AM out of range")
                                }
                            } else {
                                ERROR("Bad AM: " + s)
                            }
                        } else {
                            Parser.Ungetch()
                            assertIsChannelName(c)
                        }                       
                    } else {
                        Parser.Ungetch()
                        Parser.Ungetch()
                        Parser.Ungetch()
                        assertIsChannelName(c)
                    }


                } else if c == 'C' {
                    m := Parser.Getch()
                    if m == 'S' {
                        characterHandled = true
                        num, idx, err := assertEffectIdExistsAndChannelsActive("CS", effects.PanMacros)
                        
                        if err == nil {
                            if CurrSong.Target.SupportsPan() {
                                idx |= effects.PanMacros.GetInt(num) * 0x80
                                for _, chn := range CurrSong.Channels {
                                    if chn.Active {
                                        chn.AddCmd([]int{defs.CMD_PANMAC, idx + 1})
                                        effects.PanMacros.AddRef(num)
                                    }
                                }
                            } else {
                                WARNING("Unsupported command for this target: CS")
                            }
                        } else {
                            m = Parser.Getch()
                            t := string(byte(m)) + string(byte(Parser.Getch()))
                            if t == "OF" {
                                if CurrSong.Target.SupportsPan() {
                                    applyCmdOnAllActive("CS", []int{defs.CMD_PANMAC, 0})
                                }
                            } else {
                                ERROR("Syntax error: CS" + t)
                            }
                        }
                    } else {
                        Parser.Ungetch()
                        assertIsChannelName(c)
                    }
                
                } else if c == 'D' {
                    m := Parser.Getch()
                    if m == '-' || IsNumeric(m) {
                        characterHandled = true
                        Parser.Ungetch()
                        s := Parser.GetNumericString()
                        num, err := strconv.Atoi(s)
                        if err == nil {
                            if CurrSong.GetNumActiveChannels() == 0 {
                                ERROR("Detune requires at least one active channel")
                            } else {
                                trg := CurrSong.Target.GetID()
                                for _, chn := range CurrSong.Channels {
                                    if chn.Active {
                                        if chn.SupportsDetune() {
                                            if chn.SupportsFM() &&
                                            (trg == targets.TARGET_SMD || trg == targets.TARGET_CPS || trg == targets.TARGET_KSS) {
                                                if inRange(num, -3, 3) {
                                                    amount := num
                                                    if amount < 0 {
                                                        amount = (-amount) + 4
                                                    }
                                                    chn.AddCmd([]int{defs.CMD_DETUNE, amount})
                                                } else {
                                                    ERROR("Detune value out of range: " + s)
                                                }
                                            } else {
                                                if inRange(num, -127, 127) {
                                                    chn.AddCmd([]int{defs.CMD_DETUNE, num})
                                                } else {
                                                    ERROR("Detune value out of range: " + s)
                                                }
                                            }
                                        } else {
                                            WARNING("Unsupported command for this channel: D")
                                        }
                                    }
                                }
                            }
                        } else {
                            ERROR("Expected a number: " + s)
                        }
                    } else {
                        Parser.Ungetch()
                        assertIsChannelName(c)
                    }                       

                } else if c == 'K' {
                    m := Parser.Getch()
                    if m == '-' || IsNumeric(m) {
                        characterHandled = true
                        Parser.Ungetch()
                        s := Parser.GetNumericString()
                        num, err := strconv.Atoi(s)
                        if err == nil {
                            if CurrSong.GetNumActiveChannels() == 0 {
                                ERROR("Transpose requires at least one active channel")
                            } else {
                                if inRange(num, -127, 127) {
                                    applyCmdOnAllActive("K", []int{defs.CMD_TRANSP, num})
                                } else {
                                    ERROR("Transpose value out of range: " + s)
                                }
                            }
                        } else {
                            ERROR("Expected a number: " + s)
                        }
                    } else {
                        Parser.Ungetch()
                        assertIsChannelName(c)
                    }       
                    
                } else if c == 'E' {
                    m := Parser.Getch()
                    if m == 'N' {
                        characterHandled = true
                        num, idx, err := assertEffectIdExistsAndChannelsActive("EN", effects.Arpeggios)
                        
                        if err == nil {
                            for _, chn := range CurrSong.Channels {
                                if chn.Active {
                                    if !effects.Arpeggios.IsEmpty(num) {
                                        effects.Arpeggios.AddRef(num)
                                        idx |= effects.Arpeggios.GetInt(num) * 0x80
                                        if enRev == 0 {
                                            chn.AddCmd([]int{defs.CMD_ARPMAC, idx})
                                            //ToDo fix: usesEN[1] += 1
                                            chn.UsesEffect["EN"] = true
                                        } else {
                                            chn.AddCmd([]int{defs.CMD_APMAC2, idx})
                                            //ToDo: fix: usesEN[2] += 1
                                            chn.UsesEffect["EN2"] = true
                                        }
                                    }
                                }
                            }
                        } else {
                            assertDisablingEffect("EN", defs.CMD_ARPOFF)
                        }

                    } else if m == 'P' {
                        characterHandled = true
                        num, idx, err := assertEffectIdExistsAndChannelsActive("EP", effects.PitchMacros)

                        if err == nil {
                            for _, chn := range CurrSong.Channels {
                                if chn.Active {
                                    if !effects.PitchMacros.IsEmpty(num) {
                                        idx |= effects.PitchMacros.GetInt(num) * 0x80
                                        chn.AddCmd([]int{defs.CMD_SWPMAC, idx})
                                        effects.PitchMacros.AddRef(num)
                                        chn.UsesEffect["EP"] = true
                                    }
                                }
                            }
                        } else {
                            assertDisablingEffect("EP", defs.CMD_SWPMAC)
                        }

                    } else {
                        Parser.Ungetch()
                        assertIsChannelName(c)
                    }

                // FM feedback select
                } else if c == 'F' {
                    m := Parser.Getch()
                    if m == 'B' {
                        s := Parser.GetNumericString()
                        if len(s) > 0 {
                            characterHandled = true
                            num, err := strconv.Atoi(s)
                            if err == nil {
                                if CurrSong.GetNumActiveChannels() == 0 {
                                    ERROR("FB requires at least one active channel")
                                } else {
                                    if inRange(num, 0, 7) {
                                        applyCmdOnAllActiveFM("FB", []int{defs.CMD_FEEDBK | num})
                                    } else {
                                        ERROR(fmt.Sprintf("Feedback value out of range: %d", num))
                                    }
                                }
                            } else {
                                ERROR("Bad feedback value: " +  s)
                            }
                        } else {
                            m = Parser.Getch()
                            if m == 'M' {
                                s = Parser.GetNumericString()
                                if len(s) > 0 {
                                    characterHandled = true
                                    num, err := strconv.Atoi(s)
                                    if err == nil {
                                        if CurrSong.GetNumActiveChannels() == 0 {
                                            ERROR("FBM requires at least one active channel")
                                        } else {
                                            idx := effects.FeedbackMacros.FindKey(num)
                                            if idx >= 0 {
                                                idx |= effects.FeedbackMacros.GetInt(num) * 0x80                
                                                for _, chn := range CurrSong.Channels {
                                                    if chn.Active {
                                                        if chn.SupportsFM() {
                                                            //if o[2] >=0 and o[2] <= 7 then
                                                                chn.AddCmd([]int{defs.CMD_FBKMAC, idx})
                                                                effects.FeedbackMacros.AddRef(num)

                                                            //else
                                                            //  ERROR(sprintf("Feedback value out of range: %d", o[2]))
                                                            //end if
                                                        } else {
                                                            WARNING("FBM commands on non-FM channels are ignored")
                                                        }
                                                    }
                                                }
                                            } else {
                                                ERROR("Undefined macro: FBM" + s)
                                            }
                                        }
                                    } else {
                                        ERROR("Bad feedback value:" + s)
                                    }
                                } else {
                                    Parser.Ungetch()
                                    Parser.Ungetch()
                                    assertIsChannelName(c)
                                }
                            } else {
                                Parser.Ungetch()
                                Parser.Ungetch()
                                assertIsChannelName(c)
                            }
                        }

                    } else if m == 'T' {
                        characterHandled = true
                        s := Parser.GetNumericString()
                        num, err := strconv.Atoi(s)
                        if err == nil {
                            idx := -1
                            if CurrSong.Target.GetID() != targets.TARGET_AT8 {
                                idx = effects.Filters.FindKey(num)
                            }
                            if CurrSong.GetNumActiveChannels() == 0 {
                                ERROR("FT requires at least one active channel")
                            } else if idx >= 0 {
                                for _, chn := range CurrSong.Channels {
                                    if chn.Active {
                                        if chn.SupportsFilter() {
                                            chn.AddCmd([]int{defs.CMD_FILTER, idx + 1})
                                            effects.Filters.AddRef(num)
                                        } else {
                                            WARNING("Unsupported command for this channel: FT")
                                        }
                                    }
                                }
                            } else if CurrSong.Target.GetID() == targets.TARGET_AT8 && num == 0 {
                                applyCmdOnAllActive("FT", []int{defs.CMD_FILTER, 1})
                            } else {
                                ERROR("Undefined macro: FT" + s)
                            }
                        } else {
                            m = Parser.Getch()
                            t := string(byte(m)) + string(byte(Parser.Getch()))
                            if t == "OF" {
                                for _, chn := range CurrSong.Channels {
                                    if chn.Active {
                                        if chn.SupportsFilter() {
                                            chn.AddCmd([]int{defs.CMD_FILTER, 0})
                                        }
                                    }
                                }
                            } else {
                                ERROR("Syntax error: FT" + s)
                            }
                        }
                    } else {
                        Parser.Ungetch()
                        assertIsChannelName(c)
                    }                   

                } else if c == 'L' {
                    if CurrSong.GetNumActiveChannels() == 0 || lastWasChannelSelect {
                        if !strings.ContainsRune(CurrSong.Target.GetChannelNames(), rune(c)) {
                            writeAllPendingNotes(true)
                            WARNING("Trying to set a loop point with no active channels")
                            characterHandled = true
                        }
                    } else {
                        for _, chn := range CurrSong.Channels {
                            chn.WriteNote(true)
                        }
                        characterHandled = true
                        for _, chn := range CurrSong.Channels {
                            if chn.Active {
                                if chn.LoopPoint == -1 {
                                    chn.LoopPoint = len(chn.Cmds)
                                    chn.LoopFrames = chn.Frames
                                } else {
                                    ERROR("Loop point already defined for channel " + chn.GetName())
                                }
                            }
                        }
                    }

        
                // Vibrato select                       
                } else if c == 'M' {
                    m := Parser.Getch()
                    if m == 'P' {
                        characterHandled = true
                        num, idx, err := assertEffectIdExistsAndChannelsActive("MP", effects.Vibratos)
                        
                        if err == nil {
                            idx |= effects.Vibratos.GetInt(num) * 0x80
                            for _, chn := range CurrSong.Channels {
                                if chn.Active {
                                    chn.AddCmd([]int{defs.CMD_VIBMAC, idx + 1})
                                    effects.Vibratos.AddRef(num)
                                    chn.UsesEffect["MP"] = true
                                }
                            }
                        } else {
                            assertDisablingEffect("MP", defs.CMD_VIBMAC);
                        }
                    } else if m == 'F' {
                        s := Parser.GetNumericString()
                        if len(s) > 0 {
                            characterHandled = true
                            num, err := strconv.Atoi(s)
                            if err == nil {
                                if inRange(num, 0, 15) {
                                    if CurrSong.GetNumActiveChannels() == 0 {
                                        ERROR("MF requires at least one active channel")
                                    } else {
                                        for _, chn := range CurrSong.Channels {
                                            if chn.Active {
                                                if chn.SupportsFM() || CurrSong.Target.GetID() == targets.TARGET_AT8 {
                                                    if CurrSong.Target.GetID() == targets.TARGET_AT8 {
                                                        if num == 0 {           // 15 kHz
                                                            num = 1
                                                        } else if num == 1 {    // 64 kHz
                                                            num = 2
                                                        } else if num == 2 {
                                                            num = 3             // CPU clock    
                                                        }
                                                    }
                                                    chn.AddCmd([]int{defs.CMD_MULT, num})
                                                } else {
                                                    WARNING("MF ignored for non-FM channel")
                                                }
                                            }
                                        }
                                    }
                                } else {
                                    ERROR("MF out of range: " + s)
                                } 
                            } else {
                                ERROR("Bad MF: " + s)
                            }
                        } else {
                            Parser.Ungetch()
                            assertIsChannelName(c)
                        }
                    } else if m == 'O' {
                        m = Parser.Getch()
                        if m == 'D' {
                            s := Parser.GetNumericString()
                            if len(s) > 0 {
                                characterHandled = true
                                num, err := strconv.Atoi(s)
                                if err == nil {
                                    idx := effects.MODs.FindKey(num)
                                    if CurrSong.GetNumActiveChannels() == 0 {
                                        ERROR("MOD requires at least one active channel")
                                    } else if idx >= 0 {
                                        for _, chn := range CurrSong.Channels {
                                            if chn.Active {
                                                if chn.SupportsFM() || CurrSong.Target.GetID() == targets.TARGET_PCE {
                                                    chn.AddCmd([]int{defs.CMD_MODMAC, idx + 1})
                                                    effects.MODs.AddRef(num)
                                                }
                                            }
                                        }
                                    } else {
                                        ERROR("Undefined macro: MOD" + s)
                                    }
                                } else {
                                    ERROR("Bad MOD: " + s)
                                }
                            } else {
                                m = Parser.Getch()
                                t := string(byte(m)) + string(byte(Parser.Getch()))
                                if t == "OF" {
                                    characterHandled = true
                                    for _, chn := range CurrSong.Channels {
                                        if chn.Active {
                                            if chn.SupportsFM() || CurrSong.Target.GetID() == targets.TARGET_PCE {
                                                chn.AddCmd([]int{defs.CMD_MODMAC, 0})
                                            }
                                        }
                                    }
                                } else {
                                    Parser.Ungetch()
                                    Parser.Ungetch()
                                    Parser.Ungetch()
                                    Parser.Ungetch()
                                    assertIsChannelName(c)
                                }
                            }
                        } else {
                            Parser.Ungetch()
                            Parser.Ungetch()
                            assertIsChannelName(c)
                        }
                    } else if IsNumeric(m) {
                        characterHandled = true
                        Parser.Ungetch()
                        s := Parser.GetNumericString()
                        num, err := strconv.Atoi(s)
                        if err == nil {
                            if CurrSong.GetNumActiveChannels() == 0 {
                                ERROR("M requires at least one active channel")
                            } else if inRange(num, 0, 15) {
                                applyCmdOnAllActive("M", []int{defs.CMD_MODE | num})
                            } else {
                                ERROR("Bad mode: " + s)
                            }
                        } else {
                            ERROR("Bad mode:" + s)
                        }
                    } else {
                        Parser.Ungetch()
                        assertIsChannelName(c)
                    }

                // FM operator select
                } else if c == 'O' {
                    m := Parser.Getch()
                    if m == 'P' && !lastWasChannelSelect {
                        s := Parser.GetNumericString()
                        if len(s) > 0 {
                            characterHandled = true
                            num, err := strconv.Atoi(s)
                            if err == nil {
                                if inRange(num, 0, 4) {
                                    if CurrSong.GetNumActiveChannels() == 0 {
                                        ERROR("OP requires at least one active channel")
                                    } else {
                                        applyCmdOnAllActiveFM("OP", []int{defs.CMD_OPER | num})
                                    }
                                } else {
                                    ERROR("OP out of range: " + s)
                                }
                            } else {
                                ERROR("Bad OP: " + s)
                            }
                        } else {
                            Parser.Ungetch()
                            assertIsChannelName(c)
                        }
                    } else {
                        Parser.Ungetch()
                        assertIsChannelName(c)
                    }

                // Portamento select
                } else if c == 'P' {
                    m := Parser.Getch()
                    if m == 'T' {
                        characterHandled = true
                        s := Parser.GetNumericString()
                        num, err := strconv.Atoi(s)
                        if err == nil {
                            idx := effects.Portamentos.FindKey(num)
                            if CurrSong.GetNumActiveChannels() == 0 {
                                ERROR("PT requires at least one active channel")
                            } else if idx >= 0 {
                                // TODO: set portamento
                                WARNING("PT command not yet implemented")
                            } else {
                                ERROR("Undefined macro: PT" + s)
                            }
                        } else {
                            m = Parser.Getch()
                            t := string(byte(m)) + string(byte(Parser.Getch()))
                            if t == "OF" {
                                // TODO: deactivate portamento
                            } else {
                                ERROR("Syntax error: PT" + s)
                            }
                        }
                    } else {
                        Parser.Ungetch()
                        assertIsChannelName(c)
                    }

                // Rate scale / Ring modulation
                } else if c == 'R' {
                    m := Parser.Getch()
                    if m == 'I' {
                        s := string(byte(Parser.Getch()))
                        s += string(byte(Parser.Getch()))
                        if s == "NG" {
                            characterHandled = true
                            s = Parser.GetNumericString()
                            num, err := strconv.Atoi(s)
                            if err == nil {
                                if CurrSong.GetNumActiveChannels() == 0 {
                                    ERROR("RING requires at least one active channel")
                                } else if inRange(num, 0, 1) {
                                    for _, chn := range CurrSong.Channels {
                                        if chn.Active {
                                            if chn.SupportsRingMod() {
                                                chn.AddCmd([]int{defs.CMD_HWRM, num})
                                            } else {
                                                WARNING("Unsupported command for this channel: RING")
                                            }
                                        }
                                    }
                                } else {
                                    ERROR("RING out of range: " + s)
                                }
                            } else {
                                ERROR("Syntax error: RING" + s)
                            }
                        } else {
                            Parser.Ungetch()
                            Parser.Ungetch()
                            Parser.Ungetch()
                            assertIsChannelName(c)
                        }
                        
                    } else if m == 'S' && !lastWasChannelSelect {
                        characterHandled = true
                        s := Parser.GetNumericString()
                        num, err := strconv.Atoi(s)
                        if err == nil {
                            if CurrSong.GetNumActiveChannels() == 0 {
                                ERROR("RS requires at least one active channel")
                            } else if inRange(num, 0, 3) {
                                applyCmdOnAllActiveFM("RS", []int{defs.CMD_RSCALE, num})
                            } else {
                                ERROR("RS out of range: " + s)
                            }
                        } else {
                            ERROR("Syntax error: RS" + s)
                        }
                    } else {
                        Parser.Ungetch()
                        assertIsChannelName(c)
                    }

                // SSG-EG mode / Hard sync
                } else if c == 'S' {
                    m := Parser.Getch()
                    if m == 'Y' {
                        s := string(byte(Parser.Getch()))
                        s += string(byte(Parser.Getch()))
                        if s == "NC" {
                            characterHandled = true
                            s = Parser.GetNumericString()
                            num, err := strconv.Atoi(s)
                            if err == nil {
                                if CurrSong.GetNumActiveChannels() == 0 {
                                    ERROR("SYNC requires at least one active channel")
                                } else if inRange(num, 0, 1) {
                                    for _, chn := range CurrSong.Channels {
                                        if chn.Active {
                                            if chn.SupportsRingMod() {
                                                chn.AddCmd([]int{defs.CMD_SYNC, num})
                                            } else {
                                                WARNING("Unsupported command for this channel: SYNC")
                                            }
                                        }
                                    }
                                } else {
                                    ERROR("SYNC out of range: " + s)
                                }
                            } else {
                                ERROR("Syntax error: SYNC" + s)
                            }
                        } else {
                            Parser.Ungetch()
                            Parser.Ungetch()
                            Parser.Ungetch()
                            assertIsChannelName(c)
                        }
                    
                    } else if m == 'S' {
                        s := string(byte(m))
                        s += string(byte(Parser.Getch()))
                        if s == "SG" {
                            characterHandled = true
                            s = Parser.GetNumericString()
                            num, err := strconv.Atoi(s)
                            if err == nil {
                                if CurrSong.GetNumActiveChannels() == 0 {
                                    ERROR("SSG requires at least one active channel")
                                } else if inRange(num, 0, 7) {
                                    applyCmdOnAllActiveFM("SSG", []int{defs.CMD_SSG, num + 1})
                                } else {
                                    ERROR("SSG out of range: " + s)
                                }
                            } else {
                                m = Parser.Getch()
                                t := string(byte(m)) + string(byte(Parser.Getch()))
                                if t == "OF" {
                                    for _, chn := range CurrSong.Channels {
                                        if chn.Active {
                                            if chn.SupportsFM() {
                                                chn.AddCmd([]int{defs.CMD_SSG, 0})
                                            } else {
                                                WARNING("SSG commands not supported for channel " + chn.GetName())
                                            }
                                        }
                                    }
                                } else {
                                    ERROR("Syntax error: SSG" + s)
                                }
                            }
                        }
                    } else {
                        Parser.Ungetch()
                        assertIsChannelName(c)
                    }

                } else if c == 'W' {
                    m := Parser.Getch()
                    if m == 'T' {
                        characterHandled = true
                        isWTM := false
                        m = Parser.Getch()
                        if m == 'M' {
                            isWTM = true
                        } else {
                            Parser.Ungetch()
                        }
                        s := Parser.GetNumericString()
                        num, err := strconv.Atoi(s)
                        if err == nil {
                            if !isWTM {
                                idx := effects.Waveforms.FindKey(num)
                                if CurrSong.GetNumActiveChannels() == 0 {
                                    ERROR("WT requires at least one active channel")
                                } else if idx >= 0 {
                                    if CurrSong.GetNumActiveChannels() > 0 {
                                        applyEffectOnAllActiveSupported("WT", []int{defs.CMD_LDWAVE, idx + 1},
                                                                        func(c *channel.Channel) bool { return c.SupportsWaveTable() },
                                                                        effects.Waveforms, num)
                                    } else {
                                        WARNING("Trying to use WT with no channels active")
                                    }
                                } else if CurrSong.Target.SupportsWaveTable() {
                                    ERROR("Undefined macro: WT" + s)
                                }
                            } else {
                                idx := effects.WaveformMacros.FindKey(num)
                                if CurrSong.GetNumActiveChannels() == 0 {
                                    ERROR("WTM requires at least one active channel")
                                } else if idx >= 0 {
                                    if CurrSong.GetNumActiveChannels() > 0 {
                                        applyEffectOnAllActiveSupported("WTM", []int{defs.CMD_WAVMAC, idx + 1},
                                                                        func(c *channel.Channel) bool { return c.SupportsWaveTable() },
                                                                        effects.WaveformMacros, num)
                                        /*for i = 1 to length(supportedChannels) do
                                            if activeChannels[i] then
                                                if supportsWave[i] then
                                                    chn.AddCmd([]int{defs.CMD_WAVMAC, idx})
                                                    waveformMacros = assoc_reference(waveformMacros, o[2])
                                                else
                                                    WARNING("Unsupported command for channel " & supportedChannels[i] & ": WTM")
                                                end if
                                            end if
                                        end for*/
                                    } else {
                                        WARNING("Trying to use WTM with no channels active")
                                    }
                                } else if CurrSong.Target.SupportsWaveTable() {
                                    ERROR("Undefined macro: WTM" + s)
                                }
                            }
                        } else {
                            m = Parser.Getch()
                            t := string(byte(m)) + string(byte(Parser.Getch()))
                            if t == "OF" {
                                for _, chn := range CurrSong.Channels {
                                    if chn.Active {
                                        if chn.SupportsWaveTable() {
                                            chn.AddCmd([]int{defs.CMD_LDWAVE, 0})
                                        } else {
                                            WARNING("WT commands not supported for channel " + chn.GetName())
                                        }
                                    }
                                }
                            } else {
                                ERROR("Syntax error: WT" + s)
                            }
                        }
                    } else {
                        Parser.Ungetch()
                        assertIsChannelName(c)
                    }

                // Noise speed
                } else if c == 'n' {
                    characterHandled = true
                    s := Parser.GetNumericString()
                    num, err := strconv.Atoi(s)
                    if err == nil && inRange(num, 0, 63) {
                        if CurrSong.GetNumActiveChannels() == 0 {
                            ERROR("n requires at least one active channel")
                        } else {
                            if CurrSong.GetNumActiveChannels() > 0 {
                                for _, chn := range CurrSong.Channels {
                                    if chn.Active {
                                        switch CurrSong.Target.GetID() {
                                        case targets.TARGET_AST, targets.TARGET_KSS, targets.TARGET_CPC:
                                            chn.AddCmd([]int{defs.CMD_HWNS, num ^ 0x3F})
                                        default:
                                            WARNING("Unsupported command for this channel: n")
                                        }
                                    }
                                }
                            } else {
                                WARNING("Trying to use n with no channels active")
                            }
                        }
                    } else {
                        ERROR("Bad n: " + s)
                    }
                    
                } else if c == 'p' {
                    m := Parser.Getch()
                    if m == 'w' {
                        characterHandled = true
                        s := Parser.GetNumericString()
                        num, err := strconv.Atoi(s)
                        if err == nil && inRange(num, 0, 15) {
                            if CurrSong.GetNumActiveChannels() == 0 {
                                ERROR("pw requires at least one active channel")
                            } else {
                                if CurrSong.GetNumActiveChannels() > 0 {
                                    applyCmdOnAllActive("pw", []int{defs.CMD_PULSE | num})
                                } else {
                                    WARNING("Trying to use pw with no channels active")
                                }
                            }
                        } else {
                            ERROR("Bad pw: " + s)
                        }
                    } else {
                        Parser.Ungetch()
                        assertIsChannelName(c)
                    }
                
                } else if c == 'w' {
                    wrType := defs.CMD_WRMEM
                    m := Parser.Getch()
                    if m == '(' {
                        wrType = defs.CMD_WRPORT
                    } else {
                        Parser.Ungetch()
                    }
                    s := Parser.GetNumericString()
                    if len(s) > 0 {
                        addr, err := strconv.Atoi(s)
                        if err == nil {
                            if inRange(addr, 0, 0xFFFF) {
                                m = Parser.Getch()
                                if wrType == defs.CMD_WRPORT {
                                    if m != ')' {
                                        ERROR("Expected ')'")
                                    }
                                    m = Parser.Getch()
                                }
                                if m != ',' {
                                    ERROR("Expected ','")
                                }
                                s = Parser.GetNumericString()
                                if len(s) == 0 {
                                    ERROR("Missing second argument for w")
                                }
                                val, err := strconv.Atoi(s)
                                if err != nil {
                                    ERROR("Bad second argument for w")
                                }
                                for _, chn := range CurrSong.Channels {
                                    if chn.Active {
                                        go chn.AddCmd([]int{wrType,
                                                         (addr & 0xFF),
                                                         (addr / 0x100),
                                                         (val & 0xFF)})
                                    }
                                }
                                characterHandled = true
                            } else {
                                ERROR("Memory address out of range: " + s)
                            }
                        } else {
                            ERROR("Bad first argument for w: " + s)
                        }
                    } else {
                        ERROR("Bad first argument for w")
                    }
                }
                
                if !characterHandled {
                    for _, chn := range CurrSong.Channels {
                        if !lastWasChannelSelect {
	                    chn.Active = false
	                }
                        if chn.GetName() == string(byte(c)) {
                            chn.Active = true
                        }
                    }
                }
            } else if strings.ContainsRune("\t\r\n ", rune(c)) {
                for _, chn := range CurrSong.Channels {
                    chn.WriteNote(c == '\n')
                }
            } else {
                if c == '%' {
                    ERROR("Unexpected character: %%")
                } else {
                    ERROR("Unexpected character: " + string(byte(c)))
                }
            }

            lastWasChannelSelect = strings.ContainsRune(CurrSong.Target.GetChannelNames(), rune(c2))
        }               
                
    }
    
    Parser = OldParsers.Pop()
}
