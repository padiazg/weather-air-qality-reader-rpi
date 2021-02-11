# Weather and Air quality reporter

# Get started
Clone the repo
```bash
$ mkdir -p ~/go/src/github/padiazg
$ cd ~/go/src/github/padiazg
$ git clone https://github.com/padiazg/weather-air-qality-reader-rpi.git
$ cd weather-air-qality-reader-rpi
$ go mod tidy
```

# Build
You can build it in a regular PC, then copy the binary to the RPi
```bash
GOOS=linux GOARCH=arm go build main.go
scp main pi@rpizerow1.local:/home/pi
```
# Run
Need to figure out a good way. Keep in tunning for how we will do this.
In the meantime
```bash
$ ssh pi@rpizerow1.local:/home/pi
$ ./main >> airelibre.log  2>&1 &
```

# Diagrams & sensor wirings
WIP

# Troubleshooting
WIP
