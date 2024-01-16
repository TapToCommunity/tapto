#! /bin/bash

# if [ ! -d pcsc-lite-2.0.0 ]; then
#     wget https://pcsclite.apdu.fr/files/pcsc-lite-2.0.0.tar.bz2
#     tar xf pcsc-lite-2.0.0.tar.bz2
# fi

# cd pcsc-lite-2.0.0 || exit

# ./configure --disable-libsystemd --disable-libudev -enable-static
# make -j "$(nproc)"

cp /usr/local/sbin/pcscd .

# cd ..

# if [ ! -d ccid-1.5.4 ]; then
#     wget https://ccid.apdu.fr/files/ccid-1.5.4.tar.bz2
#     tar xf ccid-1.5.4.tar.bz2
# fi

# cd ccid-1.5.4 || exit

# ./configure --enable-static
# make -j "$(nproc)"
# make install

cp -r /usr/local/lib/pcsc/drivers .
