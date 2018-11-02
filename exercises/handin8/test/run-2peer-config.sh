#!/bin/bash

read -r -d '' skSERV << EOF
-----BEGIN KEY-----
MIIBhwKBwQC1cUbNi3VDcZhr2sp4rKprf/jxmNKg4zm5FGCA9f641UEvyeUk+Thv
mH+lDiurnuLiKOq0WTUu5P3UMOp+G2yO9OUN2ZBYTQUAe5/v5HxUGUc2myiWb2f5
qERixmrUabDV1cBR8wc2g9J4m8lygRMHDxPCIbuPm34Z2kXO2zZNPMOVvTOl7/4M
IK1hgXXMwmVA2gYFqawlwTER+QAVvr3ZYQ3shZr9Ft3IxOqVHO0y7bbga3OEY3Zj
PiKkIcZ13HcCgcB49i8zsk4s9mWdPIb7HcbyVVChEIxrQiZ7YusAo/8l44DKhpjD
UNBKZapuCXJyaeyWxfHNkM4fQ1Piy0b+vPMJbADZrK80yDwcK1zm0mVp9HR3MYJq
yHWNCZJpD2Edwo6yXdSrD6BV5GewygIxxOrZZTJOrbnqGikK7HqcyMK2k2IfVgws
eVTJKFw+RJS43kUn2vEycZOzkq3S1P0XSVSECKgpOwSTV9fhIArG/okH39rDGg7s
K72WyXLabfJnXks=
-----END KEY-----
EOF


read -r -d '' pkSERV << EOF
-----BEGIN KEY-----
MIHHAoHBALVxRs2LdUNxmGvaynisqmt/+PGY0qDjObkUYID1/rjVQS/J5ST5OG+Y
f6UOK6ue4uIo6rRZNS7k/dQw6n4bbI705Q3ZkFhNBQB7n+/kfFQZRzabKJZvZ/mo
RGLGatRpsNXVwFHzBzaD0nibyXKBEwcPE8Ihu4+bfhnaRc7bNk08w5W9M6Xv/gwg
rWGBdczCZUDaBgWprCXBMRH5ABW+vdlhDeyFmv0W3cjE6pUc7TLttuBrc4RjdmM+
IqQhxnXcdwIBAw==
-----END KEY-----
EOF

read -r -d '' pkPEER << EOF
-----BEGIN KEY-----
MIHHAoHBAMKGz7ZdnbHfp3JXHjpx4/Y2CrWysKbgGsc8bq8hLCmotyddkQD839CN
GFKor8Y73uSg0xt2dbvD25HC5N2DitY6vYrJxLO4XgXiL0hLfxLJhmueq1+f5ccS
OAKCQM7H120Ev3CLZ3OgCgnrmz9vWGSdaHmG43oq9/boV32mtKAWO5V8u83vT5YS
zwO3PBnFZbzCseFGhR+l07GfApavaDehegTUaPOBpTJVtm4YlOOiX2r8pGjZzqIp
4rvWl1E1wQIBAw==
-----END KEY-----
EOF

read -r -d '' skPEER << EOF
-----BEGIN KEY-----
MIIBiAKBwQDChs+2XZ2x36dyVx46ceP2Ngq1srCm4BrHPG6vISwpqLcnXZEA/N/Q
jRhSqK/GO97koNMbdnW7w9uRwuTdg4rWOr2KycSzuF4F4i9IS38SyYZrnqtfn+XH
EjgCgkDOx9dtBL9wi2dzoAoJ65s/b1hknWh5huN6Kvf26Fd9prSgFjuVfLvN70+W
Es8DtzwZxWW8wrHhRoUfpdOxnwKWr2g3oXoE1GjzgaUyVbZuGJTjol9q/KRo2c6i
KeK71pdRNcECgcEAga81JD5pIT/E9uS+0aFCpCQHI8x1xJVnL32fH2tyxnB6Gj5g
q1M/4F4QNxsf2X0/QxXiEk75J9fntoHt6QJcjkXlBWz9OCGqJMcx7pbNOP2r6B9U
6Lysh3iCGbAf+aGP+P9vPGpVX+MSGO8adBYC6SaNfI6Bchj7TrsR+3gDoiylJamu
vrpWxplZtTAhKLaI3DK3Q7Jrarb/Qz2zrj0ck87Srdb8ovmmQUfrT2jI30JR5RvP
kQ/MxHLK6Bqox/3L
-----END KEY-----
EOF

read -r -d '' pkACCOUNTB << EOF
-----BEGIN KEY-----
ACCOUNT_B
-----END KEY-----
EOF

read -r -d '' pkACCOUNTC << EOF
-----BEGIN KEY-----
ACCOUNT_C
-----END KEY-----
EOF


cmdServf="test/transactionServer.dat"
cmdPeerf="test/transactionPeer.dat"

resServf="test/logServer.txt"
resPeerF="test/logPeer.txt"

pkPeerF="test/pkPeer.key"
skPeerF="test/skPeer.key"
pkServF="test/pkServ.key"
skServF="test/pkServ.key"

if [[ ! -f "$cmdServf" ]] || [[ ! -f ""$cmdPeerf"" ]]; then
    touch $cmdServf
    touch $cmdPeerf
    
    for i in {1..2}
    do
      echo -e "$pkSERV\n$pkACCOUNTB\n1\n$skSERV" >> "$cmdServf"
      echo -e "$pkSERV\n$pkACCOUNTC\n1\n$skSERV" >> "$cmdPeerf"
    done
fi

cat "$cmdServf"| ./main -s "$skServF" -c "$pkServF" server > "$resServf" 
cat "$cmdPeerf"| ./main -s "$skPeerF" -c "$pkPeerF" peer "$1" "4444" > "$resPeerf"



