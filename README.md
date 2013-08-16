pget
====

Simple project for learning go.

This doesn't do anything useful yet. Work in progress.

`pget` stands for "pattern-get" and gets the naming style from `wget`. `pget`
will take in a single url and tries to detect any enumerable pattern in it. It
will then probe the server for other potential files matching this pattern and
download them. This will be parallelized to minimize protocol overhead.

In the scope of a toy project, the pattern will just be trivial integers.

Example
-------

`pget http://url.to/some/photo_9.jpg`

Detects 9 -> starts probing for files 8, 7... and 10, 11..

`pget http://url.to/some/archive.10.rar`

Detects 10 -> starts probing for files 09, 9, 8/08... and 11, 12...
