<!--
SPDX-FileCopyrightText: 2022 Peter Magnusson <me@kmpm.se>

SPDX-License-Identifier: MIT
-->

# Unifi MAC address lists

Unimac is a small command line program made for extracting
basic information about devices and clients from Unifi Controllers.

Uses https://github.com/unpoller/unifi for accessing the controller.


## Usage
```
unimac clients -output clients.xlsx

unimac devices -output devices.xlsx

unimac -h
```