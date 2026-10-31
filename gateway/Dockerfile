FROM rust:1.89-bookworm AS builder

WORKDIR /app

COPY Cargo.toml Cargo.lock ./
COPY src ./src
COPY configs ./configs

RUN cargo build --release

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/target/release/gateway ./gateway
COPY configs ./configs

EXPOSE 8080

CMD ["./gateway"]
