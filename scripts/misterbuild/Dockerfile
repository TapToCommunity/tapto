# glibc in this version matches mister
FROM arm32v7/debian:bullseye-backports

RUN apt-get update -y && apt-get upgrade -y

# make go mod download work
RUN apt-get install -y ca-certificates openssl && \
    openssl s_client -showcerts -connect github.com:443 </dev/null 2>/dev/null|openssl x509 -outform PEM > /usr/local/share/ca-certificates/github.crt && \
    openssl s_client -showcerts -connect proxy.golang.org:443 </dev/null 2>/dev/null|openssl x509 -outform PEM >  /usr/local/share/ca-certificates/proxy.golang.crt && \
    update-ca-certificates

# install go and app dependencies
RUN apt-get install build-essential git curl wget ncurses-dev -y && \
    apt-get install golang-doc/bullseye-backports golang-go/bullseye-backports golang-src/bullseye-backports golang/bullseye-backports -y

# install libnfc dependencies
RUN apt-get install -y libusb-dev libtool autoconf automake
# install custom version of libnfc
RUN mkdir /internal && cd /internal && \
    git clone --depth 1 https://github.com/sam1902/libnfc && \
    cd libnfc && \
    autoreconf -vis && \
    ./configure && \
    make -j "$(nproc)" && \
    make install

# install pcsc-lite and ccid dependencies
RUN apt-get install -y flex libusb-1.0-0-dev zlib1g-dev
# install custom version of pcsc-lite
RUN cd /internal && \
    wget https://pcsclite.apdu.fr/files/pcsc-lite-2.0.0.tar.bz2 && \
    tar xf pcsc-lite-2.0.0.tar.bz2 && \
    cd pcsc-lite-2.0.0 && \
    ./configure --disable-libsystemd --disable-libudev -enable-static && \
    make -j "$(nproc)" && \
    make install
# install custom version of ccid
RUN cd /internal && \
    wget https://ccid.apdu.fr/files/ccid-1.5.4.tar.bz2 && \
    tar xf ccid-1.5.4.tar.bz2 && \
    cd ccid-1.5.4 && \
    ./configure -enable-static && \
    make -j "$(nproc)" && \
    make install

# drop permissions on output files
RUN useradd -m -u 1000 build
USER build

# install mage
RUN cd /home/build && \
    git clone https://github.com/magefile/mage --depth 1 && \
    mkdir /home/build/go && \
    cd mage && go run bootstrap.go
ENV PATH="${PATH}:/home/build/go/bin"

WORKDIR /build