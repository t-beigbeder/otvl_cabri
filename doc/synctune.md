# Tuning synchronization parameters

Full synchronization of large file trees involve a lot of system activity (I/O for Input Output).
The system will behave differently depending on the media supporting the data storage:

- USB keys are very slow devices
- Hard drives are rather fast
- SSD are very fast
- Network I/O for object storage or remote access depends on the bandwidth and the quality of service

By default, the synchronization algorithm performs a large number of operations in parallel,
which makes synchronization rapid on fast devices.

It is possible to control the maximum number of I/O operations running in parallel
using the `--reducer <number>` which is set to 20 by default:
- on USB keys, especially as target, use a smaller size, such as 10
- even with such limitation, accessing USB keys remotely can cause network timeouts,
the option --serial forbids any parallelism and may perform even better in such circumstances
- on SSD drives, unlimit it with size 0
- with network I/O, you may want to tune it to avoid errors
