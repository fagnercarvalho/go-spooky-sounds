# go-spooky-sounds

![Go version](https://img.shields.io/github/go-mod/go-version/fagnercarvalho/go-spooky-sounds)

Play spooky sound effects for Halloween from your speaker or Bluetooth speaker.

Made in Go.

### Requirements

- This utility was tested in Go 1.13.4.
- This only supports Linux.
- If you want to play these spooky sounds in your Bluetooth speaker you will need to pair your system with the Bluetooth device first. I made a brief guide on [how to pair and connect to a Bluetooth device here](https://gist.github.com/fagnercarvalho/2755eaa492a8aa27081e0e0fe7780d14).

### Installation

```
go get -u github.com/fagnercarvalho/go-spooky-sounds
```

### Usage

This will play a spooky sound periodically from a list of 6 predefined sounds. This can be used in your Halloween Party or just to get you in the mood of the season.

You can add your own files in the sounds folder. They must be in the .wav format and have a sample rate of 44100 Hz (most tracks in .wav do, please use ffmpeg or libopus to check the rate or do a rate conversion to 44100 Hz).

### Examples

Here I played the spooky sounds directly in my Raspberry Pi audio jack.

```
go-spooky-sounds -device "hw:CARD=Headphones,DEV=0"
```

Or you can play to your Bluetooth device by using the [bluez-alsa library](https://github.com/Arkq/bluez-alsa) virtual PCM.

```
go-spooky-sounds -device "bluealsa:DEV=78:44:05:EB:93:71,PROFILE=a2dp"
```

The default value for the maximum interval between the sounds is 15 minutes but you can change this value by using the `maximumInterval` option.

```
go-spooky-sounds -device "bluealsa:DEV=78:44:05:EB:93:71,PROFILE=a2dp" -maximumInterval 10
```

### How it works?

Periodically this will read the sound effect (a .wav file) to be played and then send the byte array of sound samples to the PCM device (pulse code modulation) by using the ALSA API (Advanced Linux Sound Architecture) made in C. The ALSA API will then call the driver that will be able to call the actual hardware device to play the sound. Specifically for Bluetooth devices the [bluez-alsa library](https://github.com/Arkq/bluez-alsa) must be used to create a virtual PCM that will expose the Bluetooth device as any other sound device available in the system.