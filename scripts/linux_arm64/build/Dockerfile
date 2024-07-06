FROM arm64v8/debian:bookworm

RUN apt-get update -y && apt-get upgrade -y

# make go mod download work
RUN apt-get install -y ca-certificates openssl && \
    openssl s_client -showcerts -connect github.com:443 </dev/null 2>/dev/null|openssl x509 -outform PEM > /usr/local/share/ca-certificates/github.crt && \
    openssl s_client -showcerts -connect proxy.golang.org:443 </dev/null 2>/dev/null|openssl x509 -outform PEM >  /usr/local/share/ca-certificates/proxy.golang.crt && \
    update-ca-certificates

# install go and app dependencies
RUN apt-get install build-essential git curl wget ncurses-dev -y && \
    apt-get install golang-doc golang-go golang-src golang -y

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

# install task
RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

# drop permissions on output files
RUN useradd -m -u 1000 build
USER build

WORKDIR /build

RUN git config --global --add safe.directory /build
