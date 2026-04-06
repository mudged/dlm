FROM mcr.microsoft.com/devcontainers/go:1.25-trixie

ENV NEXT_TELEMETRY_DISABLED=1

RUN apt-get update \
	&& apt-get install -y --no-install-recommends ca-certificates curl \
	&& curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
	&& apt-get install -y --no-install-recommends nodejs \
	&& rm -rf /var/lib/apt/lists/*
