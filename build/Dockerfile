FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates

COPY ./bin/review-reminder-bot /review-reminder-bot

CMD ["/review-reminder-bot"]
