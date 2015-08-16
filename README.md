# spector

spector is a command line spectral analysis tool designed to visualise the distribution of streams of numbers representing something like latency, duration or size.

It samples data read from stdin and builds a histogram, using ANSI colour codes to display the distribution as a heat map.

It was inspired by [this Sysdig tweet](https://twitter.com/sysdig/status/618826906310324224), and follows on from my [distribution Awk script](https://github.com/mrmanc/log-ninja#distribution) which displays an actual histogram (although it also has some realtime functionality).

Please be kind… it is my first play with Go, and it is horrendous code. I’ve not tested it on anything other than OS X yet.

## Example
(dtrace example borrowed from [this HeatMap tool](https://github.com/brendangregg/HeatMap))

Using the below (after adjusting spector.go to use the exponential scale)…

```
$ sudo dtrace -qn 'syscall::read:entry { self->ts = timestamp; }
    syscall::read:return /self->ts/ {
    printf("%d\n", (timestamp - self->ts) / 1000); self->ts = 0; }' | spector
```

will display something a bit like this in your terminal:

![dtrace spectrograph](https://github.com/mrmanc/spector/blob/master/sample.png)

Using the provided normal command to generate some test data based on a normal distribution, you might see the below:

![normal distribution spectrograph](https://github.com/mrmanc/spector/blob/master/normal-distribution.png)


## Historic data

If you have historic logs with a formatted time in the line, you can use the pacemaker command to indicate to spector how to sample the data. The pacemaker command will add extra lines to the streamed output as a signal to the spector command. Feel free to leave the time text in the output, so long as the number you wish to visualise is the last thing in the line.

For example, with a log file such as below, you could run `cat test.log | pacemaker | spector`.

```
Tue Nov 11 10:14:52.130 duration=60.7
Tue Nov 11 10:14:53.130 duration=15.2
Tue Nov 11 10:14:53.131 duration=39.5
Tue Nov 11 10:14:53.140 duration=20.2
Tue Nov 11 10:14:53.237 duration=55.9
Tue Nov 11 10:14:56.845 duration=44.4
Tue Nov 11 10:14:58.493 duration=56.8
Tue Nov 11 10:14:58.510 duration=62.4
Tue Nov 11 10:14:58.510 duration=24.3
Tue Nov 11 10:14:58.510 duration=43.2
Tue Nov 11 10:14:58.510 duration=66.0
Tue Nov 11 10:14:59.199 duration=72.7
```

## Building

Once you have [installed Go](https://golang.org/doc/install) you can build the commands using 'go build spector.go && go build normal.go && go build pacemaker.go'. That should create binaries in that directory, which you can either put on your path or run in situ. I’ll get around to working out the correct way to organise the project at some point as I know it should be possible to get Go to install the commands.

## Future improvements

* Allow user to select a scale out of linear, exponential or logarithmic (perhaps at run time)
* Allow user to select whether to use grayscale or colour
* Allow user to select the update frequency
* Determine the terminal width dynamically
* Potentially split the summarisation (histogram) functionality from the rendering, since it is useful on it’s own and can allow you to use much less space to store a replay of a period in plain text