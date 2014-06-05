/*
 * Package channel
 * LoopStack implementation
 *
 * Part of XPMC.
 * Defines elements that can represent MML loops (e.g. [cde | d+g]3 ) and stacks
 * on which you can store such elements.
 *
 * /Mic, 2014
 */
 
package channel

import "../utils"

type LoopStackElem struct {
    StartPos int    /* The index within the channel's command sequence where the first command
                       following the '[' occurs. */
    StartTicks int  /* The number of ticks (32nd notes) that the channel contains at the point where
                       the loop starts. */
    Unknown int
    Skip1Pos int    /* The index within the channel's command sequence where the first command
                       following the '|' occurs. */
    Skip1Ticks int  /* The number of ticks that the channel contains at the point where the part of
                       loop following the '|' starts. */
    OrigOctave int  /* The current octave at the point where the loop starts. */
    OctChange int   /* The relative octave change within the first part of the loop. */
    HasOctCmd int   /* Set if there's an absolute octave command ('o') within the first part of the loop. */
    Skip1OctChg int /* The relative octave change within the part of the loop following the '|'. */
    Skip1OctCmd int /* Set if there's an absolute octave command ('o') within the part of the loop
                       following the '|'. */
}

type LoopStack struct {
    utils.GenericStack
}

func NewLoopStack() *LoopStack {
    return &LoopStack{*utils.NewGenericStack()}
}

func (s *LoopStack) PopLoop() LoopStackElem {
    return s.Pop().(LoopStackElem)
}

func (s *LoopStack) PeekLoop() LoopStackElem {
    return s.Peek().(LoopStackElem)
}

