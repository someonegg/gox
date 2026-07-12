# gox
gox contains small Go utility packages.

* Package pool provides object pool implementations, including a smoother pool
  for very large objects.
* Package syncx provides synchronization primitives such as events, done
  channels, semaphores, duplicate-filtering channels, and merged serial calls.
* Package racex exposes whether the binary was built with the race detector.

Documentation
-------------

- [API Reference](http://godoc.org/github.com/someonegg/gox)

Installation
------------

Install gox using the "go get" command:

    go get github.com/someonegg/gox
