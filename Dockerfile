FROM golang:1.14-alpine3.11

ENV PATH="$PATH:/bin/bash" \
	BENTO4_BIN="/opt/bento4/bin" \
	PATH="$PATH:/opt/bento4/bin"

# Install FFMPEG
RUN apk add --update ffmpeg bash curl make

# Install Bento
WORKDIR /tmp/bento4

ENV BENTO4_BASE_URL="http://zebulon.bok.net/Bento4/source/" \
	BENTO4_VERSION="1-5-0-615" \
	BENTO4_CHECKSUM="5378dbb374343bc274981d6e2ef93bce0851bda1" \
	BENTO4_TARGET="" \
	BENTO4_PATH="/opt/bento4" \
	BENTO4_TYPE="SRC"

# Download and unzip Bento4
RUN apk add --update --upgrade wget python unzip bash gcc g++ scons && \
	wget -q ${BENTO4_BASE_URL}/Bento4-${BENTO4_TYPE}-${BENTO4_VERSION}${BENTO4_TARGET}.zip && \
	mkdir -p ${BENTO4_PATH} && \
	unzip Bento4-${BENTO4_TYPE}-${BENTO4_VERSION}${BENTO4_TARGET}.zip -d ${BENTO4_PATH} && \
	rm -rf Bento4-${BENTO4_TYPE}-${BENTO4_VERSION}${BENTO4_TARGET}.zip && \
	apk del unzip && \
	cd ${BENTO4_PATH} && scons -u build_config=Release target=x86_64-unknown-linux && \
	cp -R ${BENTO4_PATH}/Build/Targets/x86_64-unknown-linux/Release ${BENTO4_PATH}/bin && \
	cp -R ${BENTO4_PATH}/Source/Python/utils ${BENTO4_PATH}/utils && \
	cp -a ${BENTO4_PATH}/Source/Python/wrappers/. ${BENTO4_PATH}/bin

WORKDIR /go/src

ENTRYPOINT [ "top" ]