FROM openjdk:17-jdk-alpine AS builder
WORKDIR /
RUN apk add git
RUN git clone https://github.com/wseemann/RPListening.git
RUN cd RPListening/RPListening && ./gradlew customFatJar

FROM golang:alpine
WORKDIR /
COPY . .
RUN apk add bluez openjdk11 ffmpeg && apk add -X http://dl-cdn.alpinelinux.org/alpine/edge/community bluez-alsa
COPY --from=builder /RPListening/RPListening/build/libs/RPListening-1.1.jar /RPListening.jar
RUN go build -o /rokubtpl cmd/main.go
ENTRYPOINT [ "/rokubtpl" ]