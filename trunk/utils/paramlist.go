/*
 * Package utils
 * ParamList functions
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
)

type ParamList struct {
    currentPart int
    currentPos int
    MainPart []interface{} 
    LoopedPart []interface{} 
}


/* Returns true if the list contains no data.
 */
func (lst *ParamList) IsEmpty() bool {
    return len(lst.MainPart) == 0 && len(lst.LoopedPart) == 0
}


/* Formats a list into a string for printing.
 */
func (lst *ParamList) Format() string {
    str := "{"
    
    for i, x := range lst.MainPart {
        switch x.(type) {
        case int:
            str += fmt.Sprintf("%d", x)
        case string:
            str += fmt.Sprintf("\"%s\"", x)
        default:
            str += "<UNKNOWN TYPE>"
        }
        if i < len(lst.MainPart)-1 {
            str += " "
        }
    }
    
    if len(lst.LoopedPart) > 0 {
        str += " | "
        for i, x := range lst.LoopedPart {
            switch x.(type) {
            case int:
                str += fmt.Sprintf("%d", x)
            case string:
                str += fmt.Sprintf("\"%s\"", x)
            default:
                str += "<UNKNOWN TYPE>"
            }
            if i < len(lst.LoopedPart)-1 {
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
