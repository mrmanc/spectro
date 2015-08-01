# Spector

Spector is a command line spectral analysis tool designed to visualise the distribution of streams of numbers representing something like latency, duration or size.

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

![Sample output](https://github.com/mrmanc/spector/blob/master/sample.png)

## Future improvements

* Allow user to select a scale (perhaps at run time)
* Use signals on stdin to determine when to sample, which could be sent through the pipeline by an upstream process, allowing them to be generated against an existing file with timestamps
* Determine the terminal width dynamically
