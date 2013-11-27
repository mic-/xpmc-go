package main

import (
    "fmt"
    "os"
    "./compiler"
    "./defs"
    "./song"
    "./targets"
    "./timing"
    "./utils"
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
    timing.UseFractionalDelays = true
    //c.WriteLength()
    //compiler.WARNING("foobar")
    

    if timing.UseFractionalDelays {
        timing.SupportedLengths = defs.EXTENDED_LENGTHS()
    }
    
    utils.OldParsers = utils.NewParserStateStack()
    
    compiler.Verbose(false)
    compiler.WarningsAreErrors(false)
    target = targets.TARGET_UNKNOWN

    for i, arg := range os.Args {
        if i >= 1 {
            if arg[0] == '-' {
                if arg == "-v" {
                    compiler.Verbose(true);
                } else if arg == "-w" {
                    compiler.WarningsAreErrors(true);
                } else if arg == "-h" || arg == "-help" {
                    showHelp("")
                    return
                } else if targets.NameToID(arg[1:]) != targets.TARGET_UNKNOWN {
                    target = targets.NameToID(arg[1:])
                } 
            } else {
                if target == targets.TARGET_UNKNOWN {
                    fmt.Printf("Error: No target platform specified.\nRun the compiler with the -h option to see a list of targets.")
                    return
                }
                compiler.Init()
                compiler.CurrSong = song.NewSong(target)
                compiler.CompileFile(arg);
                return
            }
        }
    }

    showHelp("")
}
