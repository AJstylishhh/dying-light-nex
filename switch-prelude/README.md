# Dying Light Prelude (independent prototype)

A small, independent Nintendo Switch homebrew utility for adding or removing the Dying Light 1 NEX hostname in Atmosphère DNS-MITM host files.

**This is not an official Prelude build and is not yet a tested release.** It only handles hostname redirection; it does not implement the Dying Light NEX backend.

## Build

Install devkitPro with the Nintendo Switch / libnx toolchain, then run:

```sh
make
```

The output is `dying-light-prelude.nro`.

## Use

1. Make sure Atmosphère DNS-MITM is enabled in `/atmosphere/config/system_settings.ini`:
   ```ini
   [atmosphere]
   enable_dns_mitm = u8!0x1
   ```
   Preserve any other existing settings in that file; do not replace the whole file with this snippet.
2. Copy `dying-light-prelude.nro` to `/switch/dying-light-prelude/` on the SD card.
3. Launch it from the Homebrew Menu.
4. Choose **Apply** and enter the reachable IP address of your Dying Light NEX server. It adds the explicit hostname rule to both sysMMC and emuMMC host files.
5. Reboot the console for a clean test.
6. Choose **Remove** to remove only this app's Dying Light hostname rule. Other host rules are preserved.

## Important

- The default suggestion `192.168.137.1` is only suitable while the Switch can reach your Windows PC at that address on the same local network. It is not a public server address.
- DNS redirection does not tunnel traffic or make a local server reachable from the Internet.
- Your NEX backend must return valid, reachable matchmaking/session URLs. A hostname redirect alone will not make matchmaking work.
- Test on emuMMC first. Keep a backup of the SD card before changing network configuration.
