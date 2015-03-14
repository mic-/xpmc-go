package compiler

import (
    "sort"
    "../defs"
)

func (comp *Compiler) GetShortFileName() string {
    return comp.ShortFileName
}

func (comp *Compiler) GetNumSongs() int {
    return len(comp.Songs)
}

func (comp *Compiler) GetPatterns() []defs.IMmlPattern {
    patterns := make([]defs.IMmlPattern, len(comp.patterns.data))
    for i, pat := range comp.patterns.data {
        patterns[i] = pat
    }
    return patterns
}

func (comp *Compiler) GetCurrentSong() defs.ISong {
    return comp.CurrSong
}

func (comp *Compiler) GetSong(num int) defs.ISong {
    return comp.Songs[num]
}

func (comp *Compiler) GetSongs() []defs.ISong {
    songs := make([]defs.ISong, len(comp.Songs))
    keys := []int{}
    for key := range comp.Songs {
        keys = append(keys, key)
    }
    sort.Ints(keys)
    for i, key := range keys {
        songs[i] = comp.Songs[key]
    }
    return songs
}

func (comp *Compiler) GetCallbacks() []string {
    return comp.callbacks
}