package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

	asla "github.com/cocoonlife/goalsa"
	"github.com/youpy/go-wav"
)

var lastSpookySound string = ""

func main() {
	// get the devices in your system by using the `aplay -L` command
	// Example of a headphone jack input in raspberrypi
	// 	- hw:CARD=Headphones,DEV=0
	// Example of a virtual PCM to a bluetooth speaker by using the bluealsa library
	// - bluealsa:DEV=78:44:05:EB:93:71,PROFILE=a2dp
	deviceName := flag.String("device", "bluealsa:DEV=78:44:05:EB:93:71,PROFILE=a2dp", "Device name where the spooky sounds will be played")
	maximumInterval := flag.Int("maximumInterval", 15, "Maximum interval for when the spooky sound can be played")
	flag.Parse()

	// we want our spooky sounds to be unpredictable
	rand.Seed(time.Now().UnixNano())

	for {
		// keep playing spooky sounds unless stopped
		sound := nextSpookySound()
		samples := readSpookySound(sound)
		playSpookySound(*deviceName, samples)

		// wait an unexpected amount of time for the next spooky sound
		sleepTime := time.Duration(rand.Int31n(int32(*maximumInterval))) * time.Second
		fmt.Printf("Play next spooky sound on %s\n", sleepTime)
		time.Sleep(sleepTime)
	}
}

func nextSpookySound() string {
	spookySounds := getSpookySounds()
	candidate := spookySounds[rand.Intn(len(spookySounds))]

	// never play the same spooky sound twice
	if candidate == lastSpookySound {
		return nextSpookySound()
	} else {
		lastSpookySound = candidate
	}

	return candidate
}

func getSpookySounds() []string {
	return []string{"bell", "cat", "laugh", "hauntedhouse", "raven", "witches"}
}

func readSpookySound(path string) []int16 {
	file, _ := os.Open("sounds/" + path + ".wav")
	reader := wav.NewReader(file)

	defer file.Close()

	// http://soundfile.sapp.org/doc/WaveFormat/
	format, err := reader.Format()
	checkErr(err)

	fmt.Printf("Reading %v.%v\n", path, "wav")
	fmt.Printf("File info: %v\n", format)

	var data []int16
	for {
		samples, err := reader.ReadSamples()
		if err == io.EOF {
			break
		}

		for _, sample := range samples {
			// convert stereo to mono by only getting the first channel
			value := int16(reader.IntValue(sample, 0))
			data = append(data, value)
		}
	}

	return data
}

func playSpookySound(deviceName string, data []int16) {
	fmt.Printf("Playing sound on %v\n", deviceName)

	// only supports 44100 Hz sample rate
	// any other wav file must be converted to 44100 Hz for this to work
	// you can use ffmpeg or libopus in any OS to do the rate conversion
	p, err := asla.NewPlaybackDevice(deviceName, 1, asla.FormatS16LE, 44100,
		asla.BufferParams{})
	checkErr(err)

	// send the data to the PCM (pulse code modulation) real or virtual device that will play the spooky sound
	_, err = p.Write(data)
	checkErr(err)

	p.Close()
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
