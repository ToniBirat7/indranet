# Linux uinput Notes

## Relevance
Phase 3 (Linux host support). uinput is the Linux equivalent of ViGEm.

## Overview
`uinput` is a Linux kernel module that allows user-space programs to create virtual input devices (keyboards, mice, gamepads).

```c
int fd = open("/dev/uinput", O_WRONLY | O_NONBLOCK);
// Configure device capabilities
ioctl(fd, UI_SET_EVBIT, EV_KEY);
ioctl(fd, UI_SET_KEYBIT, KEY_A);
// Create device
ioctl(fd, UI_DEV_CREATE);
// Inject event
struct input_event ev = { .type = EV_KEY, .code = KEY_A, .value = 1 };
write(fd, &ev, sizeof(ev));
```

## TODO
Defer to Phase 3.
