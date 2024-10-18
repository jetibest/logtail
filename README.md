# Usage

When using:

    my-program >>/var/log/my-program.log 2>&1

The log-file will continue to grow endlessly.

Logtail provides a simple and portable way to keep the log-file small, by occasionally removing the first part of the file:

    my-program | logtail -n 1000:1100 /var/log/my-program.log

Whenever the file reaches more than 1100 lines, logtail will automatically trim the first 100 lines, so that the file is now back to 1000 lines.

You may also specify the maximum size in bytes, like so: `-c 1MiB:1.1MiB`.
Note that logtail is line-buffered, so it will never break lines, even when the size is specified in bytes.

By default, logtail will read most of the file into RAM during the trimming.
For relatively small log files this should not pose an issue.
In order to avoid high memory usage on a system with limited RAM or very large files, logtail can use a limited buffer size to move chunks of the file around to achieve the same effect.
However, this means more i/o operations which is slower, more CPU intensive, and results in more cycles/wear of the disk.
Usage is `-m 4096` to set for instance a maximum chunk size of 4 KiB.

