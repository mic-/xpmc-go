/*
 * Package utils
 * ParamList functions
 *
 * Part of XPMC.
 * Contains related to ParamLists (all the lists on the
 * form { foo bar ... | baz ... } found in MML code).
 *
 * /Mic, 2012-2014
 */
 
package utils

import (
    "fmt"
)

type ParamList struct {
    currentPart int
    currentPos int
    MainPart []interface{} 
    LoopedPart []interface{} 
}

const (
    MAIN_PART = 0
    LOOPED_PART = 1
)


func NewParamList() *ParamList {
    return &ParamList{0, 0, []interface{}{}, []interface{}{}}
}


/* Returns true if the list contains no data.
 */
func (lst *ParamList) IsEmpty() bool {
    return len(lst.MainPart) == 0 && len(lst.LoopedPart) == 0
}


func (lst *ParamList) AppendToPart(part int, val int) {
    if part == MAIN_PART {
        lst.MainPart = append(lst.MainPart, val)
    } else if part == LOOPED_PART {
        lst.LoopedPart = append(lst.LoopedPart, val)
    }
}

func (lst *ParamList) GetPart(part int) *[]interface{} {
    if part == MAIN_PART {
        return &lst.MainPart
    } else if part == LOOPED_PART {
        return &lst.LoopedPart
    }
    return nil
}


/* Formats a list into a string for printing.
 */
func (lst *ParamList) Format() string {
    str := "{"
      
    for part := MAIN_PART; part <= LOOPED_PART; part++ {
        listPart := lst.GetPart(part)
        for i, x := range *listPart {
            switch x.(type) {
            case int:
                str += fmt.Sprintf("%d", x)
            case string:
                str += fmt.Sprintf("\"%s\"", x)
            default:
                if itfSlice, isItfSlice := x.([]interface{}); isItfSlice {
                    for _, innerElem := range itfSlice {
                        switch innerElem.(type) {
                        case int:
                            str += fmt.Sprintf("%d", innerElem)
                        case string:
                            str += fmt.Sprintf("%s", innerElem)
                        }
                    }
                } else {
                    str += "<UNKNOWN TYPE>"
                }
            }
            if i < len(*listPart)-1 {
                str += " "
            }
        }
    }

    str += "}"
    return str
}


func (lst *ParamList) MoveToStart() {
    lst.currentPart = 0
    lst.currentPos = 0
}


func (lst *ParamList) Step() {
    lst.currentPos++
    if lst.currentPart == 0 {
        if lst.currentPos >= len(lst.MainPart) {
            if len(lst.LoopedPart) > 0 {
                lst.currentPart++
                lst.currentPos = 0
            } else {
                lst.currentPos = len(lst.MainPart) - 1
            }
        }
    } else if lst.currentPos >= len(lst.LoopedPart) {
        lst.currentPos = 0
    }
}

func (lst *ParamList) Peek() int {
    if lst.currentPart == 1 {
        return lst.LoopedPart[lst.currentPos].(int)
    }
    return lst.MainPart[lst.currentPos].(int)
}


func (lst *ParamList) Equal(comparedTo *ParamList) bool {
    if len(lst.MainPart) != len(comparedTo.MainPart) ||
       len(lst.LoopedPart) != len(comparedTo.LoopedPart) {
        return false
    }
    
    for i, elem := range lst.MainPart {
        if elem != comparedTo.MainPart[i] {
            return false
        }
    }
    
    for i, elem := range lst.LoopedPart {
        if elem != comparedTo.LoopedPart[i] {
            return false
        }
    }
    
    return true
}
