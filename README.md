dumpserver
==========

Stupid HTTP server that dumps all request to an output file (or not if you don't want, but then it's even more useless).

I wanted a server to set up as a sinkhole on my personal server so I can look at what the script kiddies and other shady assholes send.

Building
--------

Get [Bazel](https://bazel.build) and run `bazel build :dumpserver`
