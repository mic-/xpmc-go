import (
    "fmt"
    "math"
    "../utils"
)

var fileData []byte
var fileDataPos

// Get an unsigned word (16 bits)
func getWord() int {
    var w int
    w = fileData[fileDataPos++]
    w += (fileData[fileDataPos++] * 0x100)
    return w
}

// Get a signed word (16 bits) from file fh
func getSword() int {
    var w int
    if fileDataPos >= len(fileData) {
        return 65536
    }
    w = fileData[fileDataPos++]
    w += (fileData[fileDataPos++] * 0x100)
    if (w & 0x8000) != 0 {
        return -(32768 - (w & 0x7FFF))
    }
    return w
}

// Get a dword (32 bits)
func getDword() int {
    var dw int
    
    dw = fileData[fileDataPos++]
    dw += (fileData[fileDataPos++] * 0x100)
    dw += (fileData[fileDataPos++] * 0x10000)
    dw += (fileData[fileDataPos++] * 0x1000000)
    return dw
}


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

    fileData, err = ioutil.ReadFile(fname)

    s = get_bytes(fn, 4)
    if not equal(s, "RIFF") then
        ERROR("No RIFF tag found in " & fname, -1)
    end if
    
    _ := getDword()
    
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
        dw = getDword()
    
        if equal(s, "fmt ") {
            wavFormat         = getWord()
            wavChannels       = getWord()
            wavSamplesPerSec  = getDword()
            wavAvgBytesPerSec = getDword()
            wavBlockAlign     = getWord()
            wavBitsPerSample  = getWord()
        
            if wavFormat != 1 || (wavBitsPerSample != 8 && wavBitsPerSample != 16) || wavChannels > 2 {
                utils.ERROR("Unsupported wav format in " & fname, -1)
            }
        
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
            
        } else if equal(s, "data") {
            dataStart = where(fn)
            dataSize = dw
            
            sampleDiv = 0
                       
            if wavBitsPerSample == 8 {
                if wavChannels == 1 {
                    samplesInChunk = dataSize
                    wavData = []float64{}
                    for floor(pos) <= samplesInChunk {
                        nextPos = floor(pos + deltaPos)
                        if floor(pos) < nextPos {
                            b = 0
                            sampleDiv = 0
                            for floor(pos) < nextPos {
                                b += getc(fn)
                                pos += 1 
                                sampleDiv += 1
                            }
                        }
                        if sampleDiv != 0 {
                            wavData = append(wavData, math.Floor(b / sampleDiv))
                        }
                    }
                } else if wavChannels == 2 {
                    samplesInChunk = dataSize / 2
                    wavData = []float64{}
                    for floor(pos) <= samplesInChunk {
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
                        if sampleDiv != 0 {
                            wavData = append(wavData, math.Floor(b / (sampleDiv * 2.0)))
                        }
                    }
                end if
            else
                if wavChannels == 1 {
                    samplesInChunk = floor(dataSize / 2)
                    wavData = []float64{}
                    for floor(pos) <= samplesInChunk {
                        nextPos = floor(pos + deltaPos)
                        if floor(pos) < nextPos {
                            w = 0
                            sampleDiv = 0
                            for floor(pos) < nextPos {
                                w += floor((getSword() + 32768) / 256)
                                pos += 1 
                                sampleDiv += 1
                            }
                        }
                        if sampleDiv != 0 {
                            wavData = append(wavData, math.Floor(w / sampleDiv))
                        }
                    }
                } else if wavChannels == 2 {
                    samplesInChunk = floor(dataSize / 4)
                    wavData = []float64{}
                    for floor(pos) <= samplesInChunk {
                        nextPos = floor(pos + deltaPos)
                        if floor(pos) < nextPos then
                            w = 0
                            sampleDiv = 0
                            while floor(pos) < nextPos do
                                w += floor((getSword() + 32768) / 256) + floor((getSword() + 32768) / 256) 
                                pos += 1 
                                sampleDiv += 1
                            end while
                        end if
                        if sampleDiv != 0 {
                            wavData = append(wavData, math.Floor(w / (sampleDiv * 2.0)))
                        }
                    }
                }
            
            }

        } else {
            // Unhandled chunk type, just skip it.
            fileDataPos += dw
        }
    }
       
    utils.INFO(fmt.Sprintf("Size of converted sample: %d bytes\n", len(wavData)))
    
    wavDataInt = make([]int, len(wavData))
    for i := range wavData {
        wavDataInt[i] = int(math.Floor((wavData[i] * float64(volume)) / 100.0))
    }

    return wavDataInt
}
