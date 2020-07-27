# Use alpine based Go image
FROM golang:alpine

# Move workdir
WORKDIR /build

# Set logging folder && assign volume
RUN mkdir log
VOLUME [ "/build/log" ]
VOLUME [ "/data" ]

# Copy all files
COPY . .

# Get go modules
RUN go mod download

# Build project
RUN go build -o go-file-processing-daemon .

# Execute the daemon
CMD [ "./go-file-processing-daemon" ]
