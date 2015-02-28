package targets

import (
    "os"
    "../effects"
)
import . "../defs"

const (
    SYNTAX_WLA_DX = 0
    SYNTAX_GAS_68K = 1
)

type ICodeGenerator interface {
    OutputCallbacks(outFile *os.File) int
    OutputChannelData(outFile *os.File) int
    OutputEffectFlags(outFile *os.File)
    OutputPatterns(outFile *os.File) int
    OutputTable(outFile *os.File, tblName string, effMap *effects.EffectMap, canLoop bool, scaling int, loopDelim int) int
}
    
type CodeGenerator struct {
    itarget ITarget
}

type CodeGeneratorWla struct {
    CodeGenerator
}

type CodeGeneratorGas68k struct {
    CodeGenerator
}

func NewCodeGenerator(cgID int, itarget ITarget) ICodeGenerator {
    var cg ICodeGenerator = ICodeGenerator(nil)
    
    switch cgID {
    case SYNTAX_WLA_DX:
        cg = &CodeGeneratorWla{CodeGenerator: CodeGenerator{itarget}}

    case SYNTAX_GAS_68K:
        cg = &CodeGeneratorGas68k{CodeGenerator: CodeGenerator{itarget}}
    }
      
    return cg
}