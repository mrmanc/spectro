# Spector

Spector is a command line spectral analysis tool designed to visualise the distribution of streams of numbers representing something like latency, duration or size.

It samples data read from stdin and builds a histogram, using ANSI colour codes to display the distribution as a heat map.

It was inspired by [this Sysdig tweet](https://twitter.com/sysdig/status/618826906310324224), and follows on from my [distribution Awk script](https://github.com/mrmanc/log-ninja#distribution) which displays an actual histogram (although it also has some realtime functionality).

## Example
(dtrace example borrowed from [this HeatMap tool](https://github.com/brendangregg/HeatMap))

```
$ sudo dtrace -qn 'syscall::read:entry { self->ts = timestamp; }
    syscall::read:return /self->ts/ {
    printf("%d\n", (timestamp - self->ts) / 1000); self->ts = 0; }' | spector
```

will display something a bit like this in your terminal:

![Sample output](https://github.com/mrmanc/spector/blob/master/sample.png)

## Future improvements

* Output labels to describe the scale
* Improve the colours to use a prettier scale
* Allow user to select a scale (perhaps at run time)
* Use signals on stdin to determine when to sample, which could be sent through the pipeline by an upstream process, allowing them to be generated against an existing file with timestamps
* Determine the terminal width dynamically
