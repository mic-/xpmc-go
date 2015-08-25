/*
 * Package vgm
 *
 * Part of XPMC.
 * Contains data/functions related to VGM file generation
 *
 * TODO:
 *  - Add proper support for the YM2413.
 *  - Add support for the YM3812.
 *  - Add support for the RF5C68 
 *
 * /Mic, 2012
 */
 

package vgm

import "../utils"

const (
    // VGM commands as given in the VGM specification
    VGM_CMD_W_PSG       = 0x50
    VGM_CMD_W_YM2413    = 0x51
    VGM_CMD_W_YM2612L   = 0x52
    VGM_CMD_W_YM2612H   = 0x53
    VGM_CMD_W_YM2151    = 0x54
    VGM_CMD_W_YM3812    = 0x55
    VGM_CMD_W_RF5C68    = 0x5F
}


const (
    YM2413_MODULATOR= 0
    YM2413_CARRIER  = 1
    
    YM2413_RHYTHM_ENABLE = 0x20
)

const (
    // SN76489 command masks
    SN76489_TONE_DATA = 0x00
    SN76489_TONE_LATCH= 0x80
    SN76489_VOL_DATA  = 0x10
    SN76489_VOL_LATCH = 0x90
)

const (
    // YM2151 registers
    R_YM2151_KEYON  = 0x08
    R_YM2151_NOISE  = 0x0F
    R_YM2151_LFO_F  = 0x18
    R_YM2151_PH_AM_D= 0x19
    R_YM2151_CT_LFOW= 0x1B
    R_YM2151_CONN_FB= 0x20
    R_YM2151_KEYCODE= 0x28
    R_YM2151_PH_AM_S= 0x38
    R_YM2151_DT_MUL = 0x40
    R_YM2151_TL     = 0x60
    R_YM2151_EG_ATK = 0x80
    R_YM2151_EG_DEC1= 0xA0
    R_YM2151_EG_DEC2= 0xC0
    R_YM2151_EG_SR  = 0xE0

    // YM2413 registers
    R_YM2413_MODEMUL= 0x00  
    R_YM2413_MOD_TL = 0x02
    R_YM2413_ATK_DEC= 0x04
    R_YM2413_SUS_REL= 0x06
    R_YM2413_RHYTHM = 0x0E
    R_YM2413_FLO    = 0x10
    R_YM2413_FHI_CTL= 0x20
    R_YM2413_INS_VOL= 0x30

    // YM2612 registers
    R_YM2612_LFO    = 0x22
    R_YM2612_CH3_6  = 0x27
    R_YM2612_KEYON  = 0x28
    R_YM2612_DAC_EN = 0x2B
    R_YM2612_DT_MUL = 0x30
    R_YM2612_TL     = 0x40
    R_YM2612_EG_SR  = 0x80
    R_YM2612_SSG_EG = 0x90
    R_YM2612_FLO    = 0xA0
    R_YM2612_FHI_BLK= 0xA4
    R_YM2612_CONN_FB= 0xB0
    R_YM2612_PH_AM_S= 0xB4
)




type LastChannelSettings struct {
    Freq    int
    Volume volume
    Panning int
    DacOn bool
}




// Write a VGM file based on the compiled song data.
//
// Arguments:
//
//  fname:  Filename of the VGM
//  song:   Song number (1..number of songs in mml)
//  psg:    Non-zero if PSG is used in the song
//  ym2151: Non-zero if YM2151 is used in the song
//  ym2413: Non-zero if YM2413 is used in the song
//  ym2612: Non-zero if YM2612 is used in the song
//
func WriteVGM(fname string, Song *song, psg bool, ym2151 bool, ym2413 bool, ym2612 bool)
    atom factor, totalWaits, loopWaits
    integer f, fhi, machineSpeed, songSize,
            cmd, vol, delay, nChannels,
            chnUpdateDelay, pcmDelay, pcmDelayReload, pcmDataPos,
            pcmDataLen, rhythm, compressVgm
    sequence freqTbl, oct1, fileEnding, s, 
             channelDone, channelLooped, iterations, gd3,
             cmdPos, loopPos, pcmData

    if ym2413 {
        /*if length(supportedChannels)-1 = 13 {
            if length(songs[song][5])  = 1 and
               length(songs[song][6])  = 1 and
               length(songs[song][7])  = 1 and
               length(songs[song][8])  = 1 and
               length(songs[song][9])  = 1 and
               length(songs[song][10]) = 1 and
               length(songs[song][11]) = 1 and
               length(songs[song][12]) = 1 and
               length(songs[song][13]) = 1 then
                // Disable the YM2413 if none of its channels are used
                ym2413 = false
            }
        } else {
            ym2413 = false
        }*/
        ym2413 = song.UsesChip(CHIP_YM2413)
    }

    if ym2612 {
        /*if length(supportedChannels)-1 = 10 {
            if length(songs[song][5]) = 1 and
               length(songs[song][6]) = 1 and
               length(songs[song][7]) = 1 and
               length(songs[song][8]) = 1 and
               length(songs[song][9]) = 1 and
               length(songs[song][10]) = 1 then
                // Disable the YM2612 if none of its channels are used
                ym2612 = false
            }
        } else {
            ym2612 = false
        }*/
        ym2413 = song.UsesChip(CHIP_YM2612)
    }
            
    nChannels = length(supportedChannels)-1
    
    compressVgm = false
    if (fname[length(fname)]='z' or fname[length(fname)]='Z') then
        compressVgm = true
    end if
    
    if compressVgm {
        outFile = gzopen(fname, "wb9 ")
        if outFile = Z_NULL then
            compiler.ERROR("Unable to open file: " + fname)
        end if
    } else {
        outFile = open(fname, "wb")
        if outFile = -1 then
            compiler.ERROR("Unable to open file: " + fname)
        end if
    }
    
    compiler.INFO("Generating VGM data")

    if song.Target.GetID() == TARGET_SMS && song.TuneSmsPitch {
        oct1 = OCTAVE1_ALT_SMS
    } else {
        oct1 = OCTAVE1
    }
    
    // Set up the frequency table for the PSG
    factor = 2.0
    sn76489FreqTable := make([]int, 12*6)
    for i = 2; i <= 7; i++ {
        for n = 1; n <= 12; n++ {
            sn76489FreqTable[(i - 2) * 12 + n] = int(math.Floor(song.target.GetMachineSpeed() / (oct1[n] * factor * 32)))
        }
        factor += factor
    }

    // Set up the frequency table for the YM2413
    ym2413FreqTable := []int{}
    if ym2413 {
        oct1 = OCTAVE1
        ym2413FreqTable = repeat(0, 12)
        for n = 1; n <= 12; n++ {
            ym2413FreqTable[n] = floor((oct1[n] * power(2, 18) / 50000))
        }
    }
    
    ym2612FreqTable := []int{}
    if ym2612 {
        ym2612FreqTable = {649, 688, 729, 772, 818, 867, 918, 973, 1031, 1092, 1157, 1226}
    }

    ym2151FreqTable := []int{}
    if ym2151 {
        ym2151FreqTable = {14,0,1,2,4,5,6,8,9,10,12,13}
    }




    vgmData = {0x56, 0x67, 0x6D, 0x20,  // "Vgm "
               0, 0, 0, 0,              // EOF offset, filled in later
               0x50, 0x01, 0, 0}        // VGM version (1.50)
    if psg {
        utils.AppendUint32(&vgmData, machineSpeed)
    } else {
        utils.AppendUint32(&vgmData, 0)
    }
    if ym2413 {
        utils.AppendUint32(&vgmData, machineSpeed)
    } else {
        utils.AppendUint32(&vgmData, 0)  
    }
    utils.AppendUint32(&vgmData, 0)  // GD3 tag offset
    utils.AppendUint32(&vgmData, 0)  // Total samples
    utils.AppendUint32(&vgmData, 0)  // Loop offset
    utils.AppendUint32(&vgmData, 0)  // Loop samples
    utils.AppendUint32(&vgmData, (floor(updateFreq))
    vgmData = append(vgmData, []byte{9, 0, 8, 0}...)     // Noise feedback pattern / shift width
    if ym2612 {
        vgmData &= int_to_bytes(7.6*1000000)
    } else {
        utils.AppendUint32(&vgmData, 0)  
    }
    if ym2151 {
        utils.AppendUint32(&vgmData, machineSpeed)
    } else {
        utils.AppendUint32(&vgmData, 0)  
    }
    vgmData &= int_to_bytes(0x0C)   // Data offset
    vgmData &= {0, 0, 0, 0, 0, 0, 0, 0}

    totalWaits  = 0
    loopWaits   = 0
    channelDone     = repeat(0, nChannels)
    channelLooped   = repeat(0, nChannels)
    iterations  = {0, 0}

    // Keeps track of the numbers of waits and current size of the VGM data
    // at the point of each command for each channel. Used for calculating
    // the offset and length of loops in the VGM.
    cmdPos = repeat(0, nChannels)
    for i = 1 to nChannels do
        cmdPos[i] = repeat(0, length(songs[song][i]))
    end for
    loopPos = {}

    // Used for avoiding unnecessary writes to the PSG port.
    // Contains the last value that have been output ({frequency, volume, panning, dac_on}) for
    // each channel.
    lastChannelSetting = repeat({-1, -1, #FF, 0}, nChannels)

    --  1 dataPtr,
    --  2 dataPos,
    --  3 delay,
    --  4 note,
    --  5 noteOffs,
    --  6 octave,
    --  7 duty,
    --  8 freq,
    --  9 freqOffs,
    -- 10 freqOffsLatch,
    -- 11 volume,
    -- 12 volOffs,
    -- 13 volOffsLatch,
    -- 14 vMac,
    -- 15 enMac,
    -- 16 en2Mac,
    -- 17 epMac,
    -- 18 mpMac,
    -- 19 loops,
    -- 20 loopIndex,
    -- 21 detune,
    -- 22 csMac,
    -- 23 mode,
    -- 24 feedback,
    -- 25 adsr,
    -- 26 operator,
    -- 27 mult,
    -- 28 ams,
    -- 29 fms,
    -- 30 lfo,
    -- 31 rateScale,
    -- 32 dtMac,
    -- 33 fbMac,
    -- 34 pattern,
    -- 35 chnNum,
    -- 36 oldPos,
    -- 37 prevNote,
    -- 38 delayLatch
    -- 39 transpose
    channel = repeat({song,
              1,
              #100,
              0,
              0,
              0,
              4,
              0,
              0,
              0,
              0,
              0,
              0,
                      {0, 0, {1, 2}},   -- vMac
                      {0, 0, {1, 2}},   -- enMac
                      {0, 0, {1, 2}},   -- en2Mac
                      {0, 0, {1, 2}},   -- epMac
                      {0, 0, 0},        -- mpMac
                      {0, 0},       -- loops
                      0,            -- loopIndex
                      0,            -- detune
                      {0, #FF, {1, 2}}, -- csMac
                      0,            -- mode
                      0,            -- feedback
                      {0, 0, 0, 0, 0},  -- adsr
                      0,            -- operator
                      0,            -- mult
                      0,            -- ams
                      0,            -- fms
                      {0, 0, 0},        -- lfo
                      0,            -- rateScale
                      {0, 0, {1, 2}},   -- dtMac
                      {0, 0, {1, 2}},   -- fbMac
                      0,            -- pattern
                      0,            -- chnNum
                      0,            -- oldPos
                      {0, 0},       -- prevNote
                      0,            -- delayLatch
                      0         -- transpose
                     }, nChannels)
    
    if ym2413 {
        vgmData = append(vgmData, {VGM_CMD_W_YM2413, 0x0F, 0x08}...)
        vgmData = append(vgmData, {VGM_CMD_W_YM2413, R_YM2413_MOD_TL, 0x00}...)
        vgmData = append(vgmData, {VGM_CMD_W_YM2413, R_YM2413_RHYTHM, YM2413_RHYTHM_ENABLE}...)
        rhythm = 0
    }
    
    if ym2612 {
        // Write PCM data bank if needed
        if length(pcms[1]) {
            pcmDataLen = 0
            for i = 1 to length(pcms[1]) do
                pcmDataLen += length(pcms[2][i][3])
            end for
            vgmData &= {0x67, 0x66, 0x00} & int_to_bytes(pcmDataLen)
            for i = 1 to length(pcms[1]) do
                vgmData &= pcms[2][i][3]
            end for
            if verbose then
                compiler.INFO("Total size of PCM data bank: %d bytes\n", pcmDataLen)
            end if
        }
    
        for i = 5 to nChannels do
            channel[i][CHN_VOLUME] = {0, 0, 0, 0}
            lastChannelSetting[i][2] = {-1, -1, -1, -1}
        end for
        
        // Turn on left and right output for all channels
        vgmData = append(vgmData, {VGM_CMD_W_YM2612L, R_YM2612_PH_AM_S,   0xC0}...)
        vgmData = append(vgmData, {VGM_CMD_W_YM2612L, R_YM2612_PH_AM_S+1, 0xC0}...)
        vgmData = append(vgmData, {VGM_CMD_W_YM2612L, R_YM2612_PH_AM_S+2, 0xC0}...)
        vgmData = append(vgmData, {VGM_CMD_W_YM2612H, R_YM2612_PH_AM_S,   0xC0}...)
        vgmData = append(vgmData, {VGM_CMD_W_YM2612H, R_YM2612_PH_AM_S+1, 0xC0}...)
        vgmData = append(vgmData, {VGM_CMD_W_YM2612H, R_YM2612_PH_AM_S+2, 0xC0}...)
        
        // Turn off DAC, LFO
        vgmData = append(vgmData, {VGM_CMD_W_YM2612L, R_YM2612_SSG_EG, 0x00}...)
        vgmData = append(vgmData, {VGM_CMD_W_YM2612L, R_YM2612_LFO,    0x00}...)
        vgmData = append(vgmData, {VGM_CMD_W_YM2612L, R_YM2612_CH3_6,  0x00}...)
        vgmData = append(vgmData, {VGM_CMD_W_YM2612L, R_YM2612_DAC_EN, 0x00}...)
    }
    
    if ym2151 {
        for i = 1 to nChannels do
            channel[i][CHN_VOLUME] = {0, 0, 0, 0}
            lastChannelSetting[i][2] = {-1, -1, -1, -1}
        end for

        // Turn on left and right output for all channels
        vgmData &= {VGM_CMD_W_YM2151, R_YM2413_FHI_CTL,   0xC0}
        vgmData &= {VGM_CMD_W_YM2151, R_YM2413_FHI_CTL+1, 0xC0}
        vgmData &= {VGM_CMD_W_YM2151, R_YM2413_FHI_CTL+2, 0xC0}
        vgmData &= {VGM_CMD_W_YM2151, R_YM2413_FHI_CTL+3, 0xC0}
        vgmData &= {VGM_CMD_W_YM2151, R_YM2413_FHI_CTL+4, 0xC0}
        vgmData &= {VGM_CMD_W_YM2151, R_YM2413_FHI_CTL+5, 0xC0}
        vgmData &= {VGM_CMD_W_YM2151, R_YM2413_FHI_CTL+6, 0xC0}
        vgmData &= {VGM_CMD_W_YM2151, R_YM2413_FHI_CTL+7, 0xC0}
        
        vgmData &= {VGM_CMD_W_YM2151, #18, #00}
        vgmData &= {VGM_CMD_W_YM2151, #1B, #C0}
    }
        
    pcmDelay = -1
    pcmDelayReload = -1
    

    while sum(channelDone) != nChannels do
        for i, c := range vgmChannels {
            c.Num = i
            if !c.Done {
                c.FreqChange = 0
                c.VolChange = 0
                c.Delay -= 0x100

                // Check if the whole part of the delay has reached 0
                if (c.Delay & 0xFFFF00) == 0 {
                    iterations[2] = 0
                    // Repeat until a note command has been read
                    for c.FreqChange != NEW_NOTE {
                        if c.Pattern >= 0 {
                            cmd = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                        } else {
                            cmd = song.GetChannelData(c.Num, c.DataPos)
                        }
                        if c.Pattern < 0 {
                            cmdPos[i][channel[i][CHN_DATAPOS]] = {length(vgmData), totalWaits}
                        }
                        cmdHi = cmd & 0xF0
                        
                        if cmdHi == defs.CMD_NOTE ||
                           cmdHi == defs.CMD_OCTUP ||
                           cmdHi == defs.CMD_OCTDN ||
                           cmdHi == defs.CMD_NOTE2 {
                            // Do the octave change if specified
                            if cmdHi == defs.CMD_OCTUP {
                                switch song.GetChannelType(c.Num) {
                                case specs.CHIP_SN76489:
                                    c.Octave += 12
                                case specs.CHIP_YM2413:
                                    if ym2413 {
                                        c.Octave++
                                    }
                                case specs.CHIP_YM2151, specs.CHIP_YM2612:
                                    c.Octave++
                                }
                                cmd (cmd & 0x0F) | defs.CMD_NOTE2
                            } else if cmdHi == defs.CMD_OCTDN {
                                switch song.GetChannelType(c.Num) {
                                case specs.CHIP_SN76489:
                                    c.Octave -= 12
                                case specs.CHIP_YM2413:
                                    if ym2413 {
                                        c.Octave--
                                    }
                                case specs.CHIP_YM2151, specs.CHIP_YM2612:
                                    c.Octave--
                                }
                                cmd (cmd & 0x0F) | defs.CMD_NOTE2
                            }

                            if cmd == defs.CMD_VOLUP {
                                c.DataPos++
                                if c.Pattern >= 0 {
                                    vol = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                                } else {
                                    vol = song.GetChannelData(c.Num, c.DataPos)
                                }
                                if vol & 0x80 {
                                    vol -= 0x100
                                }
                                if len(c.Volume.Op) > 0 {
                                    if c.Operator != 0 {
                                        c.Volume.Op[c.Operator - 1] += vol
                                    } else {
                                        for j := 0; j < len(c.Volume.Op); j++ {
                                            c.Volume.Op[j] += vol
                                        }
                                    }
                                } else {
                                    c.Volume.Vol += vol
                                }
                                c.VolMac.Disable()
                                c.VolChange = 1
                                
                            } else if cmd == defs.CMD_VOLUPC {
                            } else if cmd == defs.CMD_VOLDNC {
                            
                            } else
                                // If the previous note was a rest we need to trigger
                                // a volume change since the channel is currently muted.
                                if c.Note == defs.CMD_REST {
                                    c.VolChange = 1
                                }

                                // Note number is stored in low 4 bits of the command byte.
                                c.Note = cmd & 0x0F

                                if (cmd & 0xF0) == defs.CMD_NOTE2 {
                                    c.Delay += c.DelayLatch
                                else
                                    c.DataPos++

                                    // Delays are 16.8 unsigned fixed point. In the song data
                                    // they are stored either in two bytes (0-7FFF, for short delays)
                                    // or three bytes (0-3FFFFF, for long delays).
                                    if c.Pattern >= 0 {
                                        delay = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                                        if delay & 0x80 {
                                            c.DataPos++
                                            delay = (delay & 0x7F) * 0x80 + patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                                        }
                                        c.DataPos++
                                        delay = delay * 0x100 + patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                                    } else {
                                        delay = song.GetChannelData(c.Num, c.DataPos)
                                        if delay & 0x80 {
                                            c.DataPos++
                                            delay = (delay & 0x7F) * 0x80 + song.GetChannelData(c.Num, c.DataPos)
                                        }
                                        c.DataPos++
                                        delay = delay * 0x100 + song.GetChannelData(c.Num, c.DataPos)
                                    }

                                    c.Delay += delay
                                }

                                // The note can only be heard if the whole part of the delay
                                // is greater than zero.
                                if c.Delay > 0xFF {
                                    c.FreqChange = NEW_NOTE
                                }

                                if c.Note < defs.CMD_REST {
                                    c.VolMac.Step(&c,  EFFECT_STEP_EVERY_NOTE)
                                    c.ArpMac.Step(&c,  EFFECT_STEP_EVERY_NOTE)
                                    c.Arp2Mac.Step(&c, EFFECT_STEP_EVERY_NOTE)
                                    c.EpMac.Step(&c,   EFFECT_STEP_EVERY_NOTE)
                                    c.MpMac.Step(&c,   EFFECT_STEP_EVERY_NOTE)
                                    c.DutyMac.Step(&c, EFFECT_STEP_EVERY_NOTE)
                                    c.FbkMac.Step(&c,  EFFECT_STEP_EVERY_NOTE)
                                    c.PanMac.Step(&c,  EFFECT_STEP_EVERY_NOTE)


                                    // Reset vibrato
                                    if channel[i][CHN_MPMAC][1] then
                                        if and_bits(channel[i][CHN_MPMAC][1], #80) == EFFECT_STEP_EVERY_NOTE then
                                            if channel[i][CHN_MPMAC][3] = 0 then
                                                channel[i][9] = channel[i][10]
                                                channel[i][10] = -channel[i][10]
                                                channel[i][CHN_MPMAC][3] = vibratos[2][and_bits(channel[i][CHN_MPMAC][1], #7F)][2][2]
                                                freqChange = 1
                                            end if
                                            channel[i][CHN_MPMAC][3] -= 1   -- Decrease vibrato delay
                                        else
                                            channel[i][CHN_MPMAC][LIST_POS] = vibratos[2][and_bits(channel[i][CHN_MPMAC][1], #7F)][2][1]
                                            channel[i][9] = 0
                                            channel[i][10] = vibratos[2][and_bits(channel[i][CHN_MPMAC][1], #7F)][2][3]
                                        end if
                                    end if

                                }
                            }
                        } else if (cmd & 0xF0) == defs.CMD_OCTAVE {
                            cmd -= song.Target.GetMinOctave(c.Num)
                            switch song.GetChannelType(c.Num) {
                            case specs.CHIP_SN76489:
                                c.Octave = (cmd & 0x0F) * 12
                            case specs.CHIP_YM2413, specs.CHIP_YM2151, specs.CHIP_YM2612:
                                c.Octave = cmd & 0x0F
                            }

                        } else if (cmd & 0xF0) == defs.CMD_DUTY {
                            c.DutyMac.Disable()
                            if psg && (c.Num == 3) {
                                c.Duty = ((cmd ^ 1) &  1) * 4
                            
                            } else song.GetChannelType(c.Num) == specs.CHIP_YM2413 {
                                c.Duty = (cmd & 0x0F) * 16  
                            
                            } else if song.GetChannelType(c.Num) == specs.CHIP_YM2612 {
                                c.Duty = cmd & 7
                                vgmData &= {VGM_CMD_W_YM2612L + int(math.Floor((c.Num - 5) / 3)),
                                            R_YM2612_CONN_FB + ((c.Num - 5) % 3),
                                            c.Duty + c.Feedback}
                                            
                            } else if song.GetChannelType(c.Num) == specs.CHIP_YM2151 {
                                c.Duty = cmd & 7
                                vgmData &= {VGM_CMD_W_YM2151,
                                            R_YM2151_CONN_FB + c.Num - 1,
                                            and_bits(#C0, channel[i][22][2]) + c.Duty + c.Feedback}
                            }


                        } else if and_bits(cmd, #F0) = CMD_VOL2 then
                            if len(c.Volume.Op) {
                                if c.Operator != 0 {
                                    c.Volume.Op[c.Operator - 1] = cmd & 0x0F;
                                } else {
                                    for j := 0; j < len(c.Volume.Op); j++ {
                                        c.Volume.Op[j] = cmd & 0x0F
                                    }
                                }
                            } else {
                                c.Volume.Vol = cmd & 0x0F
                            }
                            c.VolMac.Disable()
                            volChange = 1

                        } else if cmd == defs.CMD_HWTE
                            if song.GetChannelType(c.Num) == specs.CHIP_YM2413 {
                                c.DataPos++
                                if c.Pattern >= 0 {
                                    c.Detune = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                                } else {
                                    c.Detune = song.GetChannelData(c.Num, c.DataPos)
                                }
                                if c.Operator == 0 || c.Operator == 1 {
                                    vgmData &= {VGM_CMD_W_YM2413, R_YM2413_MODEMUL,   c.Detune * 0x20 + c.Mult}
                                }
                                if c.Operator == 0 || c.Operator == 2 {
                                    vgmData &= {VGM_CMD_W_YM2413, R_YM2413_MODEMUL+1, c.Detune * 0x20 + c.Mult}
                                }
                            }

                        } else if cmd == defs.CMD_HWVE {
                            if song.GetChannelType(c.Num) == specs.CHIP_YM2413 {
                                c.DataPos++
                                if c.Pattern >= 0 {
                                    vgmData &= {VGM_CMD_W_YM2413, R_YM2413_MOD_TL, patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]}
                                } else {
                                    vgmData &= {VGM_CMD_W_YM2413, R_YM2413_MOD_TL, song.GetChannelData(c.Num, c.DataPos)}
                                }
                            }
                            
                        } else if cmd == defs.CMD_JSR {
                            c.DataPos++
                            c.Pattern = song.GetChannelData(c.Num, c.DataPos)
                            c.OldPos = c.DataPos
                            c.DataPos = 0
                        
                        } else if cmd == defs.CMD_RTS {
                            c.DataPos = c.OldPos
                            c.Pattern = -1
                            
                        } else if cmd == defs.CMD_ARPOFF then
                            c.ArpMac.Disable()
                            c.Arp2Mac.Disable()
                            c.NoteOffs = 0

                        } else if cmd == defs.CMD_CBONCE || cmd == defs.CMD_CBEVNT {
                            // Callbacks are ignored when outputting to VGM
                            c.DataPos++

                        } else if cmd == defs.CMD_MULT {
                            c.DataPos++
                            if c.Pattern >= 0 {
                                c.Mult = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                            } else {
                                c.Mult = song.GetChannelData(c.Num, c.DataPos)
                            }
                            switch song.GetChannelType(c.Num) {
                            case specs.CHIP_YM2413:
                                if c.Operator == 0 || c.Operator == 1 {
                                    vgmData &= {VGM_CMD_W_YM2413, R_YM2413_MODEMUL,   c.Detune * 0x20 + c.Mult}
                                }
                                if c.Operator == 0 || c.Operator == 2 {
                                    vgmData &= {VGM_CMD_W_YM2413, R_YM2413_MODEMUL+1, c.Detune * 0x20 + c.Mult}
                                }
                            
                            case specs.CHIP_YM2612:
                                if c.Operator != 0 {
                                    vgmData &= {VGM_CMD_W_YM2612L         + (song.target.ChipChannel(c.Num, specs.CHIP_YM2612) / 3),
                                                R_YM2612_DT_MUL + (c.Operator - 1)*4 + (song.target.ChipChannel(c.Num, specs.CHIP_YM2612) % 3),
                                                c.Detune * 0x10 + c.Mult}
                                else
                                    for j = 0; j < 3; j++ {
                                        vgmData &= {VGM_CMD_W_YM2612L + (song.target.ChipChannel(c.Num, specs.CHIP_YM2612) / 3),
                                                R_YM2612_DT_MUL + j*4 + (song.target.ChipChannel(c.Num, specs.CHIP_YM2612) % 3),
                                                c.Detune * 0x10 + c.Mult}
                                    }
                                }
                            case specs.CHIP_YM2151:
                                if c.Operator != 0 {
                                    vgmData &= {VGM_CMD_W_YM2151,
                                                R_YM2151_DT_MUL + (c.Operator - 1)*8 + song.target.ChipChannel(c.Num, specs.CHIP_YM2151),
                                                c.Detune * 0x10 + c.Mult}
                                else
                                    for op := 0; op < 4; op++ {
                                        vgmData &= {VGM_CMD_W_YM2151,
                                                R_YM2151_DT_MUL + op*8 + song.target.ChipChannel(c.Num, specs.CHIP_YM2151),
                                                c.Detune * 0x10 + c.Mult}
                                    }
                                }
                            }
                        
                        } else if (cmd & 0xF0) == defs.CMD_MODE {
                            c.Mode = cmd & 0x0F
                            switch song.GetChannelType(c.Num) {
                            case specs.CHIP_YM2413:
                                if song.target.ChipChannel(c.Num, specs.CHIP_YM2413) >= 6 {
                                    vgmChannels[song.target.FirstChipChannel(specs.CHIP_YM2413)+0].Mode = cmd & 0x0F
                                    vgmChannels[song.target.FirstChipChannel(specs.CHIP_YM2413)+1].Mode = cmd & 0x0F
                                    vgmChannels[song.target.FirstChipChannel(specs.CHIP_YM2413)+2].Mode = cmd & 0x0F
                                    rhythm = (cmd & 0x0F) * 0x20
                                    vgmData &= {VGM_CMD_W_YM2413, R_YM2413_FLO+6, 0x20, VGM_CMD_W_YM2413, R_YM2413_FHI_CTL+6, 0x05}
                                    vgmData &= {VGM_CMD_W_YM2413, R_YM2413_FLO+7, 0x57, VGM_CMD_W_YM2413, R_YM2413_FHI_CTL+7, 0x01}
                                    vgmData &= {VGM_CMD_W_YM2413, R_YM2413_FLO+8, 0x57, VGM_CMD_W_YM2413, R_YM2413_FHI_CTL+8, 0x01}
                                }
                            case specs.CHIP_YM2612:
                                if song.target.ChipChannel(c.Num, specs.CHIP_YM2612) == 5 {
                                    if c.Mode == 2 {
                                    }
                                } else {
                                    vgmData &= {VGM_CMD_W_YM2612L, R_YM2612_DAC_EN, 0x00}
                                }
                            case specs.CHIP_YM2151:
                                if song.target.ChipChannel(c.Num, specs.CHIP_YM2151) == 7 && c.Mode == 0 {
                                    vgmData &= {VGM_CMD_W_YM2151, R_YM2151_NOISE, 0x00}
                                }
                            }
            
                        } else if (cmd & 0xF0) == defs.CMD_OPER {
                            c.Operator = cmd & 0x0F

                        } else if (cmd & 0xF0) == defs.CMD_FEEDBK {
                            switch song.GetChannelType(c.Num) {
                            case specs.CHIP_YM2413:
                                c.DataPos++
                                if c.Pattern >= 0 {
                                    c.Feedback = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                                } else {
                                    c.Feedback = song.GetChannelData(c.Num, c.DataPos)
                                }
                                vgmData &= {VGM_CMD_W_YM2413, cmd & 0x0F, c.Feedback}   // ToDo: cmd & 0x0F. Correct?
                                
                            case specs.CHIP_YM2612:
                                c.Feedback = (cmd & 7) * 8
                                vgmData &= {VGM_CMD_W_YM2612L + (song.target.ChipChannel(c.Num, specs.CHIP_YM2612) / 3),
                                            R_YM2612_CONN_FB + (song.target.ChipChannel(c.Num, specs.CHIP_YM2612) % 3),
                                            c.Duty + c.Feedback}
                                            
                            case specs.CHIP_YM2151:
                                c.Feedback = (cmd & 7) * 8
                                vgmData &= {VGM_CMD_W_YM2151,
                                            R_YM2151_CONN_FB + song.target.ChipChannel(c.Num, specs.CHIP_YM2151),
                                            and_bits(#C0, channel[i][22][2]) + c.Duty + c.Feedback}
                            
                            }
                            
                        } else if cmd == defs.CMD_ADSR:
                            c.DataPos++
                            switch song.GetChannelType(c.Num) {
                            case specs.CHIP_YM2413:
                                if c.Pattern >= 0 {
                                    channel[i][25] = adsrs[2][patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]][2]
                                else
                                    c.ADSR = song.GetChannelData(c.Num, c.DataPos) // ToDo: use the channeldata as an index into ADSRs?
                                }
                                if c.Operator == 0 || c.Operator == 1 {
                                    vgmData &= {VGM_CMD_W_YM2413, R_YM2413_ATK_DEC+YM1413_MODULATOR, channel[i][25][1]}
                                    vgmData &= {VGM_CMD_W_YM2413, R_YM2413_SUS_REL+YM1413_MODULATOR, channel[i][25][2]}
                                }
                                if c.Operator == 0 c.Operator == 2 {
                                    vgmData &= {VGM_CMD_W_YM2413, R_YM2413_ATK_DEC+YM1413_CARRIER, channel[i][25][1]}
                                    vgmData &= {VGM_CMD_W_YM2413, R_YM2413_SUS_REL+YM1413_CARRIER, channel[i][25][2]}
                                }
                                
                            case specs.CHIP_YM2612:
                                f = and_bits(channel[i][25][2], #80)
                                if c.Pattern >= 0 {
                                    channel[i][25] = adsrs[2][patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]][2]
                                else
                                    channel[i][25] = adsrs[2][songs[channel[i][1]][i][channel[i][CHN_DATAPOS]]][2]
                                }
                                channel[i][25][2] = or_bits(channel[i][25][2], f)
                                cc := song.target.ChipChannel(c.Num, specs.CHIP_YM2612)
                                if c.Operator != 0 {
                                    vgmData &= {VGM_CMD_W_YM2612L + (cc / 3),
                                                R_YM2612_EG_ATK + (c.Operator - 1)*4 + (cc % 3),
                                                channel[i][25][1] + c.RateScale}
                                    
                                    vgmData &= {VGM_CMD_W_YM2612L + (cc / 3),
                                                R_YM2612_EG_DEC1 + (c.Operator - 1)*4 + (cc % 3),
                                                channel[i][25][2]}
                                    
                                    vgmData &= {VGM_CMD_W_YM2612L + (cc / 3),
                                                R_YM2612_EG_DEC2 + (c.Operator - 1)*4 + (cc % 3),
                                                channel[i][25][3]}
                                    
                                    vgmData &= {VGM_CMD_W_YM2612L + (cc / 3),
                                                R_YM2612_EG_SR + (c.Operator - 1)*4 + (cc % 3),
                                                channel[i][25][4]}
                                } else {
                                    for op := 0; op < 4; op++ {
                                        vgmData &= {VGM_CMD_W_YM2612L + (cc / 3),
                                                    R_YM2612_EG_ATK + op*4 + (cc % 3),
                                                    channel[i][25][1] + c.RateScale}
                                        
                                        vgmData &= {VGM_CMD_W_YM2612L + (cc / 3),
                                                    R_YM2612_EG_DEC1 + op*4 + (cc % 3),
                                                    channel[i][25][2]}
                                        
                                        vgmData &= {VGM_CMD_W_YM2612L + (cc / 3),
                                                    R_YM2612_EG_DEC2 + op*4 + (cc % 3),
                                                    channel[i][25][3]}
                                        
                                        vgmData &= {VGM_CMD_W_YM2612L + (cc / 3),
                                                    R_YM2612_EG_SR + op*4 + (cc % 3),
                                                    channel[i][25][4]}
                                    }
                                }
                                
                            case specs.CHIP_YM2151:
                                f = and_bits(channel[i][25][2], #80)
                                if channel[i][CHN_PATTERN] then
                                    channel[i][25] = adsrs[2][patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]][2]
                                else
                                    channel[i][25] = adsrs[2][songs[channel[i][1]][i][channel[i][CHN_DATAPOS]]][2]
                                }
                                channel[i][25][2] = or_bits(channel[i][25][2], f)
                                cc := song.target.ChipChannel(c.Num, specs.CHIP_YM2151)
                                if c.Operator != 0 {
                                    vgmData &= {VGM_CMD_W_YM2151,
                                                R_YM2151_EG_ATK + (c.Operator - 1)*8 + cc,
                                                channel[i][25][1] + c.RateScale}
                                    vgmData &= {VGM_CMD_W_YM2151,
                                                R_YM2151_EG_DEC1 + (c.Operator - 1)*8 + cc,
                                                channel[i][25][2]}
                                    vgmData &= {VGM_CMD_W_YM2151,
                                                R_YM2151_EG_DEC2 + (c.Operator - 1)*8 + cc,
                                                channel[i][25][3]}
                                    vgmData &= {VGM_CMD_W_YM2151,
                                                R_YM2151_EG_SR + (c.Operator - 1)*8 + cc,
                                                channel[i][25][4]}
                                } else {
                                    for j := 0; j < 4; j++ {
                                        vgmData &= {VGM_CMD_W_YM2151,
                                                R_YM2151_EG_ATK + j*8 + cc,
                                                channel[i][25][1] + c.RateScale}
                                        vgmData &= {VGM_CMD_W_YM2151,
                                                R_YM2151_EG_DEC1 + j*8 + cc,
                                                channel[i][25][2]}
                                        vgmData &= {VGM_CMD_W_YM2151,
                                                R_YM2151_EG_DEC2 + j*8 + cc,
                                                channel[i][25][3]}
                                        vgmData &= {VGM_CMD_W_YM2151,
                                                R_YM2151_EG_SR + j*8 + cc,
                                                channel[i][25][4]}
                                    }
                                }
                            }
                        
                        } else if cmd == defs.CMD_LEN:
                            c.DataPos++ 
                            if c.Pattern >= 0 {
                                delay = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                                if delay & 0x80 {
                                    c.DataPos++
                                    delay = (delay & 0x7F) * 0x80 + patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                                end if
                                c.DataPos++
                                delay = delay * #100 + patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                            } else {
                                delay = song.GetChannelData(c.Num, c.DataPos)
                                if delay & 0x80 {
                                    c.DataPos++
                                    delay = (delay & 0x7F) * 0x80 + song.GetChannelData(c.Num, c.DataPos)
                                }
                                c.DataPos++
                                delay = delay * 0x100 + song.GetChannelData(c.Num, c.DataPos)
                            }
                            c.DelayLatch = delay

                            
                        } else if cmd == defs.CMD_RSCALE {
                            c.DataPos++
                            if channel[i][CHN_PATTERN] then
                                c.RateScale = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]] * 0x40
                            } else {
                                c.RateScale = song.GetChannelData(c.Num, c.DataPos) * 0x40
                            }
                            switch song.GetChannelType(c.Num) {
                            case specs.CHIP_YM2612:
                                cc := song.target.ChipChannel(c.Num, specs.CHIP_YM2612)
                                if c.Operator != 0 {
                                    vgmData &= {VGM_CMD_W_YM2612L + (cc / 3),
                                                R_YM2612_EG_ATK + (channel[i][CHN_OPER] - 1)*4 + (cc % 3),
                                                channel[i][25][1] + c.RateScale}
                                else
                                    for j := 0; j < 4; j++ {
                                        vgmData &= {VGM_CMD_W_YM2612L + (cc / 3),
                                                    R_YM2612_EG_ATK + j*4 + (cc % 3),
                                                    channel[i][25][1] + c.RateScale}
                                    }
                                }
                            case specs.CHIP_YM2151:
                                cc := song.target.ChipChannel(c.Num, specs.CHIP_YM2151)
                                if c.Operator != 0 {
                                    vgmData &= {VGM_CMD_W_YM2151,
                                                R_YM2151_EG_ATK + (c.Operator - 1)*8 + cc,
                                                channel[i][25][1] + c.RateScale}
                                } else {
                                    for j := 0; j < 4; j++ {
                                        vgmData &= {VGM_CMD_W_YM2151,
                                                    R_YM2151_EG_ATK + j*8 + cc,
                                                    channel[i][25][1] + c.RateScale}
                                    }
                                }
                            }
                            
                        } else if cmd == defs.CMD_VOLSET {
                            c.DataPos++
                            if c.Pattern >= 0 {
                                if sequence(channel[i][CHN_VOLUME]) then
                                    if channel[i][CHN_OPER] then
                                        channel[i][CHN_VOLUME][channel[i][CHN_OPER]] = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                                    else
                                        channel[i][CHN_VOLUME] = repeat(patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]], length(channel[i][CHN_VOLUME]))
                                    end if
                                else
                                    c.Volume.Vol = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                                end if
                            } else {
                                if sequence(channel[i][CHN_VOLUME]) then
                                    if channel[i][CHN_OPER] then
                                        channel[i][CHN_VOLUME][channel[i][CHN_OPER]] = songs[channel[i][1]][i][channel[i][CHN_DATAPOS]]
                                    else
                                        channel[i][CHN_VOLUME] = repeat(songs[channel[i][1]][i][channel[i][CHN_DATAPOS]], length(channel[i][CHN_VOLUME]))
                                    end if
                                } else {
                                    c.Volume.Vol = songs[channel[i][1]][i][channel[i][CHN_DATAPOS]]
                                }
                            }
                            // An explicit volume set command overrides any ongoing volume effect macro
                            c.VolMac.Disable()
                            c.VolChange = 1

                        } else if cmd == defs.CMD_VOLMAC {
                            c.DataPos++
                            if c.Pattern >= 0 {
                                channel[i][CHN_VOLMAC][1] = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                            } else {
                                c.VolMac.SetID(song.GetChannelData(c.Num, c.DataPos))
                            }
                            c.ArpMac.MoveToStart()
                            if len(c.Volume.Op) {
                                if c.Operator != 0 {
                                    c.Volume.Op[c.Operator - 1] = c.VolMac.Params.Peek()
                                } else {
                                    channel[i][CHN_VOLUME] = repeat(channel[i][CHN_VOLMAC][3][3], length(channel[i][CHN_VOLUME]))
                                }
                            } else {
                                c.Volume.Vol = c.VolMac.Params.Peek()
                            }
                            //channel[i][CHN_VOLUME] = channel[i][CHN_VOLMAC][3][3]
                            c.VolChange = 1

                        } else if cmd == defs.CMD_PANMAC {
                            c.DataPos++
                            if channel[i][CHN_PATTERN] then
                                channel[i][CHN_PANMAC][1] = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                            else
                                channel[i][CHN_PANMAC][1] = songs[channel[i][1]][i][channel[i][CHN_DATAPOS]]
                            end if
                            channel[i][CHN_PANMAC][3] = step_list(panMacros[2][and_bits(channel[i][CHN_PANMAC][1], #7F)], {1, 2})
                            if channel[i][CHN_PANMAC][3][3] then
                                if and_bits(channel[i][CHN_PANMAC][3][3], #80) then
                                    if psg and i < 5 then
                                        channel[1][CHN_PANMAC][2] = clear_bit(channel[1][CHN_PANMAC][2], i - 1)
                                        channel[1][CHN_PANMAC][2] =   set_bit(channel[1][CHN_PANMAC][2], i + 3)
                                    elsif channelType[i] = TYPE_YM2612 then
                                        channel[i][CHN_PANMAC][2] = clear_bit(channel[i][CHN_PANMAC][2], 6)
                                        channel[i][CHN_PANMAC][2] =   set_bit(channel[i][CHN_PANMAC][2], 7)
                                    elsif channelType[i] = TYPE_YM2151 then
                                        channel[i][CHN_PANMAC][2] = clear_bit(channel[i][CHN_PANMAC][2], 7)
                                        channel[i][CHN_PANMAC][2] =   set_bit(channel[i][CHN_PANMAC][2], 6)
                                    end if
                                else
                                    if psg and i < 5 then
                                        channel[1][CHN_PANMAC][2] = clear_bit(channel[1][CHN_PANMAC][2], i + 3)
                                        channel[1][CHN_PANMAC][2] =   set_bit(channel[1][CHN_PANMAC][2], i - 1)
                                    elsif channelType[i] = TYPE_YM2612 then
                                        channel[i][CHN_PANMAC][2] = clear_bit(channel[i][CHN_PANMAC][2], 7)
                                        channel[i][CHN_PANMAC][2] =   set_bit(channel[i][CHN_PANMAC][2], 6)
                                    elsif channelType[i] = TYPE_YM2151 then
                                        channel[i][CHN_PANMAC][2] = clear_bit(channel[i][CHN_PANMAC][2], 6)
                                        channel[i][CHN_PANMAC][2] =   set_bit(channel[i][CHN_PANMAC][2], 7)
                                    end if
                                end if
                            } else {
                                switch song.GetChannelType(c.Num) {
                                case specs.CHIP_SN76489:
                                    channel[1][CHN_PANMAC][2] = set_bit(channel[1][CHN_PANMAC][2], i - 1)
                                    channel[1][CHN_PANMAC][2] = set_bit(channel[1][CHN_PANMAC][2], i + 3)
                                case specs.CHIP_YM2612, specs.CHIP_YM2151:
                                    channel[i][CHN_PANMAC][2] = set_bit(channel[i][CHN_PANMAC][2], 6)
                                    channel[i][CHN_PANMAC][2] = set_bit(channel[i][CHN_PANMAC][2], 7)
                                }
                            }
                            switch song.GetChannelType(c.Num) {
                            case specs.CHIP_SN76489:
                                if channel[1][CHN_PANMAC][2] != lastChannelSetting[1][3] and not supportsPAL then
                                    vgmData &= {0x4F, channel[1][CHN_PANMAC][2]}
                                    lastChannelSetting[1][3] = channel[1][CHN_PANMAC][2]
                                }
                            case specs.CHIP_YM2612:
                                if sequence(channel[i][30][3]) then
                                    vgmData &= {VGM_CMD_W_YM2612L + floor((i - 5) / 3),
                                                R_YM2612_PH_AM_S + remainder((i - 5), 3),
                                                and_bits(#C0, channel[i][22][2]) + channel[i][30][3][2]}
                                else
                                    vgmData &= {VGM_CMD_W_YM2612L + floor((i - 5) / 3),
                                                R_YM2612_PH_AM_S + remainder((i - 5), 3),
                                                and_bits(#C0, channel[i][22][2])}
                                end if
                            case specs.CHIP_YM2151:
                                vgmData &= {VGM_CMD_W_YM2151,
                                            R_YM2151_CONN_FB + i - 1,
                                            and_bits(#C0, channel[i][22][2]) + c.Duty + c.Feedback}
                            }

                        } else if cmd == defs.CMD_ARPMAC {
                            c.DataPos++
                            if c.Pattern >= 0 {
                                channel[i][15][1] = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                            } else {
                                c.ArpMac.SetID(song.GetChannelData(c.Num, c.DataPos))
                            }
                            c.ArpMac.MoveToStart()
                            c.Duty = c.ArpMac.Peek()
                            c.Arp2Mac.Disable()
                            c.FreqChange = NEW_EFFECT_VALUE

                        } else if cmd == defs.CMD_DUTMAC {
                            c.DataPos++
                            if c.Pattern >= 0 {
                                channel[i][32][1] = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                            } else {
                                c.DutyMac.SetID(song.GetChannelData(c.Num, c.DataPos))
                            }
                            c.DutyMac.MoveToStart()

                        } else if cmd == defs.CMD_FBKMAC {
                            c.DataPos++
                            if c.Pattern >= 0 {
                                channel[i][33][1] = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                            } else {
                                c.FbkMac.SetID(song.GetChannelData(c.Num, c.DataPos))
                            }
                            c.FbkMac.MoveToStart()

                        } else if cmd == defs.CMD_TRANSP {
                            c.DataPos++
                            if c.Pattern >= 0 {
                                c.Transpose = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                            } else {
                                c.Transpose = song.GetChannelData(c.Num, c.DataPos)
                            }
                        
                        } else if cmd == defs.CMD_DETUNE {
                            c.DataPos++
                            if c.Pattern >= 0 {
                                c.Detune = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                            } else {
                                c.Detune = song.GetChannelData(c.Num, c.DataPos)
                            }
                            switch song.GetChannelType(c.Num) {
                            case specs.CHIP_YM2612:
                                if c.Operator != 0 {
                                    vgmData &= {VGM_CMD_W_YM2612L + floor((i - 5) / 3),
                                                0x30 + (channel[i][CHN_OPER] - 1)*4 + remainder((i - 5), 3),
                                                c.Detune * 0x10 + c.Mult}
                                } else {
                                    for j = 0; j < 4; j++ {
                                        vgmData &= {VGM_CMD_W_YM2612L + floor((i - 5) / 3),
                                                    0x30 + j*4 + remainder((i - 5), 3),
                                                    c.Detune * 0x10 + c.Mult}
                                    }
                                }
                            case specs.CHIP_YM2151:
                                if c.Operator != 0 {
                                    vgmData &= {VGM_CMD_W_YM2151,
                                                R_YM2151_DT_MUL + (channel[i][CHN_OPER] - 1)*8 + i - 1,
                                                c.Detune * 0x10 + c.Mult}
                                } else {
                                    for j = 0; j < 4; j++ {
                                        vgmData &= {VGM_CMD_W_YM2151,
                                                    R_YM2151_DT_MUL + j*8 + i - 1,
                                                    c.Detune * 0x10 + c.Mult}
                                    }
                                }
                            }

                        } else if cmd == defs.CMD_MODMAC {
                            c.DataPos++
                            if c.Pattern >= 0 {
                                channel[i][30][1] = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                            } else {
                                channel[i][30][1] = songs[channel[i][1]][i][channel[i][CHN_DATAPOS]]
                            }
                            switch song.GetChannelType(c.Num) {
                            case specs.CHIP_YM2612:
                                if channel[i][30][1] then
                                    channel[i][30][3] = mods[2][channel[i][30][1]][2]
                                    vgmData &= {VGM_CMD_W_YM2612L, R_YM2612_LFO, channel[i][30][3][1] + 8}
                                    vgmData &= {VGM_CMD_W_YM2612L + floor((i - 5) / 3),
                                                R_YM2612_PH_AM_S + remainder((i - 5), 3),
                                                and_bits(#C0, channel[i][22][2]) + channel[i][30][3][2]}
                                } else {
                                    vgmData &= {VGM_CMD_W_YM2612L, R_YM2612_LFO, 0}
                                }
                            case specs.CHIP_YM2151:
                                if channel[i][30][1] then
                                    channel[i][30][3] = mods[2][channel[i][30][1]][2]
                                    vgmData &= {VGM_CMD_W_YM2151, R_YM2151_LFO_F, channel[i][30][3][1]}
                                    vgmData &= {VGM_CMD_W_YM2151, 0x19, channel[i][30][3][2]}
                                    vgmData &= {VGM_CMD_W_YM2151, 0x19, channel[i][30][3][3]}
                                    vgmData &= {VGM_CMD_W_YM2151, R_YM2151_CT_LFOW, channel[i][30][3][5]}
                                    vgmData &= {VGM_CMD_W_YM2151, R_YM2151_PH_AM_S + i - 1, channel[i][30][3][4]}
                                } else {
                                    vgmData &= {VGM_CMD_W_YM2151, R_YM2151_LFO_F, 0}
                                }
                            }

                        } else if cmd == defs.CMD_SSG {
                            c.DataPos++
                            if song.GetChannelType(c.Num) == specs.CHIP_YM2612 {
                                if c.Pattern >= 0 {
                                    // TODO: Handle this case
                                } else {
                                    if channel[i][CHN_OPER] then
                                        if songs[channel[i][1]][i][channel[i][CHN_DATAPOS]] then
                                            vgmData &= {VGM_CMD_W_YM2612L + floor((i - 5) / 3),
                                                        R_YM2612_SSG_EG + (channel[i][CHN_OPER] - 1)*4 + remainder((i - 5), 3),
                                                        0x8 + songs[channel[i][1]][i][channel[i][CHN_DATAPOS]] - 1}
                                        } else {
                                            vgmData &= {VGM_CMD_W_YM2612L + floor((i - 5) / 3),
                                                        R_YM2612_SSG_EG + (channel[i][CHN_OPER] - 1)*4 + remainder((i - 5), 3),
                                                        0}
                                        }
                                    } else {
                                        for j := 0; j < 4; j++ {
                                            if songs[channel[i][1]][i][channel[i][CHN_DATAPOS]] then
                                                vgmData &= {VGM_CMD_W_YM2612L + floor((i - 5) / 3),
                                                            R_YM2612_SSG_EG + j*4 + remainder((i - 5), 3),
                                                            0x8 + songs[channel[i][1]][i][channel[i][CHN_DATAPOS]] - 1}
                                            } else {
                                                vgmData &= {VGM_CMD_W_YM2612L + floor((i - 5) / 3),
                                                            R_YM2612_SSG_EG + j*4 + remainder((i - 5), 3),
                                                            0}
                                            }
                                        }
                                    }
                                }
                            }
                            
                        } else if cmd == defs.CMD_HWAM {
                            c.DataPos++
                            channel[i][25][2] = and_bits(channel[i][25][2], #7F)
                            if c.Pattern >= 0 {
                                channel[i][25][2] = or_bits(channel[i][25][2], patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]] * #80)
                            } else {
                                channel[i][25][2] = or_bits(channel[i][25][2], songs[channel[i][1]][i][channel[i][CHN_DATAPOS]] * #80)
                            }
                            
                            switch song.GetChannelType(c.Num) {
                            case specs.CHIP_YM2612:
                                if c.Operator != 0 {
                                    vgmData &= {VGM_CMD_W_YM2612L + floor((i - 5) / 3),
                                                R_YM2612_EG_DEC1 + (channel[i][CHN_OPER] - 1)*4 + remainder((i - 5), 3),
                                                channel[i][25][2]}
                                } else {
                                    for j := 0; j < 4; j++ {
                                        vgmData &= {VGM_CMD_W_YM2612L + floor((i - 5) / 3),
                                                    R_YM2612_EG_DEC1 + j*4 + remainder((i - 5), 3),
                                                    channel[i][25][2]}
                                    }
                                }
                            case specs.CHIP_YM2151:
                                if c.Operator != 0 {
                                    vgmData &= {VGM_CMD_W_YM2151,
                                                R_YM2151_EG_DEC1 + (channel[i][CHN_OPER] - 1)*8 + i - 1,
                                                channel[i][25][2]}
                                } else {
                                    for j := 0; j < 4; j++ {
                                        vgmData &= {VGM_CMD_W_YM2151,
                                                    R_YM2151_EG_DEC1 + j*8 + i - 1,
                                                    channel[i][25][2]}
                                    }
                                }
                            }
                            
                        } else if cmd == defs.CMD_APMAC2 {
                            c.DataPos++
                            if c.Pattern >= 0 {
                                channel[i][16][1] = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                            } else {
                                channel[i][16][1] = songs[channel[i][1]][i][channel[i][CHN_DATAPOS]]
                            }
                            channel[i][CHN_EN2MAC][3] = step_list(arpeggios[2][and_bits(channel[i][CHN_EN2MAC][1], #7F)], {1, 2})
                            c.ArpMac.Disable()
                            c.FreqChange = NEW_EFFECT_VALUE

                        } else if cmd == defs.CMD_SWPMAC {
                            c.DataPos++
                            if c.Pattern >= 0 {
                                channel[i][CHN_EPMAC][1] = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                            else
                                channel[i][CHN_EPMAC][1] = songs[channel[i][1]][i][channel[i][CHN_DATAPOS]]
                            }
                            if channel[i][17][1] then
                                channel[i][17][3] = step_list(pitchMacros[2][and_bits(channel[i][CHN_EPMAC][1], #7F)], {1, 2})
                                c.FreqOffs = 0
                                c.FreqChange = NEW_EFFECT_VALUE
                            } else {
                                c.FreqOffs = 0
                                c.FreqChange = NEW_EFFECT_VALUE
                            }

                        } else if cmd == defs.CMD_VIBMAC {
                            c.DataPos++
                            if c.Pattern >= 0 {
                                channel[i][CHN_MPMAC][1] = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                            } else {
                                channel[i][CHN_MPMAC][1] = songs[channel[i][1]][i][channel[i][CHN_DATAPOS]]
                            }
                            if channel[i][CHN_MPMAC][1] then
                                channel[i][CHN_MPMAC][3] = vibratos[2][and_bits(channel[i][CHN_MPMAC][1], #7F)][2][1]
                                c.FreqOffs = 0
                                channel[i][10] = vibratos[2][and_bits(channel[i][CHN_MPMAC][1], #7F)][2][3]
                                c.FreqChange = NEW_EFFECT_VALUE
                            } else {
                                c.FreqChange = NEW_EFFECT_VALUE
                                c.FreqOffs = 0
                            }

                        } else if cmd == defs.CMD_JMP {
                            if c.Pattern >= 0 {
                                channel[i][CHN_DATAPOS] = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS] + 1] +
                                                          patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS] + 2] * #100
                            } else {
                                channel[i][CHN_DATAPOS] = songs[channel[i][1]][i][channel[i][CHN_DATAPOS] + 1] +
                                                      songs[channel[i][1]][i][channel[i][CHN_DATAPOS] + 2] * #100
                            }                     
                            c.HasLooped = true
                            if sum(channelLooped) = nChannels then
                                loopPos = cmdPos[i][channel[i][CHN_DATAPOS] + 1]
                                channelDone = repeat(1, nChannels)
                                c.FreqChange = NEW_NOTE
                            }

                        } else if cmd == defs.CMD_J1 {
                            if channel[i][19][channel[i][20]] = 1 then
                                channel[i][20] -= 1
                                if c.Pattern >= 0 {
                                    channel[i][CHN_DATAPOS] = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS] + 1] +
                                                              patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS] + 2] * #100
                                } else {
                                    channel[i][CHN_DATAPOS] = songs[channel[i][1]][i][channel[i][CHN_DATAPOS] + 1] +
                                                      songs[channel[i][1]][i][channel[i][CHN_DATAPOS] + 2] * #100
                                }
                            } else {
                                c.DataPos += 2
                            }
                        
                        } else if cmd == defs.CMD_DJNZ {
                            channel[i][19][channel[i][20]] -= 1
                            if channel[i][19][channel[i][20]] != 0 then
                                if c.Pattern >= 0 {
                                    channel[i][CHN_DATAPOS] = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS] + 1] +
                                                              patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS] + 2] * #100
                                } else {
                                    channel[i][CHN_DATAPOS] = songs[channel[i][1]][i][channel[i][CHN_DATAPOS] + 1] +
                                                      songs[channel[i][1]][i][channel[i][CHN_DATAPOS] + 2] * #100
                                }
                            } else {
                                channel[i][20] -= 1
                                c.DataPos += 2
                            }

                        } else if cmd == defs.CMD_LOPCNT {
                            c.DataPos++
                            channel[i][20] += 1
                            if c.Pattern >= 0 {
                                channel[i][19][channel[i][20]] = patterns[2][channel[i][CHN_PATTERN]][channel[i][CHN_DATAPOS]]
                            } else {
                                channel[i][19][channel[i][20]] = songs[channel[i][1]][i][channel[i][CHN_DATAPOS]]
                            }
                        
                        } else if cmd == defs.CMD_WRMEM {
                            // Ignored
                            c.DataPos += 3
                        } else if cmd == defs.CMD_WRPORT {
                            // Ignored
                            c.DataPos += 3
                            
                        } else if cmd == defs.CMD_END {
                            c.Note = cmd
                            c.Done = true
                            c.FreqChange = NEW_NOTE

                        }

                        c.DataPos++
                        iterations[2] += 1
                        if iterations[2] = 131072 then
                            puts(1, "Internal error: Appears to be stuck in an infinite loop. Exiting\n")
                            //? {i, channel[1][2], channel[2][2], channel[3][2], channel[4][2]}
                            break
                        }
                    } // end for freqChange != 2
                } else {
                    // Update effects as needed
                    c.VolMac.Step( &c, VGM_STEP_FRAME)
                    c.ArpMac.Step( &c, VGM_STEP_FRAME)
                    c.Arp2Mac.Step(&c, VGM_STEP_FRAME)
                    c.EpMac.Step(  &c, VGM_STEP_FRAME)
                    c.MpMac.Step(  &c, VGM_STEP_FRAME)
                    c.DutyMac.Step(&c, VGM_STEP_FRAME)
                    c.FbkMac.Step( &c, VGM_STEP_FRAME)
                    c.PanMac.Step( &c, VGM_STEP_FRAME)
                                    
                    if channel[i][CHN_MPMAC][1] then
                        if and_bits(channel[i][CHN_MPMAC][1], #80) = VGM_STEP_FRAME then
                            if channel[i][CHN_MPMAC][3] = 0 then
                                channel[i][9] = channel[i][10]
                                channel[i][10] = -channel[i][10]
                                channel[i][CHN_MPMAC][3] = vibratos[2][and_bits(channel[i][CHN_MPMAC][1], #7F)][2][2]
                                freqChange = 1
                            end if
                            channel[i][18][3] -= 1
                        end if
                    end if
                }
                
                if c.Note == defs.CMD_END || c.Note == defs.CMD_REST {
                    // Mute the channel if necessary
                    if (len(currSettings.Volume.Op) != 0 && sum(lastChannelSetting[i][2]) != 0) ||
                       (len(currSettings.Volume.Op) == 0 && currSettings.Volume.Vol != 0)) ||
                       (ym2612 and i = 10 and c.Mode == 2) {
                        
                        switch song.GetChannelType(c.Num) {
                        case specs.CHIP_SN76489:
                            // Set max attenuation
                            vgmData &= {VGM_CMD_W_PSG, SN76489_VOL_LATCH | 0x0F | (c.Num * 0x20)}
                            vgmData &= {VGM_CMD_W_PSG, SN76489_VOL_DATA  | 0x0F | (c.Num * 0x20)}
                            currSettings.Freq = -1
                            currSettings.Volume.Vol = 0
                        case specs.CHIP_YM2413:
                            if c.Mode == 0 {
                                //vgmData &= {#51, #30 + (i - 5), channel[i][7] + #0F}
                                vgmData &= {VGM_CMD_W_YM2413, R_YM2413_FHI_CTL + (i - 5), c.Freq / 100)}
                                //lastChannelSetting[i] = {-1, 0, lastChannelSetting[i][3]}
                            } else if i = 11 && c.Mode == 1 {
                                rhythm = rhythm & 20 
                                vgmData &= {VGM_CMD_W_YM2413, 0x0E, rhythm}
                            } else if i = 12 && c.Mode == 1 {
                                //rhythm = and_bits(rhythm, #36)
                                //vgmData &= {#51, #0E, rhythm}
                            } else if i = 13 && c.Mode == 1 {
                                //rhythm = and_bits(rhythm, #39)
                                //vgmData &= {#51, #0E, rhythm}
                            }
                        case specs.CHIP_YM2612:
                            if c.Mode == 0 {
                                vgmData &= {VGM_CMD_W_YM2612L,
                                            R_YM2612_KEYON,
                                            i + floor((i - 5) / 3)- 5} 
                                //lastChannelSetting[i] = {-1, lastChannelSetting[i][2], lastChannelSetting[i][3]}
                            } else if c.Mode == 2 and i = 10 then
                                if currSettings.DacOn {
                                    // Turn off DAC
                                    vgmData &= {VGM_CMD_W_YM2612L, R_YM2612_DAC_EN, 0x00}
                                    currSettings.DacOn = false
                                }
                            }
                        case specs.CHIP_YM2151:
                            vgmData &= {VGM_CMD_W_YM2151, R_YM2151_KEYON, i - 1}
                        }
                    }
                    if c.Note == defs.CMD_END {
                        c.Done = true
                        c.HasLooped = true
                    }
                } else {    
                    if c.FreqChange != 0 && !c.Done {
                        switch song.GetChannelType(c.Num) {
                        case specs.CHIP_SN76489:
                            if psg && c.Num < 3 {
                                if c.Octave + c.Note + c.NoteOffs + c.Transpose < 0 {
                                    c.Freq = sn76489FreqTable[0]
                                } else if c.Octave + c.Note + c.NoteOffs + c.Transpose > 71 {
                                    c.Freq = sn76489FreqTable[71] 
                                } else {
                                    c.Freq = sn76489FreqTable[c.Octave + c.Note + c.NoteOffs + c.Transpose]
                                }
                                c.Freq -= c.FreqOffs + c.Detune
    
                                // Clamp the period to the highest/lowest supported values
                                if c.Freq >= 0x03EF {
                                    c.Freq = 0x03EF
                                } else if c.Freq < 0x001C {
                                    c.Freq = 0x001C
                                }
                                if c.Freq != currSettings.Freq then
                                    vgmData &= {VGM_CMD_W_PSG, SN76489_TONE_LATCH | (c.Num * 0x20) | (c.Freq & 0x0F)}
                                    vgmData &= {VGM_CMD_W_PSG, SN76489_TONE_DATA  | ((c.Freq & 0x3F0) / 0x10)}
                                    currSettings.Freq = c.Freq
                                }
                            } else if psg && c.Num == 3 {
                                // The noise channel is a special case
                                vgmData &= {VGM_CMD_W_PSG, 0xE0 | c.Duty | ((c.Note + c.NoteOffs + c.Transpose) & 3)}
                                vgmData &= {VGM_CMD_W_PSG, 0x60 | c.Duty | ((c.Note + c.NoteOffs + c.Transpose) & 3)}
                            }
                        case specs.CHIP_YM2413:
                            if c.FreqChange == NEW_NOTE {
                                // KEY_OFF
                                vgmData &= {VGM_CMD_W_YM2413, R_YM2413_FHI_CTL + (i - 5), (c.Freq / 0x100)}
                            }
                            if i = 11 and c.Mode == 1 {
                                vgmData &= {VGM_CMD_W_YM2413, R_YM2413_RHYTHM, YM2413_RHYTHM_ENABLE} 
                                //? {#51, #0E, and_bits(rhythm,#2F)}
                                rhythm = or_bits(rhythm, #10)
                                vgmData &= {VGM_CMD_W_YM2413, R_YM2413_RHYTHM, YM1413_RHYTHM_ENABLE | 0x1F} 
                                //? {#51, #0E, rhythm}
                                //vgmData &= {#51, #16, #20, #51, #26, #05}
                                //? {1 , 2, 3}
                            } else if i = 12 and channel[i][CHN_MODE] = 1 then
                                //vgmData &= {#51, #0E, and_bits(rhythm,#36)}
                                //rhythm = or_bits(rhythm, #08)
                                //vgmData &= {#51, #0E, rhythm}
                                //vgmData &= {#51, #17, #50, #51, #27, #05}
                            } else if i = 13 and channel[i][CHN_MODE] = 1 then
                                //vgmData &= {#51, #0E, and_bits(rhythm,#39)}
                                //rhythm = or_bits(rhythm, #02)
                                //vgmData &= {#51, #0E, rhythm}
                                //vgmData &= {#51, #18, #C0, #51, #28, #01}
                            } else {
                                f = (c.Octave + 1) * 12 + c.Note + c.NoteOffs + c.Transpose
                                if f < 0 {
                                    c.Freq = ym2413FreqTable[0]
                                    fhi = 0
                                } else if f > 95 {
                                    c.Freq = ym2413FreqTable[11]
                                    fhi = 7 * 0x200
                                } else {
                                    c.Freq = ym2413FreqTable[f % 12]
                                    fhi = int(f / 12.0) * 0x200
                                }
                                c.Freq += c.FreqOffs + c.Detune

                                if c.Freq >= 511 {
                                    c.Freq = 511
                                } else if c.Freq < 0 {
                                    c.Freq = 0
                                }
                                c.Freq += fhi

                                if c.Freq != currSettings.Freq {
                                    vgmData &= {VGM_CMD_W_YM2413, R_YM1413_FLO + (i - 5), (c.Freq & 0xFF)}
                                    currSettings.Freq = c.Freq
                                }
                                vgmData &= {VGM_CMD_W_YM2413, R_YM2413_FHI_CTL + (i - 5), (c.Freq / 0x100) + 0x30}
                                //printf(1, "Ch%d: Writing %02x to %02x, ", {i-5, and_bits(channel[i][8], #FF), #10 + (i - 5)})
                                //printf(1, "%02x to %02x\n", {floor(channel[i][8] / #100) + #30, #20 + (i - 5)})
                            }                           
                        case specs.CHIP_YM2612:
                            if i == 10 and c.Mode == 2 {
                                // PCM mode
                                pcmDelay = 0
                                pcmDelayReload = 5
                                f = c.Octave*12 + c.Note + c.NoteOffs + c.Transpose
                                if assoc_find_key(pcms, f) then
                                    pcmDataLen = length(pcms[2][assoc_find_key(pcms, f)][3])
                                    pcmDataPos = 0
                                    for j = 1 to length(pcms[1]) do
                                        if pcms[1][j] < f then
                                            pcmDataPos += length(pcms[2][j][3])
                                        end if
                                    end for
                                    //printf(1,"Found matching PCM data for note %d. Offset = %d, length = %d\n",{f, pcmDataPos, pcmDataLen})
                                    vgmData &= #E0 & int_to_bytes(pcmDataPos)
                                    if !currSettings.DacOn then
                                        vgmData &= {VGM_CMD_W_YM2612L, R_YM2612_DAC_EN, 0x80}
                                        currSettings.DacOn = true
                                    }
                                    pcmDataPos = 1
                                } else {
                                    compiler.WARNING(sprintf("No matching PCM sample found for note %d", f), -1)
                                }
                            } else {    
                                // FM mode
                                if c.FreqChange == NEW_NOTE {
                                    // KEY_OFF
                                    vgmData &= {VGM_CMD_W_YM2612L, 
                                                R_YM2612_KEYON,
                                                i + floor((i - 5) / 3) - 5} 
                                }
                                f = c.Octave * 12 + c.Note + c.NoteOffs + c.Transpose 
                                if f < 0 {
                                    c.Freq = ym2612FreqTable[0]
                                } else if f > 95 {
                                    c.Freq = ym2612FreqTable[11] + 7 * 0x800
                                } else {
                                    c.Freq = ym2612FreqTable[f % 12] + (f / 12) * 0x800
                                }
                                if c.Freq != currSettings.Freq {
                                    vgmData &= {VGM_CMD_W_YM2612L + floor((i - 5) / 3),
                                                R_YM2612_FHI_BLK + remainder((i - 5), 3),
                                                (c.Freq / 0x100)}
                                    vgmData &= {VGM_CMD_W_YM2612L + floor((i - 5) / 3),
                                                R_YM2612_FLO + remainder((i - 5), 3),
                                                (c.Freq & 0xFF)}
                                    currSettings.Freq = c.Freq
                                }
                                vgmData &= {VGM_CMD_W_YM2612L, 
                                            R_YM2612_KEYON,
                                            0xF0 + i + floor((i - 5) / 3)- 5} 
                            }
                        case specs.CHIP_YM2151:
                            if c.FreqChange == NEW_NOTE {
                                // KEY_OFF
                                vgmData &= {VGM_CMD_W_YM2151, R_YM2151_KEYON, i - 1}
                            }
                            f = (c.Octave + 1) * 12 + c.Note + c.NoteOffs + c.Transpose
                            if f < 0 {
                                c.Freq = ym2151FreqTable[0]
                            } else if f > 95 {
                                c.Freq = ym2151FreqTable[11] + 7 * 0x10
                            } else {
                                c.Freq = ym2151FreqTable[f % 12] + (f / 12) * 0x10
                                if (f % 12) == 0 {
                                    c.Freq -= 0x10
                                }
                            }
                            if i == 8 && c.Mode == 1 {
                                // Noise mode
                                if c.Freq != currSettings.Freq {
                                    vgmData &= {VGM_CMD_W_YM2151,
                                                R_YM2151_NOISE,
                                                0x80 + (c.Freq & 0x1F)}
                                    //lastChannelSetting[i][1] = channel[i][8]
                                }
                            }
                            // FM mode
                            if c.Freq != currSettings.Freq {
                                vgmData &= {VGM_CMD_W_YM2151,
                                        R_YM2151_KEYCODE + i - 1,
                                        channel[i][8]}
                                currSettings.Freq = c.Freq
                            }
                            // KEY_ON
                            vgmData &= {VGM_CMD_W_YM2151, R_YM2151_KEYON, 0x78 + i - 1}  // ToDo: was 0xE8. Correct? 
                        }
                    }
                    
                    if c.VolChange != 0 {
                        switch song.GetChannelType(c.Num) {
                         case specs.CHIP_SN76489:
                            vol = c.Volume.Vol + c.VolOffs
                            if vol != currSettings.Volume.Vol {
                                if vol > 15 {
                                    vol = 15
                                } else if vol < 0 {
                                    vol = 0
                                }
                                vgmData &= {VGM_CMD_W_PSG, SN76489_VOL_LATCH | (vol ^ 15) | (c.Num * 0x20)}
                                vgmData &= {VGM_CMD_W_PSG, SN76489_VOL_DATA  | (vol ^ 15) | (c.Num * 0x20)}
                                currSettings.Volume.Vol = vol
                            }
                            
                        case specs.CHIP_YM2413:
                            vol = c.Volume.Vol + c.VolOffs
                            if vol != currSettings.Volume.Vol {
                                if vol > 15 {
                                    vol = 15
                                } else if vol < 0 {
                                    vol = 0
                                }
                                //if channel[i][CHN_MODE] = 0 then
                                //? vol
                                currSettings.Volume.Vol = vol
                                vol = ^= 15
                                if c.Mode != 1 {
                                    vol += c.Duty
                                    vgmData &= {VGM_CMD_W_YM2413, R_YM2413_INS_VOL + (i - 5), vol}
                                    //printf(1, "*Ch%d: Writing %02x to %02x\n", {i-5, vol, #30 + (i - 5)})
                                } else {
                                    //vgmData &= {#51, #30 + (i - 5), vol}
                                    vgmData &= {VGM_CMD_W_YM2413, R_YM2413_INS_VOL+6, 0x00}
                                    vgmData &= {VGM_CMD_W_YM2413, R_YM2413_INS_VOL+7, 0x00}
                                    vgmData &= {VGM_CMD_W_YM2413, R_YM2413_INS_VOL+8, 0x00}
                                    //printf(1, "#Ch%d: Writing %02x to %02x\n", {i-5, #F0 + vol, #30 + (i - 5)})                                   
                                }
                                //elsif channel[i][CHN_MODE] = 1 and in_range(i, 11, 13) then
                                //  vgmData &= {#51, #30 + (i - 2), xor_bits(vol, 15)}
                                //  printf(1, "Ch%d: Writing %02x to %02x\n", {i-5, xor_bits(vol, 15), #30 + (i - 2)})
                                //end if
                                //lastChannelSetting[i][2] = vol
                            }
                            
                        case specs.CHIP_YM2612:
                            for j := 0; j < 4; j++ {
                                //if j = channel[i][CHN_OPER] or channel[i][CHN_OPER] = 0 then
                                    vol = c.Volume.Op[j] + c.VolOffs
                                    if vol != currSettings.Volume.Op[j] then
                                        if vol > 127 {
                                            vol = 127
                                        } else if vol < 0 {
                                            vol = 0
                                        }
                                        if c.Mode == 0 {
                                            vgmData &= {VGM_CMD_W_YM2612L + floor((i - 5) / 3),
                                                        R_YM2612_TL + (j - 1) * 4 + remainder((i - 5), 3),
                                                        (vol ^ 127)}
                                        //? {i, j, xor_bits(vol, 127)}
                                        }
                                        currSettings.Volume.Op[j] = vol
                                    }
                                //end if
                            }
                            
                        case specs.CHIP_YM2151:
                            for j := 0; j < 4; j++ {
                                //if j = channel[i][CHN_OPER] or channel[i][CHN_OPER] = 0 then
                                    //? {i, j, channel[i][CHN_VOLUME][j]}
                                    vol = c.Volume.Op[j] + c.VolOffs
                                    if vol != currSettings.Volume.Op[j] {
                                        if vol > 127 {
                                            vol = 127
                                        } else if vol < 0 {
                                            vol = 0
                                        }
                                        vgmData &= {VGM_CMD_W_YM2151,
                                                    R_YM2151_TL + (j - 1) * 8 + i - 1,
                                                    (vol ^ 127)}
                                        currSettings.Volume.Op[j] = vol
                                    }
                                //end if
                            }
                        }
                    }
                end if
            end if
        end for
        
        
        if timing.UpdateFreq == 50 {
            chnUpdateDelay = 882
        } else {
            chnUpdateDelay = 735
        }
        
        if ym2612 && channel[10].Mode == 2 && pcmDelay >= 0 {
            if pcmDelay > 0 && pcmDelay <= chnUpdateDelay {
                if pcmDelay < 16 {
                    vgmData &= (0x70 | pcmDelay)
                } else {
                    vgmData &= {0x61, (pcmDelay & 0xFF), (pcmDelay / 0x100)}
                end if
                chnUpdateDelay -= pcmDelay
                totalWaits += pcmDelay
            }
            pcmDelay = pcmDelayReload
            for pcmDelay <= chnUpdateDelay {
                if pcmDataPos <= pcmDataLen {
                    vgmData &= (0x80 | pcmDelay)
                    chnUpdateDelay -= pcmDelay
                    totalWaits += pcmDelay
                    pcmDataPos += 1
                } else {
                    //vgmData &= {#52, #2A, #00}
                    pcmDelay = -1
                    break
                }
            }
            if chnUpdateDelay > 0 && chnUpdateDelay < pcmDelay {
                pcmDelay -= chnUpdateDelay
            }
        }
        
        if chnUpdateDelay == 882 {
            vgmData &= 0x63
        } else if chnUpdateDelay == 735 {
            vgmData &= 0x62
        } else if chnUpdateDelay > 0 {
            if chnUpdateDelay < 16 {
                vgmData &= (0x70 | chnUpdateDelay)
            } else {
                vgmData &= {0x61, (chnUpdateDelay & 0xFF), (chnUpdateDelay / 0x100)}
            }
        }
        
        totalWaits += chnUpdateDelay
        
        iterations[1] += 1
        if iterations[1] == 54000 then
            compiler.ERROR("Internal error: Appears to be stuck in an infinite loop. Exiting\n")
            //? {channel[1][2], channel[2][2], channel[3][2], channel[4][2]}
            break
        }
    end while
    //? vgmData[#40..length(vgmData)]
    
    if length(loopPos) {
        vgmData = vgmData[1..#18] &
              int_to_bytes(totalWaits) &
              int_to_bytes(loopPos[1] - #1C) &
              int_to_bytes(totalWaits - loopPos[2]) &
              vgmData[#25..length(vgmData)]
        loopPos[2] = totalWaits - loopPos[2]
    } else {
        vgmData = vgmData[1..#18] & int_to_bytes(totalWaits) & vgmData[#1D..length(vgmData)]
        loopPos = {0, 0}
    }   

    vgmData &= 0x66
    
    // Set EOF offset and GD3 offset
    vgmData = vgmData[1..4] & int_to_bytes(length(vgmData) - 4) & vgmData[9..length(vgmData)]
    vgmData = vgmData[1..#14] & int_to_bytes(length(vgmData) - #14) & vgmData[#19..length(vgmData)]


    if compressVgm then
        if gzwrite(outFile, vgmData) != length(vgmData) then
        end if
    else
        puts(outFile, vgmData)
    end if

    // Create GD3 tag
    s = date()
    gd3 = ascii_to_wide(songTitle & 0 & 0)
    if equal(songGame, "Unknown") then
        gd3 &= ascii_to_wide(songAlbum & 0 & 0)
    else
        gd3 &= ascii_to_wide(songGame & 0 & 0)
    end if
    if ym2612 then
        gd3 &= ascii_to_wide("Sega Genesis" & 0 & 0)
    elsif ym2151 then
        gd3 &= ascii_to_wide("Capcom Play system" & 0 & 0)
    elsif supportsPAL then
        gd3 &= ascii_to_wide("Sega Master System" & 0 & 0)
    else
        gd3 &= ascii_to_wide("Sega Game Gear" & 0 & 0)
    end if
    gd3 &= ascii_to_wide(songProgrammer & 0 & 0)
    gd3 &= ascii_to_wide(sprintf("%04d/%02d/%02d", {s[1] + 1900, s[2], s[3]}) & 0)
    gd3 &= ascii_to_wide("XPMC" & 0)
    gd3 &= ascii_to_wide("Composer: " & songComposer & 0)
    gd3 = "Gd3 " &
          {0, 1, 0, 0} &
          int_to_bytes(length(gd3)) &
          gd3


    if compressVgm {
        if gzwrite(outFile, gd3) != length(gd3) then
        end if
        if gzclose(outFile) != Z_OK then
        end if
    } else {
        puts(outFile, gd3)
        close(outFile)
    }

    if verbose {
        for i = 1 to nChannels do
            printf(1, "Song %d, Channel " & supportedChannels[i] & ": %d / %d ticks\n", {song, round2(songLen[song][i]), round2(songLoopLen[song][i])})
        end for
        printf(1,"VGM size: %d bytes + %d bytes GD3\nVGM length: %d / %d seconds\n", {length(vgmData), length(gd3), floor(totalWaits / 44100), floor(loopPos[2] / 44100)})
    }
}
