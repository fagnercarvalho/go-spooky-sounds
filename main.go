package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	asla "github.com/cocoonlife/goalsa"
	_ "github.com/fagnercarvalho/go-spooky-sounds/statik"
	"github.com/pkg/errors"
	"github.com/rakyll/statik/fs"
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
		samples, err := readSpookySound(sound)
		checkErr(err)
		err = playSpookySound(*deviceName, samples)
		checkErr(err)

		// wait an unexpected amount of time for the next spooky sound
		sleepTime := time.Duration(rand.Int31n(int32(*maximumInterval))) * time.Minute
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

func readSpookySound(path string) ([]int16, error) {

	// sounds are embedded into the executable so we need to use a virtual file system to get the files
	s, err := fs.New()
	if err != nil {
		return nil, errors.Wrap(err, "Error to create virtual file system")
	}

	file, err := s.Open("//sounds//" + path + ".wav")
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error to get `%s` sound file from file system", path))
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error to get `%s` sound file stats", file))
	}

	size := stat.Size()

	// after getting the embedded file we need to create a temporary file into the real file system because the virtual file
	// doesn't have the ReadAt method implemented that needs to be used by the WAV file reader library
	buffer := make([]byte, size)
	_, err = file.Read(buffer)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error when reading `%s` sound file contents", stat.Name()))
	}

	tmpFile, err := createTempFile(buffer)
	if err != nil {
		return nil, err
	}

	defer os.Remove(tmpFile.Name())

	reader := wav.NewReader(tmpFile)

	defer file.Close()

	// http://soundfile.sapp.org/doc/WaveFormat/
	format, err := reader.Format()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error when getting `%s` .wav file format", stat.Name()))
	}

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

	return data, nil
}

func createTempFile(buffer []byte) (*os.File, error) {
	tmpfile, err := ioutil.TempFile("", "track.wav")
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error when creating sound temp file"))
	}

	_, err = tmpfile.Write(buffer)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error when writing to sound temp file"))
	}

	return tmpfile, nil
}

func playSpookySound(deviceName string, data []int16) error {
	fmt.Printf("Playing sound on %v\n", deviceName)

	// only supports 44100 Hz sample rate
	// any other wav file must be converted to 44100 Hz for this to work
	// you can use ffmpeg or libopus in any OS to do the rate conversion
	p, err := asla.NewPlaybackDevice(deviceName, 1, asla.FormatS16LE, 44100,
		asla.BufferParams{})
	if err != nil {
		return errors.Wrap(err, "Error when instantiating audio playback device")
	}

	defer p.Close()

	// send the data to the PCM (pulse code modulation) real or virtual device that will play the spooky sound
	_, err = p.Write(data)
	if err != nil {
		return errors.Wrap(err, "Error when playing sound in playback device")
	}

	return nil
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
