package main

import (
    "fmt"
    "./compiler"
    "./defs"
    "./targets"
    "./timing"
)


var target int


func showHelp(what string) {
    switch what {
    case "EN":
        fmt.Println("EN:\tArpeggio macro: { num | num}. num is zero or more numbers in\n\tthe range -63-63.\n")
    default:
        fmt.Println("Usage: xpmc.exe [options] target input [output]\n")
        fmt.Println("Options:")
        fmt.Println("\t-h\tShow this information") 
        fmt.Println("\t-v\tVerbose mode")
        fmt.Println("\t-w\tTreat warnings as errors")
        fmt.Println("\nTarget:")
        fmt.Println("\t-at8\tAtari 8-bit")
        fmt.Println("\t-c64\tCommodore 64")
        fmt.Println("\t-clv\tColecoVision")
        fmt.Println("\t-cpc\tAmstrad CPC")
        fmt.Println("\t-cps\tCPS-1")
        //fmt.Println(1, "\t-gba\tGameboy Advance")
        fmt.Println("\t-gbc\tGameboy / Gameboy Color")
        fmt.Println("\t-gen\tSEGA Genesis")
        fmt.Println("\t-kss\tKSS")
        fmt.Println("\t-nes\tNintendo Entertainment System")
        fmt.Println("\t-pce\tPC-Engine")
        //fmt.Println(1, "\t-nds\tNintendo DS")
        //fmt.Println(1, "\t-sat\tSEGA Saturn")
        fmt.Println("\t-sgg\tSEGA Game Gear")
        fmt.Println("\t-sms\tSEGA Master System")

    }
}


func main() {
    timing.UpdateFreq = 60.0
    //c.WriteLength()
    //compiler.WARNING("foobar")
    

    if timing.UseFractionalDelays {
        timing.SupportedLengths = defs.EXTENDED_LENGTHS()
    }
    
    compiler.Verbose(false)
    compiler.WarningsAreErrors(false)
    target = targets.TARGET_UNKNOWN

    showHelp("")
}
