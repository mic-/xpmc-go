import (
    "fmt"
    "math"
    "../utils"
)


// Extract the sample data from a WAV file and convert it to mono with a given sample rate and volume
func ConvertWav(fname string, sampleRate int, volume int) []int {
    integer fn
    integer b, w, wavFormat, wavChannels, wavSamplesPerSec, wavAvgBytesPerSec,
            wavBlockAlign, wavBitsPerSample, samplesInChunk, nextPos, sampleDiv
    sequence s, wavData
    atom dw, dataStart, dataSize, pos, deltaPos
    
    fn = open(fname,"rb")
    if fn < 0 then
        ERROR("Unable to open file: " & fname, lineNum)
    end if

    s = get_bytes(fn, 4)
    if not equal(s, "RIFF") then
        ERROR("No RIFF tag found in " & fname, -1)
    end if
    
    if get_dword(fn) then end if
    
    s = get_bytes(fn,4)
    if not equal(s, "WAVE") then
        ERROR("No WAVE tag found in " & fname, -1)
    end if

    dataSize := -1
    wavData := []float64{}
    
    // Read the chunks
    for {
        // Get the chunk ID
        s = get_bytes(fn, 4)
        if length(s) < 4 then
            exit
        end if

        // Get the chunk size
        dw = get_dword(fn)
    
        if equal(s, "fmt ") then
            wavFormat = get_word(fn)
            wavChannels = get_word(fn)
            wavSamplesPerSec = get_dword(fn)
            wavAvgBytesPerSec = get_dword(fn)
            wavBlockAlign = get_word(fn)
            wavBitsPerSample = get_word(fn)
        
            if wavFormat!=1 or (wavBitsPerSample!=8 and wavBitsPerSample!=16) or wavChannels > 2 then
                utils.ERROR("Unsupported wav format in " & fname, -1)
            end if
        
            deltaPos := 1.0
            if sampleRate > 0 {
                deltaPos = wavSamplesPerSec / sampleRate
            }
            pos := 1.0
            
            if deltaPos != 1.0 {
                utils.INFO(fmt.Sprintf("Resampling from %d to %d Hz", wavSamplesPerSec, sampleRate))
            }
            
            if wavChannels > 1 {
                utils.INFO("Converting sample to mono")
            }
            
            if wavBitsPerSample != 8 {
                utils.INFO("Converting sample to 8-bit unsigned")
            }
            
        elsif equal(s, "data") then
            dataStart = where(fn)
            dataSize = dw
            
            sampleDiv = 0
                       
            if wavBitsPerSample = 8 then
                if wavChannels = 1 then
                    samplesInChunk = dataSize
                    wavData = []float64{}
                    while floor(pos) <= samplesInChunk do
                        nextPos = floor(pos + deltaPos)
                        if floor(pos) < nextPos then
                            b = 0
                            sampleDiv = 0
                            while floor(pos) < nextPos do
                                b += getc(fn)
                                pos += 1 
                                sampleDiv += 1
                            end while
                        end if
                        if sampleDiv then
                            wavData = append(wavData, math.Floor(b / sampleDiv))
                        end if
                    end while
                elsif wavChannels = 2 then
                    samplesInChunk = floor(dataSize / 2)
                    wavData = []float64{}
                    while floor(pos) <= samplesInChunk do
                        nextPos = floor(pos + deltaPos)
                        if floor(pos) < nextPos then
                            b = 0
                            sampleDiv = 0
                            while floor(pos) < nextPos do
                                b += getc(fn) + getc(fn)
                                pos += 1 
                                sampleDiv += 1
                            end while
                        end if
                        if sampleDiv then
                            wavData = append(wavData, math.Floor(b / (sampleDiv * 2.0)))
                        end if
                    end while
                end if
            else
                if wavChannels = 1 then
                    samplesInChunk = floor(dataSize / 2)
                    wavData = []float64{}
                    while floor(pos) <= samplesInChunk do
                        nextPos = floor(pos + deltaPos)
                        if floor(pos) < nextPos then
                            w = 0
                            sampleDiv = 0
                            while floor(pos) < nextPos do
                                w += floor((get_sword(fn) + 32768) / 256)
                                pos += 1 
                                sampleDiv += 1
                            end while
                        end if
                        if sampleDiv then
                            wavData &= floor(w / sampleDiv)
                        end if
                    end while
                elsif wavChannels = 2 then
                    samplesInChunk = floor(dataSize / 4)
                    wavData = []float64{}
                    while floor(pos) <= samplesInChunk do
                        nextPos = floor(pos + deltaPos)
                        if floor(pos) < nextPos then
                            w = 0
                            sampleDiv = 0
                            while floor(pos) < nextPos do
                                w += floor((get_sword(fn) + 32768) / 256) + floor((get_sword(fn) + 32768) / 256) 
                                pos += 1 
                                sampleDiv += 1
                            end while
                        end if
                        if sampleDiv then
                            wavData &= floor(w / (sampleDiv * 2))
                        end if
                    end while
                end if
            
            end if

        } else {
            // Unhandled chunk type, just skip it.
            for i = 1 to dw do
                b = getc(fn)
            end for
        end if
    }
    
    close(fn)
    
    utils.INFO("Size of converted sample: %d bytes\n", length(wavData))
    
    wavDataInt = make([]int, len(wavData))
    for i := range wavData {
        wavDataInt[i] = int(math.Floor((wavData[i] * float64(volume)) / 100.0))
    }

    return wavDataInt
}
