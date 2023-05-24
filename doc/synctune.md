# Tuning synchronization parameters

Full synchronization of large file trees involve a lot of system activity (I/O for Input Output).
The system will behave differently depending on the media supporting the data storage:

- USB keys are very slow devices
- Hard drives are rather fast
- SSD are very fast
- Network I/O for object storage depends on the bandwidth and the quality of service

By default, the synchronization algorithm performs a large number of operations in parallel,
which makes synchronization rapid on fast devices.

For writing on USB keys, it is recommended to serialize operations using the `--serial` flag.

For writing large file trees on hard drives, you may need to increase the maximum number of system threads
which is set to 10,000 by default, for instance using `--maxt 50000` will make synchronization
on semi-fast drives most likely to succeed.
