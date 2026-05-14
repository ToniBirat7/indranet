# ViGEm Bus Driver Notes

## Question
How do we inject gamepad input into a Windows application using ViGEm?

## Overview
ViGEm (Virtual Gamepad Emulation) is an open-source Windows kernel driver that creates virtual Xbox 360 and DualShock 4 controllers. Games treat the virtual controller as a real device.

## Key API

```cpp
#include <ViGEm/Client.h>

// Connect to ViGEm bus
PVIGEM_CLIENT client = vigem_alloc();
vigem_connect(client);

// Create virtual Xbox 360 controller
PVIGEM_TARGET target = vigem_target_x360_alloc();
vigem_target_add(client, target);

// Update controller state
XUSB_REPORT report;
XUSB_REPORT_INIT(&report);
report.bLeftTrigger = 0;
report.bRightTrigger = 255;
report.wButtons = XUSB_GAMEPAD_A;  // A button pressed
vigem_target_x360_update(client, target, report);
```

## Requirements
- ViGEm Bus driver must be installed on the host machine
- The Tauri installer should bundle the ViGEm driver installer
- Driver requires admin install but does NOT require admin at runtime

## TODO
- Measure injection latency: `vigem_target_x360_update()` to game registering input
- Test whether ViGEm virtual controllers are visible inside Windows Sandbox (they may not be)
- If not visible in sandbox: investigate virtual HID device bridging via the IPC pipe

## Open Questions
- Can ViGEm inject input into a process running inside Windows Sandbox?
- If not: does Windows Sandbox have its own input injection mechanism?
