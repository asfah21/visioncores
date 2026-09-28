FROM golang:1.24

# Install system deps PALING ATAS supaya layer ini di-cache
# dan tidak diulang setiap ada perubahan code Go.
RUN apt-get update && apt-get install -y --no-install-recommends ffmpeg && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

ENV RECORDING_PATH=/recordings
ENV GO2RTC_CONFIG_PATH=/config/go2rtc.yaml
ENV CAMERA_SEED_PATH=/seed/cameras.yaml

CMD ["./main"]
