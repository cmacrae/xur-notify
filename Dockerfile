FROM alpine:latest

# Prepare to fetch the latest binary from GitHub
RUN apk update && apk add --no-cache curl openssl ca-certificates && update-ca-certificates

# Grab the latest linux binary
RUN curl -Ls -o /bin/xur-notify $(curl -s https://api.github.com/repos/cmacrae/xur-notify/releases | fgrep browser_download_url | fgrep linux | cut -d '"' -f 4) && chmod +x /bin/xur-notify

CMD '/bin/xur-notify'
