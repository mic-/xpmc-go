/*
 * Package defs
 *
 * Part of XPMC.
 * Contains common constants, types and variables.
 *
 * /Mic, 2012-2014
 */
 
package defs

const (
    CMD_NOTE   = 0x00   // cc+dd+eff+gg+aa+b
    CMD_REST   = 0x0C   // r    
    CMD_REST2  = 0x0D   // s
    CMD_VOLUP  = 0x0E   // v+
    CMD_VOLDN  = 0x0F   // v-
    CMD_OCTAVE = 0x10   // o
    CMD_DUTY   = 0x20   // @
    CMD_VOL2   = 0x30   //
    CMD_OCTUP  = 0x40   // >
    CMD_OCTDN  = 0x50   // <
    CMD_VOLOUP = 0x60   //
    CMD_NOTE2  = 0x60   // (short form)
    CMD_REST3  = 0x6D   // r (short form)
    CMD_VOLUPC = 0x6E   // v++
    CMD_VOLDNC = 0x6F   // v--
    CMD_VOLODN = 0x70   //
    CMD_PULSE  = 0x70   //
    CMD_ARP2   = 0x80   //
    CMD_ARPOFF = 0x90   //
    CMD_FBKMAC = 0x191  // FBM
    CMD_SSG    = 0x92   //
    CMD_FILTER = 0x193  // FT
    CMD_HWRM   = 0x94   //
    CMD_PULMAC = 0x195  // @pw
    CMD_JSR    = 0x196  //

    CMD_RTS    = 0x197
    CMD_SYNC   = 0x98
    CMD_HWES   = 0x98   // es
    CMD_HWNS   = 0x99   // n
    CMD_LEN    = 0x9A   // l
    CMD_WRMEM  = 0x9B   // w
    CMD_WRPORT = 0x9C   // w()
    CMD_RLEN   = 0x9D
    CMD_WAVMAC = 0x19E  // WTM
    CMD_TRANSP = 0x9F   // K
    CMD_MODE   = 0xA0   // M
    CMD_FEEDBK = 0xB0   // FB
    CMD_OPER   = 0xC0   // O
    CMD_RSCALE = 0xD0   // RS
    CMD_CBOFF  = 0xE0
    CMD_CBONCE = 0xE1
    CMD_CBEVNT = 0xE2
    CMD_CBEVVC = 0xE3
    CMD_CBEVVM = 0xE4
    CMD_CBEVOC = 0xE5
    CMD_CBNOTE = 0xE6
    CMD_HWTE   = 0xE8   // @te
    CMD_HWVE   = 0xE9   // @ve
    CMD_HWAM   = 0xEA
    CMD_MODMAC = 0x1EB  //- MOD
    CMD_LDWAVE = 0x1EC  // WT
    CMD_DETUNE = 0xED   // D
    CMD_ADSR   = 0x1EE  // ADSR
        
    CMD_MULT   = 0xEF   // MF
    CMD_VOLSET = 0xF0   // v
    CMD_VOLMAC = 0x1F1  // @v
    CMD_DUTMAC = 0x1F2  // @@
    CMD_PORMAC = 0xF3
    CMD_PANMAC = 0x1F4  // CS
    CMD_VIBMAC = 0x1F5  // MP
    CMD_SWPMAC = 0x1F6  // EP
    CMD_VSLMAC = 0xF7
    CMD_ARPMAC = 0x1F8  // EN
    CMD_JMP    = 0xF9
    CMD_DJNZ   = 0xFA
    CMD_LOPCNT = 0xFB
    CMD_APMAC2 = 0x1FC
    CMD_J1     = 0xFD
    CMD_END    = 0xFF
)

        
// Cutoff types
const (
    CT_NORMAL = 1
    CT_FRAMES = 2
    CT_NEG = 3
    CT_NEG_FRAMES = 4
)

const (
    EFFECT_STEP_EVERY_FRAME = 0
    EFFECT_STEP_EVERY_NOTE = 1
)

var notes []int = []int{'c', -2, 'd', -2, 'e', 'f', -2, 'g', -2, 'a', -2, 'b', 'r', 's'}

func NoteIndex(n int) int {
    for i, c := range notes {
        if c == n {
            return i
        }
    }
    return -1
}

func NoteVal(idx int) int {
    if idx >= 0 {
        return notes[idx]
    }
    return 0
}


var extendedLengths []int = []int{32, 24, 16, 12, 8, 6, 4, 3, 2, 1}

func EXTENDED_LENGTHS() []int {
    return extendedLengths
}

var EFFECT_STRINGS [6]string = [6]string{"EN", "EN2", "EP", "MP", "DM", "PM"}

const NON_NOTE_TUPLE_CMD = 0xFFFF

var Rest, Rest2 int

/////////////

type ISpecs interface {
    GetDuty() []int    
    GetVolChange() []int     
    GetFM() []int            
    GetADSR() []int          
    GetFilter() []int        
    GetRingMod() []int      
    GetWaveTable() []int     
    GetPCM() []int          
    GetToneEnv() []int      
    GetVolEnv() []int        
    GetDetune() []int       
    GetMinOct() []int     
    GetMaxOct() []int      
    GetMaxVol() []int       
    GetMinNote() []int       
    GetID() int             
    GetIDs() []int           
}

type ITarget interface {
    GetAdsrLen() int        // Number of parameters used for ADSR envelopes on this target
    GetAdsrMax() int        // Max value for ADSR parameters on this target
    GetChannelSpecs() ISpecs
    GetChannelNames() string
    GetCompilerItf() ICompiler
    GetID() int             // The ID of this target (one of the TARGET_* constants)
    GetMaxLoopDepth() int   // Max nesting of [] loops on this target
    GetMaxTempo() int
    GetMaxVolume() int
    GetMaxWavLength() int   // Max length of WT samples on this target
    GetMaxWavSample() int   // Max amplitude for WT samples on this target
    GetMinVolume() int
    GetMinWavLength() int
    GetMinWavSample() int
    Init()
    Output(outputFormat int)
    SupportsPAL() bool
    SupportsPan() bool      // Whether this target supports panning effects (CS)
    SupportsPCM() bool      // Whether this target supports one-shot PCM samples (XPCM)
    SupportsWaveTable() bool
    SetCompilerItf(icomp ICompiler)
    PutExtraInt(name string, val int)
}

type ISong interface {
    GetChannels() []IChannel
    GetComposer() string
    GetNum() int            // The song's number (e.g. 2 for "#SONG 2")
    GetProgrammer() string
    GetSmsTuning() bool
    GetTitle() string
    UsesChip(chipId int) bool
}

type IChannel interface {
    GetCommands() []int
    GetName() string
    GetNum() int
    GetTicks() int
    GetLoopTicks() int
    GetChipID() int
    IsUsed() bool
    IsUsingEffect(effName string) bool
    IsVirtual() bool
    SetMaxOctave(maxOct int)
}

type IMmlPattern interface {
    GetCommands() []int
}

type ICompiler interface {
    GetCallbacks() []string
    GetCurrentSong() ISong
    GetGbNoiseType() int
    GetGbVolCtrlType() int
    GetNumSongs() int
    GetPatterns() []IMmlPattern
    GetShortFileName() string
    GetSong(num int) ISong
    GetSongs() []ISong
    SetCommandHandler(cmd string, handler func(string, ITarget))
    SetMetaCommandHandler(cmd string, handler func(string, ITarget))
}
