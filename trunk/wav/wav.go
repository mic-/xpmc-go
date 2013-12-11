package wav

import (
    "fmt"
    "io/ioutil"
    "math"
    "../utils"
)

var fileData []byte
var fileDataPos int


// Get an unsigned word (16 bits)
func getWord() int {
    var w int
    w = int(fileData[fileDataPos])
    fileDataPos++
    w += (int(fileData[fileDataPos]) * 0x100)
    fileDataPos++
    return w
}

// Get a signed word (16 bits) from file fh
func getSword() int {
    var w int
    if fileDataPos >= len(fileData) {
        return 65536
    }
    w = int(fileData[fileDataPos])
    fileDataPos++
    w += (int(fileData[fileDataPos]) * 0x100)
    fileDataPos++
    if (w & 0x8000) != 0 {
        return -(32768 - (w & 0x7FFF))
    }
    return w
}

// Get a dword (32 bits)
func getDword() int {
    var dw int
    
    dw = int(fileData[fileDataPos])
    dw += (int(fileData[fileDataPos+1]) * 0x100)
    dw += (int(fileData[fileDataPos+2]) * 0x10000)
    dw += (int(fileData[fileDataPos+3]) * 0x1000000)
    fileDataPos += 4
    return dw
}


// Extract the sample data from a WAV file and convert it to mono with a given sample rate and volume
func ConvertWav(fname string, sampleRate int, volume int) []int {
    var accum int
    var wavFormat, wavChannels, wavSamplesPerSec, wavBitsPerSample int
    var err error
    var pos, deltaPos float64
    
    fileData, err = ioutil.ReadFile(fname)
    if err != nil {
        utils.ERROR("Unable to read from " + fname)
    }
    fileDataPos = 0
    
    s := string(fileData[fileDataPos : fileDataPos+4])
    fileDataPos += 4
    if s != "RIFF" {
        utils.ERROR("No RIFF tag found in " + fname)
    }
    
    _ = getDword()
    
    s = string(fileData[fileDataPos : fileDataPos+4])
    fileDataPos += 4
    if s != "WAVE" {
        utils.ERROR("No WAVE tag found in " + fname)
    }

    dataSize := -1
    wavData := []float64{}
    
    // Read the chunks
    for {
        // Get the chunk ID
        if (fileDataPos + 3) >= len(fileData) {
            break
        }
        chunkId := string(fileData[fileDataPos : fileDataPos+4])
        fileDataPos += 4

        // Get the chunk size
        chunkSize := getDword()
    
        if chunkId == "fmt " {
            wavFormat         = getWord()
            wavChannels       = getWord()
            wavSamplesPerSec  = getDword()
            _                 = getDword()      // Average bytes per second
            _                 = getWord()       // Block alignment
            wavBitsPerSample  = getWord()
        
            if wavFormat != 1 || (wavBitsPerSample != 8 && wavBitsPerSample != 16) || wavChannels > 2 {
                utils.ERROR("Unsupported wav format in " + fname)
            }
        
            deltaPos = 1.0
            if sampleRate > 0 {
                deltaPos = float64(wavSamplesPerSec) / float64(sampleRate)
            }
            pos = 1.0
            
            if deltaPos != 1.0 {
                utils.INFO(fmt.Sprintf("Resampling from %d to %d Hz", wavSamplesPerSec, sampleRate))
            }
            
            if wavChannels > 1 {
                utils.INFO("Converting sample to mono")
            }
            
            if wavBitsPerSample != 8 {
                utils.INFO("Converting sample to 8-bit unsigned")
            }
            
        } else if chunkId == "data" {
            //dataStart = where(fn)
            //dataSize := chunkSize
            
            sampleDiv := 0
                       
            if wavBitsPerSample == 8 {
                if wavChannels == 1 {
                    samplesInChunk := dataSize
                    wavData = []float64{}
                    for int(pos) <= samplesInChunk {
                        nextPos := int(pos + deltaPos)
                        if int(pos) < nextPos {
                            accum = 0
                            sampleDiv = 0
                            for int(pos) < nextPos {
                                accum += int(fileData[fileDataPos])
                                fileDataPos++
                                pos += 1.0 
                                sampleDiv++
                            }
                        }
                        if sampleDiv != 0 {
                            wavData = append(wavData, math.Floor(float64(accum) / float64(sampleDiv)))
                        }
                    }
                } else if wavChannels == 2 {
                    samplesInChunk := dataSize / 2
                    wavData = []float64{}
                    for int(pos) <= samplesInChunk {
                        nextPos := int(pos + deltaPos)
                        if int(pos) < nextPos {
                            accum = 0
                            sampleDiv = 0
                            for int(pos) < nextPos {
                                accum += int(fileData[fileDataPos])
                                accum += int(fileData[fileDataPos+1])
                                fileDataPos += 2
                                pos += 1.0 
                                sampleDiv++
                            }
                        }
                        if sampleDiv != 0 {
                            wavData = append(wavData, math.Floor(float64(accum) / (float64(sampleDiv) * 2.0)))
                        }
                    }
                }
            } else {
                if wavChannels == 1 {
                    samplesInChunk := dataSize / 2
                    wavData = []float64{}
                    for int(pos) <= samplesInChunk {
                        nextPos := int(pos + deltaPos)
                        if int(pos) < nextPos {
                            accum = 0
                            sampleDiv = 0
                            for int(pos) < nextPos {
                                accum += int(float64(getSword() + 32768) / 256.0)
                                pos += 1.0 
                                sampleDiv++
                            }
                        }
                        if sampleDiv != 0 {
                            wavData = append(wavData, math.Floor(float64(accum) / float64(sampleDiv)))
                        }
                    }
                } else if wavChannels == 2 {
                    samplesInChunk := dataSize / 4
                    wavData = []float64{}
                    for int(pos) <= samplesInChunk {
                        nextPos := int(pos + deltaPos)
                        if int(pos) < nextPos {
                            accum = 0
                            sampleDiv = 0
                            for int(pos) < nextPos {
                                accum += int(float64(getSword() + 32768) / 256.0) + int(float64(getSword() + 32768) / 256.0) 
                                pos += 1.0 
                                sampleDiv++
                            }
                        }
                        if sampleDiv != 0 {
                            wavData = append(wavData, math.Floor(float64(accum) / (float64(sampleDiv) * 2.0)))
                        }
                    }
                }
            
            }

        } else {
            // Unhandled chunk type, just skip it.
            fileDataPos += chunkSize
        }
    }
       
    utils.INFO(fmt.Sprintf("Size of converted sample: %d bytes\n", len(wavData)))
    
    wavDataInt := make([]int, len(wavData))
    for i := range wavData {
        wavDataInt[i] = int(math.Floor((wavData[i] * float64(volume)) / 100.0))
    }

    return wavDataInt
}
