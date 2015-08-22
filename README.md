# spector

spector is a command line spectral analysis tool designed to visualise the distribution of streams of numbers representing something like latency, duration or size on the command line.

It samples data read from stdin and builds a histogram, using ANSI colour codes to display the distribution as an ascii heat map.

It was inspired by [this Sysdig tweet](https://twitter.com/sysdig/status/618826906310324224), and follows on from my [distribution Awk script](https://github.com/mrmanc/log-ninja#distribution) which displays an actual histogram (although it also has some realtime functionality).

It is similar to [terminal-heatmap](https://github.com/jclulow/terminal-heatmap) which I found after writing spector, and which is a much more mature C based horizontal format heatmap tool (i.e. time along the bottom axis and magnitude on the vertical axis, so check that out too. It does not have a pacemaker like function but you could approximate something similar by [piping your logs through sed and awk and using it to do the batching (click for demo)](http://showterm.io/b0cc7e9da5547c59e24fc).

Please be kind… it is my first play with Go, and I’m not proud of the code. I’d love to hear advice and critique from other Go developers. I’ve not tested this on anything other than OS X yet, so let me know how you get on.

## Example
(dtrace example borrowed from [this HeatMap tool](https://github.com/brendangregg/HeatMap))

Using the below (after adjusting spector.go to use the exponential scale)…

```
$ sudo dtrace -qn 'syscall::read:entry { self->ts = timestamp; }
    syscall::read:return /self->ts/ {
    printf("%d\n", (timestamp - self->ts) / 1000); self->ts = 0; }' | spector
```

will display something a bit like this in your terminal:

![dtrace spectrograph](https://dl.dropboxusercontent.com/u/4117956/spector/dtrace.gif)

You can use the provided normal command to generate some test data based on a normal distribution:

![normal distribution spectrograph](https://dl.dropboxusercontent.com/u/4117956/spector/normal.gif)

Using the provided sample.log and the pacemaker command you can play back activity:

![sample log file spectrograph](https://dl.dropboxusercontent.com/u/4117956/spector/sample.gif)

## Historic data

If you have historic logs with a formatted time in the line, you can use the pacemaker command to indicate to spector how to sample the data. The pacemaker command will add extra lines to the streamed output as a signal to the spector command. Feel free to leave the time text in the output, so long as the number you wish to visualise is the last thing in the line. Pacemaker will look for a time matching something like this: `10:14:52`. It’s tolerant of times out of order, but this will result in repeated periods.

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

## Installing

If you wish to use the commands, you will need to have [Go](http://golang.org/) set up. Once you have [installed Go](https://golang.org/doc/install) you can get the commands by running:

```
go get github.com/mrmanc/spector/normal
go get github.com/mrmanc/spector/pacemaker
go get github.com/mrmanc/spector/spector
```

## Contributing

I would love to welcome improvements to spector, no matter how large or small! If you want to further develop spector fork the repo and see [How to Write Go Code](https://golang.org/doc/code.html) for advice if you are not familiar with Go.

## Future improvements

* Potentially split the summarisation (histogram) functionality from the rendering, since it is useful on it’s own and can allow you to use much less space to store a replay of a period in plain text
* Normalise the amplitude using the time since last sample to smooth out results when processing is slow
* Try using a static logarithmic scale for the amplitude scaling to provide consistency and better resolution on small amplitudes
* Allow the user to switch scales dynamically when the command is running
* Use rank based rendering as suggested by this [blog post by Dave Pacheco](http://dtrace.org/blogs/dap/2011/06/20/heatmap-coloring/).
