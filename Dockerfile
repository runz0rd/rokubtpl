FROM golang:buster
WORKDIR /
COPY . .
RUN go build -o /rokubtpl cmd/main.go

FROM openjdk:17-slim
WORKDIR /
RUN apt-get update && apt-get install -y git bluez bluetooth
RUN git clone https://github.com/wseemann/RPListening.git
RUN cd RPListening/RPListening && ./gradlew customFatJar
COPY entrypoint.sh entrypoint.sh
COPY --from=0 /rokubtpl /rokubtpl
RUN cp /RPListening/RPListening/build/libs/RPListening-1.1.jar /RPListening.jar
ENTRYPOINT sh entrypoint.sh