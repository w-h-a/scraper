FROM golang AS build
WORKDIR /workflow
COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/scraper ./

FROM alpine
RUN apk --no-cache add ca-certificates
COPY --from=build /go/bin/scraper /bin/scraper
ENTRYPOINT [ "/bin/scraper" ]